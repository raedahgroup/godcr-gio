package page

import (
	"image/color"

	"gioui.org/text"

	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"
)

const ValidateAddressPageID = "ValidateAddress"

const (
	none = iota
	valid
	invalid
	notOwned
)

type ValidateAddressPage struct {
	*load.Load
	addressEditor         decredmaterial.Editor
	clearBtn, validateBtn decredmaterial.Button
	wallet                *wallet.Wallet
	stateValidate         int
	walletName            string

	backButton decredmaterial.IconButton
}

func NewValidateAddressPage(l *load.Load) *ValidateAddressPage {
	pg := &ValidateAddressPage{
		Load:        l,
		validateBtn: l.Theme.Button(new(widget.Clickable), "Validate"),
		clearBtn:    l.Theme.Button(new(widget.Clickable), "Clear"),
		wallet:      l.WL.Wallet,
	}

	pg.backButton, _ = subpageHeaderButtons(l)

	pg.addressEditor = l.Theme.Editor(new(widget.Editor), "Address")
	pg.addressEditor.IsRequired = false
	pg.addressEditor.Editor.SetText("")
	pg.addressEditor.Editor.SingleLine = true

	pg.validateBtn.TextSize, pg.clearBtn.TextSize = values.TextSize14, values.TextSize14
	pg.validateBtn.Background = pg.Theme.Color.Primary
	pg.validateBtn.Font.Weight = text.Bold
	pg.clearBtn.Color = pg.Theme.Color.Primary
	pg.clearBtn.Font.Weight = text.Bold
	pg.clearBtn.Background = color.NRGBA{0, 0, 0, 0}

	pg.stateValidate = none

	return pg
}

func (pg *ValidateAddressPage) OnResume() {

}

func (pg *ValidateAddressPage) Layout(gtx layout.Context) layout.Dimensions {
	body := func(gtx C) D {
		sp := SubPage{
			Load:       pg.Load,
			title:      "Validate address",
			backButton: pg.backButton,
			back: func() {
				pg.ChangePage(*pg.ReturnPage)
			},
			body: func(gtx C) D {
				return layout.Inset{Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
						layout.Rigid(pg.addressSection()),
					)
				})
			},
		}
		return sp.Layout(gtx)
	}
	return uniformPadding(gtx, body)
}

func (pg *ValidateAddressPage) addressSection() layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(pg.description()),
				layout.Rigid(func(gtx C) D {
					return pg.addressEditor.Layout(gtx)
				}),
				layout.Rigid(pg.actionButtons()),
				layout.Rigid(pg.showDisplayResult()),
			)
		})
	}
}

func (pg *ValidateAddressPage) description() layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		desc := pg.Theme.Caption("Enter an address to validate:")
		desc.Color = pg.Theme.Color.Gray
		return layout.Inset{Bottom: values.MarginPadding20}.Layout(gtx, func(gtx C) D {
			return desc.Layout(gtx)
		})
	}
}

func (pg *ValidateAddressPage) actionButtons() layout.Widget {
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

func (pg *ValidateAddressPage) lineSeparator(gtx layout.Context) layout.Dimensions {
	m := values.MarginPadding10
	return layout.Inset{Top: m, Bottom: m}.Layout(gtx, func(gtx C) D {
		return pg.Theme.Separator().Layout(gtx)
	})
}

func (pg *ValidateAddressPage) showDisplayResult() layout.Widget {
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
								return pg.Icons.NavigationCancel.Layout(gtx, values.MarginPadding25)
							}
							return pg.Icons.ActionCheckCircle.Layout(gtx, values.MarginPadding25)
						})
					}),
					layout.Rigid(func(gtx C) D {
						if pg.stateValidate == invalid {
							txt := pg.Theme.Body1("Invalid Address")
							txt.Color = pg.Theme.Color.Danger
							txt.TextSize = values.TextSize16
							return txt.Layout(gtx)
						}
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								txt := pg.Theme.Body1("Valid address")
								txt.Color = pg.Theme.Color.Success
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
										txt := pg.Theme.Body1(text)
										txt.TextSize = values.TextSize14
										txt.Color = pg.Theme.Color.Gray
										return txt.Layout(gtx)
									}),
									layout.Rigid(func(gtx C) D {
										if pg.stateValidate == valid {
											if pg.walletName != "" {
												return layout.Inset{Left: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
													return decredmaterial.Card{
														Color: pg.Theme.Color.Surface,
													}.Layout(gtx, func(gtx C) D {
														walletText := pg.Theme.Caption(pg.walletName)
														walletText.Color = pg.Theme.Color.Gray
														return walletText.Layout(gtx)
													})
												})
											}
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

func (pg *ValidateAddressPage) pageSections(gtx layout.Context, body layout.Widget) layout.Dimensions {
	return layout.Inset{Bottom: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
		return pg.Theme.Card().Layout(gtx, func(gtx C) D {
			return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceAround}.Layout(gtx,
					layout.Rigid(body),
				)
			})
		})
	})
}

func (pg *ValidateAddressPage) Handle() {
	pg.updateColors()

	if pg.validateBtn.Button.Clicked() {
		pg.validateAddress()
	}

	if pg.clearBtn.Button.Clicked() {
		pg.clearPage()
	}
}

func (pg *ValidateAddressPage) clearPage() {
	pg.stateValidate = none
	pg.addressEditor.Editor.SetText("")
}

func (pg *ValidateAddressPage) validateAddress() {
	address := pg.addressEditor.Editor.Text()
	pg.addressEditor.SetError("")

	if address == "" {
		pg.addressEditor.SetError("Please enter a valid address")
		return
	}

	if address != "" {
		isValid, _ := pg.wallet.IsAddressValid(address)
		if !isValid {
			pg.stateValidate = invalid
			return
		}

		exist, walletName := pg.wallet.HaveAddress(address)

		if !exist {
			pg.stateValidate = notOwned
			return
		}
		pg.stateValidate = valid
		pg.walletName = walletName
		return
	}
}

func (pg *ValidateAddressPage) updateColors() {
	if pg.addressEditor.Editor.Text() == "" {
		pg.validateBtn.Background = pg.Theme.Color.Hint
	} else {
		pg.validateBtn.Background = pg.Theme.Color.Primary
	}
}

func (pg *ValidateAddressPage) clearInputs() {
	pg.validateBtn.Background = pg.Theme.Color.Hint
	pg.addressEditor.Editor.SetText("")
}

func (pg *ValidateAddressPage) OnClose() {}
