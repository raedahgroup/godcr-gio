package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const PageHelp = "Help"

type helpPage struct {
	theme         *decredmaterial.Theme
	documentation *widget.Clickable
	common        pageCommon

	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton
}

func HelpPage(common pageCommon) Page {
	pg := &helpPage{
		theme:         common.theme,
		documentation: new(widget.Clickable),
		common:        common,
	}

	pg.backButton, pg.infoButton = common.SubPageHeaderButtons()

	return pg
}

func (pg *helpPage) pageID() string {
	return PageHelp
}

// main settings layout
func (pg *helpPage) Layout(gtx layout.Context) layout.Dimensions {
	body := func(gtx C) D {
		page := SubPage{
			title:    "Help",
			subTitle: "For more information, please visit the Decred documentation.",
			back: func() {
				pg.common.popPage()
			},
			backButton: pg.backButton,
			infoButton: pg.infoButton,
			body: func(gtx C) D {
				return layout.Inset{Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Spacing: layout.SpaceBetween, WeightSum: 2}.Layout(gtx,
						layout.Flexed(1, pg.document(pg.common)),
					)
				})
			},
		}
		return pg.common.SubPageLayout(gtx, page)
	}
	return pg.common.UniformPadding(gtx, body)
}

func (pg *helpPage) document(common pageCommon) layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, common.icons.documentationIcon, pg.documentation, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(common.theme.Body1("Documentation").Layout),
			)
		})
	}
}

func (pg *helpPage) pageSections(gtx layout.Context, icon *widget.Image, action *widget.Clickable, body layout.Widget) layout.Dimensions {
	return layout.Inset{Bottom: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
		return pg.theme.Card().Layout(gtx, func(gtx C) D {
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

func (pg *helpPage) handle()  {}
func (pg *helpPage) onClose() {}
