// components contain layout code that are shared by multiple pages but aren't widely used enough to be defined as
// widgets

package ui

import (
	"time"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"github.com/decred/dcrd/dcrutil"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"
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
)

// layoutBalance aligns the main and sub DCR balances horizontally, putting the sub
// balance at the baseline of the row.
func (page pageCommon) layoutBalance(gtx layout.Context, amount string, isSwitchColor bool) layout.Dimensions {
	// todo: make "DCR" symbols small when there are no decimals in the balance
	mainText, subText := breakBalance(page.printer, amount)
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			label := page.theme.Label(values.TextSize20, mainText)
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
													if txConfirmations(common, &row.transaction) <= 1 {
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
												if txConfirmations(common, &row.transaction) <= 1 {
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

func (page pageCommon) Modal(gtx layout.Context, body layout.Dimensions, modal layout.Dimensions) layout.Dimensions {
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

// ticketCard layouts out ticket info with the shadow box, use for list horizontal or list grid
func ticketCard(gtx layout.Context, c pageCommon, t *wallet.Ticket) layout.Dimensions {
	var itemWidth int
	st := ticketStatusIcon(&c, t.Info.Status)
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
										return c.theme.Label(values.TextSize14, "10h 47m").Layout(gtx)
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
										return c.theme.Label(values.MarginPadding14, t.WalletName).Layout(gtx)
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
											txt.Text = t.DaysBehind
											return txt.Layout(gtx)
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

func (page pageCommon) handleToast() {
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

func (page pageCommon) handler() {
	page.handleToast()
}
