package ui

import (
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/wallet"
)

func (win *Window) OnSyncStarted(wasRestarted bool) {
	win.syncStatusUpdate <- wallet.SyncStatusUpdate{
		Stage: wallet.SyncStarted,
	}
}

func (win *Window) OnPeerConnectedOrDisconnected(numberOfConnectedPeers int32) {
	win.syncStatusUpdate <- wallet.SyncStatusUpdate{
		Stage:          wallet.PeersConnected,
		ConnectedPeers: numberOfConnectedPeers,
	}
}

func (win *Window) OnCFiltersFetchProgress(pro *dcrlibwallet.CFiltersFetchProgressReport) {
}

func (win *Window) OnHeadersFetchProgress(headersFetchProgress *dcrlibwallet.HeadersFetchProgressReport) {
	win.syncStatusUpdate <- wallet.SyncStatusUpdate{
		Stage: wallet.HeadersFetchProgress,
		ProgressReport: wallet.SyncHeadersFetchProgress{
			Progress: headersFetchProgress,
		},
	}
}
func (win *Window) OnAddressDiscoveryProgress(addressDiscoveryProgress *dcrlibwallet.AddressDiscoveryProgressReport) {
	win.syncStatusUpdate <- wallet.SyncStatusUpdate{
		Stage: wallet.AddressDiscoveryProgress,
		ProgressReport: wallet.SyncAddressDiscoveryProgress{
			Progress: addressDiscoveryProgress,
		},
	}
}

func (win *Window) OnHeadersRescanProgress(headersRescanProgress *dcrlibwallet.HeadersRescanProgressReport) {
	win.syncStatusUpdate <- wallet.SyncStatusUpdate{
		Stage: wallet.HeadersRescanProgress,
		ProgressReport: wallet.SyncHeadersRescanProgress{
			Progress: headersRescanProgress,
		},
	}
}
func (win *Window) OnSyncCompleted() {
	win.syncStatusUpdate <- wallet.SyncStatusUpdate{
		Stage: wallet.SyncCompleted,
	}
}

func (win *Window) OnSyncCanceled(willRestart bool) {
	win.syncStatusUpdate <- wallet.SyncStatusUpdate{
		Stage: wallet.SyncCanceled,
	}
}
func (win *Window) OnSyncEndedWithError(err error)          {}
func (win *Window) Debug(debugInfo *dcrlibwallet.DebugInfo) {}
