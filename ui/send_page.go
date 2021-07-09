package ui

import (
	"fmt"
	"image/color"
	"reflect"
	"strconv"
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/decred/dcrd/dcrutil"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/page"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"
)

const (
	PageSend               = "Send"
	invalidPassphraseError = "error broadcasting transaction: " + dcrlibwallet.ErrInvalidPassphrase
)

type amountValue struct {
	sendAmountDCR            string
	sendAmountUSD            string
	leftTransactionFeeValue  string
	rightTransactionFeeValue string
	leftTotalCostValue       string
	rightTotalCostValue      string
}

type sendPage struct {
	pageContainer layout.List
	common        *pageCommon
	theme         *decredmaterial.Theme

	txAuthor        *dcrlibwallet.TxAuthor
	broadcastResult *wallet.Broadcast
	wallet          *wallet.Wallet

	destinationAddressEditor decredmaterial.Editor
	leftAmountEditor         decredmaterial.Editor
	rightAmountEditor        decredmaterial.Editor

	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton
	moreOption decredmaterial.IconButton

	nextButton   decredmaterial.Button
	maxButton    decredmaterial.Button
	sendToButton decredmaterial.Button
	clearAllBtn  decredmaterial.Button

	accountSwitch    *decredmaterial.SwitchButtonText
	confirmModal     *decredmaterial.Modal
	txFeeCollapsible *decredmaterial.Collapsible
	currencySwap     *widget.Clickable

	remainingBalance int64
	amountAtoms      int64
	txFee            int64
	spendableBalance int64

	count int

	*comfirmModalData
	confirmTxModal *sendConfirmModal

	usdExchangeRate float64
	inputAmount     float64
	amountUSDtoDCR  float64
	amountDCRtoUSD  float64

	txFeeSize          string
	txFeeEstimatedTime string

	leftExchangeValue  string
	rightExchangeValue string

	balanceAfterSendValue string
	activeTotalAmount     string

	LastTradeRate    string
	exchangeErr      string
	noExchangeErrMsg string

	amountErrorText    string
	calculateErrorText string

	isBroadcastingTransaction bool
	isMoreOption              bool

	shouldInitializeTxAuthor bool

	txAuthorErrChan  chan error
	broadcastErrChan chan error
}

// shared between send page and confirm modal
type comfirmModalData struct {
	closeConfirmationModalButton decredmaterial.Button
	confirmButton                decredmaterial.Button
	passwordEditor               decredmaterial.Editor

	sourceAccountSelector      *accountSelector
	destinationAccountSelector *accountSelector

	totalCostDCR int64

	sendAmountDCR string
	sendAmountUSD string

	leftTransactionFeeValue  string
	rightTransactionFeeValue string
	leftTotalCostValue       string
	rightTotalCostValue      string

	sendToOption   string
	usdExchangeSet bool

	// others
	destinationAddress string //pg.destinationAddressEditor.Editor.Text()
}

func SendPage(common *pageCommon) Page {
	pg := &sendPage{
		pageContainer: layout.List{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		},

		common:          common,
		theme:           common.theme,
		wallet:          common.wallet,
		txAuthor:        common.txAuthor,
		broadcastResult: common.broadcastResult,

		currencySwap: new(widget.Clickable),

		leftExchangeValue:  "DCR",
		rightExchangeValue: "USD",
		noExchangeErrMsg:   "Exchange rate not fetched",
		maxButton:          common.theme.Button(new(widget.Clickable), "MAX"),
		clearAllBtn:        common.theme.Button(new(widget.Clickable), "Clear all fields"),
		txFeeCollapsible:   common.theme.Collapsible(),

		confirmModal:              common.theme.Modal(),
		isBroadcastingTransaction: false,

		broadcastErrChan: make(chan error),
		txAuthorErrChan:  make(chan error),
		comfirmModalData: &comfirmModalData{},
	}

	pg.closeConfirmationModalButton = common.theme.Button(new(widget.Clickable), "Cancel")
	pg.confirmButton = common.theme.Button(new(widget.Clickable), "")

	pg.confirmTxModal = newSendConfirmModal(common, pg.comfirmModalData)

	pg.accountSwitch = common.theme.SwitchButtonText([]decredmaterial.SwitchItem{{Text: "Address"}, {Text: "My account"}})

	pg.balanceAfterSendValue = "- DCR"

	pg.nextButton = common.theme.Button(new(widget.Clickable), "Next")
	pg.nextButton.Background = pg.theme.Color.InactiveGray

	activeEditorHint := fmt.Sprintf("Amount (%s)", pg.leftExchangeValue)
	pg.leftAmountEditor = common.theme.Editor(new(widget.Editor), activeEditorHint)
	pg.leftAmountEditor.Editor.SetText("")
	pg.leftAmountEditor.IsCustomButton = true
	pg.leftAmountEditor.Editor.SingleLine = true
	pg.leftAmountEditor.CustomButton.Background = common.theme.Color.Gray
	pg.leftAmountEditor.CustomButton.Inset = layout.UniformInset(values.MarginPadding2)
	pg.leftAmountEditor.CustomButton.Text = "Max"
	pg.leftAmountEditor.CustomButton.CornerRadius = values.MarginPadding0

	inactiveEditorHint := fmt.Sprintf("Amount (%s)", pg.rightExchangeValue)
	pg.rightAmountEditor = common.theme.Editor(new(widget.Editor), inactiveEditorHint)
	pg.rightAmountEditor.Editor.SetText("")
	pg.rightAmountEditor.IsCustomButton = true
	pg.rightAmountEditor.Editor.SingleLine = true
	pg.rightAmountEditor.CustomButton.Background = common.theme.Color.Gray
	pg.rightAmountEditor.CustomButton.Inset = layout.UniformInset(values.MarginPadding2)
	pg.rightAmountEditor.CustomButton.Text = "Max"
	pg.rightAmountEditor.CustomButton.CornerRadius = values.MarginPadding0

	pg.passwordEditor = common.theme.EditorPassword(new(widget.Editor), "Spending password")
	pg.passwordEditor.Editor.SetText("")
	pg.passwordEditor.Editor.SingleLine = true
	pg.passwordEditor.Editor.Submit = true

	pg.destinationAddressEditor = common.theme.Editor(new(widget.Editor), "Address")
	pg.destinationAddressEditor.Editor.SingleLine = true
	pg.destinationAddressEditor.Editor.SetText("")

	pg.closeConfirmationModalButton.Background = color.NRGBA{}
	pg.closeConfirmationModalButton.Color = common.theme.Color.Primary

	pg.backButton, pg.infoButton = common.SubPageHeaderButtons()
	pg.backButton.Icon = common.icons.contentClear

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

	// Source account picker
	pg.sourceAccountSelector = newAccountSelector(common).
		title("Sending account").
		accountSelected(func(selectedAccount *dcrlibwallet.Account) {
			pg.shouldInitializeTxAuthor = true
		}).
		accountValidator(func(account *dcrlibwallet.Account) bool {
			wal := pg.common.multiWallet.WalletWithID(account.WalletID)

			// Imported and watch only wallet accounts are invalid for sending
			accountIsValid := account.Number != MaxInt32 && !wal.IsWatchingOnlyWallet()

			if wal.ReadBoolConfigValueForKey(dcrlibwallet.AccountMixerConfigSet, false) {
				// privacy is enabled for selected wallet

				if pg.sendToOption == "Address" { //Todo
					// only mixed can send to address
					accountIsValid = account.Number == wal.MixedAccountNumber()
				} else {
					// send to account, check if selected destination account belongs to wallet
					destinationAccount := pg.destinationAccountSelector.selectedAccount
					if destinationAccount.WalletID != account.WalletID {
						accountIsValid = account.Number == wal.MixedAccountNumber()
					}
				}
			}
			return accountIsValid
		})

	// Destination account picker
	pg.destinationAccountSelector = newAccountSelector(common).
		title("Receiving account").
		accountSelected(func(selectedAccount *dcrlibwallet.Account) {
			pg.shouldInitializeTxAuthor = true
			pg.sourceAccountSelector.selectFirstWalletValidAccount() // refresh source account
		}).
		accountValidator(func(account *dcrlibwallet.Account) bool {

			// Filter out imported account and mixed.
			wal := pg.common.multiWallet.WalletWithID(account.WalletID)
			if account.Number == MaxInt32 ||
				account.Number == wal.MixedAccountNumber() {
				return false
			}

			return true
		})

	return pg
}

func (pg *sendPage) OnResume() {
	pg.destinationAccountSelector.selectFirstWalletValidAccount()
	pg.sourceAccountSelector.selectFirstWalletValidAccount()

	pg.fetchExchangeValue()
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
									return pg.topNav(gtx)
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
							return border.Layout(gtx, pg.clearAllBtn.Layout)
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

	return dims
}

func (pg *sendPage) topNav(gtx layout.Context) layout.Dimensions {
	m := values.MarginPadding20
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
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

func (pg *sendPage) toSection(gtx layout.Context, common *pageCommon) layout.Dimensions {
	return pg.pageSections(gtx, "To", func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Bottom: values.MarginPadding16,
				}.Layout(gtx, func(gtx C) D {
					if pg.sendToOption == "My account" {
						return pg.destinationAccountSelector.Layout(gtx)
					}
					return pg.destinationAddressEditor.Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx C) D {
				if pg.usdExchangeSet {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Flexed(0.45, func(gtx C) D {
							pg.leftAmountEditor.Hint = fmt.Sprintf("Amount (%s)", pg.leftExchangeValue)
							return pg.leftAmountEditor.Layout(gtx)
						}),
						layout.Flexed(0.1, func(gtx C) D {
							return layout.Center.Layout(gtx, func(gtx C) D {
								return layout.Inset{Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
									return decredmaterial.Clickable(gtx, pg.currencySwap, func(gtx C) D {
										icon := common.icons.currencySwapIcon
										icon.Scale = 0.45
										return icon.Layout(gtx)
									})
								})
							})
						}),
						layout.Flexed(0.45, func(gtx C) D {
							pg.rightAmountEditor.Hint = fmt.Sprintf("Amount (%s)", pg.rightExchangeValue)
							return pg.rightAmountEditor.Layout(gtx)
						}),
					)
				}
				return pg.leftAmountEditor.Layout(gtx)
			}),
		)
	})
}

func (pg *sendPage) feeSection(gtx layout.Context) layout.Dimensions {
	collapsibleHeader := func(gtx C) D {
		return layout.Flex{}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return pg.theme.Body1(pg.leftTransactionFeeValue).Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				b := pg.theme.Body1(pg.confirmTxModal.rightTransactionFeeValue)
				b.Color = pg.theme.Color.Gray
				inset := layout.Inset{
					Left: values.MarginPadding5,
				}
				if pg.usdExchangeSet {
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
							return pg.contentRow(gtx, "Estimated time", "-")
						}),
						layout.Rigid(func(gtx C) D {
							inset := layout.Inset{
								Top:    values.MarginPadding5,
								Bottom: values.MarginPadding5,
							}
							return inset.Layout(gtx, func(gtx C) D {
								return pg.contentRow(gtx, "Estimated size", pg.txFeeSize)
							})
						}),
						layout.Rigid(func(gtx C) D {
							return pg.contentRow(gtx, "Fee rate", "10 atoms/Byte")
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

func (pg *sendPage) balanceSection(gtx layout.Context, common *pageCommon) layout.Dimensions {
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
									if pg.usdExchangeSet {
										return pg.contentRow(gtx, "Total cost", pg.confirmTxModal.leftTotalCostValue+" "+pg.confirmTxModal.rightTotalCostValue)
									}
									return pg.contentRow(gtx, "Total cost", pg.leftTotalCostValue)
								})
							}),
							layout.Rigid(func(gtx C) D {
								return pg.contentRow(gtx, "Balance after send", pg.balanceAfterSendValue)
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

func (pg *sendPage) contentRow(gtx layout.Context, leftValue, rightValue string) layout.Dimensions {
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
						return layout.Dimensions{}
					}),
				)
			})
		}),
	)
}

func (pg *sendPage) validate() bool {
	if pg.sendToOption == "Address" {
		isAmountValid := pg.validateLeftAmount()
		if pg.rightAmountEditor.Editor.Focused() {
			isAmountValid = pg.validateRightAmount()
		}

		if pg.usdExchangeSet && !isAmountValid {
			if pg.rightAmountEditor.Editor.Focused() {
				pg.leftAmountEditor.Editor.SetText("")
			} else {
				pg.rightAmountEditor.Editor.SetText("")
			}
		}

		if !pg.validateDestinationAddress() {
			pg.nextButton.Background = pg.theme.Color.Hint
			return false
		}

		if !isAmountValid {
			pg.nextButton.Background = pg.theme.Color.Hint
			return false
		}

		if pg.calculateErrorText != "" {
			pg.nextButton.Background = pg.theme.Color.Hint
			return false
		}
	}

	pg.nextButton.Background = pg.theme.Color.Primary
	return true
}

func (pg *sendPage) validateDestinationAddress() bool {
	if !pg.inputsNotEmpty(pg.destinationAddressEditor.Editor) {
		if pg.destinationAddressEditor.Editor.Focused() {
			pg.destinationAddressEditor.SetError("Input address")
			return false
		}
	}

	if pg.inputsNotEmpty(pg.destinationAddressEditor.Editor) {
		isValid, _ := pg.wallet.IsAddressValid(pg.destinationAddressEditor.Editor.Text())
		if !isValid {
			pg.destinationAddressEditor.SetError("Invalid address")
			return false
		}

		pg.destinationAddressEditor.SetError("")
		return true
	}

	pg.balanceAfterSend(true)
	pg.destinationAddressEditor.SetError("Input address")
	return false
}

func (pg *sendPage) validateLeftAmount() bool {
	if pg.inputsNotEmpty(pg.leftAmountEditor.Editor) {
		_, err := strconv.ParseFloat(pg.leftAmountEditor.Editor.Text(), 64)
		if err != nil {
			pg.leftAmountEditor.SetError("Invalid amount")
			return false
		}
		pg.leftAmountEditor.SetError("")
		return true
	}
	pg.leftAmountEditor.SetError("")
	return false
}

func (pg *sendPage) validateRightAmount() bool {
	if pg.inputsNotEmpty(pg.rightAmountEditor.Editor) {
		_, err := strconv.ParseFloat(pg.rightAmountEditor.Editor.Text(), 64)
		if err != nil {
			pg.rightAmountEditor.SetError("Invalid amount")
			return false
		}
		pg.rightAmountEditor.SetError("")
		return true
	}
	pg.rightAmountEditor.SetError("")
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

func (pg *sendPage) calculateValues(isUpdateAmountInput bool) {
	defaultLeftValues := fmt.Sprintf("- %s", "DCR")
	defaultRightValues := "($ -)"

	pg.leftTransactionFeeValue = defaultLeftValues
	pg.rightTransactionFeeValue = defaultRightValues

	pg.leftTotalCostValue = defaultLeftValues
	pg.rightTotalCostValue = defaultRightValues
	pg.calculateErrorText = ""
	pg.txFeeSize = "-"
	pg.txFeeEstimatedTime = "-"
	pg.sendAmountDCR = defaultLeftValues
	pg.confirmTxModal.sendAmountUSD = defaultRightValues

	if reflect.DeepEqual(pg.txAuthor, &dcrlibwallet.TxAuthor{}) || !pg.validate() {
		return
	}

	pg.inputAmount, _ = strconv.ParseFloat(pg.leftAmountEditor.Editor.Text(), 64)
	if pg.usdExchangeSet && pg.rightAmountEditor.Editor.Focused() {
		pg.inputAmount, _ = strconv.ParseFloat(pg.rightAmountEditor.Editor.Text(), 64)
	}

	if pg.usdExchangeSet && pg.LastTradeRate != "" {
		pg.usdExchangeRate, _ = strconv.ParseFloat(pg.LastTradeRate, 64)
		pg.amountUSDtoDCR = pg.inputAmount / pg.usdExchangeRate
		pg.amountDCRtoUSD = pg.inputAmount * pg.usdExchangeRate
	}

	pg.updateAmountInputsValues(isUpdateAmountInput)
	pg.getTxFee()
	pg.updateDefaultValues()
	pg.balanceAfterSend(false)
}

func (pg *sendPage) updateAmountInputsValues(isUpdateAmountInput bool) {
	switch {
	case pg.leftExchangeValue == "USD" && pg.LastTradeRate != "" && pg.leftAmountEditor.Editor.Focused():
		pg.rightAmountEditor.Editor.SetText(fmt.Sprintf("%f", pg.amountUSDtoDCR))
		pg.setDestinationAddr(pg.amountUSDtoDCR)
	case pg.leftExchangeValue == "USD" && pg.LastTradeRate != "" && pg.rightAmountEditor.Editor.Focused():
		pg.leftAmountEditor.Editor.SetText(fmt.Sprintf("%f", pg.amountDCRtoUSD))
		pg.setDestinationAddr(pg.inputAmount)
	case pg.leftExchangeValue == "DCR" && pg.LastTradeRate != "" && pg.rightAmountEditor.Editor.Focused():
		pg.leftAmountEditor.Editor.SetText(fmt.Sprintf("%f", pg.amountUSDtoDCR))
		pg.setDestinationAddr(pg.amountUSDtoDCR)
	case pg.leftExchangeValue == "DCR" && pg.LastTradeRate != "" && pg.leftAmountEditor.Editor.Focused():
		pg.rightAmountEditor.Editor.SetText(fmt.Sprintf("%f", pg.amountDCRtoUSD))
		pg.setDestinationAddr(pg.inputAmount)
	default:
		if isUpdateAmountInput {
			if pg.rightAmountEditor.Editor.Focused() {
				pg.leftAmountEditor.Editor.SetText(pg.rightAmountEditor.Editor.Text())
			} else {
				pg.rightAmountEditor.Editor.SetText(pg.leftAmountEditor.Editor.Text())
			}
		}
		pg.setDestinationAddr(pg.inputAmount)
	}
}

func (pg *sendPage) updateExchangeError() {
	pg.rightAmountEditor.SetError("")
	if pg.LastTradeRate == "" && pg.usdExchangeSet {
		pg.rightAmountEditor.SetError(pg.noExchangeErrMsg)
	}
}

func (pg *sendPage) setDestinationAddr(sendAmount float64) {
	pg.amountErrorText = ""
	amount, err := dcrutil.NewAmount(sendAmount)
	if err != nil {
		pg.feeEstimationError(err.Error(), "amount")
		return
	}

	pg.amountAtoms = int64(amount)
	if pg.amountAtoms == 0 {
		return
	}

	pg.txAuthor.RemoveSendDestination(0)
	addr := pg.destinationAddressEditor.Editor.Text()
	if pg.sendToOption == "My account" {
		selectedAccount := pg.destinationAccountSelector.selectedAccount
		wal := pg.common.multiWallet.WalletWithID(selectedAccount.WalletID)
		address, err := wal.CurrentAddress(selectedAccount.Number)
		if err != nil {
			pg.feeEstimationError(err.Error(), "destination address")
		} else {
			addr = address
		}
	}
	pg.txAuthor.AddSendDestination(addr, pg.amountAtoms, false)
}

func (pg *sendPage) amountValues() amountValue {
	pg.confirmTxModal.totalCostDCR = pg.txFee + pg.amountAtoms
	txFeeValueUSD := dcrutil.Amount(pg.txFee).ToCoin() * pg.usdExchangeRate
	switch {
	case pg.leftExchangeValue == "USD" && pg.LastTradeRate != "":
		return amountValue{
			sendAmountDCR:            dcrutil.Amount(pg.amountAtoms).String(),
			sendAmountUSD:            fmt.Sprintf("$ %f", dcrutil.Amount(pg.amountAtoms).ToCoin()*pg.usdExchangeRate),
			leftTransactionFeeValue:  fmt.Sprintf("%f USD", txFeeValueUSD),
			rightTransactionFeeValue: fmt.Sprintf("(%s)", dcrutil.Amount(pg.txFee).String()),
			leftTotalCostValue:       fmt.Sprintf("%s USD", strconv.FormatFloat(pg.inputAmount+txFeeValueUSD, 'f', 7, 64)),
			rightTotalCostValue:      fmt.Sprintf("(%s )", dcrutil.Amount(pg.totalCostDCR).String()),
		}
	case pg.leftExchangeValue == "DCR" && pg.LastTradeRate != "":
		return amountValue{
			sendAmountDCR:            dcrutil.Amount(pg.amountAtoms).String(),
			sendAmountUSD:            fmt.Sprintf("$ %s", strconv.FormatFloat(pg.amountDCRtoUSD, 'f', 2, 64)),
			leftTransactionFeeValue:  dcrutil.Amount(pg.txFee).String(),
			rightTransactionFeeValue: fmt.Sprintf("($ %s)", strconv.FormatFloat(txFeeValueUSD, 'f', 2, 64)),
			leftTotalCostValue:       dcrutil.Amount(pg.totalCostDCR).String(),
			rightTotalCostValue:      fmt.Sprintf("($ %s)", strconv.FormatFloat(pg.amountDCRtoUSD+txFeeValueUSD, 'f', 2, 64)),
		}
	default:
		return amountValue{
			sendAmountDCR:           dcrutil.Amount(pg.amountAtoms).String(),
			sendAmountUSD:           "",
			leftTransactionFeeValue: dcrutil.Amount(pg.txFee).String(),
			leftTotalCostValue:      dcrutil.Amount(pg.totalCostDCR).String(),
		}
	}
}

func (pg *sendPage) getTxFee() {
	// calculate transaction fee
	feeAndSize, err := pg.txAuthor.EstimateFeeAndSize()
	if err != nil {
		pg.feeEstimationError(err.Error(), "fee")
		return
	}

	pg.txFee = feeAndSize.Fee.AtomValue
	pg.txFeeSize = fmt.Sprintf("%v Bytes", feeAndSize.EstimatedSignedSize)
}

func (pg *sendPage) balanceAfterSend(isInputAmountEmpty bool) {
	sendAcct := pg.sourceAccountSelector.selectedAccount

	pg.remainingBalance = 0
	if isInputAmountEmpty {
		pg.remainingBalance = sendAcct.Balance.Spendable
	} else {
		pg.remainingBalance = sendAcct.Balance.Spendable - pg.totalCostDCR
	}
	pg.balanceAfterSendValue = dcrutil.Amount(pg.remainingBalance).String()
}

func (pg *sendPage) feeEstimationError(err, errorPath string) {
	if err == dcrlibwallet.ErrInsufficientBalance {
		pg.amountErrorText = "Not enough funds"
		return
	}
	if strings.Contains(err, "invalid amount") {
		pg.amountErrorText = "Invalid amount"
		return
	}
	pg.calculateErrorText = fmt.Sprintf("error estimating transaction %s: %s", errorPath, err)
}

func (pg *sendPage) watchForBroadcastResult(c *pageCommon) {
	if pg.broadcastResult == nil {
		return
	}

	if pg.broadcastResult.TxHash != "" {
		*c.page = page.OverviewPageID
		c.notify("1 Transaction Sent", true)

		if pg.remainingBalance != -1 {
			pg.spendableBalance = pg.remainingBalance
		}
		pg.remainingBalance = -1

		pg.confirmTxModal.Dismiss()
		pg.isBroadcastingTransaction = false
		pg.resetFields()
		pg.broadcastResult.TxHash = ""
		pg.calculateValues(true)
		pg.destinationAddressEditor.Editor.SetText("")
	}
}

func (pg *sendPage) handleEditorChange(evt widget.EditorEvent) {
	switch evt.(type) {
	case widget.ChangeEvent:
		pg.fetchExchangeValue()
		pg.calculateValues(true)
	case widget.SubmitEvent:
		pg.sendFund()
	}
}

func (pg *sendPage) updateDefaultValues() {
	v := pg.amountValues()
	pg.sendAmountDCR = v.sendAmountDCR
	pg.sendAmountUSD = v.sendAmountUSD
	pg.activeTotalAmount = pg.leftExchangeValue
	pg.leftTransactionFeeValue = v.leftTransactionFeeValue
	pg.rightTransactionFeeValue = v.rightTransactionFeeValue
	pg.leftTotalCostValue = v.leftTotalCostValue
	pg.rightTotalCostValue = v.rightTotalCostValue
}

func (pg *sendPage) resetFields() {
	pg.destinationAddressEditor.SetError("")
	pg.leftAmountEditor.Editor.SetText("")
	pg.rightAmountEditor.Editor.SetText("")
	pg.passwordEditor.Editor.SetText("")
	pg.leftTotalCostValue = ""
	pg.rightTotalCostValue = ""
}

func (pg *sendPage) resetErrorText() {
	pg.amountErrorText = ""
	pg.calculateErrorText = ""
	pg.destinationAddressEditor.SetError("")
	pg.leftAmountEditor.SetError("")
	pg.rightAmountEditor.SetError("")
	pg.passwordEditor.SetError("")
}

func (pg *sendPage) fetchExchangeValue() {
	go func() {
		err := pg.wallet.GetUSDExchangeValues(&pg)
		if err != nil {
			pg.updateExchangeError()
		}
	}()
}

func (pg *sendPage) setMaxAmount() {

	// Get spendable balance
	sendAcct := pg.sourceAccountSelector.selectedAccount
	atomValue := sendAcct.Balance.Spendable

	pg.updateAmountField(dcrutil.Amount(0).ToCoin())
	pg.calculateValues(false)

	if atomValue > 0 {
		// Estimate max send value
		amount, err := pg.txAuthor.EstimateMaxSendAmount()
		if err == nil {
			atomValue = amount.AtomValue
			dcrValue := amount.DcrValue
			pg.updateAmountField(dcrValue)
			pg.calculateValues(false)
		}

		// Adjust value
		step := int64(10)
		for {
			_, err := pg.txAuthor.EstimateFeeAndSize()
			if err != nil {
				atomValue -= step
				pg.updateAmountField(dcrutil.Amount(atomValue).ToCoin())
				pg.calculateValues(false)
			} else {
				pg.updateAmountField(dcrutil.Amount(atomValue).ToCoin())
				pg.calculateValues(false)
				break
			}
		}
	}
}

func (pg *sendPage) updateAmountField(spendableBalanceDCR float64) {
	if !pg.usdExchangeSet {
		pg.leftAmountEditor.Editor.SetText(strconv.FormatFloat(spendableBalanceDCR, 'f', 7, 64))
	} else {
		pg.fetchExchangeValue()
		pg.usdExchangeRate, _ = strconv.ParseFloat(pg.LastTradeRate, 64)
		spendableBalanceUSD := spendableBalanceDCR * pg.usdExchangeRate

		switch {
		case pg.leftExchangeValue == "USD":
			pg.leftAmountEditor.Editor.SetText(strconv.FormatFloat(spendableBalanceUSD, 'f', 7, 64))
			pg.rightAmountEditor.Editor.SetText(strconv.FormatFloat(spendableBalanceDCR, 'f', 7, 64))
		case pg.leftExchangeValue == "DCR":
			pg.leftAmountEditor.Editor.SetText(strconv.FormatFloat(spendableBalanceDCR, 'f', 7, 64))
			pg.rightAmountEditor.Editor.SetText(strconv.FormatFloat(spendableBalanceUSD, 'f', 7, 64))
		}
	}
}

func (pg *sendPage) sendFund() {
	if !pg.inputsNotEmpty(pg.passwordEditor.Editor) {
		return
	}
	pg.isBroadcastingTransaction = true
	pg.wallet.BroadcastTransaction(pg.txAuthor, []byte(pg.passwordEditor.Editor.Text()), pg.broadcastErrChan)
}

func (pg *sendPage) Handle() {
	c := pg.common
	sendAcct := pg.sourceAccountSelector.selectedAccount

	if len(c.info.Wallets) == 0 {
		return
	}

	if pg.LastTradeRate == "" && pg.count == 0 {
		pg.count = 1
		pg.shouldInitializeTxAuthor = true
		pg.calculateValues(true)
	}

	if (pg.LastTradeRate != "" && pg.count == 0) || (pg.LastTradeRate != "" && pg.count == 1) {
		pg.count = 2
		pg.shouldInitializeTxAuthor = true
		pg.calculateValues(true)
	}

	pg.updateExchangeError()

	if pg.exchangeErr != "" {
		c.notify(pg.exchangeErr, false)
	}

	pg.sendToOption = pg.accountSwitch.SelectedOption()

	if pg.backButton.Button.Clicked() {
		pg.resetErrorText()
		pg.resetFields()
		c.changePage(*c.returnPage)
	}

	if pg.infoButton.Button.Clicked() {
		info := newInfoModal(c).
			title("Send DCR").
			body("Input or scan the destination wallet address and input the amount to send funds.").
			positiveButton("Got it", func() {})
		c.showModal(info)
	}

	for pg.moreOption.Button.Clicked() {
		pg.isMoreOption = !pg.isMoreOption
	}

	currencyExchangeValue := pg.wallet.ReadStringConfigValueForKey(dcrlibwallet.CurrencyConversionConfigKey)
	pg.usdExchangeSet = false
	if currencyExchangeValue == page.USDExchangeValue {
		pg.usdExchangeSet = true
	}

	for range pg.destinationAddressEditor.Editor.Events() {
		pg.calculateValues(true)
	}

	for pg.currencySwap.Clicked() {
		if pg.LastTradeRate != "" {
			if pg.leftExchangeValue == "DCR" {
				pg.leftExchangeValue = "USD"
				pg.rightExchangeValue = "DCR"
			} else {
				pg.leftExchangeValue = "DCR"
				pg.rightExchangeValue = "USD"
			}
		}
		pg.calculateValues(true)
	}

	for _, evt := range pg.leftAmountEditor.Editor.Events() {
		if pg.leftAmountEditor.Editor.Focused() {
			pg.handleEditorChange(evt)
		}
	}

	for _, evt := range pg.rightAmountEditor.Editor.Events() {
		if pg.rightAmountEditor.Editor.Focused() {
			pg.handleEditorChange(evt)
		}
	}

	for _, evt := range pg.passwordEditor.Editor.Events() {
		if pg.passwordEditor.Editor.Focused() {
			pg.handleEditorChange(evt)
		}
	}

	if pg.calculateErrorText != "" {
		pg.leftAmountEditor.LineColor, pg.leftAmountEditor.TitleLabelColor = pg.theme.Color.Danger, pg.theme.Color.Danger
		pg.rightAmountEditor.LineColor, pg.rightAmountEditor.TitleLabelColor = pg.theme.Color.Danger, pg.theme.Color.Danger
		c.notify(pg.calculateErrorText, false)
	} else {
		pg.leftAmountEditor.LineColor, pg.leftAmountEditor.TitleLabelColor = pg.theme.Color.Gray1, pg.theme.Color.Gray3
		pg.rightAmountEditor.LineColor, pg.rightAmountEditor.TitleLabelColor = pg.theme.Color.Gray1, pg.theme.Color.Gray3
	}

	if pg.amountErrorText != "" {
		pg.leftAmountEditor.SetError(pg.amountErrorText)
	}

	if pg.shouldInitializeTxAuthor {
		pg.shouldInitializeTxAuthor = false
		pg.leftAmountEditor.Editor.SetText("")
		pg.rightAmountEditor.Editor.SetText("")
		pg.calculateErrorText = ""
		c.wallet.CreateTransaction(sendAcct.WalletID, sendAcct.Number, pg.txAuthorErrChan)
	}

	activeAmountEditor := pg.leftAmountEditor.Editor
	if pg.rightAmountEditor.Editor.Focused() {
		activeAmountEditor = pg.rightAmountEditor.Editor
	}
	if !pg.inputsNotEmpty(pg.destinationAddressEditor.Editor, activeAmountEditor) {
		pg.balanceAfterSend(true)
	}

	pg.watchForBroadcastResult(c)

	for pg.confirmButton.Button.Clicked() {
		pg.sendFund()
	}

	for pg.nextButton.Button.Clicked() {
		if pg.validate() && pg.calculateErrorText == "" {
			pg.comfirmModalData.destinationAddress = pg.destinationAddressEditor.Editor.Text()
			pg.confirmTxModal.Show()
			pg.passwordEditor.Editor.Focus()
		}
	}

	for pg.closeConfirmationModalButton.Button.Clicked() {
		pg.confirmTxModal.Dismiss()
	}

	for pg.clearAllBtn.Button.Clicked() {
		pg.resetFields()
	}

	select {
	case err := <-pg.txAuthorErrChan:
		pg.calculateErrorText = err.Error()
		c.notify(pg.calculateErrorText, false)
	case err := <-pg.broadcastErrChan:
		if err.Error() == invalidPassphraseError {
			pg.passwordEditor.SetError("Wrong password")
		} else {
			c.notify(err.Error(), false)
			pg.confirmTxModal.Dismiss()
		}
		pg.isBroadcastingTransaction = false
	default:
	}

	if pg.leftAmountEditor.CustomButton.Button.Clicked() {
		pg.leftAmountEditor.Editor.Focus()
		pg.setMaxAmount()
	}
	if pg.rightAmountEditor.CustomButton.Button.Clicked() {
		pg.rightAmountEditor.Editor.Focus()
		pg.setMaxAmount()
	}
}

func (pg *sendPage) OnClose() {

}
