package page

import (
	"image"

	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/values"
)

const HelpPageID = "Help"

type HelpPage struct {
	*load.Load
	documentation *widget.Clickable

	backButton decredmaterial.IconButton
}

func NewHelpPage(l *load.Load) *HelpPage {
	pg := &HelpPage{
		Load:          l,
		documentation: new(widget.Clickable),
	}

	pg.backButton, _ = subpageHeaderButtons(l)

	return pg
}

func (pg *HelpPage) OnResume() {

}

// main settings layout
func (pg *HelpPage) Layout(gtx layout.Context) layout.Dimensions {
	body := func(gtx C) D {
		sp := SubPage{
			Load:       pg.Load,
			title:      "Help",
			subTitle:   "For more information, please visit the Decred documentation.",
			backButton: pg.backButton,
			back: func() {
				pg.ChangePage(MorePageID)
			},
			body: func(gtx C) D {
				return layout.Inset{Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Spacing: layout.SpaceBetween, WeightSum: 2}.Layout(gtx,
						layout.Flexed(1, pg.document()),
					)
				})
			},
		}
		return sp.Layout(gtx)
	}
	return uniformPadding(gtx, body)
}

func (pg *HelpPage) document() layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, pg.Icons.DocumentationIcon, pg.documentation, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(pg.Theme.Body1("Documentation").Layout),
			)
		})
	}
}

func (pg *HelpPage) pageSections(gtx layout.Context, icon *widget.Image, action *widget.Clickable, body layout.Widget) layout.Dimensions {
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

func (pg *HelpPage) Handle() {
	if pg.documentation.Clicked() {
		goToURL("https://docs.decred.org")
	}
}

func (pg *HelpPage) OnClose() {}
