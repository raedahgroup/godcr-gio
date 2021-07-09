package ui

import (
	"strconv"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/decred/dcrd/dcrutil"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/page"
	"github.com/planetdecred/godcr/ui/values"
)

const PageMain = "Main"

type mainPage struct {
	*pageCommon

	load                    *load.Load
	appBarNavItems          []navHandler
	drawerNavItems          []navHandler
	isNavDrawerMinimized    bool
	minimizeNavDrawerButton decredmaterial.IconButton
	maximizeNavDrawerButton decredmaterial.IconButton

	autoSync bool

	current, previous string
	pages             map[string]Page

	// page state variables
	usdExchangeSet  bool
	totalBalance    dcrutil.Amount
	totalBalanceUSD string
}

func newMainPage(common *pageCommon, l *load.Load) *mainPage {

	mp := &mainPage{
		load:       l,
		pageCommon: common,
		autoSync:   true,
		pages:      loadPages(common, l),

		minimizeNavDrawerButton: common.theme.PlainIconButton(new(widget.Clickable), common.icons.navigationArrowBack),
		maximizeNavDrawerButton: common.theme.PlainIconButton(new(widget.Clickable), common.icons.navigationArrowForward),
	}

	// init shared page functions
	// todo: common methods will be removed when all pages have been moved to the page package
	common.changeFragment, l.ChangeFragment = mp.changeFragment, mp.changeFragment
	common.changePage, l.ChangePage = mp.changePage, mp.changePage
	common.setReturnPage, l.SetReturnPage = mp.setReturnPage, mp.setReturnPage
	common.returnPage, l.ReturnPage = &mp.previous, &mp.previous
	common.page, l.Page = &mp.current, &mp.current

	toggleSync := func() {
		if mp.multiWallet.IsConnectedToDecredNetwork() {
			mp.multiWallet.CancelSync()
		} else {
			mp.startSyncing()
		}
	}
	// todo: to be removed when all pages have been migrated
	common.toggleSync = toggleSync
	l.ToggleSync = toggleSync

	iconColor := common.theme.Color.Gray3
	mp.minimizeNavDrawerButton.Color, mp.maximizeNavDrawerButton.Color = iconColor, iconColor

	mp.initNavItems()

	mp.OnResume()

	return mp
}

func (mp *mainPage) initNavItems() {
	mp.appBarNavItems = []navHandler{
		{
			clickable: new(widget.Clickable),
			image:     mp.icons.sendIcon,
			page:      values.String(values.StrSend),
		},
		{
			clickable: new(widget.Clickable),
			image:     mp.icons.receiveIcon,
			page:      values.String(values.StrReceive),
		},
	}

	mp.drawerNavItems = []navHandler{
		{
			clickable:     new(widget.Clickable),
			image:         mp.icons.overviewIcon,
			imageInactive: mp.icons.overviewIconInactive,
			page:          values.String(values.StrOverview),
		},
		{
			clickable:     new(widget.Clickable),
			image:         mp.icons.transactionIcon,
			imageInactive: mp.icons.transactionIconInactive,
			page:          values.String(values.StrTransactions),
		},
		{
			clickable:     new(widget.Clickable),
			image:         mp.icons.walletIcon,
			imageInactive: mp.icons.walletIconInactive,
			page:          values.String(values.StrWallets),
		},
		{
			clickable:     new(widget.Clickable),
			image:         mp.icons.proposalIconActive,
			imageInactive: mp.icons.proposalIconInactive,
			page:          PageProposals,
		},
		{
			clickable:     new(widget.Clickable),
			image:         mp.icons.ticketIcon,
			imageInactive: mp.icons.ticketIconInactive,
			page:          values.String(values.StrTickets),
		},
		{
			clickable:     new(widget.Clickable),
			image:         mp.icons.moreIcon,
			imageInactive: mp.icons.moreIconInactive,
			page:          values.String(values.StrMore),
		},
	}
}

func (mp *mainPage) OnResume() {
	// register for notifications
	mp.multiWallet.AddAccountMixerNotificationListener(mp, PageMain)
	mp.multiWallet.Politeia.AddNotificationListener(mp, PageMain)
	mp.multiWallet.AddTxAndBlockNotificationListener(mp, PageMain)
	mp.multiWallet.AddSyncProgressListener(mp, PageMain)

	mp.updateBalance()

	if pg, ok := mp.pages[mp.current]; ok {
		pg.OnResume()
	} else {
		mp.changeFragment(page.NewOverviewPage(mp.load), page.OverviewPageID)
	}

	if mp.autoSync {
		mp.autoSync = false
		mp.startSyncing()
	}
}

func (mp *mainPage) updateBalance() {
	currencyExchangeValue := mp.wallet.ReadStringConfigValueForKey(dcrlibwallet.CurrencyConversionConfigKey)
	mp.usdExchangeSet = currencyExchangeValue == page.USDExchangeValue

	totalBalance, err := mp.calculateTotalWalletsBalance()
	if err == nil {
		mp.totalBalance = totalBalance

		if mp.usdExchangeSet && mp.dcrUsdtBittrex.LastTradeRate != "" {
			usdExchangeRate, err := strconv.ParseFloat(mp.dcrUsdtBittrex.LastTradeRate, 64)
			if err == nil {
				balanceInUSD := totalBalance.ToCoin() * usdExchangeRate
				mp.totalBalanceUSD = formatUSDBalance(mp.printer, balanceInUSD)
			}
		}

	}
}

func (mp *mainPage) calculateTotalWalletsBalance() (dcrutil.Amount, error) {
	totalBalance := int64(0)
	for _, wallet := range mp.sortedWalletList() {
		accountsResult, err := wallet.GetAccountsRaw()
		if err != nil {
			return 0, err
		}

		for _, account := range accountsResult.Acc {
			totalBalance += account.TotalBalance
		}
	}

	return dcrutil.Amount(totalBalance), nil
}

func (mp *mainPage) startSyncing() {
	for _, wal := range mp.sortedWalletList() {
		if !wal.HasDiscoveredAccounts && wal.IsLocked() {
			mp.unlockWalletForSyncing(wal)
			return
		}
	}

	err := mp.multiWallet.SpvSync()
	if err != nil {
		// show error dialog
		log.Info("Error starting sync:", err)
	}
}

func (mp *mainPage) unlockWalletForSyncing(wal *dcrlibwallet.Wallet) {
	newPasswordModal(mp.pageCommon).
		title(values.String(values.StrResumeAccountDiscoveryTitle)).
		hint(wal.Name+" Spending password").
		negativeButton(values.String(values.StrCancel), func() {}).
		positiveButton(values.String(values.StrUnlock), func(password string, pm *passwordModal) bool {
			go func() {
				err := mp.multiWallet.UnlockWallet(wal.ID, []byte(password))
				if err != nil {
					errText := err.Error()
					if err.Error() == "invalid_passphrase" {
						errText = "Invalid passphrase"
					}
					pm.setError(errText)
					pm.setLoading(false)
					return
				}
				pm.Dismiss()
				mp.startSyncing()
			}()

			return false
		}).Show()

}

func (mp *mainPage) Handle() {
	for mp.minimizeNavDrawerButton.Button.Clicked() {
		mp.isNavDrawerMinimized = true
	}

	for mp.maximizeNavDrawerButton.Button.Clicked() {
		mp.isNavDrawerMinimized = false
	}

	for i := range mp.appBarNavItems {
		for mp.appBarNavItems[i].clickable.Clicked() {
			mp.setReturnPage(mp.current)
			mp.changePage(mp.appBarNavItems[i].page)
		}
	}

	for i := range mp.drawerNavItems {
		for mp.drawerNavItems[i].clickable.Clicked() {
			if i == 0 {
				mp.changeFragment(page.NewOverviewPage(mp.load), page.OverviewPageID)
			} else if i == 1 {
				mp.changeFragment(page.NewTransactionsPage(mp.load), page.Transactions)
			} else if i == 2 {
				mp.changeFragment(page.NewWalletPage(mp.load), page.WalletPageID)
			} else if i == 3 {
				mp.changeFragment(ProposalsPage(mp.pageCommon), PageProposals)
			} else {
				mp.changePage(mp.drawerNavItems[i].page)
			}
		}
	}
}

func (mp *mainPage) OnClose() {
	if pg, ok := mp.pages[mp.current]; ok {
		pg.OnClose()
	}
	mp.multiWallet.RemoveAccountMixerNotificationListener(PageMain)
	mp.multiWallet.Politeia.RemoveNotificationListener(PageMain)
	mp.multiWallet.RemoveTxAndBlockNotificationListener(PageMain)
	mp.multiWallet.RemoveSyncProgressListener(PageMain)
}

func (mp *mainPage) changeFragment(page Page, id string) {
	mp.pages[id] = page
	mp.changePage(id)
}

func (mp *mainPage) changePage(page string) {
	if pg, ok := mp.pages[mp.current]; ok {
		pg.OnClose()
	}

	if pg, ok := mp.pages[page]; ok {
		pg.OnResume()
		mp.current = page
	}
}

func (mp *mainPage) setReturnPage(from string) {
	mp.previous = from
}

func (mp *mainPage) Layout(gtx layout.Context) layout.Dimensions {
	mp.handler() // pageCommon
	mp.pages[mp.current].Handle()

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			// fill the entire window with a color if a user has no wallet created
			if mp.current == page.CreateRestorePageID {
				return decredmaterial.Fill(gtx, mp.theme.Color.Surface)
			}

			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(mp.layoutTopBar),
				layout.Rigid(func(gtx C) D {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							card := mp.theme.Card()
							card.Radius = decredmaterial.CornerRadius{}
							return card.Layout(gtx, mp.layoutNavDrawer)
						}),
						layout.Rigid(mp.pages[mp.current].Layout),
					)
				}),
			)
		}),
		layout.Stacked(func(gtx C) D {
			// stack the page content on the entire window if a user has no wallet
			if mp.current == page.CreateRestorePageID {
				return mp.pages[mp.current].Layout(gtx)
			}
			return layout.Dimensions{}
		}),
		layout.Stacked(func(gtx C) D {
			// global toasts. Stack toast on all pages and contents
			if *mp.toast == nil {
				return layout.Dimensions{}
			}
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return layout.Center.Layout(gtx, func(gtx C) D {
				return layout.Inset{Top: values.MarginPadding65}.Layout(gtx, func(gtx C) D {
					return displayToast(mp.theme, gtx, *mp.toast)
				})
			})
		}),
	)
}

func (mp *mainPage) layoutTopBar(gtx layout.Context) layout.Dimensions {
	card := mp.theme.Card()
	card.Radius = decredmaterial.CornerRadius{}
	return card.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.W.Layout(gtx, func(gtx C) D {
							h := values.MarginPadding24
							v := values.MarginPadding16
							// Balance container
							return Container{padding: layout.Inset{Right: h, Left: h, Top: v, Bottom: v}}.Layout(gtx,
								func(gtx C) D {
									return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											img := mp.icons.logo
											img.Scale = 1.0
											return layout.Inset{Right: values.MarginPadding16}.Layout(gtx,
												func(gtx C) D {
													return img.Layout(gtx)
												})
										}),
										layout.Rigid(func(gtx C) D {
											return layout.Center.Layout(gtx, func(gtx C) D {
												return mp.layoutBalance(gtx, mp.totalBalance.String(), true)
											})
										}),
										layout.Rigid(func(gtx C) D {
											return mp.layoutUSDBalance(gtx)
										}),
									)
								})
						})
					}),
					layout.Rigid(func(gtx C) D {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return layout.E.Layout(gtx, func(gtx C) D {
							return layout.Inset{Right: values.MarginPadding8}.Layout(gtx, func(gtx C) D {
								list := layout.List{Axis: layout.Horizontal}
								return list.Layout(gtx, len(mp.appBarNavItems), func(gtx C, i int) D {
									// header buttons container
									return Container{layout.UniformInset(values.MarginPadding16)}.Layout(gtx, func(gtx C) D {
										return decredmaterial.Clickable(gtx, mp.appBarNavItems[i].clickable, func(gtx C) D {
											return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
												layout.Rigid(func(gtx C) D {
													return layout.Inset{Right: values.MarginPadding8}.Layout(gtx,
														func(gtx C) D {
															return layout.Center.Layout(gtx, func(gtx C) D {
																img := mp.appBarNavItems[i].image
																img.Scale = 1.0
																return mp.appBarNavItems[i].image.Layout(gtx)
															})
														})
												}),
												layout.Rigid(func(gtx C) D {
													return layout.Inset{
														Left: values.MarginPadding0,
													}.Layout(gtx, func(gtx C) D {
														return layout.Center.Layout(gtx, func(gtx C) D {
															return mp.theme.Body1(mp.appBarNavItems[i].page).Layout(gtx)
														})
													})
												}),
											)
										})
									})
								})
							})
						})
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return mp.theme.Separator().Layout(gtx)
			}),
		)
	})
}

func (mp *mainPage) layoutUSDBalance(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if mp.usdExchangeSet && mp.dcrUsdtBittrex.LastTradeRate != "" {
				inset := layout.Inset{
					Top:  values.MarginPadding3,
					Left: values.MarginPadding8,
				}
				border := widget.Border{Color: mp.theme.Color.Gray, CornerRadius: unit.Dp(8), Width: unit.Dp(0.5)}
				return inset.Layout(gtx, func(gtx C) D {
					padding := layout.Inset{
						Top:    values.MarginPadding3,
						Bottom: values.MarginPadding3,
						Left:   values.MarginPadding6,
						Right:  values.MarginPadding6,
					}
					return border.Layout(gtx, func(gtx C) D {
						return padding.Layout(gtx, func(gtx C) D {
							return mp.theme.Body2(mp.totalBalanceUSD).Layout(gtx)
						})
					})
				})
			}
			return D{}
		}),
	)
}

func (mp *mainPage) layoutNavDrawer(gtx layout.Context) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx C) D {
			list := layout.List{Axis: layout.Vertical}
			return list.Layout(gtx, len(mp.drawerNavItems), func(gtx C, i int) D {
				background := mp.theme.Color.Surface
				if mp.drawerNavItems[i].page == mp.current {
					background = mp.theme.Color.ActiveGray
				}
				txt := mp.theme.Label(values.TextSize16, mp.drawerNavItems[i].page)
				return decredmaterial.Clickable(gtx, mp.drawerNavItems[i].clickable, func(gtx C) D {
					card := mp.theme.Card()
					card.Color = background
					card.Radius = decredmaterial.CornerRadius{NE: 0, NW: 0, SE: 0, SW: 0}
					return card.Layout(gtx, func(gtx C) D {
						return Container{
							layout.Inset{
								Top:    values.MarginPadding16,
								Right:  values.MarginPadding24,
								Bottom: values.MarginPadding16,
								Left:   values.MarginPadding24,
							},
						}.Layout(gtx, func(gtx C) D {
							axis := layout.Horizontal
							leftInset := values.MarginPadding15
							width := navDrawerWidth
							if mp.isNavDrawerMinimized {
								axis = layout.Vertical
								txt.TextSize = values.TextSize10
								leftInset = values.MarginPadding0
								width = navDrawerMinimizedWidth
							}

							gtx.Constraints.Min.X = gtx.Px(width)
							return layout.Flex{Axis: axis}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									img := mp.drawerNavItems[i].imageInactive
									if mp.drawerNavItems[i].page == mp.current {
										img = mp.drawerNavItems[i].image
									}
									return layout.Center.Layout(gtx, func(gtx C) D {
										img.Scale = 1.0
										return img.Layout(gtx)
									})
								}),
								layout.Rigid(func(gtx C) D {
									return layout.Inset{
										Left: leftInset,
										Top:  values.MarginPadding4,
									}.Layout(gtx, func(gtx C) D {
										textColor := mp.theme.Color.Gray4
										if mp.drawerNavItems[i].page == mp.current {
											textColor = mp.theme.Color.DeepBlue
										}
										txt.Color = textColor
										return layout.Center.Layout(gtx, txt.Layout)
									})
								}),
							)
						})
					})
				})
			})
		}),
		layout.Expanded(func(gtx C) D {
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			return layout.SE.Layout(gtx, func(gtx C) D {
				btn := mp.minimizeNavDrawerButton
				if mp.isNavDrawerMinimized {
					btn = mp.maximizeNavDrawerButton
				}
				return btn.Layout(gtx)
			})
		}),
	)
}
