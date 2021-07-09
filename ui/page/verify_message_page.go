package page

import (
	"image/color"
	"strings"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"

	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/modal"
	"github.com/planetdecred/godcr/ui/values"
)

const VerifyMessagePageID = "VerifyMessage"

type VerifyMessagePage struct {
	*load.Load
	addressInput, messageInput, signInput decredmaterial.Editor
	clearBtn, verifyBtn                   decredmaterial.Button
	verifyMessage                         decredmaterial.Label

	verifyMessageStatus *widget.Icon

	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton
}

func NewVerifyMessagePage(l *load.Load) *VerifyMessagePage {
	pg := &VerifyMessagePage{
		Load:          l,
		addressInput:  l.Theme.Editor(new(widget.Editor), "Address"),
		messageInput:  l.Theme.Editor(new(widget.Editor), "Message"),
		signInput:     l.Theme.Editor(new(widget.Editor), "Signature"),
		verifyMessage: l.Theme.Body1(""),
		verifyBtn:     l.Theme.Button(new(widget.Clickable), "Verify message"),
		clearBtn:      l.Theme.Button(new(widget.Clickable), "Clear all"),
	}

	pg.addressInput.Editor.SingleLine, pg.messageInput.Editor.SingleLine = true, true
	pg.signInput.Editor.Submit, pg.addressInput.Editor.Submit, pg.messageInput.Editor.Submit = true, true, true
	pg.verifyBtn.TextSize, pg.clearBtn.TextSize, pg.clearBtn.TextSize = values.TextSize14, values.TextSize14, values.TextSize14
	pg.clearBtn.Background = color.NRGBA{0, 0, 0, 0}
	pg.verifyBtn.Font.Weight = text.Bold
	pg.clearBtn.Font.Weight = text.Bold

	pg.backButton, pg.infoButton = subpageHeaderButtons(l)

	return pg
}

func (pg *VerifyMessagePage) OnResume() {

}

func (pg *VerifyMessagePage) Layout(gtx layout.Context) layout.Dimensions {
	body := func(gtx C) D {
		sp := SubPage{
			Load:       pg.Load,
			title:      "Verify message",
			backButton: pg.backButton,
			infoButton: pg.infoButton,
			back: func() {
				pg.clearInputs()
				pg.ChangePage(WalletPageID)
				pg.ChangePage(*pg.ReturnPage)
			},
			body: func(gtx layout.Context) layout.Dimensions {
				return pg.Theme.Card().Layout(gtx, func(gtx C) D {
					return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(pg.description()),
							layout.Rigid(pg.inputRow(pg.addressInput)),
							layout.Rigid(pg.inputRow(pg.signInput)),
							layout.Rigid(pg.inputRow(pg.messageInput)),
							layout.Rigid(pg.verifyAndClearButtons()),
							layout.Rigid(pg.verifyMessageResponse()),
						)
					})
				})
			},
			infoTemplate: modal.VerifyMessageInfoTemplate,
		}
		return sp.Layout(gtx)
	}
	return uniformPadding(gtx, body)
}

func (pg *VerifyMessagePage) inputRow(editor decredmaterial.Editor) layout.Widget {
	return func(gtx C) D {
		return layout.Inset{Bottom: values.MarginPadding15}.Layout(gtx, editor.Layout)
	}
}

func (pg *VerifyMessagePage) description() layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		desc := pg.Theme.Caption("Enter the address, signature, and message to verify:")
		desc.Color = pg.Theme.Color.Gray
		return layout.Inset{Bottom: values.MarginPadding20}.Layout(gtx, desc.Layout)
	}
}

func (pg *VerifyMessagePage) verifyAndClearButtons() layout.Widget {
	return func(gtx C) D {
		dims := layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Flexed(1, func(gtx C) D {
				return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return layout.Inset{Right: values.MarginPadding10}.Layout(gtx, pg.clearBtn.Layout)
						}),
						layout.Rigid(pg.verifyBtn.Layout),
					)
				})
			}),
		)
		return dims
	}
}

func (pg *VerifyMessagePage) verifyMessageResponse() layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if pg.verifyMessageStatus != nil {
			return layout.Inset{Top: values.MarginPadding30}.Layout(gtx, func(gtx C) D {
				pg.Theme.Separator().Layout(gtx)
				return layout.Inset{Top: values.MarginPadding15}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
								return pg.verifyMessageStatus.Layout(gtx, values.MarginPadding20)
							})
						}),
						layout.Rigid(pg.verifyMessage.Layout),
					)
				})
			})
		}
		return layout.Dimensions{}
	}
}

func (pg *VerifyMessagePage) Handle() {
	pg.verifyBtn.Background, pg.clearBtn.Color = pg.Theme.Color.Hint, pg.Theme.Color.Hint
	if pg.inputsNotEmpty() {
		pg.verifyBtn.Background, pg.clearBtn.Color = pg.Theme.Color.Primary, pg.Theme.Color.Primary
		if pg.verifyBtn.Button.Clicked() || handleSubmitEvent(pg.addressInput.Editor, pg.messageInput.Editor, pg.signInput.Editor) {
			pg.verifyMessage.Text = ""
			pg.verifyMessageStatus = nil
			valid, err := pg.WL.MultiWallet.VerifyMessage(pg.addressInput.Editor.Text(), pg.messageInput.Editor.Text(), pg.signInput.Editor.Text())
			if err != nil {
				pg.signInput.SetError("Invalid signature or message")
				return
			}
			pg.signInput.SetError("")

			if !valid {
				pg.verifyMessageStatus = pg.Icons.NavigationCancel
				pg.verifyMessage.Text = "Invalid signature or message"
				pg.verifyMessage.Color = pg.Theme.Color.Danger
				return
			}

			pg.verifyMessageStatus = pg.Icons.ActionCheck
			pg.verifyMessageStatus.Color = pg.Theme.Color.Success
			pg.verifyMessage.Text = "Valid signature"
			pg.verifyMessage.Color = pg.Theme.Color.Success
		}
	}

	pg.handlerEditorEvents(pg.addressInput.Editor)
	if pg.clearBtn.Button.Clicked() {
		pg.clearInputs()
	}
}

func (pg *VerifyMessagePage) handlerEditorEvents(w *widget.Editor) {
	for _, evt := range w.Events() {
		switch evt.(type) {
		case widget.ChangeEvent:
			pg.validateAddress()
			return
		}
	}
}

func (pg *VerifyMessagePage) clearInputs() {
	pg.verifyMessageStatus = nil
	pg.verifyBtn.Background = pg.Theme.Color.Hint
	pg.addressInput.Editor.SetText("")
	pg.signInput.Editor.SetText("")
	pg.messageInput.Editor.SetText("")
	pg.verifyMessage.Text = ""
	pg.addressInput.SetError("")
	pg.signInput.SetError("")
}

func (pg *VerifyMessagePage) validateAddress() bool {
	if isValid, _ := pg.WL.Wallet.IsAddressValid(pg.addressInput.Editor.Text()); !isValid {
		pg.addressInput.SetError("Invalid address")
		return false
	}

	pg.addressInput.SetError("")
	return true
}

func (pg *VerifyMessagePage) inputsNotEmpty() bool {
	if strings.Trim(pg.addressInput.Editor.Text(), " ") == "" {
		return false
	}
	if strings.Trim(pg.messageInput.Editor.Text(), " ") == "" {
		return false
	}
	if strings.Trim(pg.signInput.Editor.Text(), " ") == "" {
		return false
	}

	pg.verifyBtn.Background = pg.Theme.Color.Primary
	return true
}

func (pg *VerifyMessagePage) OnClose() {}
