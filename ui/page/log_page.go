package page

import (
	"sync"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/values"
)

const LogPageID = "Log"

type LogPage struct {
	*load.Load

	internalLog chan string
	copyLog     *widget.Clickable
	copyIcon    *widget.Image
	backButton  decredmaterial.IconButton

	entriesList layout.List
	fullLog     string
	logEntries  []decredmaterial.Label
	entriesLock sync.Mutex
}

func NewLogPage(l *load.Load) *LogPage {
	pg := &LogPage{
		Load:        l,
		internalLog: l.Receiver.InternalLog,
		entriesList: layout.List{
			Axis:        layout.Vertical,
			ScrollToEnd: true,
		},
		copyLog:    new(widget.Clickable),
		logEntries: make([]decredmaterial.Label, 0, 20),
	}

	pg.copyIcon = pg.Icons.CopyIcon
	pg.copyIcon.Scale = 0.25

	pg.backButton, _ = subpageHeaderButtons(l)

	go pg.watchLogs(pg.internalLog)

	return pg
}

func (pg *LogPage) OnResume() {

}

func (pg *LogPage) copyLogEntries(gtx C) {
	go func() {
		pg.entriesLock.Lock()
		defer pg.entriesLock.Unlock()
		clipboard.WriteOp{Text: pg.fullLog}.Add(gtx.Ops)
	}()
}

func (pg *LogPage) watchLogs(internalLog chan string) {
	for l := range internalLog {
		entry := l[:len(l)-1]
		pg.entriesLock.Lock()
		pg.fullLog += l
		pg.logEntries = append(pg.logEntries, pg.Theme.Body1(entry))
		pg.entriesLock.Unlock()
	}
}

func (pg *LogPage) Layout(gtx C) D {
	container := func(gtx C) D {
		sp := SubPage{
			Load:       pg.Load,
			title:      "Wallet log",
			backButton: pg.backButton,
			back: func() {
				pg.ChangePage(DebugPageID)
			},
			extraItem: pg.copyLog,
			extra: func(gtx C) D {
				return layout.Center.Layout(gtx, func(gtx C) D {
					return decredmaterial.Clickable(gtx, pg.copyLog, func(gtx C) D {
						sz := gtx.Constraints.Max.X
						pg.copyIcon.Scale = float32(sz) / float32(gtx.Px(unit.Dp(float32(sz))))
						return pg.copyIcon.Layout(gtx)
					})

				})
			},
			handleExtra: func() {
				pg.copyLogEntries(gtx)
			},
			body: func(gtx C) D {
				background := pg.Theme.Color.Surface
				card := pg.Theme.Card()
				card.Color = background
				return card.Layout(gtx, func(gtx C) D {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
					return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
						return pg.entriesList.Layout(gtx, len(pg.logEntries), func(gtx C, i int) D {
							pg.entriesLock.Lock()
							defer pg.entriesLock.Unlock()
							return pg.logEntries[i].Layout(gtx)
						})
					})

				})
			},
		}
		return sp.Layout(gtx)
	}
	return uniformPadding(gtx, container)
}

func (pg *LogPage) Handle()  {}
func (pg *LogPage) OnClose() {}
