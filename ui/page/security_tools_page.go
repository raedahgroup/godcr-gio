package page

import (
	"image"

	"github.com/planetdecred/godcr/ui/modal"

	"github.com/planetdecred/godcr/ui/load"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/planetdecred/godcr/ui/values"

	"github.com/planetdecred/godcr/ui/decredmaterial"
)

const SecurityToolsPageID = "SecurityTools"

type SecurityToolsPage struct {
	*load.Load
	verifyMessage   *widget.Clickable
	validateAddress *widget.Clickable

	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton
}

func NewSecurityToolsPage(l *load.Load) *SecurityToolsPage {
	pg := &SecurityToolsPage{
		Load:            l,
		verifyMessage:   new(widget.Clickable),
		validateAddress: new(widget.Clickable),
	}

	pg.backButton, pg.infoButton = subpageHeaderButtons(l)

	return pg
}

func (pg *SecurityToolsPage) OnResume() {

}

// main settings layout
func (pg *SecurityToolsPage) Layout(gtx layout.Context) layout.Dimensions {
	body := func(gtx C) D {
		sp := SubPage{
			Load:       pg.Load,
			title:      "Security Tools",
			backButton: pg.backButton,
			infoButton: pg.infoButton,
			back: func() {
				pg.ChangePage(MorePageID)
			},
			body: func(gtx C) D {
				return layout.Inset{Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
						layout.Flexed(.5, pg.message()),
						layout.Rigid(func(gtx C) D {
							size := image.Point{X: 15, Y: gtx.Constraints.Min.Y}
							return layout.Dimensions{Size: size}
						}),
						layout.Flexed(.5, pg.address()),
					)
				})
			},
			infoTemplate: modal.SecurityToolsInfoTemplate,
		}
		return sp.Layout(gtx)
	}
	return uniformPadding(gtx, body)
}

func (pg *SecurityToolsPage) message() layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, pg.Icons.VerifyMessageIcon, pg.verifyMessage, pg.Theme.Body1("Verify Message").Layout)
	}
}

func (pg *SecurityToolsPage) address() layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, pg.Icons.LocationPinIcon, pg.validateAddress, pg.Theme.Body1("Validate Address").Layout)
	}
}

func (pg *SecurityToolsPage) pageSections(gtx layout.Context, icon *widget.Image, action *widget.Clickable, body layout.Widget) layout.Dimensions {
	return layout.Inset{Bottom: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
		return pg.Theme.Card().Layout(gtx, func(gtx C) D {
			return decredmaterial.Clickable(gtx, action, func(gtx C) D {
				return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceAround}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							icon := icon
							icon.Scale = 1
							return icon.Layout(gtx)
						}),
						layout.Rigid(body),
						layout.Rigid(func(gtx C) D {
							size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Min.Y}
							return layout.Dimensions{Size: size}
						}),
					)
				})
			})
		})
	})
}

func (pg *SecurityToolsPage) Handle() {
	if pg.verifyMessage.Clicked() {
		pg.SetReturnPage(SecurityToolsPageID)
		pg.ChangePage(VerifyMessagePageID)
	}

	if pg.validateAddress.Clicked() {
		pg.SetReturnPage(SecurityToolsPageID)
		pg.ChangePage(ValidateAddressPageID)
	}
}

func (pg *SecurityToolsPage) OnClose() {}
