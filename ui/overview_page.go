package ui

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/raedahgroup/godcr/ui/values"

	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"

	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/raedahgroup/dcrlibwallet"
	"github.com/raedahgroup/godcr/ui/decredmaterial"
	"github.com/raedahgroup/godcr/wallet"
)

const PageOverview = "overview"

type overviewPageText struct {
	balanceTitle,
	statusTitle,
	stepsTitle,
	transactionsTitle,
	connectedPeersTitle,
	headersFetchedTitle,
	syncingProgressTitle,
	latestBlockTitle,
	lastSyncedTitle,
	noTransaction,
	offlineStatus,
	onlineStatus,
	syncingStatus,
	notSyncedStatus,
	syncedStatus,
	fetchingBlockHeaders,
	reconnect,
	disconnect,
	noWallet,
	cancel,
	viewAllTx,
	connectedPeersInfo,
	noConnectedPeers string
}

// walletSyncDetails contains sync data for each wallet when a sync
// is in progress.
type walletSyncDetails struct {
	name               decredmaterial.Label
	status             decredmaterial.Label
	blockHeaderFetched decredmaterial.Label
	syncingProgress    decredmaterial.Label
}

type transactionWidgets struct {
	wallet      decredmaterial.Label
	balance     string
	direction   *decredmaterial.Icon
	mainBalance decredmaterial.Label
	subBalance  decredmaterial.Label
	date        decredmaterial.Label
	status      decredmaterial.Label
}

type overviewPage struct {
	listContainer, walletSyncList *layout.List
	gtx                           *layout.Context
	theme                         *decredmaterial.Theme
	tab                           *decredmaterial.Tabs

	walletInfo                    *wallet.MultiWalletInfo
	walletSyncStatus              *wallet.SyncStatus
	walletTransactions            **wallet.Transactions
	walletTransaction             **wallet.Transaction
	toTransactions, sync          decredmaterial.Button
	toTransactionsW, syncW        widget.Button
	syncedIcon, notSyncedIcon,
	walletStatusIcon *decredmaterial.Icon
	syncingIcon          image.Image
	toTransactionDetails []*gesture.Click
	line                 *decredmaterial.Line

	text             overviewPageText
	syncButtonHeight int
	syncButtonWidth  int
	moreButtonWidth  int
	moreButtonHeight int
	gray             color.RGBA
}

func (win *Window) OverviewPage(c pageCommon) layout.Widget {
	page := overviewPage{
		gtx:   c.gtx,
		theme: c.theme,
		tab:   c.navTab,

		walletInfo:         win.walletInfo,
		walletSyncStatus:   win.walletSyncStatus,
		walletTransactions: &win.walletTransactions,
		walletTransaction:  &win.walletTransaction,
		listContainer:      &layout.List{Axis: layout.Vertical},
		walletSyncList:     &layout.List{Axis: layout.Horizontal},
		line:               c.theme.Line(),

		syncButtonHeight: 70,
		syncButtonWidth:  145,
		moreButtonWidth:  115,
		moreButtonHeight: 70,

		iconPadding: unit.Dp(0),
		iconSize:    unit.Dp(25),
		gray:        color.RGBA{137, 151, 165, 255},
	}
	page.text = overviewPageText{
		balanceTitle:         "Current Total Balance",
		statusTitle:          "Wallet Status",
		stepsTitle:           "Step",
		transactionsTitle:    "Recent Transactions",
		connectedPeersTitle:  "Connected peers count",
		connectedPeersInfo:   "Connected to",
		noConnectedPeers:     "No connected peers",
		headersFetchedTitle:  "Block header fetched",
		syncingProgressTitle: "Syncing progress",
		latestBlockTitle:     "Last Block Height",
		lastSyncedTitle:      "Last Block Mined",
		noTransaction:        "no transactions",
		noWallet:             "No wallet loaded",
		offlineStatus:        "Offline",
		onlineStatus:         "Online",
		syncingStatus:        "Syncing...",
		notSyncedStatus:      "Not Synced",
		syncedStatus:         "Synced",
		fetchingBlockHeaders: "Fetching block headers",
		reconnect:            "Connect",
		disconnect:           "Disconnect",
		cancel:               "Cancel",
		viewAllTx:            "See All",
	}

	page.toTransactions = c.theme.Button(page.text.viewAllTx)
	page.toTransactions.TextSize = values.TextSize14
	page.toTransactions.Background = color.RGBA{}
	page.toTransactions.Color = c.theme.Color.Primary
	page.toTransactions.Inset = layout.Inset{
		Top: unit.Dp(10), Bottom: unit.Dp(0),
		Left: unit.Dp(0), Right: unit.Dp(0),
	}

	page.sync = c.theme.Button(page.text.reconnect)
	page.sync.TextSize = values.TextSize10

	page.sync.Background = c.theme.Color.Background
	page.sync.Color = c.theme.Color.Text

	page.syncedIcon = c.icons.actionCheckCircle
	page.syncedIcon.Color = c.theme.Color.Success

	page.syncingIcon = c.icons.syncingIcon

	page.notSyncedIcon = c.icons.navigationCancel
	page.notSyncedIcon.Color = c.theme.Color.Danger

	page.walletStatusIcon = c.icons.imageBrightness1

	page.line.Color = c.theme.Color.Gray

	return func() {
		page.Layout(c)
		page.Handler(c)
	}
}

// Layout lays out the entire content for overview page.
func (page *overviewPage) Layout(c pageCommon) {
	if c.info.LoadedWallets == 0 {
		c.Layout(c.gtx, func() {
			layout.Center.Layout(c.gtx, func() {
				c.theme.H3(page.text.noWallet).Layout(c.gtx)
			})
		})
		return
	}

	gtx := page.gtx
	walletInfo := page.walletInfo
	theme := page.theme

	pageContent := []func(){
		func() {
			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func() {
					mainBalance := theme.H4("")
					subBalance := theme.H6("")
					page.layoutBalance(walletInfo.TotalBalance, mainBalance, subBalance)
				}),
				layout.Rigid(func() {
					theme.Caption(page.text.balanceTitle).Layout(gtx)
				}),
			)
		},
		func() {
			page.recentTransactionsColumn(c)
		},
		func() {
			layout.Inset{Bottom: values.MarginPadding20}.Layout(gtx, func() {
				page.syncStatusColumn()
			})
		},
	}

	c.Layout(c.gtx, func() {
		page.listContainer.Layout(gtx, len(pageContent), func(i int) {
			layout.UniformInset(values.MarginPadding5).Layout(gtx, pageContent[i])
		})
	})
}

// syncDetail returns a walletSyncDetails object containing data of a single wallet sync box
func (page *overviewPage) syncDetail(name, status, headersFetched, progress string) walletSyncDetails {
	return walletSyncDetails{
		name:               page.theme.Caption(name),
		status:             page.theme.Caption(status),
		blockHeaderFetched: page.theme.Caption(headersFetched),
		syncingProgress:    page.theme.Caption(progress),
	}
}

// recentTransactionsColumn lays out the list of recent transactions.
func (page *overviewPage) recentTransactionsColumn(c pageCommon) {
	theme := page.theme
	gtx := page.gtx
	var transactionRows []func()

	if len((*page.walletTransactions).Txs) > 0 {
		page.updateToTransactionDetailsButtons()

		for index, txn := range (*page.walletTransactions).Recent {
			txnWidgets := transactionWidgets{
				wallet:      theme.Body1(txn.WalletName),
				balance:     txn.Balance,
				mainBalance: theme.Body1(""),
				subBalance:  theme.Caption(""),
				date:        theme.Body1(txn.DateTime),
				status:      theme.Body1(txn.Status),
			}
			if txn.Txn.Direction == dcrlibwallet.TxDirectionSent {
				txnWidgets.direction = c.icons.contentRemove
				txnWidgets.direction.Color = c.theme.Color.Danger
			} else {
				txnWidgets.direction = c.icons.contentAdd
				txnWidgets.direction.Color = c.theme.Color.Success
			}

			click := page.toTransactionDetails[index]

			transactionRows = append(transactionRows, func() {
				page.recentTransactionRow(txnWidgets)
				pointer.Rect(image.Rectangle{Max: gtx.Dimensions.Size}).Add(gtx.Ops)
				click.Add(gtx.Ops)
			})
		}
	} else {
		transactionRows = append(transactionRows, func() {
			page.centralize(func() {
				label := theme.Caption(page.text.noTransaction)
				label.Color = page.gray
				label.Layout(gtx)
			})
		})
	}

	page.drawlayout(func() {
		layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func() {
				theme.Caption(page.text.transactionsTitle).Layout(page.gtx)
			}),
			layout.Rigid(func() {
				list := &layout.List{Axis: layout.Vertical}
				page.centralize(func() {
					list.Layout(page.gtx, len(transactionRows), func(i int) {
						layout.Inset{Top: values.MarginPadding5}.Layout(page.gtx, transactionRows[i])
					})
				})

			}),
			layout.Rigid(func() {
				page.line.Width = gtx.Constraints.Width.Max
				page.line.Layout(gtx)
			}),
			layout.Rigid(func() {
				page.centralize(func() {
					page.toTransactions.Layout(gtx, &page.toTransactionsW)
				})
			}),
		)
	})
}

func (page *overviewPage) centralize(content layout.Widget) {
	layout.Flex{Axis: layout.Horizontal}.Layout(page.gtx,
		layout.Flexed(1, func() {
			layout.Center.Layout(page.gtx, content)
		}),
	)
}

// recentTransactionRow lays out a single row of a recent transaction.
func (page *overviewPage) recentTransactionRow(txn transactionWidgets) {
	gtx := page.gtx
	margin := layout.UniformInset(values.MarginPadding10)

	layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func() {
			gtx.Constraints.Width.Min = gtx.Px(unit.Dp(50))
			layout.Inset{Top: unit.Dp(12)}.Layout(gtx, func() {
				txn.direction.Layout(gtx, unit.Dp(16))
			})
		}),
		layout.Rigid(func() {
			gtx.Constraints.Width.Min = gtx.Px(unit.Dp(100))
			margin.Layout(gtx, func() {
				page.layoutBalance(txn.balance, txn.mainBalance, txn.subBalance)
			})
		}),
		layout.Rigid(func() {
			gtx.Constraints.Width.Min = gtx.Px(unit.Dp(100))
			margin.Layout(gtx, func() {
				txn.wallet.Layout(gtx)
			})
		}),
		layout.Rigid(func() {
			gtx.Constraints.Width.Min = gtx.Px(unit.Dp(100))
			margin.Layout(gtx, func() {
				txn.date.Layout(gtx)
			})
		}),
		layout.Rigid(func() {
			margin.Layout(gtx, func() {
				txn.status.Layout(gtx)
			})
		}),
	)
}

// syncStatusColumn lays out content for displaying sync status.
func (page *overviewPage) syncStatusColumn() {
	gtx := page.gtx
	uniform := layout.UniformInset(values.MarginPadding5)
	page.drawlayout(func() {
		layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func() {
				page.syncBoxTitleRow(uniform)
			}),
			layout.Rigid(func() {
				page.syncStatusTextRow(uniform)
			}),
			layout.Rigid(func() {
				if page.walletInfo.Syncing {
					page.syncActiveContent(uniform)
				} else {
					page.syncDormantContent()
				}
			}),
		)
	})
}

// drawlayout wraps the page tx and sync section in a card layout
func (page *overviewPage) drawlayout(body layout.Widget) {
	decredmaterial.Card{Color: page.theme.Color.Surface}.Layout(page.gtx, func() {
		layout.UniformInset(values.MarginPadding20).Layout(page.gtx, body)
	})
}

// syncingContent lays out sync status content when the wallet is syncing
func (page *overviewPage) syncActiveContent(uniform layout.Inset) {
	layout.Flex{Axis: layout.Vertical}.Layout(page.gtx,
		layout.Rigid(func() {
			page.progressBarRow(uniform)
		}),
		layout.Rigid(func() {
			page.progressStatusRow(uniform)
		}),
		layout.Rigid(func() {
			page.walletSyncRow(uniform)
		}),
	)
}

// syncDormantContent lays out sync status content when the wallet is synced or not connected
func (page *overviewPage) syncDormantContent() {
	layout.Inset{Left: unit.Dp(45)}.Layout(page.gtx, func() {
		layout.Flex{Axis: layout.Vertical}.Layout(page.gtx,
			layout.Rigid(func() {
				layout.Inset{Bottom: unit.Dp(12)}.Layout(page.gtx, func() {
					page.blockInfoRow()
				})
			}),
			layout.Rigid(func() {
				if page.walletInfo.Synced {
					page.connectionPeer()
				} else {
					latestBlockTitleLabel := page.theme.Body1(page.text.noConnectedPeers)
					latestBlockTitleLabel.Color = page.gray
					latestBlockTitleLabel.Layout(page.gtx)
				}
			}),
		)
	})
}

func (page *overviewPage) blockInfoRow() {
	layout.Flex{Axis: layout.Horizontal}.Layout(page.gtx,
		layout.Rigid(func() {
			latestBlockTitleLabel := page.theme.Body1(page.text.latestBlockTitle)
			latestBlockTitleLabel.Color = page.gray
			latestBlockTitleLabel.Layout(page.gtx)
		}),
		layout.Rigid(func() {
			layout.Inset{Left: values.MarginPadding5, Right: values.MarginPadding5}.Layout(page.gtx, func() {
				page.theme.Body1(fmt.Sprintf("%v", page.walletInfo.BestBlockHeight)).Layout(page.gtx)
			})
		}),
		layout.Rigid(func() {
			page.walletStatusIcon.Color = page.gray
			layout.Inset{Right: unit.Dp(9), Top: unit.Dp(8)}.Layout(page.gtx, func() {
				page.walletStatusIcon.Layout(page.gtx, values.MarginPadding5)
			})
		}),
		layout.Rigid(func() {
			layout.Inset{Right: values.MarginPadding5}.Layout(page.gtx, func() {
				page.theme.Body1(fmt.Sprintf("%v", page.walletInfo.LastSyncTime)).Layout(page.gtx)
			})
		}),
		layout.Rigid(func() {
			lastSyncedLabel := page.theme.Body1("ago")
			lastSyncedLabel.Color = page.gray
			lastSyncedLabel.Layout(page.gtx)
		}),
	)
}

func (page *overviewPage) connectionPeer() {
	layout.Flex{Axis: layout.Horizontal}.Layout(page.gtx,
		layout.Rigid(func() {
			connectedPeersInfoLabel := page.theme.Body1(page.text.connectedPeersInfo)
			connectedPeersInfoLabel.Color = page.gray
			connectedPeersInfoLabel.Layout(page.gtx)
		}),
		layout.Rigid(func() {
			layout.Inset{Left: values.MarginPadding5, Right: values.MarginPadding5}.Layout(page.gtx, func() {
				page.theme.Body1(fmt.Sprintf("%d", page.walletSyncStatus.ConnectedPeers)).Layout(page.gtx)
			})
		}),
		layout.Rigid(func() {
			peersLabel := page.theme.Body1("peers")
			peersLabel.Color = page.gray
			peersLabel.Layout(page.gtx)
		}),
	)
}

// endToEndRow layouts out its content on both ends of its horizontal layout.
func (page *overviewPage) endToEndRow(inset layout.Inset, LeftLabel, rightLabel decredmaterial.Label) {
	gtx := page.gtx
	inset.Layout(gtx, func() {
		layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func() {
				LeftLabel.Layout(gtx)
			}),
			layout.Flexed(1, func() {
				layout.E.Layout(gtx, func() {
					rightLabel.Layout(gtx)
				})
			}),
		)
	})
}

// syncBoxTitleRow lays out widgets in the title row inside the sync status box.
func (page *overviewPage) syncBoxTitleRow(inset layout.Inset) {
	statusTitleLabel := page.theme.Caption(page.text.statusTitle)
	statusTitleLabel.Color = page.theme.Color.Text
	statusLabel := page.theme.Body1(page.text.offlineStatus)
	page.walletStatusIcon.Color = page.theme.Color.Danger
	if page.walletInfo.Synced || page.walletInfo.Syncing {
		statusLabel.Text = page.text.onlineStatus
		page.walletStatusIcon.Color = page.theme.Color.Success
	}

	gtx := page.gtx
	inset.Layout(gtx, func() {
		layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func() {
				statusTitleLabel.Layout(gtx)
			}),
			layout.Flexed(1, func() {
				layout.E.Layout(gtx, func() {
					layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func() {
							layout.Inset{Top: unit.Dp(6), Right: unit.Dp(20)}.Layout(gtx, func() {
								page.walletStatusIcon.Layout(gtx, unit.Dp(10))
							})
						}),
						layout.Rigid(func() {
							statusLabel.Layout(gtx)
						}),
					)
				})
			}),
		)
	})
}

// syncStatusTextRow lays out sync status text and sync button.
func (page *overviewPage) syncStatusTextRow(inset layout.Inset) {
	gtx := page.gtx
	syncStatusLabel := page.theme.H6(page.text.notSyncedStatus)
	syncStatusIcon := page.notSyncedIcon
	if page.walletInfo.Syncing {
		syncStatusLabel.Text = page.text.syncingStatus
	} else if page.walletInfo.Synced {
		syncStatusLabel.Text = page.text.syncedStatus
		syncStatusIcon = page.syncedIcon
	}

	inset.Layout(gtx, func() {
		layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func() {
				if page.walletInfo.Syncing {
					layout.Inset{Right: unit.Dp(20)}.Layout(gtx, func() {
						page.theme.ImageIcon(gtx, page.syncingIcon, 50)
					})
				} else {
					layout.Inset{Right: unit.Dp(40)}.Layout(gtx, func() {
						syncStatusIcon.Layout(gtx, page.iconSize)
					})
				}
			}),
			layout.Flexed(0.5, func() {
				layout.W.Layout(gtx, func() {
					syncStatusLabel.Layout(page.gtx)
				})
			}),
			layout.Flexed(1, func() {
				// stack a button on a card widget to produce a transparent button.
				layout.E.Layout(gtx, func() {
					gtx.Constraints.Width.Min = page.syncButtonWidth
					gtx.Constraints.Width.Min = page.syncButtonWidth
					gtx.Constraints.Height.Max = page.syncButtonHeight
					if page.walletInfo.Synced {
						page.sync.Text = page.text.disconnect
					}
					page.sync.Layout(gtx, &page.syncW)
				})
			}),
		)
	})
}

// syncBoxTitleRow lays out the progress bar.
func (page *overviewPage) progressBarRow(inset layout.Inset) {
	inset.Layout(page.gtx, func() {
		progress := page.walletSyncStatus.Progress
		page.gtx.Constraints.Height.Max = 20
		page.theme.ProgressBar(float64(progress)).Layout(page.gtx)
	})
}

// syncBoxTitleRow lays out the progress status when the wallet is syncing.
func (page *overviewPage) progressStatusRow(inset layout.Inset) {
	timeLeft := page.walletSyncStatus.RemainingTime
	if timeLeft == "" {
		timeLeft = "0s"
	}

	percentageLabel := page.theme.Body1(fmt.Sprintf("%v%%", page.walletSyncStatus.Progress))
	timeLeftLabel := page.theme.Body1(fmt.Sprintf("%v Left", timeLeft))
	page.endToEndRow(inset, percentageLabel, timeLeftLabel)
}

//	walletSyncRow layouts a list of wallet sync boxes horizontally.
func (page *overviewPage) walletSyncRow(inset layout.Inset) {
	gtx := page.gtx
	layout.Inset{Top: values.MarginPadding30}.Layout(gtx, func() {
		layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func() {
				completedSteps := page.walletSyncStatus.Steps
				totalSteps := page.walletSyncStatus.TotalSteps
				completedStepsLabel := page.theme.Caption(fmt.Sprintf("%s %d/%d", page.text.stepsTitle, completedSteps, totalSteps))
				completedStepsLabel.Color = page.gray
				headersFetchedLabel := page.theme.Body1(fmt.Sprintf("%s. %v%%", page.text.fetchingBlockHeaders,
					page.walletSyncStatus.HeadersFetchProgress))
				headersFetchedLabel.Color = page.gray
				page.endToEndRow(inset, completedStepsLabel, headersFetchedLabel)
			}),
			layout.Rigid(func() {
				connectedPeersTitleLabel := page.theme.Caption(page.text.connectedPeersTitle)
				connectedPeersTitleLabel.Color = page.gray
				connectedPeersLabel := page.theme.Body1(fmt.Sprintf("%d", page.walletSyncStatus.ConnectedPeers))
				page.endToEndRow(inset, connectedPeersTitleLabel, connectedPeersLabel)
			}),
			layout.Rigid(func() {
				var overallBlockHeight int32
				var walletSyncBoxes []func()

				if page.walletSyncStatus != nil {
					overallBlockHeight = page.walletSyncStatus.HeadersToFetch
				}

				for i := 0; i < len(page.walletInfo.Wallets); i++ {
					w := page.walletInfo.Wallets[i]
					if w.BestBlockHeight > overallBlockHeight {
						overallBlockHeight = w.BestBlockHeight
					}
					blockHeightProgress := fmt.Sprintf("%v of %v", w.BestBlockHeight, overallBlockHeight)
					details := page.syncDetail(w.Name, w.Status, blockHeightProgress, w.DaysBehind)
					uniform := layout.UniformInset(values.MarginPadding5)
					walletSyncBoxes = append(walletSyncBoxes,
						func() {
							page.walletSyncBox(uniform, details)
						})
				}

				page.walletSyncList.Layout(gtx, len(walletSyncBoxes), func(i int) {
					if i == 0 {
						walletSyncBoxes[i]()
					} else {
						layout.Inset{Left: values.MarginPadding30}.Layout(gtx, walletSyncBoxes[i])
					}
				})
			}),
		)
	})
}

// walletSyncBox lays out the wallet syncing details of a single wallet.
func (page *overviewPage) walletSyncBox(inset layout.Inset, details walletSyncDetails) {
	gtx := page.gtx
	layout.Inset{Top: values.MarginPadding30}.Layout(gtx, func() {
		gtx.Constraints.Width.Min = gtx.Px(values.WalletSyncBoxContentWidth)
		gtx.Constraints.Width.Max = gtx.Constraints.Width.Min
		decredmaterial.Card{}.Layout(gtx, func() {
			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func() {
					page.endToEndRow(inset, details.name, details.status)
				}),
				layout.Rigid(func() {
					headersFetchedTitleLabel := page.theme.Caption(page.text.headersFetchedTitle)
					headersFetchedTitleLabel.Color = page.gray
					page.endToEndRow(inset, headersFetchedTitleLabel, details.blockHeaderFetched)
				}),
				layout.Rigid(func() {
					progressTitleLabel := page.theme.Caption(page.text.syncingProgressTitle)
					progressTitleLabel.Color = page.gray
					page.endToEndRow(inset, progressTitleLabel, details.syncingProgress)
				}),
			)
		})
	})
}

// layoutBalance aligns the main and sub DCR balances horizontally, putting the sub
// balance at the baseline of the row.
func (page *overviewPage) layoutBalance(amount string, main, sub decredmaterial.Label) {
	mainText, subText := page.breakBalance(amount)
	layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline}.Layout(page.gtx,
		layout.Rigid(func() {
			main.Text = mainText
			main.Layout(page.gtx)
		}),
		layout.Rigid(func() {
			sub.Text = subText
			sub.Layout(page.gtx)
		}),
	)
}

// breakBalance takes the balance string and returns it in two slices
func (page *overviewPage) breakBalance(balance string) (b1, b2 string) {
	balanceParts := strings.Split(balance, ".")
	if len(balanceParts) == 1 {
		return balanceParts[0], ""
	}
	b1 = balanceParts[0]
	b2 = balanceParts[1]
	b1 = b1 + "." + b2[:2]
	b2 = b2[2:]
	return
}

func (page *overviewPage) updateToTransactionDetailsButtons() {
	recentTxs := (*page.walletTransactions).Recent
	if len(recentTxs) != len(page.toTransactionDetails) {
		page.toTransactionDetails = make([]*gesture.Click, len(recentTxs))
		for i := range recentTxs {
			page.toTransactionDetails[i] = &gesture.Click{}
		}
	}
}

func (page *overviewPage) Handler(c pageCommon) {
	if page.syncW.Clicked(page.gtx) {
		if page.walletInfo.Synced || page.walletInfo.Syncing {
			c.wallet.CancelSync()
			page.sync.Text = page.text.reconnect
		} else {
			c.wallet.StartSync()
			page.sync.Text = page.text.cancel
		}
	}
	if page.toTransactionsW.Clicked(page.gtx) {
		page.tab.ChangeTab(4)
	}

	for index, click := range page.toTransactionDetails {
		for _, e := range click.Events(page.gtx) {
			if e.Type == gesture.TypeClick {
				txn := (*page.walletTransactions).Recent[index]
				*page.walletTransaction = &txn
				*c.page = PageTransactionDetails
				return
			}
		}
	}
}
