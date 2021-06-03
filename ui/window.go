package ui

import (
	"errors"
	"image"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"
)

// Window represents the app window (and UI in general). There should only be one.
// Window uses an internal state of booleans to determine what the window is currently displaying.
type Window struct {
	window *app.Window
	theme  *decredmaterial.Theme
	ops    *op.Ops

	wallet             *wallet.Wallet
	syncStatusUpdate   chan wallet.SyncStatusUpdate
	walletSyncStatus   *wallet.SyncStatus
	walletTransactions *wallet.Transactions
	walletTransaction  *wallet.Transaction
	walletTickets      *wallet.Tickets
	vspInfo            *wallet.VSP
	proposals          *wallet.Proposals
	proposal           chan *wallet.Proposal

	walletUnspentOutputs *wallet.UnspentOutputs

	signatureResult *wallet.Signature

	selectedAccount int
	txAuthor        dcrlibwallet.TxAuthor
	broadcastResult wallet.Broadcast

	selected int
	states

	err string

	// TODO: protect with mutex
	pageBackStack             []Page
	currentPage, previousPage Page
	common                    pageCommon

	walletTabs, accountTabs *decredmaterial.Tabs
	keyEvents               chan *key.Event
	toast                   *toast
	modal                   chan *modalLoad
	sysDestroyWithSync      bool
	walletAcctMixerStatus   chan *wallet.AccountMixer
	internalLog             chan string
	refreshPage             bool
}

type WriteClipboard struct {
	Text string
}

// CreateWindow creates and initializes a new window with start
// as the first page displayed.
// Should never be called more than once as it calls
// app.NewWindow() which does not support being called more
// than once.
func CreateWindow(wal *wallet.Wallet, decredIcons map[string]image.Image, collection []text.FontFace, internalLog chan string) (*Window, error) {
	win := new(Window)
	var netType string
	if wal.Net == "testnet3" {
		netType = "testnet"
	} else {
		netType = wal.Net
	}
	win.window = app.NewWindow(app.Size(values.AppWidth, values.AppHeight), app.Title(values.StringF(values.StrAppTitle, netType)))
	theme := decredmaterial.NewTheme(collection, decredIcons, false)
	if theme == nil {
		return nil, errors.New("Unexpected error while loading theme")
	}
	win.theme = theme
	win.ops = &op.Ops{}

	win.syncStatusUpdate = make(chan wallet.SyncStatusUpdate, 10)
	win.walletSyncStatus = new(wallet.SyncStatus)
	win.walletTransactions = new(wallet.Transactions)
	win.walletUnspentOutputs = new(wallet.UnspentOutputs)
	win.walletAcctMixerStatus = make(chan *wallet.AccountMixer)
	win.walletTickets = new(wallet.Tickets)
	win.vspInfo = new(wallet.VSP)
	win.proposals = new(wallet.Proposals)
	win.proposal = make(chan *wallet.Proposal)

	win.wallet = wal
	win.states.loading = false

	win.keyEvents = make(chan *key.Event)
	win.modal = make(chan *modalLoad)

	win.internalLog = internalLog

	wal.MultiWallet().AddSyncProgressListener(win, "window") // register for sync notifications
	win.common = win.loadPageCommon(decredIcons, wal.MultiWallet())
	win.currentPage = MainPage(win.common) // TODO

	return win, nil
}

func (win *Window) changePage(page Page) {
	win.currentPage.onClose()
	win.pageBackStack = append(win.pageBackStack, win.currentPage)

	win.currentPage = page
	win.refresh()
}

func (win *Window) changePageAndRefresh(page Page) {
	win.refreshPage = true
	win.changePage(page)
}

func (win *Window) refresh() {
	win.window.Invalidate()
}

// popPage goes back to the previous page
func (win *Window) popPage() {
	if len(win.pageBackStack) > 0 {
		// get and remove last page
		previousPage := win.pageBackStack[len(win.pageBackStack)-1]
		win.pageBackStack = win.pageBackStack[:len(win.pageBackStack)-1]

		win.currentPage.onClose()
		win.currentPage = previousPage
		win.refresh()
	}
}

func (win *Window) popToPage(pageID string) error {
	// TODO
	return errors.New("not implemented")
}

func (win *Window) setReturnPage(from Page) {
	win.previousPage = from
	win.refresh()
}

func (win *Window) unloaded() {
	lbl := win.theme.H3("Multiwallet not loaded\nIs another instance open?")
	for {
		e := <-win.window.Events()
		switch evt := e.(type) {
		case system.DestroyEvent:
			return
		case system.FrameEvent:
			gtx := layout.NewContext(win.ops, evt)
			lbl.Layout(gtx)
			evt.Frame(win.ops)
		}
	}
}

func (win *Window) layoutPage(gtx C, page Page) {
	layout.Stack{
		Alignment: layout.N,
	}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			return decredmaterial.Fill(gtx, win.theme.Color.LightGray)
		}),
		layout.Stacked(func(gtx C) D {
			page.handle()
			return page.Layout(gtx)
		}),
	)
}

// Loop runs main event handling and page rendering loop
func (win *Window) Loop(shutdown chan int) {
	for {
		select {
		case e := <-win.wallet.Send:
			if e.Err != nil {
				err := e.Err.Error()
				log.Error("Wallet Error: " + err)
				if err == dcrlibwallet.ErrWalletDatabaseInUse {
					close(shutdown)
					win.unloaded()
					return
				}
				win.err = err
				if win.states.loading {
					log.Warn("NOT Attemping to get multiwallet info")
					// win.wallet.GetMultiWalletInfo()
				}

				win.window.Invalidate()
				break
			}

			win.updateStates(e.Resp)

		case update := <-win.wallet.Sync:
			switch update.Stage {
			case wallet.SyncCompleted:
				if win.sysDestroyWithSync {
					close(shutdown)
					return
				}
			case wallet.SyncStarted:
				// dcrlibwallet triggers the SyncStart method several times
				// without sending a SyncComplete signal when sync is done.
			case wallet.SyncCanceled:
				if win.sysDestroyWithSync {
					close(shutdown)
					return
				}
			case wallet.HeadersFetchProgress:
				win.updateSyncProgress(update.ProgressReport)
			case wallet.AddressDiscoveryProgress:
				win.updateSyncProgress(update.ProgressReport)
			case wallet.HeadersRescanProgress:
				win.updateSyncProgress(update.ProgressReport)
			case wallet.PeersConnected:
				win.updateConnectedPeers(update.ConnectedPeers)
			case wallet.BlockAttached:
				if win.wallet.IsSynced() {
					win.wallet.GetAllTickets()
					win.updateSyncProgress(update.BlockInfo)
				}
			case wallet.BlockConfirmed:
				win.updateSyncProgress(update.ConfirmedTxn)
			case wallet.AccountMixerStarted, wallet.AccountMixerEnded:
				go func() {
					win.walletAcctMixerStatus <- &update.AcctMixerInfo
				}()
			case wallet.ProposalAdded, wallet.ProposalVoteFinished, wallet.ProposalVoteStarted, wallet.ProposalSynced:
				win.wallet.GetAllProposals()
				go func() {
					win.proposal <- &update.Proposal
				}()
			}
			win.window.Invalidate()
		case e := <-win.window.Events():
			switch evt := e.(type) {
			case system.DestroyEvent:
				if win.wallet.IsConnectedToDecredWallet() {
					win.sysDestroyWithSync = true
					win.wallet.CancelSync()
				} else {
					close(shutdown)
				}
			case system.FrameEvent:
				gtx := layout.NewContext(win.ops, evt)
				// ts := int64(time.Since(time.Unix(win.walletInfo.BestBlockTime, 0)).Seconds())
				// win.walletInfo.LastSyncTime = wallet.SecondsToDays(ts)
				s := win.states
				// if win.walletInfo.LoadedWallets == 0 {
				// 	win.changePage(PageCreateRestore)
				// }

				if s.loading {
					win.Loading(gtx)
				} else {
					win.layoutPage(gtx, win.currentPage)
				}

				evt.Frame(win.ops)
			case key.Event:
				go func() {
					win.keyEvents <- &evt
				}()
			case nil:
				// Ignore
			default:
				log.Tracef("Unhandled window event %+v\n", e)
			}
		}
	}
}
