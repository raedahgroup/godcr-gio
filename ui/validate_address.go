package ui

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const ValidateAddress = "ValidateAddress"

const (
	none = iota
	valid
	invalid
	notOwned
)

type validateAddressPage struct {
	common                pageCommon
	theme                 *decredmaterial.Theme
	addressEditor         decredmaterial.Editor
	clearBtn, validateBtn decredmaterial.Button
	stateValidate         int

	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton
}

func ValidateAddressPage(common pageCommon) Page {
	pg := &validateAddressPage{
		theme:       common.theme,
		validateBtn: common.theme.Button(new(widget.Clickable), "Validate"),
		clearBtn:    common.theme.Button(new(widget.Clickable), "Clear"),
		common:      common,
	}

	pg.addressEditor = common.theme.Editor(new(widget.Editor), "Address")
	pg.addressEditor.IsRequired = false
	pg.addressEditor.Editor.SetText("")
	pg.addressEditor.Editor.SingleLine = true

	pg.validateBtn.TextSize, pg.clearBtn.TextSize = values.TextSize14, values.TextSize14
	pg.validateBtn.Background = pg.theme.Color.Primary
	pg.clearBtn.Color = pg.theme.Color.Primary
	pg.clearBtn.Background = color.NRGBA{0, 0, 0, 0}

	pg.stateValidate = none

	pg.backButton, pg.infoButton = common.SubPageHeaderButtons()

	return pg
}

func (pg *validateAddressPage) pageID() string {
	return ValidateAddress
}

func (pg *validateAddressPage) Layout(gtx layout.Context) layout.Dimensions {
	common := pg.common
	body := func(gtx C) D {
		page := SubPage{
			title: ValidateAddress,
			back: func() {
				common.popPage()
			},
			backButton: pg.backButton,
			infoButton: pg.infoButton,
			body: func(gtx C) D {
				return layout.Inset{Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
						layout.Rigid(pg.addressSection(common)),
					)
				})
			},
		}
		return common.SubPageLayout(gtx, page)
	}

	return body(gtx)
}

func (pg *validateAddressPage) addressSection(common pageCommon) layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(pg.description()),
				layout.Rigid(func(gtx C) D {
					return pg.addressEditor.Layout(gtx)
				}),
				layout.Rigid(pg.actionButtons()),
				layout.Rigid(pg.showDisplayResult(common)),
			)
		})
	}
}

func (pg *validateAddressPage) description() layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		desc := pg.theme.Caption("Enter an address to validate:")
		desc.Color = pg.theme.Color.Gray
		return layout.Inset{Bottom: values.MarginPadding20}.Layout(gtx, func(gtx C) D {
			return desc.Layout(gtx)
		})
	}
}

func (pg *validateAddressPage) actionButtons() layout.Widget {
	return func(gtx C) D {
		dims := layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Flexed(1, func(gtx C) D {
				return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Top: values.MarginPadding15}.Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								return layout.Inset{Right: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
									return pg.clearBtn.Layout(gtx)
								})
							}),
							layout.Rigid(func(gtx C) D {
								return pg.validateBtn.Layout(gtx)
							}),
						)
					})
				})
			}),
		)
		return dims
	}
}

func (pg *validateAddressPage) lineSeparator(gtx layout.Context) layout.Dimensions {
	m := values.MarginPadding10
	return layout.Inset{Top: m, Bottom: m}.Layout(gtx, func(gtx C) D {
		return pg.theme.Separator().Layout(gtx)
	})
}

func (pg *validateAddressPage) showDisplayResult(c pageCommon) layout.Widget {
	if pg.stateValidate == none {
		return func(gtx C) D {
			return layout.Dimensions{}
		}
	}
	return func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return pg.lineSeparator(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Right: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
							if pg.stateValidate == invalid {
								return c.icons.navigationCancel.Layout(gtx, values.MarginPadding25)
							}
							return c.icons.actionCheckCircle.Layout(gtx, values.MarginPadding25)
						})
					}),
					layout.Rigid(func(gtx C) D {
						if pg.stateValidate == invalid {
							txt := pg.theme.Body1("Invalid Address")
							txt.Color = pg.theme.Color.Danger
							txt.TextSize = values.TextSize16
							return txt.Layout(gtx)
						}
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								txt := pg.theme.Body1("Valid address")
								txt.Color = pg.theme.Color.Success
								txt.TextSize = values.TextSize16
								return txt.Layout(gtx)
							}),
							layout.Rigid(func(gtx C) D {
								return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										var text string
										if pg.stateValidate == valid {
											text = "Owned by you in"
										} else {
											text = "Not owned by you"
										}
										txt := pg.theme.Body1(text)
										txt.TextSize = values.TextSize14
										txt.Color = pg.theme.Color.Gray
										return txt.Layout(gtx)
									}),
									layout.Rigid(func(gtx C) D {
										if pg.stateValidate == valid {
											return layout.Inset{Left: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
												return decredmaterial.Card{
													Color: pg.theme.Color.Surface,
												}.Layout(gtx, func(gtx C) D {
													walletText := pg.theme.Caption("TODO")
													walletText.Color = pg.theme.Color.Gray
													return walletText.Layout(gtx)
												})
											})
										}
										return layout.Dimensions{}
									}),
								)
							}),
						)
					}),
				)
			}),
		)
	}
}

func (pg *validateAddressPage) pageSections(gtx layout.Context, body layout.Widget) layout.Dimensions {
	return layout.Inset{Bottom: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
		return pg.theme.Card().Layout(gtx, func(gtx C) D {
			return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceAround}.Layout(gtx,
					layout.Rigid(body),
				)
			})
		})
	})
}

func (pg *validateAddressPage) handle() {
	c := pg.common
	pg.updateColors(c)

	if pg.validateBtn.Button.Clicked() {
		pg.validateAddress()
	}

	if pg.clearBtn.Button.Clicked() {
		pg.clearPage()
	}
}

func (pg *validateAddressPage) clearPage() {
	pg.stateValidate = none
	pg.addressEditor.Editor.SetText("")
}

func (pg *validateAddressPage) validateAddress() {
	address := pg.addressEditor.Editor.Text()
	pg.addressEditor.SetError("")

	if address == "" {
		pg.addressEditor.SetError("Please enter a valid address")
		return
	}

	if address != "" {
		isValid := pg.common.multiWallet.IsAddressValid(address)
		if !isValid {
			pg.stateValidate = invalid
			return
		}

		// exist, err := pg.wallet.HaveAddress(pg.walletID, address)
		// if err != nil {
		// 	return
		// }
		// if !exist {
		// 	pg.stateValidate = notOwned
		// 	return
		// }
		// TODO
		pg.stateValidate = valid
		return
	}
}

func (pg *validateAddressPage) updateColors(common pageCommon) {
	if pg.addressEditor.Editor.Text() == "" {
		pg.validateBtn.Background = common.theme.Color.Hint
	} else {
		pg.validateBtn.Background = common.theme.Color.Primary
	}
}

func (pg *validateAddressPage) clearInputs(c *pageCommon) {
	pg.validateBtn.Background = c.theme.Color.Hint
	pg.addressEditor.Editor.SetText("")
}

func (pg *validateAddressPage) onClose() {}
