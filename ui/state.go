package ui

import (
	"gioui.org/op"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
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
	case wallet.MultiWalletInfo:
		*win.walletInfo = e
		win.states.loading = false

		if e.LoadedWallets > 0 {
			// set wallets and accounts tab when wallet info is updated
			wallets := make([]decredmaterial.TabItem, len(e.Wallets))
			for i := range e.Wallets {
				wallets[i] = decredmaterial.TabItem{
					Title: e.Wallets[i].Name,
				}
			}

			accounts := make([]decredmaterial.TabItem, len(e.Wallets[win.selected].Accounts))
			for i, account := range e.Wallets[win.selected].Accounts {
				if account.Name == "imported" {
					continue
				}
				accounts[i] = decredmaterial.TabItem{
					Title: e.Wallets[win.selected].Accounts[i].Name,
				}
			}
		}
		return
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
		return
	case *wallet.VSP:
		win.vspInfo = e
		return
	case *wallet.Proposals:
		win.states.loading = false
		win.proposals = e
		return
	case wallet.Restored:
		win.states.creating = false
		op.InvalidateOp{}.Add(win.ops)
	case wallet.DeletedWallet:
		win.selected = 0
		win.notifyOnSuccess("Wallet removed")
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
	case wallet.SetupAccountMixer:
		win.notifyOnSuccess("Mixer setup completed")
	case *wallet.TicketPurchase:
		win.notifyOnSuccess("Ticket(s) purchased, attempting to pay fee")
	}

	win.states.loading = true
	win.wallet.GetMultiWalletInfo()
	win.wallet.GetAllTransactions(0, 0, 0)
	win.wallet.GetAllTickets()
	// win.wallet.ClearProposals()
	win.wallet.GetAllProposals()
	op.InvalidateOp{}.Add(win.ops)
	log.Debugf("Updated with multiwallet info: %+v\n and window state %+v", win.walletInfo, win.states)
}

func (win *Window) notifyOnSuccess(text string) {
	win.toast = &toast{
		text:    text,
		success: true,
	}
}
