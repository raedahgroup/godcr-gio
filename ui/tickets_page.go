package ui

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"

	"gioui.org/unit"

	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/decred/dcrd/dcrutil"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"

	"gioui.org/layout"
	"github.com/planetdecred/godcr/ui/decredmaterial"
)

const PageTickets = "Tickets"

type ticketPage struct {
	th     *decredmaterial.Theme
	wal    *wallet.Wallet
	vspd   *dcrlibwallet.VSPD
	common pageCommon

	ticketPageContainer layout.List
	ticketsLive         layout.List
	ticketsActivity     layout.List

	purchaseTicket        decredmaterial.Button
	cancelPurchase        decredmaterial.Button
	reviewPurchase        decredmaterial.Button
	cancelConfirmPurchase decredmaterial.Button
	submitPurchase        decredmaterial.Button
	tickets               *wallet.Tickets
	ticketPrice           string
	totalCost             string
	remainingBalance      string
	ticketAmount          decredmaterial.Editor
	showPurchaseOptions   bool
	showPurchaseConfirm   bool

	selector              *accountSelector
	purchaseOptions       *decredmaterial.Modal
	autoPurchaseEnabled   *widget.Bool
	toTickets             decredmaterial.TextAndIconButton
	toTicketsActivity     decredmaterial.TextAndIconButton
	purchaseErrChan       chan error
	walletSelectedID      int
	accountSelectedNumber int32

	loadedVSP        []wallet.VSPInfo
	vspHosts         layout.List
	rememberVSP      decredmaterial.CheckBoxStyle
	showVSPHosts     bool
	showVSP          *widget.Clickable
	selectVSP        []*gesture.Click
	selectedVSP      wallet.VSPInfo
	inputVSP         decredmaterial.Editor
	spendingPassword decredmaterial.Editor
	addVSP           decredmaterial.Button
	vspErrChan       chan error

	isPurchaseLoading bool
}

type vspHosts struct {
	Remember string
	List     []string
}

func TicketPage(c pageCommon) Page {
	pg := &ticketPage{
		th:     c.theme,
		wal:    c.wallet,
		common: c,

		tickets:               new(wallet.Tickets),
		ticketsLive:           layout.List{Axis: layout.Horizontal},
		ticketsActivity:       layout.List{Axis: layout.Vertical},
		ticketPageContainer:   layout.List{Axis: layout.Vertical},
		purchaseTicket:        c.theme.Button(new(widget.Clickable), "Purchase"),
		cancelPurchase:        c.theme.Button(new(widget.Clickable), "Cancel"),
		cancelConfirmPurchase: c.theme.Button(new(widget.Clickable), "Cancel"),
		submitPurchase:        c.theme.Button(new(widget.Clickable), "Purchase ticket"),
		reviewPurchase:        c.theme.Button(new(widget.Clickable), "Review purchase"),
		autoPurchaseEnabled:   new(widget.Bool),
		toTickets:             c.theme.TextAndIconButton(new(widget.Clickable), "See All", c.icons.navigationArrowForward),
		toTicketsActivity:     c.theme.TextAndIconButton(new(widget.Clickable), "See All", c.icons.navigationArrowForward),
		purchaseOptions:       c.theme.Modal(),
		ticketAmount:          c.theme.Editor(new(widget.Editor), ""),
		purchaseErrChan:       make(chan error),
		vspHosts:              layout.List{Axis: layout.Vertical},
		showVSP:               new(widget.Clickable),
		rememberVSP:           c.theme.CheckBox(new(widget.Bool), "Remember VSP"),
		inputVSP:              c.theme.Editor(new(widget.Editor), "Add a new VSP..."),
		addVSP:                c.theme.Button(new(widget.Clickable), "Save"),
		spendingPassword:      c.theme.EditorPassword(new(widget.Editor), "Spending password"),
		vspErrChan:            make(chan error),
	}
	pg.ticketAmount.Editor.SetText("1")

	pg.purchaseTicket.TextSize = values.TextSize12
	pg.purchaseTicket.Background = c.theme.Color.Primary

	pg.cancelPurchase.Background = color.NRGBA{}
	pg.cancelPurchase.Color = c.theme.Color.Primary
	pg.cancelConfirmPurchase.Background = color.NRGBA{}
	pg.cancelConfirmPurchase.Color = c.theme.Color.Primary

	pg.toTickets.Color = c.theme.Color.Primary
	pg.toTickets.BackgroundColor = c.theme.Color.Surface

	pg.toTicketsActivity.Color = c.theme.Color.Primary
	pg.toTicketsActivity.BackgroundColor = c.theme.Color.Surface

	pg.selector = newAccountSelector(c).
		title("Purchase account").
		accountSelected(func(selectedAccount *dcrlibwallet.Account) {
			pg.walletSelectedID = selectedAccount.WalletID
			pg.accountSelectedNumber = selectedAccount.Number
			if pg.selectedVSP.Host != "" {
				pg.createNewVSPD()
			}
		}).
		accountValidator(func(account *dcrlibwallet.Account) bool {
			return account.Number == MaxInt32
		})

	go func() {
		tickets, err := pg.common.wallet.GetAllTickets() //TODO
		if err == nil {
			pg.tickets = tickets
		}
		pg.getAllVSP()
	}()

	return pg
}

func (pg *ticketPage) pageID() string {
	return PageTickets
}

func (pg *ticketPage) getAllVSP() {
	var valueOut vspHosts

	pg.common.multiWallet.ReadUserConfigValue(dcrlibwallet.VSPHostConfigKey, &valueOut)
	var loadedVSP []wallet.VSPInfo

	for _, host := range valueOut.List {
		v, err := wallet.GetVSPInfo(host)
		if err == nil {
			loadedVSP = append(loadedVSP, wallet.VSPInfo{
				Host: host,
				Info: v,
			})
		}
	}

	l, _ := wallet.GetInitVSPInfo("https://api.decred.org/?c=vsp")
	for h, v := range l {
		if strings.Contains("testnet", v.Network) { //todo
			loadedVSP = append(loadedVSP, wallet.VSPInfo{
				Host: fmt.Sprintf("https://%s", h),
				Info: v,
			})
		}
	}

	pg.loadedVSP = loadedVSP
}

func (pg *ticketPage) Layout(gtx layout.Context) layout.Dimensions {
	c := pg.common
	dims := c.UniformPadding(gtx, func(gtx layout.Context) layout.Dimensions {
		sections := []func(gtx C) D{
			func(ctx layout.Context) layout.Dimensions {
				return pg.ticketPriceSection(gtx, c)
			},
			func(ctx layout.Context) layout.Dimensions {
				return pg.ticketsLiveSection(gtx, c)
			},
			func(ctx layout.Context) layout.Dimensions {
				return pg.ticketsActivitySection(gtx, c)
			},
			func(ctx layout.Context) layout.Dimensions {
				return pg.stackingRecordSection(gtx, c)
			},
		}

		return pg.ticketPageContainer.Layout(gtx, len(sections), func(gtx C, i int) D {
			return sections[i](gtx)
		})
	})

	if pg.showPurchaseConfirm {
		return pg.confirmPurchaseModal(gtx)
	}

	if pg.showVSPHosts {
		return pg.vspHostModalLayout(gtx, c)
	}

	if pg.showPurchaseOptions {
		return pg.purchaseModal(gtx, c)
	}

	return dims
}

func (pg *ticketPage) pageSections(gtx layout.Context, body layout.Widget) layout.Dimensions {
	return layout.Inset{
		Bottom: values.MarginPadding8,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return pg.th.Card().Layout(gtx, func(gtx C) D {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return layout.UniformInset(values.MarginPadding16).Layout(gtx, body)
		})
	})
}

func (pg *ticketPage) titleRow(gtx layout.Context, leftWidget, rightWidget func(C) D) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(leftWidget),
		layout.Rigid(rightWidget),
	)
}

func (pg *ticketPage) ticketPriceSection(gtx layout.Context, c pageCommon) layout.Dimensions {
	return pg.pageSections(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Bottom: values.MarginPadding11,
				}.Layout(gtx, func(gtx C) D {
					tit := c.theme.Label(values.TextSize14, "Ticket Price")
					tit.Color = c.theme.Color.Gray2
					return pg.titleRow(gtx, tit.Layout, material.Switch(pg.th.Base, pg.autoPurchaseEnabled).Layout)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Bottom: values.MarginPadding8,
				}.Layout(gtx, func(gtx C) D {
					ic := c.icons.ticketPurchasedIcon
					ic.Scale = 1.2
					return layout.Center.Layout(gtx, ic.Layout)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Bottom: values.MarginPadding16,
				}.Layout(gtx, func(gtx C) D {
					return layout.Center.Layout(gtx, func(gtx C) D {
						mainText, subText := breakBalance(c.printer, pg.ticketPrice)
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								label := pg.th.Label(values.TextSize28, mainText)
								label.Color = pg.th.Color.DeepBlue
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx C) D {
								label := pg.th.Label(values.TextSize16, subText)
								label.Color = pg.th.Color.DeepBlue
								return label.Layout(gtx)
							}),
						)
					})
				})
			}),
			layout.Rigid(pg.purchaseTicket.Layout),
		)
	})
}

func (pg *ticketPage) ticketsLiveSection(gtx layout.Context, c pageCommon) layout.Dimensions {
	return pg.pageSections(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Bottom: values.MarginPadding14}.Layout(gtx, func(gtx C) D {
					tit := c.theme.Label(values.TextSize14, "Live Tickets")
					tit.Color = c.theme.Color.Gray2
					return pg.titleRow(gtx, tit.Layout, func(gtx C) D {
						ticketLiveCounter := (*pg.tickets).LiveCounter
						var elements []layout.FlexChild
						for i := 0; i < len(ticketLiveCounter); i++ {
							item := ticketLiveCounter[i]
							elements = append(elements, layout.Rigid(func(gtx C) D {
								return layout.Inset{Right: values.MarginPadding14}.Layout(gtx, func(gtx C) D {
									return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											st := ticketStatusIcon(&c, item.Status)
											if st == nil {
												return layout.Dimensions{}
											}
											st.icon.Scale = .5
											return st.icon.Layout(gtx)
										}),
										layout.Rigid(func(gtx C) D {
											return layout.Inset{Left: values.MarginPadding4}.Layout(gtx, func(gtx C) D {
												label := pg.th.Label(values.TextSize14, fmt.Sprintf("%d", item.Count))
												label.Color = pg.th.Color.DeepBlue
												return label.Layout(gtx)
											})
										}),
									)
								})
							}))
						}
						elements = append(elements, layout.Rigid(pg.toTickets.Layout))
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx, elements...)
					})
				})
			}),
			layout.Rigid(func(gtx C) D {
				tickets := (*pg.tickets).LiveRecent
				return pg.ticketsLive.Layout(gtx, len(tickets), func(gtx C, index int) D {
					return layout.Inset{Right: values.MarginPadding8}.Layout(gtx, func(gtx C) D {
						return ticketCard(gtx, c, &tickets[index])
					})
				})
			}),
		)
	})
}

func (pg *ticketPage) ticketsActivitySection(gtx layout.Context, c pageCommon) layout.Dimensions {
	tickets := (*pg.tickets).RecentActivity
	if len(tickets) == 0 {
		return layout.Dimensions{}
	}

	return pg.pageSections(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Bottom: values.MarginPadding14,
				}.Layout(gtx, func(gtx C) D {
					tit := c.theme.Label(values.TextSize14, "Recent Activity")
					tit.Color = c.theme.Color.Gray2
					return pg.titleRow(gtx, tit.Layout, pg.toTicketsActivity.Layout)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return pg.ticketsActivity.Layout(gtx, len(tickets), func(gtx C, index int) D {
					return pg.ticketActivityItemnInfo(gtx, c, tickets[index], index)
				})
			}),
		)
	})
}

func (pg *ticketPage) ticketActivityItemnInfo(gtx layout.Context, c pageCommon, t wallet.Ticket, index int) layout.Dimensions {
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Right: values.MarginPadding16}.Layout(gtx, func(gtx C) D {
				st := ticketStatusIcon(&c, t.Info.Status)
				if st == nil {
					return layout.Dimensions{}
				}
				st.icon.Scale = 0.6
				return st.icon.Layout(gtx)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if index == 0 {
						return layout.Dimensions{}
					}
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					separator := pg.th.Separator()
					separator.Width = gtx.Constraints.Max.X
					return layout.E.Layout(gtx, separator.Layout)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top:    values.MarginPadding8,
						Bottom: values.MarginPadding8,
					}.Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								labelStatus := pg.th.Label(values.TextSize18, strings.Title(strings.ToLower(t.Info.Status)))
								labelStatus.Color = pg.th.Color.DeepBlue

								labelDaysBehind := pg.th.Label(values.TextSize14, t.DaysBehind)
								labelDaysBehind.Color = pg.th.Color.DeepBlue

								return endToEndRow(gtx,
									labelStatus.Layout,
									labelDaysBehind.Layout)
							}),
							layout.Rigid(func(gtx C) D {
								return layout.Flex{
									Alignment: layout.Middle,
								}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										txt := pg.th.Label(values.TextSize14, t.WalletName)
										txt.Color = pg.th.Color.Gray2
										return txt.Layout(gtx)
									}),
									layout.Rigid(func(gtx C) D {
										return layout.Inset{
											Left:  values.MarginPadding4,
											Right: values.MarginPadding4,
										}.Layout(gtx, func(gtx C) D {
											ic := c.icons.imageBrightness1
											ic.Color = pg.th.Color.Gray2
											return c.icons.imageBrightness1.Layout(gtx, values.MarginPadding5)
										})
									}),
									layout.Rigid(func(gtx C) D {
										return layout.Inset{
											Right: values.MarginPadding4,
										}.Layout(gtx, func(gtx C) D {
											ic := c.icons.ticketIconInactive
											ic.Scale = 0.5
											return ic.Layout(gtx)
										})
									}),
									layout.Rigid(func(gtx C) D {
										txt := pg.th.Label(values.TextSize14, t.Amount)
										txt.Color = pg.th.Color.Gray2
										return txt.Layout(gtx)
									}),
								)
							}),
						)
					})
				}),
			)
		}),
	)
}

func (pg *ticketPage) stackingRecordSection(gtx layout.Context, c pageCommon) layout.Dimensions {
	return pg.pageSections(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Bottom: values.MarginPadding14,
				}.Layout(gtx, func(gtx C) D {
					tit := c.theme.Label(values.TextSize14, "Staking Record")
					tit.Color = c.theme.Color.Gray2
					return pg.titleRow(gtx, tit.Layout, func(gtx C) D { return layout.Dimensions{} })
				})
			}),
			layout.Rigid(func(gtx C) D {
				stackingRecords := (*pg.tickets).StackingRecordCounter
				return decredmaterial.GridWrap{
					Axis:      layout.Horizontal,
					Alignment: layout.End,
				}.Layout(gtx, len(stackingRecords), func(gtx layout.Context, i int) layout.Dimensions {
					item := stackingRecords[i]
					width := unit.Value{U: unit.UnitDp, V: 118}
					gtx.Constraints.Min.X = gtx.Px(width)

					return layout.Inset{Bottom: values.MarginPadding16}.Layout(gtx, func(gtx C) D {
						return layout.Flex{}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								st := ticketStatusIcon(&c, item.Status)
								if st == nil {
									return layout.Dimensions{}
								}
								st.icon.Scale = 0.6
								return st.icon.Layout(gtx)
							}),
							layout.Rigid(func(gtx C) D {
								return layout.Inset{Left: values.MarginPadding4}.Layout(gtx, func(gtx C) D {
									return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											label := pg.th.Label(values.TextSize16, fmt.Sprintf("%d", item.Count))
											label.Color = pg.th.Color.DeepBlue
											return label.Layout(gtx)
										}),
										layout.Rigid(func(gtx C) D {
											return layout.Inset{Right: values.MarginPadding40}.Layout(gtx, func(gtx C) D {
												txt := pg.th.Label(values.TextSize12, strings.Title(strings.ToLower(item.Status)))
												txt.Color = pg.th.Color.Gray2
												return txt.Layout(gtx)
											})
										}),
									)
								})
							}),
						)
					})
				})
			}),
			layout.Rigid(func(gtx C) D {
				wrapper := pg.th.Card()
				wrapper.Color = pg.th.Color.Success2
				return wrapper.Layout(gtx, func(gtx C) D {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return layout.Center.Layout(gtx, func(gtx C) D {
						return layout.Inset{
							Top:    values.MarginPadding16,
							Bottom: values.MarginPadding16,
						}.Layout(gtx, func(gtx C) D {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return layout.Inset{Bottom: values.MarginPadding4}.Layout(gtx, func(gtx C) D {
										txt := pg.th.Label(values.TextSize14, "Rewards Earned")
										txt.Color = pg.th.Color.Success
										return txt.Layout(gtx)
									})
								}),
								layout.Rigid(func(gtx C) D {
									return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											ic := c.icons.stakeyIcon
											ic.Scale = 1.0
											return ic.Layout(gtx)
										}),
										layout.Rigid(func(gtx C) D {
											return c.layoutBalance(gtx, "16.5112316", false)
										}),
									)
								}),
							)
						})
					})
				})
			}),
		)
	})
}

func (pg *ticketPage) purchaseModal(gtx layout.Context, c pageCommon) layout.Dimensions {
	return pg.purchaseOptions.Layout(gtx, []layout.Widget{
		func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Center.Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								ic := c.icons.ticketPurchasedIcon
								ic.Scale = 1.2
								return ic.Layout(gtx)
							}),
							layout.Rigid(func(gtx C) D {
								return layout.Inset{Top: values.MarginPadding8}.Layout(gtx, func(gtx C) D {
									return c.layoutBalance(gtx, pg.ticketPrice, true)
								})
							}),
						)
					})
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Flex{}.Layout(gtx,
						layout.Flexed(.5, func(gtx C) D {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									tit := pg.th.Label(values.TextSize14, "Total")
									tit.Color = pg.th.Color.Gray2
									return tit.Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									return pg.th.Label(values.TextSize16, pg.ticketPrice).Layout(gtx)
								}),
							)
						}),
						layout.Flexed(.5, pg.ticketAmount.Layout),
					)
				}),
			)
		},
		func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return pg.selector.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: values.MarginPadding16}.Layout(gtx, func(gtx C) D {
						return pg.vspHostSelectorLayout(gtx, c)
					})
				}),
			)
		},
		func(gtx C) D {
			return pg.rememberVSP.Layout(gtx)
		},
		func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Right: values.MarginPadding4}.Layout(gtx, pg.cancelPurchase.Layout)
					}),
					layout.Rigid(pg.reviewPurchase.Layout),
				)
			})
		},
	}, 900)
}

func (pg *ticketPage) confirmPurchaseModal(gtx layout.Context) layout.Dimensions {
	return pg.purchaseOptions.Layout(gtx, []layout.Widget{
		func(gtx C) D {
			return pg.th.Label(values.TextSize20, "Confirm to purchase tickets").Layout(gtx)
		},
		func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					tleft := pg.th.Label(values.TextSize14, "Amount")
					tleft.Color = pg.th.Color.Gray2
					tright := pg.th.Label(values.TextSize14, pg.ticketAmount.Editor.Text())
					return endToEndRow(gtx, tleft.Layout, tright.Layout)
				}),
				layout.Rigid(func(gtx C) D {
					tleft := pg.th.Label(values.TextSize14, "Total cost")
					tleft.Color = pg.th.Color.Gray2
					tright := pg.th.Label(values.TextSize14, pg.totalCost)
					return endToEndRow(gtx, tleft.Layout, tright.Layout)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top:    values.MarginPadding16,
						Bottom: values.MarginPadding16,
					}.Layout(gtx, pg.th.Separator().Layout)
				}),
				layout.Rigid(func(gtx C) D {
					tleft := pg.th.Label(values.TextSize14, "Account")
					tleft.Color = pg.th.Color.Gray2
					tright := pg.th.Label(values.TextSize14, pg.selector.selectedAccount.Name)
					return endToEndRow(gtx, tleft.Layout, tright.Layout)
				}),
				layout.Rigid(func(gtx C) D {
					tleft := pg.th.Label(values.TextSize14, "Remaining")
					tleft.Color = pg.th.Color.Gray2
					tright := pg.th.Label(values.TextSize14, pg.remainingBalance)
					return endToEndRow(gtx, tleft.Layout, tright.Layout)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top:    values.MarginPadding16,
						Bottom: values.MarginPadding16,
					}.Layout(gtx, pg.th.Separator().Layout)
				}),
				layout.Rigid(func(gtx C) D {
					tleft := pg.th.Label(values.TextSize14, "VSP")
					tleft.Color = pg.th.Color.Gray2
					tright := pg.th.Label(values.TextSize14, pg.selectedVSP.Host)
					return endToEndRow(gtx, tleft.Layout, tright.Layout)
				}),
			)
		},
		func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(pg.spendingPassword.Layout),
			)
		},
		func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Right: values.MarginPadding4}.Layout(gtx, pg.cancelConfirmPurchase.Layout)
					}),
					layout.Rigid(pg.submitPurchase.Layout),
				)
			})
		},
	}, 900)
}

func (pg *ticketPage) vspHostSelectorLayout(gtx C, c pageCommon) layout.Dimensions {
	border := widget.Border{
		Color:        pg.th.Color.Gray1,
		CornerRadius: values.MarginPadding8,
		Width:        values.MarginPadding2,
	}
	return border.Layout(gtx, func(gtx C) D {
		return layout.UniformInset(values.MarginPadding12).Layout(gtx, func(gtx C) D {
			return decredmaterial.Clickable(gtx, pg.showVSP, func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						if pg.selectedVSP.Host == "" {
							txt := pg.th.Label(values.TextSize16, "Select VSP...")
							txt.Color = pg.th.Color.Gray2
							return txt.Layout(gtx)
						}
						return pg.th.Label(values.TextSize16, pg.selectedVSP.Host).Layout(gtx)
					}),
					layout.Flexed(1, func(gtx C) D {
						return layout.E.Layout(gtx, func(gtx C) D {
							return layout.Flex{}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									if pg.selectedVSP.Info == nil {
										return layout.Dimensions{}
									}
									txt := pg.th.Label(values.TextSize16, fmt.Sprintf("%v", pg.selectedVSP.Info.FeePercentage)+"%")
									txt.Color = pg.th.Color.DeepBlue
									return txt.Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									inset := layout.Inset{
										Left: values.MarginPadding15,
									}
									return inset.Layout(gtx, func(gtx C) D {
										return c.icons.dropDownIcon.Layout(gtx, values.MarginPadding20)
									})
								}),
							)
						})
					}),
				)
			})
		})
	})
}

func (pg *ticketPage) vspHostModalLayout(gtx C, c pageCommon) layout.Dimensions {
	return pg.purchaseOptions.Layout(gtx, []layout.Widget{
		func(gtx C) D {
			return pg.th.Label(values.TextSize20, "Voting service provider").Layout(gtx)
		},
		func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					txt := pg.th.Label(values.TextSize14, "Address")
					txt.Color = pg.th.Color.Gray2
					txtFee := pg.th.Label(values.TextSize14, "Fee")
					txtFee.Color = pg.th.Color.Gray2
					return layout.Inset{Right: values.MarginPadding40}.Layout(gtx, func(gtx C) D {
						return endToEndRow(gtx, txt.Layout, txtFee.Layout)
					})
				}),
				layout.Rigid(func(gtx C) D {
					listVSP := pg.loadedVSP
					return pg.vspHosts.Layout(gtx, len(listVSP), func(gtx C, i int) D {
						click := pg.selectVSP[i]
						pointer.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Add(gtx.Ops)
						click.Add(gtx.Ops)
						pg.handlerSelectVSP(click.Events(gtx), listVSP[i], c)

						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							layout.Flexed(0.8, func(gtx C) D {
								return layout.Inset{Top: values.MarginPadding12, Bottom: values.MarginPadding12}.Layout(gtx, func(gtx C) D {
									txt := pg.th.Label(values.TextSize14, fmt.Sprintf("%v", listVSP[i].Info.FeePercentage)+"%")
									txt.Color = pg.th.Color.Gray2
									return endToEndRow(gtx, pg.th.Label(values.TextSize16, listVSP[i].Host).Layout, txt.Layout)
								})
							}),
							layout.Rigid(func(gtx C) D {
								if pg.selectedVSP.Host != listVSP[i].Host {
									return layout.Inset{Right: values.MarginPadding40}.Layout(gtx, func(gtx C) D {
										return layout.Dimensions{}
									})
								}
								return layout.Inset{Left: values.MarginPadding20}.Layout(gtx, func(gtx C) D {
									return c.icons.navigationCheck.Layout(gtx, values.MarginPadding20)
								})
							}),
						)
					})
				}),
			)
		},
		func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, pg.inputVSP.Layout),
				layout.Rigid(pg.addVSP.Layout),
			)
		},
	}, 900)
}

func (pg *ticketPage) handlerSelectVSP(events []gesture.ClickEvent, v wallet.VSPInfo, c pageCommon) {
	for _, e := range events {
		if e.Type == gesture.TypeClick {
			pg.selectedVSP = v
			pg.createNewVSPD()
			pg.showVSPHosts = false
			if pg.rememberVSP.CheckBox.Value {
				c.wallet.RememberVSP(pg.selectedVSP.Host)
			}
		}
	}
}

func (pg *ticketPage) editorsNotEmpty(btn *decredmaterial.Button, editors ...*widget.Editor) bool {
	btn.Color = pg.th.Color.Surface
	for _, e := range editors {
		if e.Text() == "" {
			btn.Background = pg.th.Color.Hint
			return false
		}
	}

	btn.Background = pg.th.Color.Primary
	return true
}

func (pg *ticketPage) calculateAndValidCost(c pageCommon) bool {
	tprice, _ := c.wallet.TicketPrice()
	tnumber, err := strconv.ParseInt(pg.ticketAmount.Editor.Text(), 10, 64)
	pg.submitPurchase.Text = "Purchase tickets"
	pg.reviewPurchase.Background = pg.th.Color.Hint
	if err != nil || pg.selectedVSP.Info == nil {
		return false
	}
	pg.submitPurchase.Text = fmt.Sprintf("Purchase %d tickets", tnumber)

	accountBalance := pg.selector.selectedAccount.Balance.Spendable
	feePercentage := pg.selectedVSP.Info.FeePercentage
	total := tprice * tnumber
	feeDCR := int64((float64(total) / 100) * feePercentage)
	remaining := accountBalance - total + feeDCR

	if accountBalance < total+feeDCR || remaining < 0 {
		return false
	}

	pg.reviewPurchase.Background = pg.th.Color.Primary
	pg.totalCost = dcrutil.Amount(total).String()
	pg.remainingBalance = dcrutil.Amount(remaining).String()
	return true
}

func (pg *ticketPage) doPurchaseTicket(c pageCommon, password []byte, ticketAmount uint32) {
	if pg.isPurchaseLoading {
		log.Info("Please wait...")
		return
	}
	pg.isPurchaseLoading = true
	c.wallet.PurchaseTicket(pg.walletSelectedID, pg.accountSelectedNumber, ticketAmount, password, pg.vspd, pg.purchaseErrChan)
}

func (pg *ticketPage) createNewVSPD() {
	pg.vspd = pg.common.multiWallet.NewVSPD(pg.selectedVSP.Host, pg.walletSelectedID, pg.accountSelectedNumber)
}

func (pg *ticketPage) handle() {
	pg.selector.handle()

	c := pg.common
	// TODO: frefresh when ticket price update from remote
	if pg.ticketPrice == "" {
		_, priceText := c.wallet.TicketPrice()
		pg.ticketPrice = priceText
		c.wallet.GetAllVSP()
	}

	if len(pg.loadedVSP) != len(pg.selectVSP) {
		pg.selectVSP = createClickGestures(len(pg.loadedVSP))
	}

	for _, evt := range pg.ticketAmount.Editor.Events() {
		switch evt.(type) {
		case widget.ChangeEvent:
			pg.calculateAndValidCost(c)
		}
	}

	if pg.purchaseTicket.Button.Clicked() {
		var vspHost string
		pg.common.multiWallet.ReadUserConfigValue(dcrlibwallet.VSPHostConfigKey, &vspHost)
		if vspHost != "" {
			for _, vinfo := range pg.loadedVSP {
				if vinfo.Host == vspHost {
					pg.selectedVSP = vinfo
					pg.rememberVSP.CheckBox.Value = true
					pg.createNewVSPD()
					break
				}
			}
		}

		if pg.autoPurchaseEnabled.Value {
			// TODO: calculate ticket number and vsp selected
			pg.showPurchaseConfirm = true
			return
		}
		pg.showPurchaseOptions = true
	}

	if pg.cancelConfirmPurchase.Button.Clicked() {
		pg.showPurchaseConfirm = false
	}

	if pg.editorsNotEmpty(&pg.submitPurchase, pg.spendingPassword.Editor) &&
		pg.calculateAndValidCost(c) &&
		pg.submitPurchase.Button.Clicked() {
		i, err := strconv.Atoi(pg.ticketAmount.Editor.Text())
		if err != nil {
			return
		}
		pg.doPurchaseTicket(c, []byte(pg.spendingPassword.Editor.Text()), uint32(i))
	}

	if pg.cancelPurchase.Button.Clicked() {
		pg.showPurchaseOptions = false
	}

	if pg.calculateAndValidCost(c) && pg.reviewPurchase.Button.Clicked() {
		pg.showPurchaseConfirm = true
	}

	if pg.showVSP.Clicked() {
		c.wallet.GetAllVSP()
		pg.showVSPHosts = true
	}

	if pg.editorsNotEmpty(&pg.addVSP, pg.inputVSP.Editor) && pg.addVSP.Button.Clicked() {
		// c.wallet.AddVSP("http://dev.planetdecred.org:23125", pg.vspErrChan)
		c.wallet.AddVSP(pg.inputVSP.Editor.Text(), pg.vspErrChan)
	}

	if pg.rememberVSP.CheckBox.Changed() {
		if pg.rememberVSP.CheckBox.Value {
			c.wallet.RememberVSP(pg.selectedVSP.Host)
		} else {
			c.wallet.RememberVSP("")
		}
	}

	if pg.toTickets.Button.Clicked() {
		c.changePage(TicketPageList(c, pg.tickets))
	}

	select {
	case err := <-pg.vspErrChan:
		c.notify(err.Error(), false)
	case err := <-pg.purchaseErrChan:
		if err != nil {
			c.notify(err.Error(), false)
		} else {
			pg.spendingPassword.Editor.SetText("")
			pg.showPurchaseConfirm = false
			pg.showVSPHosts = false
			pg.showPurchaseOptions = false
		}
		pg.isPurchaseLoading = false
	default:
	}
}

func (pg *ticketPage) onClose() {}
