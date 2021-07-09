package page

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/modal"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/decred/dcrd/dcrutil"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const TransactionDetailsPageID = "TransactionDetails"

type TransactionDetailsPage struct {
	*load.Load
	theme                           *decredmaterial.Theme
	transactionDetailsPageContainer layout.List
	transactionInputsContainer      layout.List
	transactionOutputsContainer     layout.List
	hashClickable                   *widget.Clickable
	destAddressClickable            *widget.Clickable
	copyTextBtn                     []decredmaterial.Button
	dot                             *widget.Icon
	toDcrdata                       *widget.Clickable
	outputsCollapsible              *decredmaterial.Collapsible
	inputsCollapsible               *decredmaterial.Collapsible
	backButton                      decredmaterial.IconButton
	infoButton                      decredmaterial.IconButton
	gtx                             *layout.Context

	transaction *dcrlibwallet.Transaction
	wallet      *dcrlibwallet.Wallet

	txSourceAccount      string
	txDestinationAddress string
}

func NewTransactionDetailsPage(l *load.Load, transaction *dcrlibwallet.Transaction) *TransactionDetailsPage {
	pg := &TransactionDetailsPage{
		Load: l,
		transactionDetailsPageContainer: layout.List{
			Axis: layout.Vertical,
		},
		transactionInputsContainer: layout.List{
			Axis: layout.Vertical,
		},
		transactionOutputsContainer: layout.List{
			Axis: layout.Vertical,
		},

		theme: l.Theme,

		outputsCollapsible: l.Theme.Collapsible(),
		inputsCollapsible:  l.Theme.Collapsible(),

		hashClickable:        new(widget.Clickable),
		destAddressClickable: new(widget.Clickable),
		toDcrdata:            new(widget.Clickable),

		transaction: transaction,
		wallet:      l.WL.MultiWallet.WalletWithID(transaction.WalletID),
	}

	pg.backButton, pg.infoButton = subpageHeaderButtons(pg.Load)

	pg.copyTextBtn = make([]decredmaterial.Button, 0)

	pg.dot = l.Icons.ImageBrightness1
	pg.dot.Color = l.Theme.Color.Gray

	// find source account
	if transaction.Direction == dcrlibwallet.TxDirectionSent ||
		transaction.Direction == dcrlibwallet.TxDirectionTransferred {
		for _, input := range transaction.Inputs {
			if input.AccountNumber != -1 {
				accountName, err := pg.wallet.AccountName(input.AccountNumber)
				if err != nil {
					log.Error(err)
				} else {
					pg.txSourceAccount = accountName
				}
			}
		}
	}

	//	find destination address
	if transaction.Direction == dcrlibwallet.TxDirectionSent {
		for _, output := range transaction.Outputs {
			if output.AccountNumber == -1 {
				pg.txDestinationAddress = output.Address
			}
		}
	}

	return pg
}

func (pg *TransactionDetailsPage) OnResume() {

}

func (pg *TransactionDetailsPage) Layout(gtx layout.Context) layout.Dimensions {
	if pg.gtx == nil {
		pg.gtx = &gtx
	}

	body := func(gtx C) D {
		sp := SubPage{
			Load:       pg.Load,
			title:      dcrlibwallet.TransactionDirectionName(pg.transaction.Direction),
			backButton: pg.backButton,
			infoButton: pg.infoButton,
			back: func() {
				pg.ChangePage(*pg.ReturnPage)
			},
			body: func(gtx layout.Context) layout.Dimensions {
				widgets := []func(gtx C) D{
					func(gtx C) D {
						return pg.txnBalanceAndStatus(gtx)
					},
					func(gtx C) D {
						return pg.separator(gtx)
					},
					func(gtx C) D {
						return pg.txnTypeAndID(gtx)
					},
					func(gtx C) D {
						return pg.separator(gtx)
					},
					func(gtx C) D {
						return pg.txnInputs(gtx)
					},
					func(gtx C) D {
						return pg.separator(gtx)
					},
					func(gtx C) D {
						return pg.txnOutputs(gtx)
					},
					func(gtx C) D {
						return pg.separator(gtx)
					},
					func(gtx C) D {
						return pg.viewTxn(gtx)
					},
				}
				return pg.Theme.Card().Layout(gtx, func(gtx C) D {
					return pg.transactionDetailsPageContainer.Layout(gtx, len(widgets), func(gtx C, i int) D {
						return layout.Inset{}.Layout(gtx, widgets[i])
					})
				})
			},
			infoTemplate: modal.TransactionDetailsInfoTemplate,
		}
		return sp.Layout(gtx)
	}

	return uniformPadding(gtx, body)
}

func (pg *TransactionDetailsPage) txnBalanceAndStatus(gtx layout.Context) layout.Dimensions {
	txnWidgets := initTxnWidgets(pg.Load, pg.transaction)
	return pg.pageSections(gtx, func(gtx C) D {
		return layout.Flex{}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{
					Right: values.MarginPadding15,
					Top:   values.MarginPadding10,
				}.Layout(gtx, txnWidgets.direction.Layout)
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						amount := strings.Split(dcrutil.Amount(pg.transaction.Amount).String(), " ")
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								return layout.Inset{Right: values.MarginPadding2}.Layout(gtx, pg.Theme.H4(amount[0]).Layout)
							}),
							layout.Rigid(pg.Theme.H6(amount[1]).Layout),
						)
					}),
					layout.Rigid(func(gtx C) D {
						m := values.MarginPadding10
						return layout.Inset{
							Top:    m,
							Bottom: m,
						}.Layout(gtx, func(gtx C) D {
							txnWidgets.time.Color = pg.Theme.Color.Gray
							return txnWidgets.time.Layout(gtx)
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								return layout.Inset{
									Right: values.MarginPadding5,
									Top:   values.MarginPadding2,
								}.Layout(gtx, txnWidgets.statusIcon.Layout)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								txt := pg.Theme.Body1("")
								if pg.txConfirmations() > 1 {
									txt.Text = strings.Title("confirmed")
									txt.Color = pg.Theme.Color.Success
								} else {
									txt.Color = pg.Theme.Color.Gray
								}
								return txt.Layout(gtx)
							}),
							layout.Rigid(func(gtx C) D {
								m := values.MarginPadding10
								return layout.Inset{
									Left:  m,
									Right: m,
									Top:   m,
								}.Layout(gtx, func(gtx C) D {
									return pg.dot.Layout(gtx, values.MarginPadding2)
								})
							}),
							layout.Rigid(func(gtx C) D {
								txt := pg.Theme.Body1(values.StringF(values.StrNConfirmations, pg.txConfirmations()))
								txt.Color = pg.Theme.Color.Gray
								return txt.Layout(gtx)
							}),
						)
					}),
				)
			}),
		)
	})
}

//TODO: do this at startup
func (pg *TransactionDetailsPage) txConfirmations() int32 {
	transaction := pg.transaction
	if transaction.BlockHeight != -1 {
		return (pg.WL.MultiWallet.WalletWithID(transaction.WalletID).GetBestBlock() - transaction.BlockHeight) + 1
	}

	return 0
}

func (pg *TransactionDetailsPage) txnTypeAndID(gtx layout.Context) layout.Dimensions {
	transaction := pg.transaction
	return pg.pageSections(gtx, func(gtx C) D {
		m := values.MarginPadding10
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return pg.txnInfoSection(gtx, values.String(values.StrFrom), pg.txSourceAccount, true, nil)
			}),
			layout.Rigid(func(gtx C) D {
				if transaction.Direction == dcrlibwallet.TxDirectionSent {
					return layout.Inset{Top: m}.Layout(gtx, func(gtx C) D {
						return pg.txnInfoSection(gtx, values.String(values.StrTo), pg.txDestinationAddress, false, pg.destAddressClickable)
					})
				}
				return layout.Dimensions{}
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Bottom: m, Top: m}.Layout(gtx, func(gtx C) D {
					return pg.txnInfoSection(gtx, values.String(values.StrFee), dcrutil.Amount(transaction.Fee).String(), false, nil)
				})
			}),
			layout.Rigid(func(gtx C) D {
				if transaction.BlockHeight != -1 {
					return pg.txnInfoSection(gtx, values.String(values.StrIncludedInBlock), fmt.Sprintf("%d", transaction.BlockHeight), false, nil)
				}
				return layout.Dimensions{}
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Bottom: m, Top: m}.Layout(gtx, func(gtx C) D {
					return pg.txnInfoSection(gtx, values.String(values.StrType), transaction.Type, false, nil)
				})
			}),
			layout.Rigid(func(gtx C) D {
				trimmedHash := transaction.Hash[:24] + "..." + transaction.Hash[len(transaction.Hash)-24:]
				return layout.Inset{Bottom: m}.Layout(gtx, func(gtx C) D {
					return pg.txnInfoSection(gtx, values.String(values.StrTransactionID), trimmedHash, false, pg.hashClickable)
				})
			}),
		)
	})
}

func (pg *TransactionDetailsPage) txnInfoSection(gtx layout.Context, label, value string, showWalletBadge bool, clickable *widget.Clickable) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			t := pg.theme.Body1(label)
			t.Color = pg.theme.Color.Gray
			return t.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					if showWalletBadge {
						card := pg.theme.Card()
						card.Radius = decredmaterial.CornerRadius{
							NE: 0,
							NW: 0,
							SE: 0,
							SW: 0,
						}
						card.Color = pg.theme.Color.LightGray
						return card.Layout(gtx, func(gtx C) D {
							return layout.UniformInset(values.MarginPadding2).Layout(gtx, func(gtx C) D {
								txt := pg.theme.Body2(pg.wallet.Name)
								txt.Color = pg.theme.Color.Gray
								return txt.Layout(gtx)
							})
						})
					}
					return layout.Dimensions{}
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
						if clickable == nil {
							txt := pg.theme.Body1(value)
							return txt.Layout(gtx)
						}

						btn := pg.theme.Button(clickable, value)
						btn.Color = pg.theme.Color.Primary
						btn.Background = color.NRGBA{}
						btn.Inset = layout.UniformInset(values.MarginPadding0)
						return btn.Layout(gtx)
					})
				}),
			)
		}),
	)
}

func (pg *TransactionDetailsPage) txnInputs(gtx layout.Context) layout.Dimensions {
	transaction := pg.transaction
	x := len(transaction.Inputs) + len(transaction.Outputs)
	for i := 0; i < x; i++ {
		pg.copyTextBtn = append(pg.copyTextBtn, pg.theme.Button(new(widget.Clickable), ""))
	}

	collapsibleHeader := func(gtx C) D {
		t := pg.theme.Body1(values.StringF(values.StrXInputsConsumed, len(transaction.Inputs)))
		t.Color = pg.theme.Color.Gray
		return t.Layout(gtx)
	}

	collapsibleBody := func(gtx C) D {
		return pg.transactionInputsContainer.Layout(gtx, len(transaction.Inputs), func(gtx C, i int) D {
			input := transaction.Inputs[i]
			return pg.txnIORow(gtx, input.Amount, input.AccountNumber, input.PreviousOutpoint, i)
		})
	}
	return pg.pageSections(gtx, func(gtx C) D {
		return pg.inputsCollapsible.Layout(gtx, collapsibleHeader, collapsibleBody)
	})
}

func (pg *TransactionDetailsPage) txnOutputs(gtx layout.Context) layout.Dimensions {
	transaction := pg.transaction

	collapsibleHeader := func(gtx C) D {
		t := pg.Theme.Body1(values.StringF(values.StrXOutputCreated, len(transaction.Outputs)))
		t.Color = pg.Theme.Color.Gray
		return t.Layout(gtx)
	}

	collapsibleBody := func(gtx C) D {
		return pg.transactionOutputsContainer.Layout(gtx, len(transaction.Outputs), func(gtx C, i int) D {
			output := transaction.Outputs[i]
			x := len(transaction.Inputs)
			return pg.txnIORow(gtx, output.Amount, output.AccountNumber, output.Address, i+x)
		})
	}
	return pg.pageSections(gtx, func(gtx C) D {
		return pg.outputsCollapsible.Layout(gtx, collapsibleHeader, collapsibleBody)
	})
}

func (pg *TransactionDetailsPage) txnIORow(gtx layout.Context, amount int64, acctNum int32, address string, i int) layout.Dimensions {

	accountName := "external"
	walletName := ""
	if acctNum != -1 {
		name, err := pg.wallet.AccountName(acctNum)
		if err == nil {
			accountName = name
			walletName = pg.wallet.Name
		}
	}

	accountName = fmt.Sprintf("(%s)", accountName)
	amt := dcrutil.Amount(amount).String()

	return layout.Inset{Bottom: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
		card := pg.theme.Card()
		card.Color = pg.theme.Color.LightGray
		return card.Layout(gtx, func(gtx C) D {
			return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Flex{}.Layout(gtx,
							layout.Rigid(pg.theme.Body1(amt).Layout),
							layout.Rigid(func(gtx C) D {
								m := values.MarginPadding5
								return layout.Inset{
									Left:  m,
									Right: m,
								}.Layout(gtx, pg.theme.Body1(accountName).Layout)
							}),
							layout.Rigid(func(gtx C) D {
								card := pg.theme.Card()
								card.Radius = decredmaterial.CornerRadius{
									NE: 0,
									NW: 0,
									SE: 0,
									SW: 0,
								}
								card.Color = pg.theme.Color.LightGray
								return card.Layout(gtx, func(gtx C) D {
									return layout.UniformInset(values.MarginPadding2).Layout(gtx, func(gtx C) D {
										txt := pg.theme.Body2(walletName)
										txt.Color = pg.theme.Color.Gray
										return txt.Layout(gtx)
									})
								})
							}),
						)
					}),
					layout.Rigid(func(gtx C) D {
						pg.copyTextBtn[i].Color = pg.theme.Color.Primary
						pg.copyTextBtn[i].Background = color.NRGBA{}
						pg.copyTextBtn[i].Text = address
						pg.copyTextBtn[i].Inset = layout.UniformInset(values.MarginPadding0)

						return layout.W.Layout(gtx, pg.copyTextBtn[i].Layout)
					}),
				)
			})
		})
	})
}

func (pg *TransactionDetailsPage) viewTxn(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return pg.pageSections(gtx, func(gtx C) D {
		return layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Rigid(pg.theme.Body1(values.String(values.StrViewOnDcrdata)).Layout),
			layout.Rigid(func(gtx C) D {
				redirect := pg.Icons.RedirectIcon
				redirect.Scale = 1.0
				return decredmaterial.Clickable(gtx, pg.toDcrdata, redirect.Layout)
			}),
		)
	})
}

func (pg *TransactionDetailsPage) pageSections(gtx layout.Context, body layout.Widget) layout.Dimensions {
	m := values.MarginPadding20
	mtb := values.MarginPadding5
	return layout.Inset{Left: m, Right: m, Top: mtb, Bottom: mtb}.Layout(gtx, body)
}

func (pg *TransactionDetailsPage) separator(gtx layout.Context) layout.Dimensions {
	m := values.MarginPadding5
	return layout.Inset{Top: m, Bottom: m}.Layout(gtx, pg.theme.Separator().Layout)
}

func (pg *TransactionDetailsPage) Handle() {
	gtx := pg.gtx
	if pg.toDcrdata.Clicked() {
		goToURL(pg.WL.Wallet.GetBlockExplorerURL(pg.transaction.Hash))
	}

	for _, b := range pg.copyTextBtn {
		for b.Button.Clicked() {
			clipboard.WriteOp{Text: b.Text}.Add(gtx.Ops)
		}
	}

	for pg.hashClickable.Clicked() {
		clipboard.WriteOp{Text: pg.transaction.Hash}.Add(gtx.Ops)
	}

	for pg.destAddressClickable.Clicked() {
		clipboard.WriteOp{Text: pg.txDestinationAddress}.Add(gtx.Ops)
	}
}

func (pg *TransactionDetailsPage) OnClose() {}
