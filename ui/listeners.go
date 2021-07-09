package ui

import (
	"encoding/json"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/page"
	"github.com/planetdecred/godcr/wallet"
)

// Transaction notifications

func (mp *mainPage) OnTransaction(transaction string) {
	mp.updateBalance()

	// beeep send notification

	var tx dcrlibwallet.Transaction
	err := json.Unmarshal([]byte(transaction), &tx)
	if err == nil {
		mp.updateNotification(wallet.NewTransaction{
			Transaction: &tx,
		})
	}
}

func (mp *mainPage) OnBlockAttached(walletID int, blockHeight int32) {
	mp.updateBalance()
	mp.updateNotification(wallet.SyncStatusUpdate{
		Stage: wallet.BlockAttached,
	})
}

func (mp *mainPage) OnTransactionConfirmed(walletID int, hash string, blockHeight int32) {
	mp.updateBalance()
}

// Account mixer
func (mp *mainPage) OnAccountMixerStarted(walletID int) {}
func (mp *mainPage) OnAccountMixerEnded(walletID int)   {}

// Politeia notifications
func (mp *mainPage) OnProposalsSynced() {
	mp.notificationsUpdate <- wallet.Proposal{
		ProposalStatus: wallet.Synced,
	}
}

func (mp *mainPage) OnNewProposal(proposal *dcrlibwallet.Proposal) {
	mp.notificationsUpdate <- wallet.Proposal{
		ProposalStatus: wallet.NewProposalFound,
		Proposal:       proposal,
	}
}

func (mp *mainPage) OnProposalVoteStarted(proposal *dcrlibwallet.Proposal) {
	mp.notificationsUpdate <- wallet.Proposal{
		ProposalStatus: wallet.VoteStarted,
		Proposal:       proposal,
	}
}
func (mp *mainPage) OnProposalVoteFinished(proposal *dcrlibwallet.Proposal) {
	mp.notificationsUpdate <- wallet.Proposal{
		ProposalStatus: wallet.VoteFinished,
		Proposal:       proposal,
	}
}

// Sync notifications

func (mp *mainPage) OnSyncStarted(wasRestarted bool) {
	mp.updateNotification(wallet.SyncStatusUpdate{
		Stage: wallet.SyncStarted,
	})
}

func (mp *mainPage) OnPeerConnectedOrDisconnected(numberOfConnectedPeers int32) {
	mp.updateNotification(wallet.SyncStatusUpdate{
		Stage:          wallet.PeersConnected,
		ConnectedPeers: numberOfConnectedPeers,
	})
}

func (mp *mainPage) OnCFiltersFetchProgress(cfiltersFetchProgress *dcrlibwallet.CFiltersFetchProgressReport) {
	mp.updateNotification(wallet.SyncStatusUpdate{
		Stage:          wallet.CfiltersFetchProgress,
		ProgressReport: cfiltersFetchProgress,
	})
}

func (mp *mainPage) OnHeadersFetchProgress(headersFetchProgress *dcrlibwallet.HeadersFetchProgressReport) {
	mp.updateNotification(wallet.SyncStatusUpdate{
		Stage: wallet.HeadersFetchProgress,
		ProgressReport: wallet.SyncHeadersFetchProgress{
			Progress: headersFetchProgress,
		},
	})
}
func (mp *mainPage) OnAddressDiscoveryProgress(addressDiscoveryProgress *dcrlibwallet.AddressDiscoveryProgressReport) {
	mp.updateNotification(wallet.SyncStatusUpdate{
		Stage: wallet.AddressDiscoveryProgress,
		ProgressReport: wallet.SyncAddressDiscoveryProgress{
			Progress: addressDiscoveryProgress,
		},
	})
}

func (mp *mainPage) OnHeadersRescanProgress(headersRescanProgress *dcrlibwallet.HeadersRescanProgressReport) {
	mp.updateNotification(wallet.SyncStatusUpdate{
		Stage: wallet.HeadersRescanProgress,
		ProgressReport: wallet.SyncHeadersRescanProgress{
			Progress: headersRescanProgress,
		},
	})
}
func (mp *mainPage) OnSyncCompleted() {
	mp.updateBalance()
	mp.updateNotification(wallet.SyncStatusUpdate{
		Stage: wallet.SyncCompleted,
	})
}

func (mp *mainPage) OnSyncCanceled(willRestart bool) {
	mp.updateNotification(wallet.SyncStatusUpdate{
		Stage: wallet.SyncCanceled,
	})
}
func (mp *mainPage) OnSyncEndedWithError(err error)          {}
func (mp *mainPage) Debug(debugInfo *dcrlibwallet.DebugInfo) {}

// todo: this will be removed when all pages have been moved to the page package
// updateNotification sends notification to the notification channel depending on which channel the page uses
func (mp *mainPage) updateNotification(signal interface{}) {
	switch *mp.page {
	case page.OverviewPageID, page.Transactions:
		mp.load.Receiver.NotificationsUpdate <- signal
	default:
		mp.notificationsUpdate <- signal
	}
}
