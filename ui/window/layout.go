package window

import (
	"gioui.org/layout"
	"gioui.org/unit"

	"github.com/raedahgroup/godcr-gio/ui"
	"github.com/raedahgroup/godcr-gio/ui/page"
	"github.com/raedahgroup/godcr-gio/ui/themes/materialplus"
)

const (
	navWidth = 200
)

func (win *Window) layoutPage(gtx *layout.Context, pg page.Page) interface{} {
	if !pg.IsNavPage {
		return pg.Handler.Draw(gtx)
	}

	var evt interface{}
	layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func() {
			gtx.Constraints.Width.Min = navWidth
			materialplus.Fill(gtx, ui.LightBlueColor, navWidth, gtx.Constraints.Height.Max)
			win.layoutNavSection(gtx)
		}),
		layout.Rigid(func() {
			evt = pg.Handler.Draw(gtx)
		}),
	)

	return evt
}

func (win *Window) layoutNavSection(gtx *layout.Context) {
	w := []func(){}

	for i := range win.pages {
		page := win.pages[i]
		if page.IsNavPage {
			id := page.ID

			fn := func() {
				for page.Button.Clicked(gtx) {
					if win.current != id {
						win.current = id
					}
				}
				btn := win.theme.Button(id)
				btn.Layout(gtx, page.Button)
			}
			w = append(w, fn)
		}
	}

	list := layout.List{Axis: layout.Vertical}
	list.Layout(gtx, len(w), func(i int) {
		layout.UniformInset(unit.Dp(0)).Layout(gtx, w[i])
	})
}
