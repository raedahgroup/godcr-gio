package ui

import (
	"fmt"

	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/decred/dcrd/dcrutil"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const ModalSendConfirm = "send_confirm_modal"

type sendConfirmModal struct {
	*pageCommon
	modal decredmaterial.Modal

	*comfirmModalData
}

func newSendConfirmModal(common *pageCommon, data *comfirmModalData) *sendConfirmModal {
	scm := &sendConfirmModal{
		pageCommon: common,
		modal:      *common.theme.ModalFloatTitle(),

		comfirmModalData: data,
	}

	return scm
}

func (scm *sendConfirmModal) ModalID() string {
	return ModalSendConfirm
}

func (scm *sendConfirmModal) Show() {
	scm.pageCommon.showModal(scm)
}

func (scm *sendConfirmModal) Dismiss() {
	scm.dismissModal(scm)
}

func (scm *sendConfirmModal) OnResume() {
}

func (scm *sendConfirmModal) OnDismiss() {

}

func (scm *sendConfirmModal) Handle() {
	if scm.passwordEditor.Editor.Text() == "" {
		scm.confirmButton.Background = scm.theme.Color.InactiveGray
	} else {
		scm.confirmButton.Background = scm.theme.Color.Primary
	}
}

func (scm *sendConfirmModal) Layout(gtx layout.Context) D {
	receiveAcct := scm.destinationAccountSelector.selectedAccount
	receiveWallet := scm.multiWallet.WalletWithID(receiveAcct.WalletID)
	sendAcct := scm.sourceAccountSelector.selectedAccount
	sendWallet := scm.multiWallet.WalletWithID(sendAcct.WalletID)

	w := []layout.Widget{
		func(gtx C) D {
			return scm.theme.H6("Confim to send").Layout(gtx)
		},
		func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					icon := scm.icons.sendIcon
					icon.Scale = 0.7
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return layout.Inset{Top: values.MarginPadding2, Right: values.MarginPadding16}.Layout(gtx, icon.Layout)
						}),
						layout.Rigid(func(gtx C) D {
							return layout.Flex{}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return scm.layoutBalance(gtx, scm.sendAmountDCR, true)
								}),
								layout.Flexed(1, func(gtx C) D {
									if scm.usdExchangeSet {
										return layout.E.Layout(gtx, func(gtx C) D {
											txt := scm.theme.Body1(scm.sendAmountUSD)
											txt.Color = scm.theme.Color.Gray
											return txt.Layout(gtx)
										})
									}
									return layout.Dimensions{}
								}),
							)
						}),
					)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							icon := scm.icons.navigationArrowForward
							icon.Color = scm.theme.Color.Gray3
							return layout.Inset{Right: values.MarginPadding8}.Layout(gtx, func(gtx C) D {
								return icon.Layout(gtx, values.MarginPadding15)
							})
						}),
						layout.Rigid(func(gtx C) D {
							if scm.sendToOption == "My account" {
								return layout.E.Layout(gtx, func(gtx C) D {
									return layout.Flex{}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											return scm.theme.Body2(receiveAcct.Name).Layout(gtx)
										}),
										layout.Rigid(func(gtx C) D {
											card := scm.theme.Card()
											card.Radius = decredmaterial.CornerRadius{NE: 0, NW: 0, SE: 0, SW: 0}
											card.Color = scm.theme.Color.LightGray
											inset := layout.Inset{
												Left: values.MarginPadding5,
											}
											return inset.Layout(gtx, func(gtx C) D {
												return card.Layout(gtx, func(gtx C) D {
													return layout.UniformInset(values.MarginPadding2).Layout(gtx, func(gtx C) D {
														txt := scm.theme.Caption(receiveWallet.Name)
														txt.Color = scm.theme.Color.Gray
														return txt.Layout(gtx)
													})
												})
											})
										}),
									)
								})
							}
							return scm.theme.Body2(scm.destinationAddress).Layout(gtx)
						}),
					)
				}),
			)
		},
		func(gtx C) D {
			return scm.theme.Separator().Layout(gtx)
		},
		func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return scm.contentRow(gtx, "Sending from", sendAcct.Name, sendWallet.Name)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: values.MarginPadding8, Bottom: values.MarginPadding8}.Layout(gtx, func(gtx C) D {
						if scm.usdExchangeSet {
							return scm.contentRow(gtx, "Fee", scm.leftTransactionFeeValue+" "+scm.rightTransactionFeeValue, "")
						}
						return scm.contentRow(gtx, "Fee", scm.leftTransactionFeeValue, "")
					})
				}),
				layout.Rigid(func(gtx C) D {
					if scm.usdExchangeSet {
						return scm.contentRow(gtx, "Total cost", scm.leftTotalCostValue+" "+scm.rightTotalCostValue, "")
					}
					return scm.contentRow(gtx, "Total cost", scm.leftTotalCostValue, "")
				}),
			)
		},
		func(gtx C) D {
			return scm.passwordEditor.Layout(gtx)
		},
		func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					icon := scm.icons.actionInfo
					icon.Color = scm.theme.Color.Gray
					return layout.Inset{Right: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
						return icon.Layout(gtx, values.MarginPadding20)
					})
				}),
				layout.Rigid(func(gtx C) D {
					txt := scm.theme.Body2("Your DCR will be sent after this step.")
					txt.Color = scm.theme.Color.Gray3
					return txt.Layout(gtx)
				}),
			)
		},
		func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						inset := layout.Inset{
							Left: values.MarginPadding5,
						}
						return inset.Layout(gtx, func(gtx C) D {
							return scm.closeConfirmationModalButton.Layout(gtx)
						})
					}),
					layout.Rigid(func(gtx C) D {
						if false {
							th := material.NewTheme(gofont.Collection())
							return layout.Inset{Top: unit.Dp(7)}.Layout(gtx, func(gtx C) D {
								return material.Loader(th).Layout(gtx)
							})
						}
						scm.confirmButton.Text = fmt.Sprintf("Send %s", dcrutil.Amount(scm.totalCostDCR).String())
						return scm.confirmButton.Layout(gtx)
					}),
				)
			})
		},
	}

	return scm.modal.Layout(gtx, w, 900)
}

func (scm *sendConfirmModal) contentRow(gtx layout.Context, leftValue, rightValue, walletName string) layout.Dimensions {
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			txt := scm.theme.Body2(leftValue)
			txt.Color = scm.theme.Color.Gray
			return txt.Layout(gtx)
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(scm.theme.Body1(rightValue).Layout),
					layout.Rigid(func(gtx C) D {
						if walletName != "" {
							card := scm.theme.Card()
							card.Radius = decredmaterial.CornerRadius{NE: 0, NW: 0, SE: 0, SW: 0}
							card.Color = scm.theme.Color.LightGray
							inset := layout.Inset{
								Left: values.MarginPadding5,
							}
							return inset.Layout(gtx, func(gtx C) D {
								return card.Layout(gtx, func(gtx C) D {
									return layout.UniformInset(values.MarginPadding2).Layout(gtx, func(gtx C) D {
										txt := scm.theme.Caption(walletName)
										txt.Color = scm.theme.Color.Gray
										return txt.Layout(gtx)
									})
								})
							})
						}
						return layout.Dimensions{}
					}),
				)
			})
		}),
	)
}
