package page

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/planetdecred/godcr/ui/load"

	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"
)

const OverviewPageID = "Overview"

// walletSyncDetails contains sync data for each wallet when a sync
// is in progress.
type walletSyncDetails struct {
	name               decredmaterial.Label
	status             decredmaterial.Label
	blockHeaderFetched decredmaterial.Label
	syncingProgress    decredmaterial.Label
}

type OverviewPage struct {
	*load.Load
	pageClosing chan bool
	listContainer, walletSyncList,
	transactionsList *layout.List
	theme *decredmaterial.Theme
	tab   *decredmaterial.Tabs

	allWallets   []*dcrlibwallet.Wallet
	transactions []dcrlibwallet.Transaction

	toTransactions    decredmaterial.TextAndIconButton
	sync              decredmaterial.Button
	toggleSyncDetails decredmaterial.Button
	syncedIcon, notSyncedIcon,
	walletStatusIcon, cachedIcon *widget.Icon
	syncingIcon          *widget.Image
	toTransactionDetails []*gesture.Click

	walletSyncing bool
	walletSynced  bool
	isConnnected  bool

	bestBlock            *dcrlibwallet.BlockInfo
	connectedPeers       int32
	remainingSyncTime    string
	headersToFetchOrScan int32
	headerFetchProgress  int32
	syncProgress         int
	syncStep             int

	syncButtonHeight      int
	moreButtonWidth       int
	moreButtonHeight      int
	syncDetailsVisibility bool
	txnRowHeight          int
	queue                 event.Queue
}

func NewOverviewPage(l *load.Load) *OverviewPage {
	pg := &OverviewPage{
		Load:        l,
		pageClosing: make(chan bool, 1),
		theme:       l.Theme,

		allWallets: l.WL.SortedWalletList(),

		listContainer:    &layout.List{Axis: layout.Vertical},
		walletSyncList:   &layout.List{Axis: layout.Vertical},
		transactionsList: &layout.List{Axis: layout.Vertical},

		bestBlock: l.WL.MultiWallet.GetBestBlock(),

		syncButtonHeight: 50,
		moreButtonWidth:  115,
		moreButtonHeight: 70,
		txnRowHeight:     56,
	}

	pg.toTransactions = l.Theme.TextAndIconButton(new(widget.Clickable), values.String(values.StrSeeAll), l.Icons.NavigationArrowForward)
	pg.toTransactions.Color = l.Theme.Color.Primary
	pg.toTransactions.BackgroundColor = l.Theme.Color.Surface

	pg.sync = l.Theme.Button(new(widget.Clickable), values.String(values.StrReconnect))
	pg.sync.TextSize = values.TextSize10
	pg.sync.Background = color.NRGBA{}
	pg.sync.Color = l.Theme.Color.Text

	pg.toggleSyncDetails = l.Theme.Button(new(widget.Clickable), values.String(values.StrShowDetails))
	pg.toggleSyncDetails.TextSize = values.TextSize16
	pg.toggleSyncDetails.Background = color.NRGBA{}
	pg.toggleSyncDetails.Color = l.Theme.Color.Primary
	pg.toggleSyncDetails.Inset = layout.Inset{}

	pg.syncedIcon = l.Icons.ActionCheckCircle
	pg.syncedIcon.Color = l.Theme.Color.Success

	pg.syncingIcon = l.Icons.SyncingIcon
	pg.syncingIcon.Scale = 1

	pg.notSyncedIcon = l.Icons.NavigationCancel
	pg.notSyncedIcon.Color = l.Theme.Color.Danger

	pg.walletStatusIcon = l.Icons.ImageBrightness1
	pg.cachedIcon = l.Icons.Cached

	return pg
}

func (pg *OverviewPage) OnResume() {
	pg.walletSyncing = pg.WL.MultiWallet.IsSyncing()
	pg.walletSynced = pg.WL.MultiWallet.IsSynced()
	pg.isConnnected = pg.WL.MultiWallet.IsConnectedToDecredNetwork()
	pg.connectedPeers = pg.WL.MultiWallet.ConnectedPeers()
	pg.bestBlock = pg.WL.MultiWallet.GetBestBlock()

	pg.loadTransactions()
	pg.listenForSyncNotifications()
}

func (pg *OverviewPage) loadTransactions() {
	transactions, err := pg.WL.MultiWallet.GetTransactionsRaw(0, 5, dcrlibwallet.TxFilterAll, true)
	if err != nil {
		log.Error("Error getting transactions:", err)
		return
	}

	pg.transactions = transactions
}

// Layout lays out the entire content for overview pg.
func (pg *OverviewPage) Layout(gtx layout.Context) layout.Dimensions {
	pg.queue = gtx
	if pg.WL.Info.LoadedWallets == 0 {
		return uniformPadding(gtx, func(gtx C) D {
			return layout.Center.Layout(gtx, func(gtx C) D {
				return pg.Theme.H3(values.String(values.StrNoWalletLoaded)).Layout(gtx)
			})
		})
	}

	pageContent := []func(gtx C) D{
		func(gtx C) D {
			return pg.recentTransactionsSection(gtx)
		},
		func(gtx C) D {
			return layout.Inset{Bottom: values.MarginPadding20}.Layout(gtx, pg.syncStatusSection)
		},
	}

	return uniformPadding(gtx, func(gtx C) D {
		return pg.listContainer.Layout(gtx, len(pageContent), func(gtx C, i int) D {
			return layout.UniformInset(values.MarginPadding5).Layout(gtx, pageContent[i])
		})
	})
}

// syncDetail returns a walletSyncDetails object containing data of a single wallet sync box
func (pg *OverviewPage) syncDetail(name, status, headersFetched, progress string) walletSyncDetails {
	return walletSyncDetails{
		name:               pg.theme.Body1(name),
		status:             pg.theme.Body2(status),
		blockHeaderFetched: pg.theme.Body1(headersFetched),
		syncingProgress:    pg.theme.Body1(progress),
	}
}

// recentTransactionsSection lays out the list of recent transactions.
func (pg *OverviewPage) recentTransactionsSection(gtx layout.Context) layout.Dimensions {

	if len(pg.transactions) != len(pg.toTransactionDetails) {
		pg.toTransactionDetails = createClickGestures(len(pg.transactions))
	}

	return pg.theme.Card().Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		padding := values.MarginPadding15
		return Container{layout.Inset{Top: padding, Bottom: padding}}.Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					title := pg.theme.Body2(values.String(values.StrRecentTransactions))
					title.Color = pg.theme.Color.Gray3
					return pg.titleRow(gtx, title.Layout, pg.toTransactions.Layout)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: values.MarginPadding16}.Layout(gtx, pg.theme.Separator().Layout)
				}),
				layout.Rigid(func(gtx C) D {
					if len(pg.transactions) == 0 {
						message := pg.theme.Body1(values.String(values.StrNoTransactionsYet))
						message.Color = pg.theme.Color.Gray2
						return Container{layout.Inset{
							Left:   values.MarginPadding16,
							Bottom: values.MarginPadding3,
							Top:    values.MarginPadding18,
						}}.Layout(gtx, message.Layout)
					}

					return pg.transactionsList.Layout(gtx, len(pg.transactions), func(gtx C, i int) D {
						click := pg.toTransactionDetails[i]
						pointer.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Add(gtx.Ops)
						click.Add(gtx.Ops)
						var row = TransactionRow{
							transaction: pg.transactions[i],
							index:       i,
							showBadge:   len(pg.allWallets) > 1,
						}
						return layout.Inset{Left: values.MarginPadding16}.Layout(gtx, func(gtx C) D {
							return transactionRow(gtx, pg.Load, row)
						})
					})
				}),
			)
		})
	})
}

// syncStatusSection lays out content for displaying sync status.
func (pg *OverviewPage) syncStatusSection(gtx layout.Context) layout.Dimensions {
	uniform := layout.Inset{Top: values.MarginPadding5, Bottom: values.MarginPadding5}
	return pg.theme.Card().Layout(gtx, func(gtx C) D {
		return Container{layout.Inset{
			Top:    values.MarginPadding15,
			Bottom: values.MarginPadding16,
		}}.Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return pg.syncBoxTitleRow(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: values.MarginPadding16}.Layout(gtx, pg.theme.Separator().Layout)
				}),
				layout.Rigid(func(gtx C) D {
					return Container{layout.UniformInset(values.MarginPadding16)}.Layout(gtx, func(gtx C) D {
						return layout.Flex{}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								return pg.syncStatusIcon(gtx)
							}),
							layout.Rigid(func(gtx C) D {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										return pg.syncStatusTextRow(gtx, uniform)
									}),
									layout.Rigid(func(gtx C) D {
										if !pg.walletSyncing {
											return pg.syncDormantContent(gtx, uniform)
										}
										return layout.Dimensions{}
									}),
									layout.Rigid(func(gtx C) D {
										if pg.walletSyncing {
											return pg.progressBarRow(gtx, uniform)
										}
										return layout.Dimensions{}
									}),
									layout.Rigid(func(gtx C) D {
										if pg.walletSyncing {
											return pg.progressStatusRow(gtx, uniform)
										}
										return layout.Dimensions{}
									}),
								)
							}),
						)
					})
				}),
				layout.Rigid(func(gtx C) D {
					if pg.walletSyncing {
						return pg.theme.Separator().Layout(gtx)
					}
					return layout.Dimensions{}
				}),
				layout.Rigid(func(gtx C) D {
					if pg.walletSyncing && pg.syncDetailsVisibility {
						return Container{layout.UniformInset(values.MarginPadding16)}.Layout(gtx, func(gtx C) D {
							return pg.walletSyncRow(gtx, uniform)
						})
					}
					return layout.Dimensions{}
				}),
				layout.Rigid(func(gtx C) D {
					if pg.walletSyncing && pg.syncDetailsVisibility {
						return pg.theme.Separator().Layout(gtx)
					}
					return layout.Dimensions{}
				}),
				layout.Rigid(func(gtx C) D {
					if pg.walletSyncing {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return layout.Inset{Top: values.MarginPadding14}.Layout(gtx, func(gtx C) D {
							return pg.toggleSyncDetails.Layout(gtx)
						})
					}
					return layout.Dimensions{}
				}),
			)
		})
	})
}

func (pg OverviewPage) titleRow(gtx layout.Context, leftWidget, rightWidget func(C) D) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	titlePadding := values.MarginPadding15
	return Container{layout.Inset{
		Left:   titlePadding,
		Right:  titlePadding,
		Bottom: titlePadding,
	}}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return leftWidget(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				return rightWidget(gtx)
			}),
		)
	})
}

// syncDormantContent lays out sync status content when the wallet is synced or not connected
func (pg *OverviewPage) syncDormantContent(gtx layout.Context, uniform layout.Inset) layout.Dimensions {
	return uniform.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Bottom: values.MarginPadding12}.Layout(gtx, pg.blockInfoRow)
			}),
			layout.Rigid(func(gtx C) D {
				if pg.walletSynced {
					return pg.connectionPeer(gtx)
				}
				latestBlockTitleLabel := pg.theme.Body1(values.String(values.StrNoConnectedPeer))
				latestBlockTitleLabel.Color = pg.theme.Color.Gray
				return latestBlockTitleLabel.Layout(gtx)
			}),
		)
	})
}

func (pg *OverviewPage) blockInfoRow(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			latestBlockTitleLabel := pg.theme.Body1(values.String(values.StrLastBlockHeight))
			latestBlockTitleLabel.Color = pg.theme.Color.Gray
			return latestBlockTitleLabel.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Left:  values.MarginPadding5,
				Right: values.MarginPadding5,
			}.Layout(gtx, pg.theme.Body1(fmt.Sprintf("%d", pg.bestBlock.Height)).Layout)
		}),
		layout.Rigid(func(gtx C) D {
			pg.walletStatusIcon.Color = pg.theme.Color.Gray
			return layout.Inset{Right: values.MarginPadding10, Top: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
				return pg.walletStatusIcon.Layout(gtx, values.MarginPadding5)
			})
		}),
		layout.Rigid(func(gtx C) D {
			currentSeconds := time.Now().UnixNano() / int64(time.Second)
			return layout.Inset{Right: values.MarginPadding5}.Layout(gtx, pg.theme.Body1(wallet.SecondsToDays(currentSeconds-pg.bestBlock.Timestamp)).Layout)
		}),
		layout.Rigid(func(gtx C) D {
			lastSyncedLabel := pg.theme.Body1(values.String(values.StrAgo))
			lastSyncedLabel.Color = pg.theme.Color.Gray
			return lastSyncedLabel.Layout(gtx)
		}),
	)
}

func (pg *OverviewPage) connectionPeer(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			connectedPeersInfoLabel := pg.theme.Body1(values.String(values.StrConnectedTo))
			connectedPeersInfoLabel.Color = pg.theme.Color.Gray
			return connectedPeersInfoLabel.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Left: values.MarginPadding5, Right: values.MarginPadding5}.Layout(gtx, pg.theme.Body1(fmt.Sprintf("%d", pg.connectedPeers)).Layout)
		}),
		layout.Rigid(func(gtx C) D {
			peersLabel := pg.theme.Body1("peers")
			peersLabel.Color = pg.theme.Color.Gray
			return peersLabel.Layout(gtx)
		}),
	)
}

// syncBoxTitleRow lays out widgets in the title row inside the sync status box.
func (pg *OverviewPage) syncBoxTitleRow(gtx layout.Context) layout.Dimensions {
	title := pg.theme.Body2(values.String(values.StrWalletStatus))
	title.Color = pg.theme.Color.Gray3
	statusLabel := pg.theme.Body1(values.String(values.StrOffline))
	pg.walletStatusIcon.Color = pg.theme.Color.Danger
	if pg.isConnnected {
		statusLabel.Text = values.String(values.StrOnline)
		pg.walletStatusIcon.Color = pg.theme.Color.Success
	}

	syncStatus := func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Top: values.MarginPadding4, Right: values.MarginPadding4}.Layout(gtx, func(gtx C) D {
					return pg.walletStatusIcon.Layout(gtx, values.MarginPadding14)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return statusLabel.Layout(gtx)
			}),
		)
	}
	return pg.titleRow(gtx, title.Layout, syncStatus)
}

// syncStatusTextRow lays out sync status text and sync button.
func (pg *OverviewPage) syncStatusTextRow(gtx layout.Context, inset layout.Inset) layout.Dimensions {
	syncStatusLabel := pg.theme.H6(values.String(values.StrWalletNotSynced))
	if pg.walletSyncing {
		syncStatusLabel.Text = values.String(values.StrSyncingState)
	} else if pg.walletSynced {
		syncStatusLabel.Text = values.String(values.StrSynced)
	}

	return inset.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Flexed(1, syncStatusLabel.Layout),
			layout.Rigid(func(gtx C) D {
				// stack a button on a card widget to produce a transparent button.
				return layout.E.Layout(gtx, func(gtx C) D {
					gtx.Constraints.Max.Y = pg.syncButtonHeight
					border := widget.Border{Color: pg.theme.Color.Hint, CornerRadius: values.MarginPadding10, Width: values.MarginPadding1}
					return border.Layout(gtx, func(gtx C) D {
						pg.sync.Inset = layout.Inset{
							Top:    values.MarginPadding5,
							Bottom: values.MarginPadding5,
							Left:   values.MarginPadding10,
							Right:  values.MarginPadding10,
						}
						pg.sync.CornerRadius = values.MarginPadding10
						if pg.isConnnected {
							pg.sync.Text = values.String(values.StrDisconnect)
						} else {
							pg.sync.Text = values.String(values.StrReconnect)
							pg.sync.Inset.Left = values.MarginPadding25
							layout.Inset{Top: values.MarginPadding4, Left: values.MarginPadding7}.Layout(gtx, func(gtx C) D {
								pg.cachedIcon.Color = pg.theme.Color.Gray
								return pg.cachedIcon.Layout(gtx, values.TextSize14)
							})
						}

						return pg.sync.Layout(gtx)
					})
				})
			}),
		)
	})
}

func (pg *OverviewPage) syncStatusIcon(gtx layout.Context) layout.Dimensions {
	syncStatusIcon := pg.notSyncedIcon
	if pg.walletSynced {
		syncStatusIcon = pg.syncedIcon
	}
	i := layout.Inset{Right: values.MarginPadding16, Top: values.MarginPadding9}
	if pg.walletSyncing {
		return i.Layout(gtx, pg.syncingIcon.Layout)
	}
	return i.Layout(gtx, func(gtx C) D {
		return syncStatusIcon.Layout(gtx, values.MarginPadding24)
	})
}

// progressBarRow lays out the progress bar.
func (pg *OverviewPage) progressBarRow(gtx layout.Context, inset layout.Inset) layout.Dimensions {
	return inset.Layout(gtx, func(gtx C) D {
		p := pg.theme.ProgressBar(pg.syncProgress)
		p.Height = values.MarginPadding8
		p.Radius = values.MarginPadding4
		p.Color = pg.theme.Color.Success
		return p.Layout(gtx)
	})
}

// progressStatusRow lays out the progress status when the wallet is syncing.
func (pg *OverviewPage) progressStatusRow(gtx layout.Context, inset layout.Inset) layout.Dimensions {
	timeLeft := pg.remainingSyncTime
	if timeLeft == "" {
		timeLeft = "0s"
	}

	percentageLabel := pg.theme.Body1(fmt.Sprintf("%v%%", pg.syncProgress))
	timeLeftLabel := pg.theme.Body1(fmt.Sprintf("%v Left", timeLeft))
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return endToEndRow(gtx, percentageLabel.Layout, timeLeftLabel.Layout)
	})

}

//	walletSyncRow layouts a list of wallet sync boxes horizontally.
func (pg *OverviewPage) walletSyncRow(gtx layout.Context, inset layout.Inset) layout.Dimensions {
	return layout.Inset{Top: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				completedSteps := pg.theme.Body2(values.StringF(values.StrSyncSteps, pg.syncStep))
				completedSteps.Color = pg.theme.Color.Gray
				headersFetched := pg.theme.Body1(values.StringF(values.StrFetchingBlockHeaders, pg.headerFetchProgress))
				return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return endToEndRow(gtx, completedSteps.Layout, headersFetched.Layout)
				})
			}),
			layout.Rigid(func(gtx C) D {
				connectedPeersTitleLabel := pg.theme.Body2(values.String(values.StrConnectedPeersCount))
				connectedPeersTitleLabel.Color = pg.theme.Color.Gray
				connectedPeersLabel := pg.theme.Body1(fmt.Sprintf("%d", pg.connectedPeers))
				return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return endToEndRow(gtx, connectedPeersTitleLabel.Layout, connectedPeersLabel.Layout)
				})
			}),
			layout.Rigid(func(gtx C) D {
				var walletSyncBoxes []layout.Widget

				currentSeconds := time.Now().UnixNano() / int64(time.Second)
				for i := 0; i < len(pg.allWallets); i++ {
					w := pg.allWallets[i]

					status := "syncing..."
					if w.IsWaiting() {
						status = "waiting..."
					}

					blockHeightProgress := values.StringF(values.StrBlockHeaderFetchedCount, w.GetBestBlock(), pg.headersToFetchOrScan)
					daysBehind := wallet.SecondsToDays(currentSeconds - w.GetBestBlockTimeStamp())
					details := pg.syncDetail(w.Name, status, blockHeightProgress, daysBehind)
					uniform := layout.UniformInset(values.MarginPadding5)
					walletSyncBoxes = append(walletSyncBoxes,
						func(gtx C) D {
							return pg.walletSyncBox(gtx, uniform, details)
						})
				}

				return pg.walletSyncList.Layout(gtx, len(walletSyncBoxes), func(gtx C, i int) D {
					return walletSyncBoxes[i](gtx)
				})
			}),
		)
	})
}

// walletSyncBox lays out the wallet syncing details of a single wallet.
func (pg *OverviewPage) walletSyncBox(gtx layout.Context, inset layout.Inset, details walletSyncDetails) layout.Dimensions {
	return layout.Inset{Top: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		card := pg.theme.Card()
		card.Color = pg.theme.Color.LightGray
		return card.Layout(gtx, func(gtx C) D {
			return Container{
				layout.UniformInset(values.MarginPadding16),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return endToEndRow(gtx, details.name.Layout, details.status.Layout)
						})
					}),
					layout.Rigid(func(gtx C) D {
						headersFetchedTitleLabel := pg.theme.Body2(values.String(values.StrBlockHeaderFetched))
						headersFetchedTitleLabel.Color = pg.theme.Color.Gray
						return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return endToEndRow(gtx, headersFetchedTitleLabel.Layout, details.blockHeaderFetched.Layout)
						})
					}),
					layout.Rigid(func(gtx C) D {
						progressTitleLabel := pg.theme.Body2(values.String(values.StrSyncingProgress))
						progressTitleLabel.Color = pg.theme.Color.Gray
						return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return endToEndRow(gtx, progressTitleLabel.Layout, details.syncingProgress.Layout)
						})
					}),
				)
			})
		})
	})
}

func (pg *OverviewPage) Handle() {
	eq := pg.queue

	if pg.sync.Button.Clicked() {
		go pg.ToggleSync()
	}

	if pg.toTransactions.Button.Clicked() {
		pg.ChangeFragment(NewTransactionsPage(pg.Load), Transactions)
	}

	// todo: resolve navigation to transaction details page
	for index, click := range pg.toTransactionDetails {
		for _, e := range click.Events(eq) {
			if e.Type == gesture.TypeClick {
				txn := pg.transactions[index]

				pg.SetReturnPage(OverviewPageID)
				pg.ChangeFragment(NewTransactionDetailsPage(pg.Load, &txn), TransactionDetailsPageID)
				return
			}
		}
	}

	if pg.toggleSyncDetails.Button.Clicked() {
		pg.syncDetailsVisibility = !pg.syncDetailsVisibility
		if pg.syncDetailsVisibility {
			pg.toggleSyncDetails.Text = values.String(values.StrHideDetails)
		} else {
			pg.toggleSyncDetails.Text = values.String(values.StrShowDetails)
		}
	}
}

func (pg *OverviewPage) listenForSyncNotifications() {
	go func() {
		for {
			var notification interface{}

			select {
			case notification = <-pg.Receiver.NotificationsUpdate:
			case <-pg.pageClosing:
				return
			}

			switch n := notification.(type) {
			case wallet.NewTransaction:
				pg.loadTransactions()
			case wallet.SyncStatusUpdate:
				switch t := n.ProgressReport.(type) {
				case wallet.SyncHeadersFetchProgress:
					pg.headerFetchProgress = t.Progress.HeadersFetchProgress
					pg.headersToFetchOrScan = t.Progress.TotalHeadersToFetch
					pg.syncProgress = int(t.Progress.TotalSyncProgress)
					pg.remainingSyncTime = wallet.SecondsToDays(t.Progress.TotalTimeRemainingSeconds)
					pg.syncStep = wallet.FetchHeadersSteps
				case wallet.SyncAddressDiscoveryProgress:
					pg.syncProgress = int(t.Progress.TotalSyncProgress)
					pg.remainingSyncTime = wallet.SecondsToDays(t.Progress.TotalTimeRemainingSeconds)
					pg.syncStep = wallet.AddressDiscoveryStep
				case wallet.SyncHeadersRescanProgress:
					pg.headersToFetchOrScan = t.Progress.TotalHeadersToScan
					pg.syncProgress = int(t.Progress.TotalSyncProgress)
					pg.remainingSyncTime = wallet.SecondsToDays(t.Progress.TotalTimeRemainingSeconds)
					pg.syncStep = wallet.RescanHeadersStep
				}

				switch n.Stage {
				case wallet.PeersConnected:
					pg.connectedPeers = n.ConnectedPeers
				case wallet.SyncStarted:
					fallthrough
				case wallet.SyncCanceled:
					fallthrough
				case wallet.SyncCompleted:
					pg.loadTransactions()
					pg.walletSyncing = pg.WL.MultiWallet.IsSyncing()
					pg.walletSynced = pg.WL.MultiWallet.IsSynced()
					pg.isConnnected = pg.WL.MultiWallet.IsConnectedToDecredNetwork()
				case wallet.BlockAttached:
					pg.bestBlock = pg.WL.MultiWallet.GetBestBlock()
				}
			}

			pg.RefreshWindow()

		}
	}()
}

func (pg *OverviewPage) OnClose() {
	pg.pageClosing <- true
}
