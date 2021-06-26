// components contain layout code that are shared by multiple pages but aren't widely used enough to be defined as
// widgets

package ui

import (
	"image"
	"strings"
	"time"

	"gioui.org/gesture"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"github.com/decred/dcrd/dcrutil"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"
)

const (
	purchasingAccountTitle = "Purchasing account"
	sendingAccountTitle    = "Sending account"
	receivingAccountTitle  = "Receiving account"
	ticketAge              = "Ticket age"
	durationMsg            = "10 hrs 47 mins (118/256 blocks)"
)

type (
	TransactionRow struct {
		transaction dcrlibwallet.Transaction
		index       int
		showBadge   bool
	}
	toast struct {
		text    string
		success bool
		timer   *time.Timer
	}
	tooltips struct {
		statusTooltip     *decredmaterial.Tooltip
		walletNameTooltip *decredmaterial.Tooltip
		dateTooltip       *decredmaterial.Tooltip
		daysBehindTooltip *decredmaterial.Tooltip
		durationTooltip   *decredmaterial.Tooltip
	}
)

// layoutBalance aligns the main and sub DCR balances horizontally, putting the sub
// balance at the baseline of the row.
func (page *pageCommon) layoutBalance(gtx layout.Context, amount string, isSwitchColor bool) layout.Dimensions {
	// todo: make "DCR" symbols small when there are no decimals in the balance
	mainMsg, subText := breakBalance(page.printer, amount)
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			label := page.theme.Label(values.TextSize20, mainMsg)
			if isSwitchColor {
				label.Color = page.theme.Color.DeepBlue
			}
			return label.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			label := page.theme.Label(values.TextSize14, subText)
			if isSwitchColor {
				label.Color = page.theme.Color.DeepBlue
			}
			return label.Layout(gtx)
		}),
	)
}

var (
	navDrawerWidth          = unit.Value{U: unit.UnitDp, V: 160}
	navDrawerMinimizedWidth = unit.Value{U: unit.UnitDp, V: 72}
	title                   = ""
	mainMsg                 = ""
	mainMsgDesc             = ""
	dayBehind               = ""
	durationTitle           = ""
	durationDesc            = ""
)

// transactionRow is a single transaction row on the transactions and overview page. It lays out a transaction's
// direction, balance, status.
func transactionRow(gtx layout.Context, common *pageCommon, row TransactionRow) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	directionIconTopMargin := values.MarginPadding16

	if row.index == 0 && row.showBadge {
		directionIconTopMargin = values.MarginPadding14
	} else if row.index == 0 {
		// todo: remove top margin from container
		directionIconTopMargin = values.MarginPadding0
	}

	wal := common.multiWallet.WalletWithID(row.transaction.WalletID)

	return layout.Inset{Top: values.MarginPadding16}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				icon := common.icons.receiveIcon
				if row.transaction.Direction == dcrlibwallet.TxDirectionSent {
					icon = common.icons.sendIcon
				}
				icon.Scale = 1.0

				return layout.Inset{Top: directionIconTopMargin}.Layout(gtx, func(gtx C) D {
					return icon.Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if row.index == 0 {
							return layout.Dimensions{}
						}
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						separator := common.theme.Separator()
						separator.Width = gtx.Constraints.Max.X - gtx.Px(unit.Dp(16))
						return layout.E.Layout(gtx, func(gtx C) D {
							// Todo: add comment
							marginBottom := values.MarginPadding16
							if row.showBadge {
								marginBottom = values.MarginPadding5
							}
							return layout.Inset{Bottom: marginBottom}.Layout(gtx,
								func(gtx C) D {
									return separator.Layout(gtx)
								})
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return layout.Inset{}.Layout(gtx, func(gtx C) D {
							return layout.Flex{
								Axis:      layout.Horizontal,
								Spacing:   layout.SpaceBetween,
								Alignment: layout.Middle,
							}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return layout.Inset{Left: values.MarginPadding16}.Layout(gtx, func(gtx C) D {
										return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
											layout.Rigid(func(gtx C) D {
												return common.layoutBalance(gtx, dcrutil.Amount(row.transaction.Amount).String(), true)
											}),
											layout.Rigid(func(gtx C) D {
												if row.showBadge {
													return walletLabel(gtx, common, wal.Name)
												}
												return layout.Dimensions{}
											}),
										)
									})
								}),
								layout.Rigid(func(gtx C) D {
									return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											return layout.Inset{Right: values.MarginPadding8}.Layout(gtx,
												func(gtx C) D {
													status := common.theme.Body1("pending")
													if txConfirmations(common, row.transaction) <= 1 {
														status.Color = common.theme.Color.Gray5
													} else {
														status.Color = common.theme.Color.Gray4
														status.Text = formatDateOrTime(row.transaction.Timestamp)
													}
													status.Alignment = text.Middle
													return status.Layout(gtx)
												})
										}),
										layout.Rigid(func(gtx C) D {
											return layout.Inset{Right: values.MarginPadding16}.Layout(gtx, func(gtx C) D {
												statusIcon := common.icons.confirmIcon
												if txConfirmations(common, row.transaction) <= 1 {
													statusIcon = common.icons.pendingIcon
												}
												statusIcon.Scale = 1.0
												return statusIcon.Layout(gtx)
											})
										}),
									)
								}),
							)
						})
					}),
				)
			}),
		)
	})
}

func txConfirmations(common *pageCommon, transaction dcrlibwallet.Transaction) int32 {
	if transaction.BlockHeight != -1 {
		// TODO
		return (common.multiWallet.WalletWithID(transaction.WalletID).GetBestBlock() - transaction.BlockHeight) + 1
	}

	return 0
}

// walletLabel displays the wallet which a transaction belongs to. It is only displayed on the overview page when there
// are transactions from multiple wallets
func walletLabel(gtx layout.Context, c *pageCommon, walletName string) D {
	return decredmaterial.Card{
		Color: c.theme.Color.LightGray,
	}.Layout(gtx, func(gtx C) D {
		return Container{
			layout.Inset{
				Left:  values.MarginPadding4,
				Right: values.MarginPadding4,
			}}.Layout(gtx, func(gtx C) D {
			name := c.theme.Label(values.TextSize12, walletName)
			name.Color = c.theme.Color.Gray
			return name.Layout(gtx)
		})
	})
}

// endToEndRow layouts out its content on both ends of its horizontal layout.
func endToEndRow(gtx layout.Context, leftWidget, rightWidget func(C) D) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(leftWidget),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, rightWidget)
		}),
	)
}

func (page *pageCommon) Modal(gtx layout.Context, body layout.Dimensions, modal layout.Dimensions) layout.Dimensions {
	dims := layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			return body
		}),
		layout.Stacked(func(gtx C) D {
			return modal
		}),
	)
	return dims
}

func (page *pageCommon) initSelectAccountWidget(wallAcct map[int][]walletAccount, windex int) {
	if _, ok := wallAcct[windex]; !ok {
		accounts := page.info.Wallets[windex].Accounts
		if len(accounts) != len(wallAcct[windex]) {
			wallAcct[windex] = make([]walletAccount, len(accounts))
			for aindex := range accounts {
				if accounts[aindex].Name == "imported" {
					continue
				}

				wallAcct[windex][aindex] = walletAccount{
					walletIndex:  windex,
					accountIndex: aindex,
					evt:          &gesture.Click{},
					accountName:  accounts[aindex].Name,
					totalBalance: accounts[aindex].TotalBalance,
					spendable:    dcrutil.Amount(accounts[aindex].SpendableBalance).String(),
					number:       accounts[aindex].Number,
				}
			}
		}
	}
}

func setText(t string) {
	switch t {
	case "UNMINED":
		title = "This ticket is waiting in mempool to be included in a block."
		mainMsg, mainMsgDesc, dayBehind, durationTitle, durationDesc = "", "", ticketAge, "Live in", durationMsg
	case "IMMATURE":
		title = "This ticket will enter the ticket pool and become a live ticket after 256 blocks (~20 hrs)."
		mainMsg, mainMsgDesc, dayBehind, durationTitle, durationDesc = "", "", ticketAge, "Live in", durationMsg
	case "LIVE":
		title = "Waiting to be chosen to vote."
		mainMsg = "The average vote time is 28 days, but can take up to 142 days."
		mainMsgDesc = "There is a 0.5% chance of expiring before being chosen to vote (this expiration returns the original ticket price without a reward)."
		dayBehind, durationTitle, durationDesc = ticketAge, "Live in", durationMsg
	case "VOTED":
		title = "Congratulations! This ticket has voted."
		mainMsg = "The ticket price + reward will become spendable after 256 blocks (~20 hrs)."
		dayBehind, durationTitle, durationDesc = "Days to vote", "Spendable in", durationMsg
	case "MISSED":
		title = "This ticket was chosen to vote, but missed the voting window."
		mainMsg = "Missed tickets will be revoked to return the original ticket price to you."
		mainMsgDesc = "If a ticket is not revoked automatically, use the revoke button."
		dayBehind, durationTitle, durationDesc = "Days to miss", "Miss in", durationMsg
	case "EXPIRED":
		title = "This ticket has not been chosen to vote within 40960 blocks, and thus expired. "
		mainMsg = "Expired tickets will be revoked to return the original ticket price to you."
		mainMsgDesc = "If a ticket is not revoked automatically, use the revoke button."
		dayBehind, durationTitle, durationDesc = "Days to expire", "Expired in", durationMsg
	case "REVOKED":
		title = "This ticket has been revoked."
		dayBehind, durationTitle, durationDesc = ticketAge, "Spendable in", durationMsg
	}
}

func ticketStatusTooltip(gtx C, c *pageCommon, t *wallet.Ticket) layout.Dimensions {
	setText(t.Info.Status)
	st := ticketStatusIcon(c, t.Info.Status)
	status := c.theme.Body2(t.Info.Status)
	status.Color = st.color
	st.icon.Scale = .5

	titleLabel, mainMsgLabel, mainMsgLabel2 := c.theme.Body2(title), c.theme.Body2(mainMsg), c.theme.Body2(mainMsgDesc)
	mainMsgLabel.Color, mainMsgLabel2.Color = c.theme.Color.Gray, c.theme.Color.Gray
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(st.icon.Layout),
				layout.Rigid(toolTipContent(layout.Inset{Left: values.MarginPadding4}, status.Layout)),
			)
		}),
		layout.Rigid(toolTipContent(layout.Inset{Top: values.MarginPadding8}, titleLabel.Layout)),
		layout.Rigid(toolTipContent(layout.Inset{Top: values.MarginPadding8}, mainMsgLabel.Layout)),
		layout.Rigid(func(gtx C) D {
			if mainMsgDesc != "" {
				toolTipContent(layout.Inset{Top: values.MarginPadding8}, mainMsgLabel2.Layout)
			}
			return layout.Dimensions{}
		}),
	)
}

func ticketCardTooltip(gtx C, rectLayout layout.Dimensions, tooltip *decredmaterial.Tooltip, body layout.Widget) layout.Dimensions {
	inset := layout.Inset{
		Top:   values.MarginPadding15,
		Right: unit.Dp(-150),
		Left:  values.MarginPadding15,
	}

	rect := image.Rectangle{
		Max: image.Point{
			X: rectLayout.Size.X,
			Y: rectLayout.Size.Y,
		},
	}
	return tooltip.Layout(gtx, rect, inset, body)
}

func walletNameDateTimeTooltip(gtx C, c *pageCommon, title string, body layout.Widget) layout.Dimensions {
	walletNameLabel := c.theme.Body2(title)
	walletNameLabel.Color = c.theme.Color.Gray
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(walletNameLabel.Layout),
		layout.Rigid(body),
	)
}

func toolTipContent(inset layout.Inset, body layout.Widget) layout.Widget {
	return func(gtx C) D {
		return inset.Layout(gtx, body)
	}
}

// ticketCard layouts out ticket info with the shadow box, use for list horizontal or list grid
func ticketCard(gtx layout.Context, c *pageCommon, t *wallet.Ticket, tooltip tooltips) layout.Dimensions {
	var itemWidth int
	status := t.Info.Status
	st := ticketStatusIcon(c, status)
	if st == nil {
		return layout.Dimensions{}
	}
	st.icon.Scale = 1.0
	return c.theme.Shadow().Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				wrap := c.theme.Card()
				wrap.Radius = decredmaterial.CornerRadius{NE: 8, NW: 8, SE: 0, SW: 0}
				wrap.Color = st.background
				return wrap.Layout(gtx, func(gtx C) D {
					return layout.Stack{Alignment: layout.S}.Layout(gtx,
						layout.Expanded(func(gtx C) D {
							return layout.NE.Layout(gtx, func(gtx C) D {
								wTimeLabel := c.theme.Card()
								wTimeLabel.Radius = decredmaterial.CornerRadius{NE: 0, NW: 8, SE: 0, SW: 8}
								return wTimeLabel.Layout(gtx, func(gtx C) D {
									return layout.Inset{
										Top:    values.MarginPadding4,
										Bottom: values.MarginPadding4,
										Right:  values.MarginPadding8,
										Left:   values.MarginPadding8,
									}.Layout(gtx, func(gtx C) D {
										txt := c.theme.Label(values.TextSize14, "10h 47m")
										txtLayout := txt.Layout(gtx)
										ticketCardTooltip(gtx, txtLayout, tooltip.durationTooltip, func(gtx C) D {
											setText(status)
											return walletNameDateTimeTooltip(gtx, c, durationTitle,
												toolTipContent(layout.Inset{Top: values.MarginPadding8}, c.theme.Body2(durationMsg).Layout))
										})
										return txtLayout
									})
								})
							})
						}),

						layout.Stacked(func(gtx C) D {
							content := layout.Inset{
								Top:    values.MarginPadding24,
								Right:  values.MarginPadding62,
								Left:   values.MarginPadding62,
								Bottom: values.MarginPadding24,
							}.Layout(gtx, func(gtx C) D {
								return st.icon.Layout(gtx)
							})
							itemWidth = content.Size.X
							return content
						}),

						layout.Stacked(func(gtx C) D {
							return layout.Center.Layout(gtx, func(gtx C) D {
								return layout.Inset{Top: values.MarginPadding20}.Layout(gtx, func(gtx C) D {
									gtx.Constraints.Max.X = itemWidth
									p := c.theme.ProgressBar(20)
									p.Height, p.Radius = values.MarginPadding4, values.MarginPadding1
									p.Color = st.color
									return p.Layout(gtx)
								})
							})
						}),
					)
				})
			}),
			layout.Rigid(func(gtx C) D {
				wrap := c.theme.Card()
				wrap.Radius = decredmaterial.CornerRadius{NE: 0, NW: 0, SE: 8, SW: 8}
				return wrap.Layout(gtx, func(gtx C) D {
					gtx.Constraints.Min.X, gtx.Constraints.Max.X = itemWidth, itemWidth
					return layout.Inset{
						Left:   values.MarginPadding12,
						Right:  values.MarginPadding12,
						Bottom: values.MarginPadding8,
					}.Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								return layout.Inset{
									Top: values.MarginPadding16,
								}.Layout(gtx, func(gtx C) D {
									return c.layoutBalance(gtx, t.Amount, true)
								})
							}),
							layout.Rigid(func(gtx C) D {
								return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										txt := c.theme.Label(values.MarginPadding14, t.Info.Status)
										txt.Color = st.color
										txtLayout := txt.Layout(gtx)
										ticketCardTooltip(gtx, txtLayout, tooltip.statusTooltip, func(gtx C) D {
											setText(status)
											return ticketStatusTooltip(gtx, c, t)
										})
										return txtLayout
									}),
									layout.Rigid(func(gtx C) D {
										return layout.Inset{
											Left:  values.MarginPadding4,
											Right: values.MarginPadding4,
										}.Layout(gtx, func(gtx C) D {
											ic := c.icons.imageBrightness1
											ic.Color = c.theme.Color.Gray2
											return c.icons.imageBrightness1.Layout(gtx, values.MarginPadding5)
										})
									}),
									layout.Rigid(func(gtx C) D {
										txt := c.theme.Label(values.MarginPadding14, t.WalletName)
										txt.Color = c.theme.Color.Gray
										txtLayout := txt.Layout(gtx)
										ticketCardTooltip(gtx, txtLayout, tooltip.walletNameTooltip, func(gtx C) D {
											setText(status)
											return walletNameDateTimeTooltip(gtx, c, "Wallet name",
												toolTipContent(layout.Inset{Top: values.MarginPadding8}, c.theme.Body2(t.WalletName).Layout))
										})
										return txtLayout
									}),
								)
							}),
							layout.Rigid(func(gtx C) D {
								return layout.Inset{
									Top:    values.MarginPadding16,
									Bottom: values.MarginPadding16,
								}.Layout(gtx, func(gtx C) D {
									txt := c.theme.Label(values.TextSize14, t.MonthDay)
									txt.Color = c.theme.Color.Gray2
									return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											txtLayout := txt.Layout(gtx)
											ticketCardTooltip(gtx, txtLayout, tooltip.dateTooltip, func(gtx C) D {
												dt := strings.Split(t.DateTime, " ")
												s1 := []string{dt[0], dt[1], dt[2]}
												date := strings.Join(s1, " ")
												s2 := []string{dt[3], dt[4]}
												time := strings.Join(s2, " ")
												dateTime := fmt.Sprintf("%s at %s", date, time)
												return walletNameDateTimeTooltip(gtx, c, "Purchased",
													toolTipContent(layout.Inset{Top: values.MarginPadding8}, c.theme.Body2(dateTime).Layout))
											})
											return txtLayout
										}),
										layout.Rigid(func(gtx C) D {
											return layout.Inset{
												Left:  values.MarginPadding4,
												Right: values.MarginPadding4,
											}.Layout(gtx, func(gtx C) D {
												ic := c.icons.imageBrightness1
												ic.Color = c.theme.Color.Gray2
												return c.icons.imageBrightness1.Layout(gtx, values.MarginPadding5)
											})
										}),
										layout.Rigid(func(gtx C) D {
											txt.Text = t.DaysBehind
											txtLayout := txt.Layout(gtx)
											ticketCardTooltip(gtx, txtLayout, tooltip.daysBehindTooltip, func(gtx C) D {
												setText(status)
												return walletNameDateTimeTooltip(gtx, c, dayBehind,
													toolTipContent(layout.Inset{Top: values.MarginPadding8}, c.theme.Body2(t.DaysBehind).Layout))
											})
											return txtLayout
										}),
									)
								})
							}),
						)
					})
				})
			}),
		)
	})
}

// ticketActivityRow layouts out ticket info, display ticket activities on the tickets_page and tickets_activity_page
func ticketActivityRow(gtx layout.Context, c *pageCommon, t wallet.Ticket, index int) layout.Dimensions {
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Right: values.MarginPadding16}.Layout(gtx, func(gtx C) D {
				st := ticketStatusIcon(c, t.Info.Status)
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
					separator := c.theme.Separator()
					separator.Width = gtx.Constraints.Max.X
					return layout.E.Layout(gtx, func(gtx C) D {
						return separator.Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top:    values.MarginPadding8,
						Bottom: values.MarginPadding8,
					}.Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								labelStatus := c.theme.Label(values.TextSize18, strings.Title(strings.ToLower(t.Info.Status)))
								labelStatus.Color = c.theme.Color.DeepBlue

								labelDaysBehind := c.theme.Label(values.TextSize14, t.DaysBehind)
								labelDaysBehind.Color = c.theme.Color.DeepBlue

								return endToEndRow(gtx,
									labelStatus.Layout,
									labelDaysBehind.Layout)
							}),
							layout.Rigid(func(gtx C) D {
								return layout.Flex{
									Alignment: layout.Middle,
								}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										txt := c.theme.Label(values.TextSize14, t.WalletName)
										txt.Color = c.theme.Color.Gray2
										return txt.Layout(gtx)
									}),
									layout.Rigid(func(gtx C) D {
										return layout.Inset{
											Left:  values.MarginPadding4,
											Right: values.MarginPadding4,
										}.Layout(gtx, func(gtx C) D {
											ic := c.icons.imageBrightness1
											ic.Color = c.theme.Color.Gray2
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
										txt := c.theme.Label(values.TextSize14, t.Amount)
										txt.Color = c.theme.Color.Gray2
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

func displayToast(th *decredmaterial.Theme, gtx layout.Context, n *toast) layout.Dimensions {
	color := th.Color.Success
	if !n.success {
		color = th.Color.Danger
	}

	card := th.Card()
	card.Color = color
	return card.Layout(gtx, func(gtx C) D {
		return layout.Inset{
			Top: values.MarginPadding7, Bottom: values.MarginPadding7,
			Left: values.MarginPadding15, Right: values.MarginPadding15,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			t := th.Body1(n.text)
			t.Color = th.Color.Surface
			return t.Layout(gtx)
		})
	})
}

func (page *pageCommon) handleToast() {
	if (*page.toast) == nil {
		return
	}

	if (*page.toast).timer == nil {
		(*page.toast).timer = time.NewTimer(time.Second * 3)
	}

	select {
	case <-(*page.toast).timer.C:
		*page.toast = nil
	default:
	}
}

// createOrUpdateWalletDropDown check for len of wallets to create dropDown,
// also update the list when create, update, delete a wallet.
func (page *pageCommon) createOrUpdateWalletDropDown(dwn **decredmaterial.DropDown, wallets []*dcrlibwallet.Wallet) {
	var walletDropDownItems []decredmaterial.DropDownItem
	for _, wal := range wallets {
		item := decredmaterial.DropDownItem{
			Text: wal.Name,
			Icon: page.icons.walletIcon,
		}
		walletDropDownItems = append(walletDropDownItems, item)
	}
	*dwn = page.theme.DropDown(walletDropDownItems, 2)
}

func createOrderDropDown(c *pageCommon) *decredmaterial.DropDown {
	return c.theme.DropDown([]decredmaterial.DropDownItem{{Text: values.String(values.StrNewest)},
		{Text: values.String(values.StrOldest)}}, 1)
}

func (page *pageCommon) handler() {
	page.handleToast()
}
