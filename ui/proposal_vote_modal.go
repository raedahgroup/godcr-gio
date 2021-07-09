package ui

import (
	"fmt"
	"image/color"
	"strconv"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const ModalInputVote = "input_vote_modal"

type inputVoteOptionsWidgets struct {
	label      string
	background color.NRGBA
	input      decredmaterial.Editor
	increment  decredmaterial.IconButton
	decrement  decredmaterial.IconButton
	max        decredmaterial.Button
}

type voteModal struct {
	*pageCommon
	randomID       string
	modal          decredmaterial.Modal
	passwordEditor decredmaterial.Editor
	callback       func(password string, m *voteModal) bool // return true to dismiss dialog
	btnPositve     decredmaterial.Button
	btnNegative    decredmaterial.Button
	yesVote        inputVoteOptionsWidgets
	noVote         inputVoteOptionsWidgets
}

func newInputVoteOptions(c *pageCommon, label string) inputVoteOptionsWidgets {
	i := inputVoteOptionsWidgets{
		label:      label,
		background: c.theme.Color.LightGray,
		input:      c.theme.Editor(new(widget.Editor), ""),
		increment:  c.theme.PlainIconButton(new(widget.Clickable), c.icons.contentAdd),
		decrement:  c.theme.PlainIconButton(new(widget.Clickable), c.icons.contentRemove),
		max:        c.theme.Button(new(widget.Clickable), "MAX"),
	}
	i.max.Background = c.theme.Color.Surface
	i.max.Color = c.theme.Color.Gray2
	i.max.Font.Weight = text.Bold

	i.increment.Color, i.decrement.Color = c.theme.Color.Text, c.theme.Color.Text
	i.increment.Size, i.decrement.Size = values.TextSize18, values.TextSize18
	i.input.Bordered = false
	i.input.Editor.SetText("0")
	i.input.Editor.Alignment = text.Middle
	return i
}

func newvoteModal(common *pageCommon) *voteModal {
	cm := &voteModal{
		pageCommon:  common,
		randomID:    fmt.Sprintf("%s-%d", ModalInputVote, generateRandomNumber()),
		modal:       *common.theme.ModalFloatTitle(),
		btnPositve:  common.theme.Button(new(widget.Clickable), "Vote"),
		btnNegative: common.theme.Button(new(widget.Clickable), "Cancel"),
	}

	cm.btnPositve.TextSize, cm.btnNegative.TextSize = values.TextSize16, values.TextSize16
	cm.btnPositve.Font.Weight, cm.btnNegative.Font.Weight = text.Bold, text.Bold
	cm.btnPositve.Background = common.theme.Color.Gray1
	cm.btnPositve.Color = common.theme.Color.Surface

	cm.passwordEditor = common.theme.EditorPassword(new(widget.Editor), "Spending password")
	cm.passwordEditor.Editor.SingleLine, cm.passwordEditor.Editor.Submit = true, true

	cm.yesVote = newInputVoteOptions(common, "Yes")
	cm.yesVote.background = common.theme.Color.Success2
	cm.noVote = newInputVoteOptions(common, "No")
	return cm
}

func (cm *voteModal) ModalID() string {
	return cm.randomID
}

func (cm *voteModal) OnResume() {
}

func (cm *voteModal) OnDismiss() {

}

func (cm *voteModal) Show() {
	cm.showModal(cm)
}

func (cm *voteModal) Dismiss() {
	cm.dismissModal(cm)
}

func (i *inputVoteOptionsWidgets) handleVoteCountButtons() {
	if i.increment.Button.Clicked() {
		value, err := strconv.Atoi(i.input.Editor.Text())
		if err != nil {
			log.Error(err)
			return
		}
		value++
		i.input.Editor.SetText(fmt.Sprintf("%d", value))
	}

	if i.decrement.Button.Clicked() {
		value, err := strconv.Atoi(i.input.Editor.Text())
		if err != nil {
			log.Error(err)
			return
		}
		value--
		if value < 0 {
			return
		}
		i.input.Editor.SetText(fmt.Sprintf("%d", value))
	}

	if i.max.Button.Clicked() {
		i.input.Editor.SetText("5")
	}
}

func (cm *voteModal) Handle() {
	if cm.btnNegative.Button.Clicked() {
		cm.Dismiss()
	}

	cm.yesVote.handleVoteCountButtons()
	cm.noVote.handleVoteCountButtons()
}

func (cm *voteModal) Layout(gtx layout.Context) D {
	w := []layout.Widget{
		func(gtx C) D {
			t := cm.theme.H6("Vote")
			t.Font.Weight = text.Bold
			return t.Layout(gtx)
		},
		func(gtx C) D {
			return cm.theme.Label(values.TextSize16, "You have 5 votes").Layout(gtx)
		},

		func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return cm.inputOptions(gtx, &cm.yesVote)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top: values.MarginPadding10,
					}.Layout(gtx, func(gtx C) D {
						return cm.inputOptions(gtx, &cm.noVote)
					})
				}),
			)
		},

		func(gtx C) D {
			return cm.passwordEditor.Layout(gtx)
		},

		func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {

						cm.btnNegative.Background = cm.theme.Color.Surface
						cm.btnNegative.Color = cm.theme.Color.Primary
						return cm.btnNegative.Layout(gtx)
					}),
					layout.Rigid(func(gtx C) D {
						return cm.btnPositve.Layout(gtx)
					}),
				)
			})
		},
	}

	return cm.modal.Layout(gtx, w, 850)
}

func (cm *voteModal) inputOptions(gtx layout.Context, wdg *inputVoteOptionsWidgets) D {
	wrap := cm.theme.Card()
	wrap.Color = wdg.background
	return wrap.Layout(gtx, func(gtx C) D {
		inset := layout.Inset{
			Top:    values.MarginPadding8,
			Bottom: values.MarginPadding8,
			Left:   values.MarginPadding16,
			Right:  values.MarginPadding8,
		}
		return inset.Layout(gtx, func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(.4, func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							icon := cm.icons.imageBrightness1
							icon.Color = cm.theme.Color.Success
							return icon.Layout(gtx, values.MarginPadding8)
						}),
						layout.Rigid(func(gtx C) D {
							return layout.Inset{Left: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
								label := cm.theme.Body2(wdg.label)
								return label.Layout(gtx)
							})
						}),
					)
				}),
				layout.Flexed(.6, func(gtx C) D {
					border := widget.Border{
						Color:        cm.theme.Color.Gray1,
						CornerRadius: values.MarginPadding8,
						Width:        values.MarginPadding2,
					}

					return border.Layout(gtx, func(gtx C) D {
						card := cm.theme.Card()
						card.Color = cm.theme.Color.Surface
						return card.Layout(gtx, func(gtx C) D {
							var height int
							gtx.Constraints.Min.X = gtx.Constraints.Max.X
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Flexed(1, func(gtx C) D {
									dims := layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											return wdg.decrement.Layout(gtx)
										}),
										layout.Rigid(func(gtx C) D {
											gtx.Constraints.Min.X, gtx.Constraints.Max.X = 100, 100
											return wdg.input.Layout(gtx)
										}),
										layout.Rigid(func(gtx C) D {
											return wdg.increment.Layout(gtx)
										}),
									)
									height = dims.Size.Y
									return dims
								}),
								layout.Flexed(0.02, func(gtx C) D {
									line := cm.theme.Line(height, gtx.Px(values.MarginPadding2))
									line.Color = cm.theme.Color.Gray1
									return line.Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									return wdg.max.Layout(gtx)
								}),
							)
						})
					})
				}),
			)
		})
	})
}
