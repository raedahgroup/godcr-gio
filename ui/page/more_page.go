package page

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/planetdecred/godcr/ui/load"

	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const MorePageID = "More"

type morePageHandler struct {
	clickable *widget.Clickable
	image     *widget.Image
	page      string
}

type MorePage struct {
	*load.Load
	container         layout.Flex
	morePageListItems []morePageHandler
}

func NewMorePage(l *load.Load) *MorePage {
	morePageListItems := []morePageHandler{
		{
			clickable: new(widget.Clickable),
			image:     l.Icons.SettingsIcon,
			page:      SettingsPageID,
		},
		{
			clickable: new(widget.Clickable),
			image:     l.Icons.SecurityIcon,
			page:      SecurityToolsPageID,
		},
		{
			clickable: new(widget.Clickable),
			image:     l.Icons.HelpIcon,
			page:      HelpPageID,
		},
		{
			clickable: new(widget.Clickable),
			image:     l.Icons.AboutIcon,
			page:      AboutPageID,
		},
		{
			clickable: new(widget.Clickable),
			image:     l.Icons.DebugIcon,
			page:      DebugPageID,
		},
	}

	for i := range morePageListItems {
		morePageListItems[i].image.Scale = 1
	}

	pg := &MorePage{
		container:         layout.Flex{Axis: layout.Vertical},
		morePageListItems: morePageListItems,
		Load:              l,
	}

	return pg
}

func (pg *MorePage) OnResume() {

}

func (pg *MorePage) handleClickEvents(l *load.Load) {
	for i := range pg.morePageListItems {
		for pg.morePageListItems[i].clickable.Clicked() {
			l.ChangePage(pg.morePageListItems[i].page)
		}
	}
}

func (pg *MorePage) Layout(gtx layout.Context) layout.Dimensions {
	pg.handleClickEvents(pg.Load)

	container := func(gtx C) D {
		pg.layoutMoreItems(gtx)
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}
	return uniformPadding(gtx, container)
}

func (pg *MorePage) layoutMoreItems(gtx layout.Context) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx C) D {
			list := layout.List{Axis: layout.Vertical}
			return list.Layout(gtx, len(pg.morePageListItems), func(gtx C, i int) D {
				return layout.Inset{Bottom: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
					return decredmaterial.Clickable(gtx, pg.morePageListItems[i].clickable, func(gtx C) D {
						background := pg.Theme.Color.Surface
						card := pg.Theme.Card()
						card.Color = background
						return card.Layout(gtx, func(gtx C) D {
							gtx.Constraints.Min.X = gtx.Constraints.Max.X
							return layout.Stack{}.Layout(gtx,
								layout.Stacked(func(gtx C) D {
									return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
										gtx.Constraints.Min.X = gtx.Constraints.Max.X
										return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
											layout.Rigid(func(gtx C) D {
												return layout.Center.Layout(gtx, pg.morePageListItems[i].image.Layout)
											}),
											layout.Rigid(func(gtx C) D {
												return layout.Inset{
													Left: values.MarginPadding15,
													Top:  values.MarginPadding2,
												}.Layout(gtx, func(gtx C) D {
													return layout.Center.Layout(gtx, func(gtx C) D {
														page := pg.morePageListItems[i].page
														if page == SecurityToolsPageID {
															page = "Security Tools"
														}
														return pg.Theme.Body1(page).Layout(gtx)
													})
												})
											}),
										)
									})
								}),
							)
						})
					})
				})
			})
		}),
	)
}

func (pg *MorePage) Handle()  {}
func (pg *MorePage) OnClose() {}
