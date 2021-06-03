package ui

import (
	"errors"
	"strconv"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/decred/dcrd/dcrutil/v3"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const PageMain = "MainPage"

type navHandler struct {
	clickable     *widget.Clickable
	image         *widget.Image
	imageInactive *widget.Image
	pageID        string
	label         string
}

type mainPage struct {
	pageCommon
	multiWallet *dcrlibwallet.MultiWallet

	// TODO: protect with mutex
	pageBackStack []Page
	currentPage   Page

	modals []Modal

	appBarNavItems []navHandler
	drawerNavItems []navHandler

	isNavDrawerMinimized    bool
	minimizeNavDrawerButton decredmaterial.IconButton
	maximizeNavDrawerButton decredmaterial.IconButton

	// page state variables
	usdExchangeSet  bool
	totalBalance    dcrutil.Amount
	totalBalanceUSD string
}

func MainPage(c pageCommon) Page {
	mp := &mainPage{
		pageCommon:  c,
		multiWallet: c.multiWallet,
	}

	mp.pageCommon.popPage = mp.popPage
	mp.pageCommon.changePage = mp.changePage
	mp.pageCommon.popToPage = mp.popToPage

	mp.pageCommon.showModal = mp.showModal
	mp.pageCommon.dismissModal = mp.dismissModal

	mp.minimizeNavDrawerButton = c.theme.PlainIconButton(new(widget.Clickable), c.icons.navigationArrowBack)
	mp.minimizeNavDrawerButton.Color = c.theme.Color.Gray3

	mp.maximizeNavDrawerButton = c.theme.PlainIconButton(new(widget.Clickable), c.icons.navigationArrowForward)
	mp.maximizeNavDrawerButton.Color = c.theme.Color.Gray3

	currencyExchangeValue := mp.multiWallet.ReadStringConfigValueForKey(dcrlibwallet.CurrencyConversionConfigKey)
	mp.usdExchangeSet = currencyExchangeValue == USDExchangeValue

	mp.setupNavHandlers()
	mp.updateBalance()

	mp.currentPage = OverviewPage(mp.pageCommon)

	return mp
}

func (mp *mainPage) updateBalance() {
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

func (mp *mainPage) setupNavHandlers() {
	appBarNavItems := []navHandler{
		{
			clickable: new(widget.Clickable),
			image:     mp.icons.sendIcon,
			pageID:    PageSend,
			label:     values.String(values.StrSend),
		},
		{
			clickable: new(widget.Clickable),
			image:     mp.icons.receiveIcon,
			pageID:    PageReceive,
			label:     values.String(values.StrReceive),
		},
	}

	drawerNavItems := []navHandler{
		{
			clickable:     new(widget.Clickable),
			image:         mp.icons.overviewIcon,
			imageInactive: mp.icons.overviewIconInactive,
			pageID:        PageOverview,
			label:         values.String(values.StrOverview),
		},
		{
			clickable:     new(widget.Clickable),
			image:         mp.icons.transactionIcon,
			imageInactive: mp.icons.transactionIconInactive,
			pageID:        PageTransactions,
			label:         values.String(values.StrTransactions),
		},
		{
			clickable:     new(widget.Clickable),
			image:         mp.icons.walletIcon,
			imageInactive: mp.icons.walletIconInactive,
			pageID:        PageWallet,
			label:         values.String(values.StrWallets),
		},
		{
			clickable:     new(widget.Clickable),
			image:         mp.icons.proposalIconActive,
			imageInactive: mp.icons.proposalIconInactive,
			pageID:        PageProposals,
			label:         PageProposals,
		},
		{
			clickable:     new(widget.Clickable),
			image:         mp.icons.ticketIcon,
			imageInactive: mp.icons.ticketIconInactive,
			pageID:        PageTickets,
			label:         values.String(values.StrTickets),
		},
		{
			clickable:     new(widget.Clickable),
			image:         mp.icons.moreIcon,
			imageInactive: mp.icons.moreIconInactive,
			pageID:        PageMore,
			label:         values.String(values.StrMore),
		},
	}

	mp.appBarNavItems = appBarNavItems
	mp.drawerNavItems = drawerNavItems
}

func (mp *mainPage) onClose() {

}

func (mp *mainPage) pageID() string {
	return PageMain
}

func (mp *mainPage) handle() {
	for _, appBarItem := range mp.appBarNavItems {
		for appBarItem.clickable.Clicked() {
			switch appBarItem.pageID {
			case PageSend:
				mp.changePage(SendPage(mp.pageCommon))
			case PageReceive:
				mp.changePage(ReceivePage(mp.pageCommon))
			}
		}
	}

	for _, navBarItem := range mp.drawerNavItems {
		for navBarItem.clickable.Clicked() {
			// clear backstack since no backbutton to move between nav bars
			switch navBarItem.pageID {
			case PageOverview:
				mp.changePage(OverviewPage(mp.pageCommon))
			case PageTransactions:
				mp.changePage(TransactionsPage(mp.pageCommon))
			case PageWallet:
				mp.changePage(WalletPage(mp.pageCommon))
			case PageProposals:
				mp.changePage(ProposalsPage(mp.pageCommon))
			case PageTickets:
				mp.changePage(TicketPage(mp.pageCommon))
			case PageMore:
				mp.changePage(MorePage(mp.pageCommon))
			}
		}
	}

	for mp.minimizeNavDrawerButton.Button.Clicked() {
		mp.isNavDrawerMinimized = true
	}

	for mp.maximizeNavDrawerButton.Button.Clicked() {
		mp.isNavDrawerMinimized = false
	}

	mp.currentPage.handle()

	for _, modal := range mp.modals {
		modal.handle()
	}
}

func (mp *mainPage) calculateTotalWalletsBalance() (dcrutil.Amount, error) {
	totalBalance := int64(0)
	for _, wallet := range mp.multiWallet.AllWallets() {
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

// Page handlers

func (mp *mainPage) changePage(page Page) {
	if mp.currentPage != nil {
		mp.currentPage.onClose()
		mp.pageBackStack = append(mp.pageBackStack, mp.currentPage)
	}
	mp.currentPage = page
}

// popPage goes back to the previous page
func (mp *mainPage) popPage() {
	if len(mp.pageBackStack) > 0 {
		// get and remove last page
		previousPage := mp.pageBackStack[len(mp.pageBackStack)-1]
		mp.pageBackStack = mp.pageBackStack[:len(mp.pageBackStack)-1]

		mp.currentPage.onClose()
		mp.currentPage = previousPage
	}
}

func (mp *mainPage) popToPage(pageID string) error {
	// TODO
	return errors.New("not implemented")
}

// Modal handlers

func (mp *mainPage) showModal(modal Modal) {
	modal.OnResume() // setup display data
	mp.modals = append(mp.modals, modal)
}

func (mp *mainPage) dismissModal(modal Modal) {
	for i, m := range mp.modals {
		if m.modalID() == modal.modalID() {
			modal.OnDismiss() // do garbage collection in modal
			mp.modals = append(mp.modals[:i], mp.modals[i+1:]...)
		}
	}
}

// Layout

func (mp *mainPage) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			// fill the entire window with a color if a user has no wallet created
			if mp.currentPage != nil && mp.currentPage.pageID() == PageCreateRestore { //TODO
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
						layout.Rigid(func(gtx C) D {
							if mp.currentPage != nil {
								mp.currentPage.Layout(gtx) // todo
							}

							return layout.Dimensions{}
						}),
					)
				}),
			)
		}),
		layout.Stacked(func(gtx C) D {
			// stack the page content on the entire window if a user has no wallet
			if "page.page.pageID()" == PageCreateRestore { //TODO
				return mp.currentPage.Layout(gtx) // todo
			}
			return layout.Dimensions{}
		}),
		layout.Stacked(func(gtx C) D {
			// global modal. Stack modal on all pages and contents
		outer:
			for {
				select {
				case load := <-mp.modalReceiver:
					mp.modalLoad.template = load.template
					mp.modalLoad.title = load.title
					mp.modalLoad.confirm = load.confirm
					mp.modalLoad.confirmText = load.confirmText
					mp.modalLoad.cancel = load.cancel
					mp.modalLoad.cancelText = load.cancelText
					mp.modalLoad.isReset = false
				default:
					break outer
				}
			}

			if mp.modalLoad.cancel != nil {
				return mp.modal.Layout(gtx, mp.modalTemplate.Layout(mp.theme, mp.modalLoad),
					900)
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
					return displayToast(mp.theme, gtx, *mp.toast) // TODO
				})
			})
		}),
		layout.Stacked(func(gtx C) D {
			if len(mp.modals) > 0 {
				return mp.modals[0].Layout(gtx)
			}
			return layout.Dimensions{}
		}),
	)
}

func (mp *mainPage) layoutTopBar(gtx layout.Context) layout.Dimensions {
	common := mp.pageCommon
	card := common.theme.Card()
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
											img := common.icons.logo
											img.Scale = 1.0
											return layout.Inset{Right: values.MarginPadding16}.Layout(gtx,
												func(gtx C) D {
													return img.Layout(gtx)
												})
										}),
										layout.Rigid(func(gtx C) D {
											return layout.Center.Layout(gtx, func(gtx C) D {
												return mp.layoutBalance(gtx, true)
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
								return list.Layout(gtx, len(mp.appBarNavItems), func(gtx C, i int) D { //TODO move itemps to main page
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
															return common.theme.Body1(mp.appBarNavItems[i].label).Layout(gtx)
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
				return common.theme.Separator().Layout(gtx)
			}),
		)
	})
}

func (mp *mainPage) layoutBalance(gtx layout.Context, isSwitchColor bool) layout.Dimensions {
	common := mp.pageCommon
	// todo: make "DCR" symbols small when there are no decimals in the balance
	mainText, subText := breakBalance(common.printer, mp.totalBalance.String())
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			label := common.theme.Label(values.TextSize20, mainText)
			if isSwitchColor {
				label.Color = common.theme.Color.DeepBlue
			}
			return label.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			label := common.theme.Label(values.TextSize14, subText)
			if isSwitchColor {
				label.Color = common.theme.Color.DeepBlue
			}
			return label.Layout(gtx)
		}),
	)
}

func (mp *mainPage) layoutUSDBalance(gtx layout.Context) layout.Dimensions {
	common := mp.pageCommon
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if mp.usdExchangeSet && common.dcrUsdtBittrex.LastTradeRate != "" {
				inset := layout.Inset{
					Top:  values.MarginPadding3,
					Left: values.MarginPadding8,
				}
				border := widget.Border{Color: common.theme.Color.Gray, CornerRadius: unit.Dp(8), Width: unit.Dp(0.5)}
				return inset.Layout(gtx, func(gtx C) D {
					padding := layout.Inset{
						Top:    values.MarginPadding3,
						Bottom: values.MarginPadding3,
						Left:   values.MarginPadding6,
						Right:  values.MarginPadding6,
					}
					return border.Layout(gtx, func(gtx C) D {
						return padding.Layout(gtx, func(gtx C) D {
							return common.theme.Body2(mp.totalBalanceUSD).Layout(gtx)
						})
					})
				})
			}
			return D{}
		}),
	)
}

func (mp *mainPage) layoutNavDrawer(gtx layout.Context) layout.Dimensions {
	common := mp.pageCommon
	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx C) D {
			list := layout.List{Axis: layout.Vertical}
			// todo
			return list.Layout(gtx, len(mp.drawerNavItems), func(gtx C, i int) D {
				background := common.theme.Color.Surface
				if mp.drawerNavItems[i].pageID == "page.page.pageID()" { //TODO
					background = common.theme.Color.ActiveGray
				}
				txt := common.theme.Label(values.TextSize16, mp.drawerNavItems[i].pageID)
				return decredmaterial.Clickable(gtx, mp.drawerNavItems[i].clickable, func(gtx C) D {
					card := common.theme.Card()
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
									if mp.drawerNavItems[i].pageID == "page.page.pageID()" { // TODO
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
										textColor := common.theme.Color.Gray4
										if mp.drawerNavItems[i].pageID == "page.page.pageID()" { // TODO
											textColor = common.theme.Color.DeepBlue
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
