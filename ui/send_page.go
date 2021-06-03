package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/decred/dcrd/dcrutil"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const (
	PageSend               = "Send"
	invalidPassphraseError = "error broadcasting transaction: " + dcrlibwallet.ErrInvalidPassphrase
)

type sendPage struct {
	pageContainer layout.List
	common        pageCommon
	theme         *decredmaterial.Theme

	txAuthor *dcrlibwallet.TxAuthor

	exchangeRate float64

	destinationAddressEditor decredmaterial.Editor
	dcrAmountEditor          decredmaterial.Editor
	usdAmountEditor          decredmaterial.Editor
	passwordEditor           decredmaterial.Editor

	moreOption decredmaterial.IconButton

	nextButton                   decredmaterial.Button
	closeConfirmationModalButton decredmaterial.Button
	confirmButton                decredmaterial.Button
	maxButton                    decredmaterial.Button
	sendToButton                 decredmaterial.Button
	clearAllBtn                  decredmaterial.Button
	backButton                   decredmaterial.IconButton
	infoButton                   decredmaterial.IconButton

	accountSwitch    *decredmaterial.SwitchButtonText
	confirmModal     *decredmaterial.Modal
	txFeeCollapsible *decredmaterial.Collapsible

	sendToAddress              bool
	sourceAccountSelector      *accountSelector
	destinationAccountSelector *accountSelector
	sendMax                    bool

	txFee               string
	txFeeUSD            string
	estSignedSize       string
	totalCost           string
	totalCostUSD        string
	balanceAfterSend    string
	balanceAfterSendUSD string
	sendAmount          string
	sendAmountUSD       string

	exchangeErr string

	noExchangeErrMsg   string
	amountErrorText    string
	calculateErrorText string

	isConfirmationModalOpen bool
	isMoreOption            bool

	usdExchangeSet bool

	txAuthorErrChan  chan error
	broadcastErrChan chan error
}

func SendPage(common pageCommon) Page {
	pg := &sendPage{
		pageContainer: layout.List{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		},

		common: common,
		theme:  common.theme,

		exchangeRate: -1,

		noExchangeErrMsg:             "Exchange rate not fetched",
		closeConfirmationModalButton: common.theme.Button(new(widget.Clickable), "Cancel"),
		confirmButton:                common.theme.Button(new(widget.Clickable), ""),
		maxButton:                    common.theme.Button(new(widget.Clickable), "MAX"),
		clearAllBtn:                  common.theme.Button(new(widget.Clickable), "Clear all fields"),
		backButton:                   common.theme.PlainIconButton(new(widget.Clickable), common.icons.navigationArrowBack),
		infoButton:                   common.theme.PlainIconButton(new(widget.Clickable), common.icons.actionInfo),
		txFeeCollapsible:             common.theme.Collapsible(),

		confirmModal:            common.theme.Modal(),
		isConfirmationModalOpen: false,

		broadcastErrChan: make(chan error),
		txAuthorErrChan:  make(chan error),
	}
	pg.accountSwitch = common.theme.SwitchButtonText([]decredmaterial.SwitchItem{{Text: "Address"}, {Text: "My account"}})

	pg.nextButton = common.theme.Button(new(widget.Clickable), "Next")
	pg.nextButton.Background = pg.theme.Color.InactiveGray

	zeroInset := layout.UniformInset(values.MarginPadding0)
	pg.backButton.Color, pg.infoButton.Color = common.theme.Color.Gray3, common.theme.Color.Gray3

	m25 := values.MarginPadding25
	pg.backButton.Size, pg.infoButton.Size = m25, m25
	pg.backButton.Inset, pg.infoButton.Inset = zeroInset, zeroInset

	pg.dcrAmountEditor = common.theme.Editor(new(widget.Editor), "Amount (DCR)")
	pg.dcrAmountEditor.Editor.SetText("")
	pg.dcrAmountEditor.IsCustomButton = true
	pg.dcrAmountEditor.Editor.SingleLine = true
	pg.dcrAmountEditor.CustomButton.Background = common.theme.Color.Gray
	pg.dcrAmountEditor.CustomButton.Inset = layout.UniformInset(values.MarginPadding2)
	pg.dcrAmountEditor.CustomButton.Text = "Max"
	pg.dcrAmountEditor.CustomButton.CornerRadius = values.MarginPadding0

	pg.usdAmountEditor = common.theme.Editor(new(widget.Editor), "Amount (USD)")
	pg.usdAmountEditor.Editor.SetText("")
	pg.usdAmountEditor.IsCustomButton = false
	pg.usdAmountEditor.Editor.SingleLine = true

	pg.passwordEditor = common.theme.EditorPassword(new(widget.Editor), "Spending password")
	pg.passwordEditor.Editor.SetText("")
	pg.passwordEditor.Editor.SingleLine = true
	pg.passwordEditor.Editor.Submit = true

	pg.destinationAddressEditor = common.theme.Editor(new(widget.Editor), "Address")
	pg.destinationAddressEditor.Editor.SingleLine = true
	pg.destinationAddressEditor.Editor.SetText("")

	pg.closeConfirmationModalButton.Background = color.NRGBA{}
	pg.closeConfirmationModalButton.Color = common.theme.Color.Primary

	pg.moreOption = common.theme.PlainIconButton(new(widget.Clickable), common.icons.navMoreIcon)
	pg.moreOption.Color = common.theme.Color.Gray3
	pg.moreOption.Inset = layout.UniformInset(values.MarginPadding0)

	pg.maxButton.Background = common.theme.Color.Gray3
	pg.maxButton.Inset = layout.UniformInset(values.MarginPadding5)

	pg.sendToButton = common.theme.Button(new(widget.Clickable), "Send to account")
	pg.sendToButton.TextSize = values.TextSize14
	pg.sendToButton.Background = color.NRGBA{}
	pg.sendToButton.Color = common.theme.Color.Primary
	pg.sendToButton.Inset = layout.UniformInset(values.MarginPadding0)

	pg.clearAllBtn.Background = common.theme.Color.Surface
	pg.clearAllBtn.Color = common.theme.Color.Text
	pg.clearAllBtn.Inset = layout.UniformInset(values.MarginPadding15)

	currencyExchangeValue := common.multiWallet.ReadStringConfigValueForKey(dcrlibwallet.CurrencyConversionConfigKey)
	pg.usdExchangeSet = false
	if currencyExchangeValue == USDExchangeValue {
		pg.usdExchangeSet = true
	}

	// Source account picker
	pg.sourceAccountSelector = newAccountSelector(common).
		title("Sending account").
		accountSelected(func(selectedAccount *dcrlibwallet.Account) {
		}).
		accountValidator(func(account *dcrlibwallet.Account) bool {
			wal := pg.common.multiWallet.WalletWithID(account.WalletID)

			// Imported and watch only wallet accounts are invalid for sending
			accountIsValid := account.Number != MaxInt32 && !wal.IsWatchingOnlyWallet()

			if wal.ReadBoolConfigValueForKey(dcrlibwallet.AccountMixerConfigSet, false) {
				// privacy is enabled for selected wallet

				mixedAccountNumber := wal.ReadInt32ConfigValueForKey(dcrlibwallet.AccountMixerMixedAccount, -1)

				if pg.sendToAddress {
					// only mixed can send to address
					accountIsValid = account.Number == mixedAccountNumber
				} else {
					// send to account, check if selected destination account belongs to wallet
					destinationAccount := pg.destinationAccountSelector.selectedAccount
					if destinationAccount.WalletID != account.WalletID {
						accountIsValid = account.Number == mixedAccountNumber
					}
				}
			}
			return accountIsValid
		})

	// Destination account picker
	pg.destinationAccountSelector = newAccountSelector(common).
		title("Receiving account").
		accountSelected(func(selectedAccount *dcrlibwallet.Account) {

		}).
		accountValidator(func(account *dcrlibwallet.Account) bool {

			// Filter out imported account and mixed.
			wal := pg.common.multiWallet.WalletWithID(account.WalletID)
			mixedAccountNumber := wal.ReadInt32ConfigValueForKey(dcrlibwallet.AccountMixerMixedAccount, -1)
			if account.Number == MaxInt32 ||
				account.Number == mixedAccountNumber {
				return false
			}

			return true
		})

	if pg.usdExchangeSet {
		pg.fetchExchangeValue()
	}

	return pg
}

func (pg *sendPage) pageID() string {
	return PageSend
}

func (pg *sendPage) Layout(gtx layout.Context) layout.Dimensions {
	common := pg.common
	pageContent := []func(gtx C) D{
		func(gtx C) D {
			return pg.pageSections(gtx, "From", func(gtx C) D {
				return pg.sourceAccountSelector.Layout(gtx)
			})
		},
		func(gtx C) D {
			return pg.toSection(gtx, common)
		},
		func(gtx C) D {
			return pg.feeSection(gtx)
		},
	}

	dims := layout.Stack{Alignment: layout.S}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			return layout.Stack{Alignment: layout.NE}.Layout(gtx,
				layout.Expanded(func(gtx C) D {
					return common.UniformPadding(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								return layout.Inset{Bottom: values.MarginPadding16}.Layout(gtx, func(gtx C) D {
									return pg.topNav(gtx, common)
								})
							}),
							layout.Rigid(func(gtx C) D {
								return layout.Inset{Bottom: values.MarginPadding16}.Layout(gtx, func(gtx C) D {
									return pg.pageContainer.Layout(gtx, len(pageContent), func(gtx C, i int) D {
										return layout.Inset{Bottom: values.MarginPadding4, Top: values.MarginPadding4}.Layout(gtx, pageContent[i])
									})
								})
							}),
						)
					})
				}),
				layout.Stacked(func(gtx C) D {
					if pg.isMoreOption {
						inset := layout.Inset{
							Top:   values.MarginPadding40,
							Right: values.MarginPadding20,
						}
						return inset.Layout(gtx, func(gtx C) D {
							border := widget.Border{Color: pg.theme.Color.Background, CornerRadius: values.MarginPadding5, Width: values.MarginPadding1}
							return border.Layout(gtx, func(gtx C) D {
								return pg.clearAllBtn.Layout(gtx)
							})
						})
					}
					return layout.Dimensions{}
				}),
			)
		}),
		layout.Stacked(func(gtx C) D {
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			return layout.S.Layout(gtx, func(gtx C) D {
				return layout.Inset{Left: values.MarginPadding1}.Layout(gtx, func(gtx C) D {
					return pg.balanceSection(gtx, common)
				})
			})
		}),
	)

	if pg.isConfirmationModalOpen {
		return common.Modal(gtx, dims, pg.confirmationModal(gtx, common))
	}

	return dims
}

func (pg *sendPage) topNav(gtx layout.Context, common pageCommon) layout.Dimensions {
	m := values.MarginPadding20
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					pg.backButton.Icon = common.icons.contentClear
					return pg.backButton.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: m}.Layout(gtx, pg.theme.H6("Send DCR").Layout)
				}),
			)
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(pg.infoButton.Layout),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Left: m}.Layout(gtx, pg.moreOption.Layout)
					}),
				)
			})
		}),
	)
}

func (pg *sendPage) toSection(gtx layout.Context, common pageCommon) layout.Dimensions {
	return pg.pageSections(gtx, "To", func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Bottom: values.MarginPadding16,
				}.Layout(gtx, func(gtx C) D {
					if pg.sendToAddress {
						return pg.destinationAddressEditor.Layout(gtx)
					}

					return pg.destinationAccountSelector.Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx C) D {
				if pg.usdExchangeSet {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Flexed(0.45, func(gtx C) D {
							return pg.dcrAmountEditor.Layout(gtx)
						}),
						layout.Flexed(0.1, func(gtx C) D {
							return layout.Center.Layout(gtx, func(gtx C) D {
								return layout.Inset{Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
									return decredmaterial.Clickable(gtx, new(widget.Clickable), func(gtx C) D {
										//TODO nil clickable
										icon := common.icons.currencySwapIcon
										icon.Scale = 0.45
										return icon.Layout(gtx)
									})
								})
							})
						}),
						layout.Flexed(0.45, func(gtx C) D {
							return pg.usdAmountEditor.Layout(gtx)
						}),
					)
				}
				return pg.dcrAmountEditor.Layout(gtx)
			}),
		)
	})
}

func (pg *sendPage) feeSection(gtx layout.Context) layout.Dimensions {
	collapsibleHeader := func(gtx C) D {
		return layout.Flex{}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return pg.theme.Body1(pg.txFee).Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				b := pg.theme.Body1(pg.txFeeUSD) //todo
				b.Color = pg.theme.Color.Gray
				inset := layout.Inset{
					Left: values.MarginPadding5,
				}
				if pg.exchangeRate != -1 {
					return inset.Layout(gtx, b.Layout)
				}
				return layout.Dimensions{}
			}),
		)
	}

	collapsibleBody := func(gtx C) D {
		card := pg.theme.Card()
		card.Color = pg.theme.Color.LightGray
		inset := layout.Inset{
			Top: values.MarginPadding10,
		}
		return inset.Layout(gtx, func(gtx C) D {
			return card.Layout(gtx, func(gtx C) D {
				return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return pg.contentRow(gtx, "Estimated time", "10 mins(2 blocks)", "")
						}),
						layout.Rigid(func(gtx C) D {
							inset := layout.Inset{
								Top:    values.MarginPadding5,
								Bottom: values.MarginPadding5,
							}
							return inset.Layout(gtx, func(gtx C) D {
								return pg.contentRow(gtx, "Estimated size", pg.estSignedSize, "")
							})
						}),
						layout.Rigid(func(gtx C) D {
							return pg.contentRow(gtx, "Fee rate", "10 atoms/Byte", "")
						}),
					)
				})
			})
		})
	}
	inset := layout.Inset{
		Bottom: values.MarginPadding75,
	}
	return inset.Layout(gtx, func(gtx C) D {
		return pg.pageSections(gtx, "Fee", func(gtx C) D {
			return pg.txFeeCollapsible.Layout(gtx, collapsibleHeader, collapsibleBody)
		})
	})
}

func (pg *sendPage) balanceSection(gtx layout.Context, common pageCommon) layout.Dimensions {
	c := pg.theme.Card()
	c.Radius = decredmaterial.CornerRadius{NE: 0, NW: 0, SE: 0, SW: 0}
	return c.Layout(gtx, func(gtx C) D {
		return common.UniformPadding(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(0.6, func(gtx C) D {
					inset := layout.Inset{
						Right: values.MarginPadding15,
					}
					return inset.Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								inset := layout.Inset{
									Bottom: values.MarginPadding10,
								}
								return inset.Layout(gtx, func(gtx C) D {
									if pg.exchangeRate != -1 {
										return pg.contentRow(gtx, "Total cost", pg.totalCost+" "+pg.totalCostUSD, "")
									}
									return pg.contentRow(gtx, "Total cost", pg.totalCost, "")
								})
							}),
							layout.Rigid(func(gtx C) D {
								return pg.contentRow(gtx, "Balance after send", pg.balanceAfterSend, "")
							}),
						)
					})
				}),
				layout.Flexed(0.3, func(gtx C) D {
					pg.nextButton.Inset = layout.Inset{Top: values.MarginPadding15, Bottom: values.MarginPadding15}
					return pg.nextButton.Layout(gtx)
				}),
			)
		})
	})
}

func (pg *sendPage) confirmationModal(gtx layout.Context, common pageCommon) layout.Dimensions {
	receiveWallet := common.multiWallet.WalletWithID(pg.destinationAccountSelector.selectedAccount.WalletID)
	sendWallet := common.multiWallet.WalletWithID(pg.sourceAccountSelector.selectedAccount.WalletID)

	w := []layout.Widget{
		func(gtx C) D {
			return pg.theme.H6("Confim to send").Layout(gtx)
		},
		func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					icon := common.icons.sendIcon
					icon.Scale = 0.7
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return layout.Inset{Top: values.MarginPadding2, Right: values.MarginPadding16}.Layout(gtx, icon.Layout)
						}),
						layout.Rigid(func(gtx C) D {
							return layout.Flex{}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return common.layoutBalance(gtx, pg.sendAmount, true)
								}),
								layout.Flexed(1, func(gtx C) D {
									if pg.exchangeRate != -1 {
										return layout.E.Layout(gtx, func(gtx C) D {
											txt := pg.theme.Body1(pg.sendAmountUSD)
											txt.Color = pg.theme.Color.Gray
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
							icon := common.icons.navigationArrowForward
							icon.Color = pg.theme.Color.Gray3
							return layout.Inset{Right: values.MarginPadding8}.Layout(gtx, func(gtx C) D {
								return icon.Layout(gtx, values.MarginPadding15)
							})
						}),
						layout.Rigid(func(gtx C) D {
							// send to address
							if pg.sendToAddress {
								return pg.theme.Body2(pg.destinationAddressEditor.Editor.Text()).Layout(gtx)
							}

							// send to account
							return layout.E.Layout(gtx, func(gtx C) D {
								return layout.Flex{}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										return pg.theme.Body2(pg.destinationAccountSelector.selectedAccount.Name).Layout(gtx)
									}),
									layout.Rigid(func(gtx C) D {
										card := pg.theme.Card()
										card.Radius = decredmaterial.CornerRadius{NE: 0, NW: 0, SE: 0, SW: 0}
										card.Color = pg.theme.Color.LightGray
										inset := layout.Inset{
											Left: values.MarginPadding5,
										}
										return inset.Layout(gtx, func(gtx C) D {
											return card.Layout(gtx, func(gtx C) D {
												return layout.UniformInset(values.MarginPadding2).Layout(gtx, func(gtx C) D {
													txt := pg.theme.Caption(receiveWallet.Name)
													txt.Color = pg.theme.Color.Gray
													return txt.Layout(gtx)
												})
											})
										})
									}),
								)
							})

						}),
					)
				}),
			)
		},
		func(gtx C) D {
			return pg.theme.Separator().Layout(gtx)
		},
		func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return pg.contentRow(gtx, "Sending from", pg.sourceAccountSelector.selectedAccount.Name, sendWallet.Name)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: values.MarginPadding8, Bottom: values.MarginPadding8}.Layout(gtx, func(gtx C) D {
						if pg.usdExchangeSet {
							return pg.contentRow(gtx, "Fee", pg.txFee+" "+pg.txFeeUSD, "")
						}
						return pg.contentRow(gtx, "Fee", pg.txFee, "")
					})
				}),
				layout.Rigid(func(gtx C) D {
					if pg.exchangeRate != -1 {
						return pg.contentRow(gtx, "Total cost", pg.totalCost+" "+pg.totalCostUSD, "")
					}
					return pg.contentRow(gtx, "Total cost", pg.totalCost, "")
				}),
			)
		},
		func(gtx C) D {
			return pg.passwordEditor.Layout(gtx)
		},
		func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					icon := common.icons.actionInfo
					icon.Color = pg.theme.Color.Gray
					return layout.Inset{Right: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
						return icon.Layout(gtx, values.MarginPadding20)
					})
				}),
				layout.Rigid(func(gtx C) D {
					txt := pg.theme.Body2("Your DCR will be sent after this step.")
					txt.Color = pg.theme.Color.Gray3
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
							return pg.closeConfirmationModalButton.Layout(gtx)
						})
					}),
					layout.Rigid(func(gtx C) D {
						if common.modalLoad.loading {
							th := material.NewTheme(gofont.Collection())
							return layout.Inset{Top: unit.Dp(7)}.Layout(gtx, func(gtx C) D {
								return material.Loader(th).Layout(gtx)
							})
						}
						pg.confirmButton.Text = fmt.Sprintf("Send %s", pg.totalCost)
						return pg.confirmButton.Layout(gtx)
					}),
				)
			})
		},
	}

	return pg.confirmModal.Layout(gtx, w, 900)
}

func (pg *sendPage) pageSections(gtx layout.Context, title string, body layout.Widget) layout.Dimensions {
	return pg.theme.Card().Layout(gtx, func(gtx C) D {
		return layout.UniformInset(values.MarginPadding16).Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							inset := layout.Inset{
								Bottom: values.MarginPadding16,
							}
							return inset.Layout(gtx, pg.theme.Body1(title).Layout)
						}),
						layout.Flexed(1, func(gtx C) D {
							if title == "To" {
								return layout.E.Layout(gtx, func(gtx C) D {
									inset := layout.Inset{
										Top: values.MarginPaddingMinus5,
									}
									return inset.Layout(gtx, pg.accountSwitch.Layout)
								})
							}
							return layout.Dimensions{}
						}),
					)
				}),
				layout.Rigid(body),
			)
		})
	})
}

func (pg *sendPage) contentRow(gtx layout.Context, leftValue, rightValue, walletName string) layout.Dimensions {
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			txt := pg.theme.Body2(leftValue)
			txt.Color = pg.theme.Color.Gray
			return txt.Layout(gtx)
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(pg.theme.Body1(rightValue).Layout),
					layout.Rigid(func(gtx C) D {
						if walletName != "" {
							card := pg.theme.Card()
							card.Radius = decredmaterial.CornerRadius{NE: 0, NW: 0, SE: 0, SW: 0}
							card.Color = pg.theme.Color.LightGray
							inset := layout.Inset{
								Left: values.MarginPadding5,
							}
							return inset.Layout(gtx, func(gtx C) D {
								return card.Layout(gtx, func(gtx C) D {
									return layout.UniformInset(values.MarginPadding2).Layout(gtx, func(gtx C) D {
										txt := pg.theme.Caption(walletName)
										txt.Color = pg.theme.Color.Gray
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

func (pg *sendPage) constructTx() {
	address, err := pg.destinationAddress()
	if err != nil {
		pg.feeEstimationError(err.Error())
		return
	}

	amountAtom := int64(0)

	if !pg.sendMax {
		amount, err := strconv.ParseFloat(pg.dcrAmountEditor.Editor.Text(), 64)
		if err != nil {
			pg.feeEstimationError(err.Error())
			return
		}
		amountAtom = dcrlibwallet.AmountAtom(amount)
	}

	sourceAccount := pg.sourceAccountSelector.selectedAccount
	sourceWallet := pg.common.multiWallet.WalletWithID(sourceAccount.WalletID)

	unsignedTx := pg.common.multiWallet.NewUnsignedTx(sourceWallet, sourceAccount.Number)

	err = unsignedTx.AddSendDestination(address, amountAtom, pg.sendMax)
	if err != nil {
		pg.feeEstimationError(err.Error())
		return
	}

	feeAndSize, err := unsignedTx.EstimateFeeAndSize()
	if err != nil {
		pg.feeEstimationError(err.Error())
		return
	}

	feeAtom := feeAndSize.Fee.AtomValue
	if pg.sendMax {
		amountAtom = sourceAccount.Balance.Spendable - feeAtom
	}

	totalSendingAmount := dcrutil.Amount(amountAtom + feeAtom)
	balanceAfterSend := dcrutil.Amount(sourceAccount.Balance.Spendable - int64(totalSendingAmount))

	// populate display data
	pg.txFee = dcrutil.Amount(feeAtom).String()
	pg.estSignedSize = fmt.Sprintf("%d bytes", feeAndSize.EstimatedSignedSize)
	pg.totalCost = totalSendingAmount.String()
	pg.balanceAfterSend = balanceAfterSend.String()
	pg.sendAmount = dcrutil.Amount(amountAtom).String()

	if pg.sendMax {
		pg.dcrAmountEditor.Editor.SetText(fmt.Sprintf("%.8f", dcrutil.Amount(amountAtom).ToCoin()))
	}

	if pg.exchangeRate != -1 {
		pg.txFeeUSD = fmt.Sprintf("$%.4f", dcrTOUSD(pg.exchangeRate, feeAndSize.Fee.DcrValue))
		pg.totalCostUSD = formatUSDBalance(pg.common.printer, dcrTOUSD(pg.exchangeRate, totalSendingAmount.ToCoin()))
		pg.balanceAfterSendUSD = formatUSDBalance(pg.common.printer, dcrTOUSD(pg.exchangeRate, balanceAfterSend.ToCoin()))
		pg.sendAmountUSD = formatUSDBalance(pg.common.printer, dcrTOUSD(pg.exchangeRate, dcrutil.Amount(amountAtom).ToCoin()))

		if pg.sendMax {
			pg.usdAmountEditor.Editor.SetText(formatUSDBalance(pg.common.printer, dcrutil.Amount(amountAtom).ToCoin()))
		}
	}

	pg.txAuthor = unsignedTx
}

func (pg *sendPage) validateAndConstructTx() {
	if pg.validate() {
		pg.constructTx()
	} else {
		pg.clearEstimates()
	}
}

func (pg *sendPage) validate() bool {

	_, err := strconv.ParseFloat(pg.dcrAmountEditor.Editor.Text(), 64)
	amountIsValid := err == nil
	addressIsValid := true // default send to account
	if pg.sendToAddress {
		addressIsValid, _ = pg.validateDestinationAddress()
	}

	validForSending := (amountIsValid || pg.sendMax) && addressIsValid

	if validForSending {
		pg.nextButton.Background = pg.theme.Color.Primary
	} else {
		pg.nextButton.Background = pg.theme.Color.Hint
	}

	return validForSending
}

func (pg *sendPage) destinationAddress() (string, error) {
	if pg.sendToAddress {
		valid, address := pg.validateDestinationAddress()
		if valid {
			return address, nil
		}

		return "", fmt.Errorf("invalid address")
	}

	destinationAccount := pg.destinationAccountSelector.selectedAccount
	wal := pg.common.multiWallet.WalletWithID(destinationAccount.WalletID)

	return wal.CurrentAddress(destinationAccount.Number)
}

func (pg *sendPage) validateDestinationAddress() (bool, string) {

	address := pg.destinationAddressEditor.Editor.Text()
	address = strings.TrimSpace(address)

	if len(address) == 0 {
		pg.destinationAddressEditor.SetError("")
		return false, address
	}

	if pg.common.multiWallet.IsAddressValid(address) {
		pg.destinationAddressEditor.SetError("")
		return true, address
	}

	pg.destinationAddressEditor.SetError("Invalid address")
	return false, address
}

func (pg *sendPage) validateDCRAmount() bool {
	pg.amountErrorText = ""
	if pg.inputsNotEmpty(pg.dcrAmountEditor.Editor) {
		dcrAmount, err := strconv.ParseFloat(pg.dcrAmountEditor.Editor.Text(), 64)
		if err != nil {
			pg.dcrAmountEditor.SetError("Invalid amount")
			// empty usd input
			pg.usdAmountEditor.Editor.SetText("")
			return false
		}

		if pg.exchangeRate != -1 {
			usdAmount := dcrTOUSD(pg.exchangeRate, dcrAmount)
			pg.usdAmountEditor.Editor.SetText(fmt.Sprintf("%.2f", usdAmount)) // 2 decimal places
		}

		pg.dcrAmountEditor.SetError("")
		return true
	}

	// empty usd input
	pg.usdAmountEditor.Editor.SetText("")
	pg.dcrAmountEditor.SetError("")
	return false
}

// validateUSDAmount is called when usd text changes
func (pg *sendPage) validateUSDAmount() bool {
	pg.amountErrorText = ""
	if pg.inputsNotEmpty(pg.usdAmountEditor.Editor) {
		usdAmount, err := strconv.ParseFloat(pg.usdAmountEditor.Editor.Text(), 64)
		if err != nil {
			// empty dcr input
			pg.dcrAmountEditor.Editor.SetText("")
			pg.usdAmountEditor.SetError("Invalid amount")
			return false
		}
		pg.usdAmountEditor.SetError("")

		if pg.exchangeRate != -1 { //TODO usd amount should not be visible.
			dcrAmount := usdToDCR(pg.exchangeRate, usdAmount)
			pg.dcrAmountEditor.Editor.SetText(fmt.Sprintf("%.8f", dcrAmount)) // 8 decimal places
		}

		return true
	}

	// empty dcr input
	pg.dcrAmountEditor.Editor.SetText("")
	pg.usdAmountEditor.SetError("")
	return false
}

func (pg *sendPage) inputsNotEmpty(editors ...*widget.Editor) bool {
	for _, e := range editors {
		if e.Text() == "" {
			return false
		}
	}
	return true
}

func (pg *sendPage) updateExchangeError(err error) {
	pg.usdAmountEditor.SetError(err.Error())
}

func (pg *sendPage) feeEstimationError(err string) {
	if err == dcrlibwallet.ErrInsufficientBalance {
		pg.amountErrorText = "Not enough funds"
		return
	}
	if strings.Contains(err, "invalid amount") {
		pg.amountErrorText = "Invalid amount"
		return
	}
	pg.calculateErrorText = fmt.Sprintf("error estimating transaction: %s", err)
}

func (pg *sendPage) handleEditorChange(evt widget.EditorEvent, c pageCommon) {
	switch evt.(type) {
	case widget.ChangeEvent:
		pg.validateAndConstructTx()
	case widget.SubmitEvent:
		pg.sendFund(c)
	}
}

func (pg *sendPage) clearEstimates() {
	pg.txFee = " - "
	pg.txFeeUSD = " - "
	pg.estSignedSize = " - "
	pg.totalCost = " - "
	pg.totalCostUSD = " - "
	pg.balanceAfterSend = " - "
	pg.balanceAfterSendUSD = " - "
	pg.sendAmount = " - "
	pg.sendAmountUSD = " - "
}

func (pg *sendPage) resetFields() {
	pg.destinationAddressEditor.SetError("")
	pg.dcrAmountEditor.Editor.SetText("")
	pg.usdAmountEditor.Editor.SetText("")
	pg.passwordEditor.Editor.SetText("")
	pg.amountErrorText = ""
	pg.calculateErrorText = ""
}

func (pg *sendPage) resetErrorText() {
	pg.amountErrorText = ""
	pg.calculateErrorText = ""
	pg.destinationAddressEditor.SetError("")
	pg.dcrAmountEditor.SetError("")
	pg.usdAmountEditor.SetError("")
	pg.passwordEditor.SetError("")
}

func (pg *sendPage) fetchExchangeValue() {
	go func() {
		var dcrUsdtBittrex DCRUSDTBittrex
		err := GetUSDExchangeValues(&dcrUsdtBittrex)
		if err != nil {
			pg.updateExchangeError(err)
			return
		}

		exchangeRate, err := strconv.ParseFloat(dcrUsdtBittrex.LastTradeRate, 64)
		if err != nil {
			pg.updateExchangeError(err)
			return
		}
		pg.exchangeRate = exchangeRate
	}()
}

func (pg *sendPage) sendFund(c pageCommon) {
	if !pg.inputsNotEmpty(pg.passwordEditor.Editor) && pg.txAuthor != nil {
		return
	}
	c.modalLoad.setLoading(true)
	go func() {
		password := pg.passwordEditor.Editor.Text()
		_, err := pg.txAuthor.Broadcast([]byte(password))
		c.modalLoad.setLoading(false)
		if err != nil {
			if err.Error() == dcrlibwallet.ErrInvalidPassphrase {
				pg.passwordEditor.SetError("Wrong password")
			} else {
				c.notify(err.Error(), false)
				pg.isConfirmationModalOpen = false
			}

			return
		}

		pg.isConfirmationModalOpen = false
		pg.resetFields()

		c.popPage()
		c.notify("1 Transaction Sent", true)
	}()
}

func (pg *sendPage) handle() {

	// initialize destinationAccountSelector first as the selected account value
	// is required by sourceAccountSelector
	pg.destinationAccountSelector.handle()
	pg.sourceAccountSelector.handle()

	c := pg.common

	if pg.exchangeErr != "" {
		c.notify(pg.exchangeErr, false)
	}

	sendToAddress := pg.accountSwitch.SelectedIndex() == 1
	if sendToAddress != pg.sendToAddress { // switch changed
		pg.sendToAddress = sendToAddress
		pg.validateAndConstructTx()
	}

	pg.sendToAddress = pg.accountSwitch.SelectedIndex() == 1

	if pg.backButton.Button.Clicked() {
		pg.resetErrorText()
		pg.resetFields()
		c.popPage()
	}

	if pg.infoButton.Button.Clicked() {
		go func() {
			c.modalReceiver <- &modalLoad{
				template:   SendInfoTemplate,
				title:      "Send DCR",
				cancel:     c.closeModal,
				cancelText: "Got it",
			}
		}()
	}

	for pg.moreOption.Button.Clicked() {
		pg.isMoreOption = !pg.isMoreOption
	}

	for _, evt := range pg.destinationAddressEditor.Editor.Events() {
		if pg.destinationAddressEditor.Editor.Focused() {
			switch evt.(type) {
			case widget.ChangeEvent:
				pg.sendMax = false
				pg.validateAndConstructTx()
			}
		}
	}

	for _, evt := range pg.dcrAmountEditor.Editor.Events() {
		if pg.dcrAmountEditor.Editor.Focused() {
			switch evt.(type) {
			case widget.ChangeEvent:
				pg.sendMax = false
				pg.validateDCRAmount()
				pg.validateAndConstructTx()

			}
		}
	}

	for _, evt := range pg.usdAmountEditor.Editor.Events() {
		if pg.usdAmountEditor.Editor.Focused() {
			switch evt.(type) {
			case widget.ChangeEvent:
				pg.sendMax = false
				pg.validateUSDAmount()
				pg.validateAndConstructTx()

			}
		}
	}

	for _, evt := range pg.passwordEditor.Editor.Events() {
		if pg.passwordEditor.Editor.Focused() {
			//TODO
			pg.handleEditorChange(evt, c)
		}
	}

	if pg.calculateErrorText != "" {
		pg.dcrAmountEditor.LineColor, pg.dcrAmountEditor.TitleLabelColor = pg.theme.Color.Danger, pg.theme.Color.Danger
		pg.usdAmountEditor.LineColor, pg.usdAmountEditor.TitleLabelColor = pg.theme.Color.Danger, pg.theme.Color.Danger
		c.notify(pg.calculateErrorText, false)
	} else {
		pg.dcrAmountEditor.LineColor, pg.dcrAmountEditor.TitleLabelColor = pg.theme.Color.Gray1, pg.theme.Color.Gray3
		pg.usdAmountEditor.LineColor, pg.usdAmountEditor.TitleLabelColor = pg.theme.Color.Gray1, pg.theme.Color.Gray3
	}

	if pg.amountErrorText != "" {
		pg.dcrAmountEditor.SetError(pg.amountErrorText)
	}

	for pg.confirmButton.Button.Clicked() {
		pg.sendFund(c)
	}

	for pg.nextButton.Button.Clicked() {
		if pg.validate() && pg.calculateErrorText == "" {
			pg.isConfirmationModalOpen = true
			pg.passwordEditor.Editor.Focus()
		}
	}

	if pg.isConfirmationModalOpen {
		pg.confirmButton.Background = pg.theme.Color.Primary
		if !pg.inputsNotEmpty(pg.passwordEditor.Editor) {
			pg.confirmButton.Background = pg.theme.Color.InactiveGray
		}
	}

	for pg.closeConfirmationModalButton.Button.Clicked() {
		c.modalLoad.setLoading(false)
		pg.isConfirmationModalOpen = false
	}

	for pg.clearAllBtn.Button.Clicked() {
		pg.resetFields()
	}

	if pg.dcrAmountEditor.CustomButton.Button.Clicked() {
		// (bug) this would not work if the amount input is focussed
		pg.sendMax = true
		pg.validateAndConstructTx()
	}
	if pg.usdAmountEditor.CustomButton.Button.Clicked() {
		pg.usdAmountEditor.Editor.Focus()
		// send max
	}
}

func (pg *sendPage) onClose() {}
