package page

import (
	"image/color"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/modal"
	"github.com/planetdecred/godcr/ui/values"
)

const SignMessagePageID = "SignMessage"

type SignMessagePage struct {
	*load.Load
	theme     *decredmaterial.Theme
	container layout.List
	wallet    *dcrlibwallet.Wallet

	isSigningMessage                           bool
	titleLabel, errorLabel, signedMessageLabel decredmaterial.Label
	addressEditor, messageEditor               decredmaterial.Editor
	clearButton, signButton, copyButton        decredmaterial.Button
	copySignature                              *widget.Clickable
	copyIcon                                   *widget.Image
	gtx                                        *layout.Context

	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton
}

func NewSignMessagePage(l *load.Load, wallet *dcrlibwallet.Wallet) *SignMessagePage {
	addressEditor := l.Theme.Editor(new(widget.Editor), "Address")
	addressEditor.Editor.SingleLine, addressEditor.Editor.Submit = true, true
	messageEditor := l.Theme.Editor(new(widget.Editor), "Message")
	messageEditor.Editor.SingleLine, messageEditor.Editor.Submit = true, true
	clearButton := l.Theme.Button(new(widget.Clickable), "Clear all")
	clearButton.Background = color.NRGBA{}
	clearButton.Color = l.Theme.Color.Gray
	clearButton.Font.Weight = text.Bold
	errorLabel := l.Theme.Caption("")
	errorLabel.Color = l.Theme.Color.Danger
	copyIcon := l.Icons.CopyIcon

	pg := &SignMessagePage{
		Load:   l,
		wallet: wallet,
		container: layout.List{
			Axis: layout.Vertical,
		},

		titleLabel:         l.Theme.H5("Sign Message"),
		signedMessageLabel: l.Theme.Body1(""),
		errorLabel:         errorLabel,
		addressEditor:      addressEditor,
		messageEditor:      messageEditor,

		clearButton:   clearButton,
		signButton:    l.Theme.Button(new(widget.Clickable), "Sign message"),
		copyButton:    l.Theme.Button(new(widget.Clickable), "Copy"),
		copySignature: new(widget.Clickable),
		copyIcon:      copyIcon,
	}

	pg.signedMessageLabel.Color = l.Theme.Color.Gray
	pg.backButton, pg.infoButton = subpageHeaderButtons(l)
	pg.signButton.Font.Weight = text.Bold

	return pg
}

func (pg *SignMessagePage) OnResume() {

}

func (pg *SignMessagePage) Layout(gtx layout.Context) layout.Dimensions {
	if pg.gtx == nil {
		pg.gtx = &gtx
	}

	body := func(gtx C) D {
		sp := SubPage{
			Load:       pg.Load,
			title:      "Sign message",
			walletName: pg.wallet.Name,
			backButton: pg.backButton,
			infoButton: pg.infoButton,
			back: func() {
				pg.clearForm()
				pg.ChangePage(WalletPageID)
			},
			body: func(gtx layout.Context) layout.Dimensions {
				return pg.Theme.Card().Layout(gtx, func(gtx C) D {
					return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(pg.description()),
							layout.Rigid(pg.editors(pg.addressEditor)),
							layout.Rigid(pg.editors(pg.messageEditor)),
							layout.Rigid(pg.drawButtonsRow()),
							layout.Rigid(pg.drawResult()),
						)
					})
				})
			},
			infoTemplate: modal.SignMessageInfoTemplate,
		}
		return sp.Layout(gtx)
	}

	return uniformPadding(gtx, body)
}

func (pg *SignMessagePage) description() layout.Widget {
	return func(gtx C) D {
		desc := pg.Theme.Caption("Enter an address and message to sign:")
		desc.Color = pg.Theme.Color.Gray
		return layout.Inset{Bottom: values.MarginPadding20}.Layout(gtx, desc.Layout)
	}
}

func (pg *SignMessagePage) editors(editor decredmaterial.Editor) layout.Widget {
	return func(gtx C) D {
		return layout.Inset{Bottom: values.MarginPadding15}.Layout(gtx, editor.Layout)
	}
}

func (pg *SignMessagePage) drawButtonsRow() layout.Widget {
	return func(gtx C) D {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Flexed(1, func(gtx C) D {
				return layout.E.Layout(gtx, func(gtx C) D {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							inset := layout.Inset{
								Right: values.MarginPadding5,
							}
							return inset.Layout(gtx, pg.clearButton.Layout)
						}),
						layout.Rigid(pg.signButton.Layout),
					)
				})
			}),
		)
	}
}

func (pg *SignMessagePage) drawResult() layout.Widget {
	return func(gtx C) D {
		if pg.signedMessageLabel.Text == "" {
			return layout.Dimensions{}
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				m := values.MarginPadding30
				return layout.Inset{Top: m, Bottom: m}.Layout(gtx, pg.Theme.Separator().Layout)
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Stack{}.Layout(gtx,
					layout.Stacked(func(gtx C) D {
						border := widget.Border{Color: pg.theme.Color.LightGray, CornerRadius: values.MarginPadding10, Width: values.MarginPadding2}
						wrapper := pg.theme.Card()
						wrapper.Color = pg.theme.Color.LightGray
						return border.Layout(gtx, func(gtx C) D {
							return wrapper.Layout(gtx, func(gtx C) D {
								return layout.UniformInset(values.MarginPadding10).Layout(gtx, func(gtx C) D {
									return layout.Flex{}.Layout(gtx,
										layout.Flexed(0.9, pg.signedMessageLabel.Layout),
										layout.Flexed(0.1, func(gtx C) D {
											return layout.E.Layout(gtx, func(gtx C) D {
												return layout.Inset{Top: values.MarginPadding7}.Layout(gtx, func(gtx C) D {
													pg.copyIcon.Scale = 1.0
													return decredmaterial.Clickable(gtx, pg.copySignature, pg.copyIcon.Layout)
												})
											})
										}),
									)
								})
							})
						})
					}),
					layout.Stacked(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Top:  values.MarginPaddingMinus10,
							Left: values.MarginPadding10,
						}.Layout(gtx, func(gtx C) D {
							return pg.Theme.Card().Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								label := pg.Theme.Body1("Signature")
								label.Color = pg.Theme.Color.Gray
								return label.Layout(gtx)
							})
						})
					}),
				)
			}),
		)
	}
}

func (pg *SignMessagePage) updateColors() {
	if pg.isSigningMessage || pg.addressEditor.Editor.Text() == "" || pg.messageEditor.Editor.Text() == "" {
		pg.signButton.Background = pg.Theme.Color.Hint
	} else {
		pg.signButton.Background = pg.Theme.Color.Primary
	}
}

func (pg *SignMessagePage) Handle() {
	gtx := pg.gtx
	pg.updateColors()
	pg.validate(true)

	for pg.clearButton.Button.Clicked() {
		pg.clearForm()
	}

	for pg.signButton.Button.Clicked() || handleSubmitEvent(pg.addressEditor.Editor, pg.messageEditor.Editor) {
		if !pg.isSigningMessage && pg.validate(false) {
			address := pg.addressEditor.Editor.Text()
			message := pg.messageEditor.Editor.Text()

			modal.NewPasswordModal(pg.Load).
				Title("Confirm to sign").
				NegativeButton("Cancel", func() {}).
				PositiveButton("Confirm", func(password string, pm *modal.PasswordModal) bool {

					go func() {
						sig, err := pg.wallet.SignMessage([]byte(password), address, message)
						if err != nil {
							pm.SetError(err.Error())
							pm.SetLoading(false)
							return
						}

						pm.Dismiss()
						pg.signedMessageLabel.Text = dcrlibwallet.EncodeBase64(sig)

					}()
					return false
				}).Show()
		}
	}

	if pg.copySignature.Clicked() {
		clipboard.WriteOp{Text: pg.signedMessageLabel.Text}.Add(gtx.Ops)
	}
}

func (pg *SignMessagePage) validate(ignoreEmpty bool) bool {
	isAddressValid := pg.validateAddress(ignoreEmpty)
	isMessageValid := pg.validateMessage(ignoreEmpty)
	if !isAddressValid || !isMessageValid {
		return false
	}
	return true
}

func (pg *SignMessagePage) validateAddress(ignoreEmpty bool) bool {
	address := pg.addressEditor.Editor.Text()
	pg.addressEditor.SetError("")

	if address == "" && !ignoreEmpty {
		pg.addressEditor.SetError("Please enter a valid address")
		return false
	}

	if address != "" {
		isValid := pg.WL.MultiWallet.IsAddressValid(address)
		if !isValid {
			pg.addressEditor.SetError("Invalid address")
			return false
		}

		exist := pg.wallet.HaveAddress(address)

		if !exist {
			pg.addressEditor.SetError("Address not owned by this wallet")
			return false
		}
	}
	return true
}

func (pg *SignMessagePage) validateMessage(ignoreEmpty bool) bool {
	message := pg.messageEditor.Editor.Text()
	if message == "" && !ignoreEmpty {
		return false
	}
	return true
}

func (pg *SignMessagePage) clearForm() {
	pg.addressEditor.Editor.SetText("")
	pg.messageEditor.Editor.SetText("")
	pg.signedMessageLabel.Text = ""
	pg.errorLabel.Text = ""
}

func (pg *SignMessagePage) OnClose() {}
