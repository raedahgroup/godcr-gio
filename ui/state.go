package ui

import (
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/wallet"
)

// states represents a combination of booleans that determine what the wallet is displaying.
type states struct {
	loading  bool // true if the window is in the middle of an operation that cannot be stopped
	creating bool // true if a wallet is being created or restored
}

// updateStates changes the wallet state based on the received update
func (win *Window) updateStates(update interface{}) {
	switch e := update.(type) {
	case *wallet.Transactions:
		win.states.loading = false
		win.walletTransactions = e
		return
	case *wallet.Transaction:
		win.walletTransaction = e
		return
	case *wallet.UnspentOutputs:
		win.walletUnspentOutputs = e
	case *wallet.Tickets:
		win.states.loading = false
		win.walletTickets = e
		return
	case *wallet.VSPInfo:
		win.states.loading = false
		win.vspInfo.List = append(win.vspInfo.List, *e)
		win.refresh()
		return
	case *wallet.VSP:
		win.vspInfo = e
		win.refresh()
		return
	case *wallet.Proposals:
		win.states.loading = false
		win.proposals = e
		return
	case wallet.CreatedSeed:
		win.notifyOnSuccess("Wallet created")
		win.changePage(WalletPage(win.common))
	case wallet.Renamed:
		win.notifyOnSuccess("Wallet renamed")
	case wallet.Restored:
		win.changePage(WalletPage(win.common))
		win.states.creating = false
		win.window.Invalidate()
	case wallet.DeletedWallet:
		win.selected = 0
		win.changePageAndRefresh(WalletPage(win.common))
		win.notifyOnSuccess("Wallet removed")
	case wallet.AddedAccount:
		win.notifyOnSuccess("Account created")
	case wallet.UpdatedAccount:
		win.notifyOnSuccess("Account renamed")
	case *wallet.Signature:
		win.notifyOnSuccess("Message signed")
		win.signatureResult = update.(*wallet.Signature)
	case *dcrlibwallet.TxAuthor:
		txAuthor := update.(*dcrlibwallet.TxAuthor)
		win.txAuthor = *txAuthor
	case *wallet.Broadcast:
		broadcastResult := update.(*wallet.Broadcast)
		win.broadcastResult = *broadcastResult
	case *wallet.ChangePassword:
		win.notifyOnSuccess("Spending password changed")
	case *wallet.StartupPassphrase:
		win.notifyOnSuccess(update.(*wallet.StartupPassphrase).Msg)
	case wallet.OpenWallet:
		go func() {
			win.modal <- &modalLoad{}
		}()
	case wallet.SetupAccountMixer:
		win.notifyOnSuccess("Mixer setup completed")
	case *wallet.TicketPurchase:
		win.notifyOnSuccess("Ticket(s) purchased, attempting to pay fee")
	}

	win.states.loading = true
	win.wallet.GetAllTransactions(0, 0, 0)
	win.wallet.GetAllTickets()
	win.wallet.GetAllProposals()
	win.window.Invalidate()
	log.Debugf("Updated with window state %+v", win.states)
}

func (win *Window) notifyOnSuccess(text string) {
	win.toast = &toast{
		text:    text,
		success: true,
	}

	go func() {
		win.modal <- &modalLoad{}
	}()
}
