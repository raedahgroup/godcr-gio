package ui

import (
	"image/color"
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const PageVerifyMessage = "VerifyMessage"

type verifyMessagePage struct {
	theme                                 *decredmaterial.Theme
	common                                pageCommon
	addressInput, messageInput, signInput decredmaterial.Editor
	clearBtn, verifyBtn                   decredmaterial.Button
	verifyMessage                         decredmaterial.Label

	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton

	verifyMessageStatus *widget.Icon
}

func VerifyMessagePage(c pageCommon) Page {
	pg := &verifyMessagePage{
		theme:         c.theme,
		common:        c,
		addressInput:  c.theme.Editor(new(widget.Editor), "Address"),
		messageInput:  c.theme.Editor(new(widget.Editor), "Message"),
		signInput:     c.theme.Editor(new(widget.Editor), "Signature"),
		verifyMessage: c.theme.Body1(""),
		verifyBtn:     c.theme.Button(new(widget.Clickable), "Verify message"),
		clearBtn:      c.theme.Button(new(widget.Clickable), "Clear all"),
	}

	pg.addressInput.Editor.SingleLine, pg.messageInput.Editor.SingleLine = true, true
	pg.signInput.Editor.Submit, pg.addressInput.Editor.Submit, pg.messageInput.Editor.Submit = true, true, true
	pg.verifyBtn.TextSize, pg.clearBtn.TextSize, pg.clearBtn.TextSize = values.TextSize14, values.TextSize14, values.TextSize14
	pg.clearBtn.Background = color.NRGBA{0, 0, 0, 0}

	pg.backButton, pg.infoButton = c.SubPageHeaderButtons()

	return pg
}

func (pg *verifyMessagePage) pageID() string {
	return PageVerifyMessage
}

func (pg *verifyMessagePage) Layout(gtx layout.Context) layout.Dimensions {
	c := pg.common

	body := func(gtx C) D {
		load := SubPage{
			title:      "Verify message",
			walletName: "",
			back: func() {
				pg.clearInputs(&c)
				c.popPage() //TODO
			},
			backButton: pg.backButton,
			infoButton: pg.infoButton,
			body: func(gtx layout.Context) layout.Dimensions {
				return pg.theme.Card().Layout(gtx, func(gtx C) D {
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
			infoTemplate: VerifyMessageInfoTemplate,
		}
		return c.SubPageLayout(gtx, load)
	}

	return pg.common.UniformPadding(gtx, body)
}

func (pg *verifyMessagePage) inputRow(editor decredmaterial.Editor) layout.Widget {
	return func(gtx C) D {
		return layout.Inset{Bottom: values.MarginPadding15}.Layout(gtx, editor.Layout)
	}
}

func (pg *verifyMessagePage) description() layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		desc := pg.theme.Caption("Enter the address, signature, and message to verify:")
		desc.Color = pg.theme.Color.Gray
		return layout.Inset{Bottom: values.MarginPadding20}.Layout(gtx, desc.Layout)
	}
}

func (pg *verifyMessagePage) verifyAndClearButtons() layout.Widget {
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

func (pg *verifyMessagePage) verifyMessageResponse() layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if pg.verifyMessageStatus != nil {
			return layout.Inset{Top: values.MarginPadding30}.Layout(gtx, func(gtx C) D {
				pg.theme.Separator().Layout(gtx)
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

func (pg *verifyMessagePage) handle() {
	c := pg.common

	pg.verifyBtn.Background, pg.clearBtn.Color = c.theme.Color.Hint, c.theme.Color.Hint
	if pg.inputsNotEmpty(c) {
		pg.verifyBtn.Background, pg.clearBtn.Color = c.theme.Color.Primary, c.theme.Color.Primary
		if pg.verifyBtn.Button.Clicked() || handleSubmitEvent(pg.addressInput.Editor, pg.messageInput.Editor, pg.signInput.Editor) {
			pg.verifyMessage.Text = ""
			pg.verifyMessageStatus = nil
			valid, err := c.multiWallet.VerifyMessage(pg.addressInput.Editor.Text(), pg.messageInput.Editor.Text(), pg.signInput.Editor.Text())
			if err != nil {
				pg.signInput.SetError("Invalid signature")
				return
			}
			pg.signInput.SetError("")

			if !valid {
				pg.verifyMessageStatus = c.icons.navigationCancel
				pg.verifyMessage.Text = "Invalid signature"
				pg.verifyMessage.Color = c.theme.Color.Danger
				return
			}

			pg.verifyMessageStatus = c.icons.actionCheck
			pg.verifyMessageStatus.Color = c.theme.Color.Success
			pg.verifyMessage.Text = "Valid signature"
			pg.verifyMessage.Color = c.theme.Color.Success
		}
	}

	pg.handlerEditorEvents(&c, pg.addressInput.Editor)
	if pg.clearBtn.Button.Clicked() {
		pg.clearInputs(&c)
	}
}

func (pg *verifyMessagePage) handlerEditorEvents(c *pageCommon, w *widget.Editor) {
	for _, evt := range w.Events() {
		switch evt.(type) {
		case widget.ChangeEvent:
			pg.validateAddress(*c)
			return
		}
	}
}

func (pg *verifyMessagePage) clearInputs(c *pageCommon) {
	pg.verifyMessageStatus = nil
	pg.verifyBtn.Background = c.theme.Color.Hint
	pg.addressInput.Editor.SetText("")
	pg.signInput.Editor.SetText("")
	pg.messageInput.Editor.SetText("")
	pg.verifyMessage.Text = ""
	pg.addressInput.SetError("")
	pg.signInput.SetError("")
}

func (pg *verifyMessagePage) validateAddress(c pageCommon) bool {
	if c.multiWallet.IsAddressValid(pg.addressInput.Editor.Text()) {
		pg.addressInput.SetError("Invalid address")
		return false
	}

	pg.addressInput.SetError("")
	return true
}

func (pg *verifyMessagePage) inputsNotEmpty(c pageCommon) bool {
	if strings.Trim(pg.addressInput.Editor.Text(), " ") == "" {
		return false
	}
	if strings.Trim(pg.messageInput.Editor.Text(), " ") == "" {
		return false
	}
	if strings.Trim(pg.signInput.Editor.Text(), " ") == "" {
		return false
	}

	pg.verifyBtn.Background = c.theme.Color.Primary
	return true
}

func (pg *verifyMessagePage) onClose() {}
