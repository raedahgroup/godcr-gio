package ui

import (
	"sync"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const PageLog = "Log"

type logPage struct {
	theme  *decredmaterial.Theme
	common pageCommon

	copyLog    *widget.Clickable
	copyIcon   *widget.Image
	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton

	entriesList layout.List
	fullLog     string
	logEntries  []decredmaterial.Label
	entriesLock sync.Mutex
}

func LogPage(common pageCommon) Page {
	pg := &logPage{
		common: common,
		theme:  common.theme,
		entriesList: layout.List{
			Axis:        layout.Vertical,
			ScrollToEnd: true,
		},
		copyLog:    new(widget.Clickable),
		logEntries: make([]decredmaterial.Label, 0, 20),
	}

	pg.copyIcon = common.icons.copyIcon
	pg.copyIcon.Scale = 0.25

	pg.backButton, pg.infoButton = common.SubPageHeaderButtons()
	// go pg.watchLogs(win.internalLog) TODO

	return pg
}

func (pg *logPage) pageID() string {
	return PageLog
}

func (pg *logPage) copyLogEntries(gtx C) {
	go func() {
		pg.entriesLock.Lock()
		defer pg.entriesLock.Unlock()
		clipboard.WriteOp{Text: pg.fullLog}.Add(gtx.Ops)
	}()
}

func (pg *logPage) watchLogs(internalLog chan string) {
	for l := range internalLog {
		entry := l[:len(l)-1]
		pg.entriesLock.Lock()
		pg.fullLog += l
		pg.logEntries = append(pg.logEntries, pg.theme.Body1(entry))
		pg.entriesLock.Unlock()
	}
}

func (pg *logPage) Layout(gtx C) D {
	common := pg.common

	container := func(gtx C) D {
		page := SubPage{
			title: "Wallet log",
			back: func() {
				common.changePage(DebugPage(common))
			},
			backButton: pg.backButton,
			infoButton: pg.infoButton,
			extraItem:  pg.copyLog,
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
				background := common.theme.Color.Surface
				card := common.theme.Card()
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
		return common.SubPageLayout(gtx, page)
	}
	return container(gtx)
}

func (pg *logPage) handle()  {}
func (pg *logPage) onClose() {}
