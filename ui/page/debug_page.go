package page

import (
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/values"
)

const DebugPageID = "Debug"

type debugItem struct {
	clickable *widget.Clickable
	text      string
	page      string
}

type DebugPage struct {
	*load.Load
	debugItems []debugItem

	backButton decredmaterial.IconButton
}

func NewDebugPage(l *load.Load) *DebugPage {
	debugItems := []debugItem{
		{
			clickable: new(widget.Clickable),
			text:      "Check wallet logs",
			page:      LogPageID,
		},
		{
			clickable: new(widget.Clickable),
			text:      "Check statistics",
			page:      StatisticsPageID,
		},
	}

	pg := &DebugPage{
		Load:       l,
		debugItems: debugItems,
	}

	pg.backButton, _ = subpageHeaderButtons(l)

	return pg
}

func (pg *DebugPage) OnResume() {

}

func (pg *DebugPage) Handle() {
	for i := range pg.debugItems {
		for pg.debugItems[i].clickable.Clicked() {
			pg.ChangePage(pg.debugItems[i].page)
		}
	}
}

func (pg *DebugPage) OnClose() {}

func (pg *DebugPage) debugItem(gtx C, i int) D {
	return decredmaterial.Clickable(gtx, pg.debugItems[i].clickable, func(gtx C) D {
		return layout.Flex{}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.UniformInset(values.MarginPadding15).Layout(gtx, pg.Theme.Body1(pg.debugItems[i].text).Layout)
			}),
			layout.Flexed(1, func(gtx C) D {
				return layout.E.Layout(gtx, func(gtx C) D {
					return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
						return pg.Icons.ChevronRight.Layout(gtx, values.MarginPadding22)
					})
				})
			}),
		)
	})
}

func (pg *DebugPage) layoutDebugItems(gtx C) {
	background := pg.Theme.Color.Surface
	card := pg.Theme.Card()
	card.Color = background
	card.Layout(gtx, func(gtx C) D {
		list := layout.List{Axis: layout.Vertical}
		return list.Layout(gtx, len(pg.debugItems), func(gtx C, i int) D {
			return pg.debugItem(gtx, i)
		})
	})
}

func (pg *DebugPage) Layout(gtx C) D {
	container := func(gtx C) D {
		sp := SubPage{
			Load:       pg.Load,
			title:      "Debug",
			backButton: pg.backButton,
			back: func() {
				pg.ChangePage(MorePageID)
			},
			body: func(gtx C) D {
				pg.layoutDebugItems(gtx)
				return layout.Dimensions{Size: gtx.Constraints.Max}
			},
		}
		return sp.Layout(gtx)

	}
	return uniformPadding(gtx, container)
}
