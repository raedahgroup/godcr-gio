package ui

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gen2brain/beeep"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/dcrlibwallet/txhelper"
	"github.com/planetdecred/godcr/wallet"
)

// Transaction notifications
func (mp *mainPage) OnTransaction(transaction string) {
	mp.updateBalance()

	var tx dcrlibwallet.Transaction
	err := json.Unmarshal([]byte(transaction), &tx)
	if err == nil {
		mp.notificationsUpdate <- wallet.NewTransaction{
			Transaction: &tx,
		}
	}
	mp.desktopNotifier(tx)
}

func (mp *mainPage) OnBlockAttached(walletID int, blockHeight int32) {
	mp.updateBalance()
	mp.notificationsUpdate <- wallet.SyncStatusUpdate{
		Stage: wallet.BlockAttached,
	}
}

func (mp *mainPage) OnTransactionConfirmed(walletID int, hash string, blockHeight int32) {
	mp.updateBalance()
}

// Account mixer
func (mp *mainPage) OnAccountMixerStarted(walletID int) {}
func (mp *mainPage) OnAccountMixerEnded(walletID int)   {}

// Politeia notifications
func (mp *mainPage) OnProposalsSynced()                            {}
func (mp *mainPage) OnNewProposal(proposal *dcrlibwallet.Proposal) {}

func (mp *mainPage) OnProposalVoteStarted(proposal *dcrlibwallet.Proposal)  {}
func (mp *mainPage) OnProposalVoteFinished(proposal *dcrlibwallet.Proposal) {}

// Sync notifications

func (mp *mainPage) OnSyncStarted(wasRestarted bool) {
	mp.notificationsUpdate <- wallet.SyncStatusUpdate{
		Stage: wallet.SyncStarted,
	}
}

func (mp *mainPage) OnPeerConnectedOrDisconnected(numberOfConnectedPeers int32) {
	mp.notificationsUpdate <- wallet.SyncStatusUpdate{
		Stage:          wallet.PeersConnected,
		ConnectedPeers: numberOfConnectedPeers,
	}
}

func (mp *mainPage) OnCFiltersFetchProgress(cfiltersFetchProgress *dcrlibwallet.CFiltersFetchProgressReport) {
	mp.notificationsUpdate <- wallet.SyncStatusUpdate{
		Stage:          wallet.CfiltersFetchProgress,
		ProgressReport: cfiltersFetchProgress,
	}
}

func (mp *mainPage) OnHeadersFetchProgress(headersFetchProgress *dcrlibwallet.HeadersFetchProgressReport) {
	mp.notificationsUpdate <- wallet.SyncStatusUpdate{
		Stage: wallet.HeadersFetchProgress,
		ProgressReport: wallet.SyncHeadersFetchProgress{
			Progress: headersFetchProgress,
		},
	}
}
func (mp *mainPage) OnAddressDiscoveryProgress(addressDiscoveryProgress *dcrlibwallet.AddressDiscoveryProgressReport) {
	mp.notificationsUpdate <- wallet.SyncStatusUpdate{
		Stage: wallet.AddressDiscoveryProgress,
		ProgressReport: wallet.SyncAddressDiscoveryProgress{
			Progress: addressDiscoveryProgress,
		},
	}
}

func (mp *mainPage) OnHeadersRescanProgress(headersRescanProgress *dcrlibwallet.HeadersRescanProgressReport) {
	mp.notificationsUpdate <- wallet.SyncStatusUpdate{
		Stage: wallet.HeadersRescanProgress,
		ProgressReport: wallet.SyncHeadersRescanProgress{
			Progress: headersRescanProgress,
		},
	}
}
func (mp *mainPage) OnSyncCompleted() {
	mp.updateBalance()
	mp.notificationsUpdate <- wallet.SyncStatusUpdate{
		Stage: wallet.SyncCompleted,
	}
}

func (mp *mainPage) OnSyncCanceled(willRestart bool) {
	mp.notificationsUpdate <- wallet.SyncStatusUpdate{
		Stage: wallet.SyncCanceled,
	}
}
func (mp *mainPage) OnSyncEndedWithError(err error)          {}
func (mp *mainPage) Debug(debugInfo *dcrlibwallet.DebugInfo) {}

func (mp *mainPage) desktopNotifier(notifier interface{}) {
	var notification string
	switch t := notifier.(type) {
	case dcrlibwallet.Transaction:
		// remove trailing zeros from amount and convert to string
		amount := strconv.FormatFloat(dcrlibwallet.AmountCoin(t.Amount), 'f', -1, 64)

		defaultNotification := "Transaction notification of %s DCR"

		notificationText := "You have %s %s DCR "

		//get wallet details
		if mp.pageCommon.wallet.LoadedWalletsCount() > 0 {
			for _, w := range mp.pageCommon.info.Wallets {
				if w.ID == t.WalletID {
					switch {
					case t.Direction == txhelper.TxDirectionReceived:
						notification = fmt.Sprintf(notificationText+" in %s wallet", "received", amount, w.Name)
					case t.Direction == txhelper.TxDirectionSent:
						notification = fmt.Sprintf("You have sent %s DCR from %s wallet", amount, w.Name)
					case t.Direction == txhelper.TxDirectionTransferred:
						notification = fmt.Sprintf("You have transferred %s DCR to %s wallet", amount, w.Name)
					default:
						notification = fmt.Sprintf(defaultNotification+" to/from %s wallet", amount, w.Name)
					}
				} else {
					switch {
					case t.Direction == txhelper.TxDirectionReceived:
						notification = fmt.Sprintf("You have received %s DCR", amount)
					case t.Direction == txhelper.TxDirectionSent:
						notification = fmt.Sprintf("You have sent %s DCR from %s wallet", amount, w.Name)
					case t.Direction == txhelper.TxDirectionTransferred:
						notification = fmt.Sprintf("You have transferred %s DCR to %s wallet", w.Name, amount)
					default:
						notification = fmt.Sprintf(defaultNotification+"to/from %s wallet", w.Name, amount)
					}
				}
			}
		} else {
			notification = fmt.Sprintf(defaultNotification, amount)
		}
	case dcrlibwallet.Proposal:
	}

	err := beeep.Notify("Decred Godcr Wallet", notification, "assets/information.png")
	if err != nil {
		log.Info("could not initiate desktop notification, reason:", err.Error())
	}
}
