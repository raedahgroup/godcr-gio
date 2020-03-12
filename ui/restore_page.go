package ui

import (
	"fmt"
	"strings"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
)

var (
	inputGroupContainerLeft  = &layout.List{Axis: layout.Vertical}
	inputGroupContainerRight = &layout.List{Axis: layout.Vertical}
)

// RestorePage lays out the main restore page
func (win *Window) RestorePage() {
	body := func() {
		layout.Flex{Axis: layout.Vertical}.Layout(win.gtx,
			layout.Rigid(func() {
				txt := win.theme.H3("Restore from seed phrase")
				txt.Alignment = text.Middle
				txt.Layout(win.gtx)
			}),
			layout.Rigid(func() {
				txt := win.theme.H6("Enter your seed phrase in the correct order")
				txt.Alignment = text.Middle
				txt.Layout(win.gtx)
			}),
			layout.Rigid(func() {
				layout.Inset{Top: unit.Dp(20)}.Layout(win.gtx, func() {})
			}),
			layout.Flexed(1, func() {
				layout.Center.Layout(win.gtx, func() {
					layout.Flex{}.Layout(win.gtx,
						layout.Rigid(func() {
							inputsGroup(win, inputGroupContainerLeft, 16, 0)
						}),
						layout.Rigid(func() {
							inputsGroup(win, inputGroupContainerRight, 17, 16)
						}),
					)
				})
			}),
			layout.Rigid(func() {
				layout.Center.Layout(win.gtx, func() {
					layout.Inset{Top: unit.Dp(15), Bottom: unit.Dp(15)}.Layout(win.gtx, func() {
						win.outputs.restoreDiag.Layout(win.gtx, &win.inputs.restoreDiag)
					})
				})
			}),
		)
	}

	win.Page(body)
}

func inputsGroup(win *Window, l *layout.List, len int, startIndex int) {
	win.gtx.Constraints.Width.Min = win.gtx.Constraints.Width.Max / 2
	l.Layout(win.gtx, len, func(i int) {
		layout.Flex{Axis: layout.Vertical}.Layout(win.gtx,
			layout.Rigid(func() {
				layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline}.Layout(win.gtx,
					layout.Rigid(func() {
						win.theme.Label(unit.Dp(16), fmt.Sprintf("Word #%d", i+startIndex+1)).Layout(win.gtx)
					}),
					layout.Rigid(func() {
						layout.Inset{Left: unit.Dp(20), Bottom: unit.Dp(20)}.Layout(win.gtx, func() {
							win.outputs.seedEditors[i+startIndex].Layout(win.gtx, &win.inputs.seedEditors[i+startIndex])
						})
					}),
				)
			}),
			layout.Rigid(func() {
				autoComplete(win, i+startIndex, win.inputs.seedEditors[i+startIndex].Focused())
			}),
		)
	})
}

func autoComplete(win *Window, eIndex int, isFocused bool) {
	if !isFocused ||
		strings.Trim(win.inputs.seedEditors[eIndex].Text(), " ") == "" {
		return
	}

	(&layout.List{Axis: layout.Horizontal}).Layout(win.gtx, len(win.inputs.seedsSuggestionsBtn), func(i int) {
		layout.Inset{Right: unit.Dp(4)}.Layout(win.gtx, func() {
			win.outputs.seedsSuggestionsBtn[i].Layout(win.gtx, &win.inputs.seedsSuggestionsBtn[i].button)
		})
	})
}