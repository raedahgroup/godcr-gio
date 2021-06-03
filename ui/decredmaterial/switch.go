// SPDX-License-Identifier: Unlicense OR MIT

package decredmaterial

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type Switch struct {
	material.SwitchStyle
}

type SwitchItem struct {
	Text   string
	button Button
}

type SwitchButtonText struct {
	t                                  *Theme
	activeTextColor, inactiveTextColor color.NRGBA
	active, inactive                   color.NRGBA
	items                              []SwitchItem
	selected                           int
}

func (t *Theme) Switch(swtch *widget.Bool) Switch {
	return Switch{material.Switch(t.Base, swtch)}
}

func (t *Theme) SwitchButtonText(i []SwitchItem) *SwitchButtonText {
	sw := &SwitchButtonText{
		t:     t,
		items: make([]SwitchItem, len(i)+1),
	}

	sw.active, sw.inactive = sw.t.Color.Surface, color.NRGBA{}
	sw.activeTextColor, sw.inactiveTextColor = sw.t.Color.DeepBlue, sw.t.Color.Gray3

	for index := range i {
		i[index].button = t.Button(new(widget.Clickable), i[index].Text)
		i[index].button.Background, i[index].button.Color = sw.inactive, sw.inactiveTextColor
		i[index].button.TextSize = unit.Sp(14)
		sw.items[index+1] = i[index]
	}

	if len(sw.items) > 0 {
		sw.selected = 1
	}
	return sw
}

func (s *SwitchButtonText) Layout(gtx layout.Context) layout.Dimensions {
	s.handleClickEvent()
	m8 := unit.Dp(8)
	m4 := unit.Dp(4)
	card := s.t.Card()
	card.Color = s.t.Color.Gray1
	card.Radius = CornerRadius{NE: 8, NW: 8, SE: 8, SW: 8}
	return card.Layout(gtx, func(gtx C) D {
		return layout.UniformInset(unit.Dp(2)).Layout(gtx, func(gtx C) D {
			list := &layout.List{Axis: layout.Horizontal}
			Items := s.items[1:]
			return list.Layout(gtx, len(Items), func(gtx C, i int) D {
				return layout.UniformInset(unit.Dp(0)).Layout(gtx, func(gtx C) D {
					index := i + 1
					btn := s.items[index].button
					btn.Inset = layout.Inset{
						Left:   m8,
						Bottom: m4,
						Right:  m8,
						Top:    m4,
					}
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(btn.Layout),
					)
				})
			})
		})
	})
}

func (s *SwitchButtonText) handleClickEvent() {
	for index := range s.items {
		if index != 0 {
			if s.items[index].button.Button.Clicked() {
				s.selected = index
			}
		}

		if s.selected == index {
			s.items[s.selected].button.Background = s.active
			s.items[s.selected].button.Color = s.activeTextColor
		} else {
			s.items[index].button.Background = s.inactive
			s.items[index].button.Color = s.inactiveTextColor
		}
	}
}

func (s *SwitchButtonText) SelectedIndex() int {
	return s.selected
}

func (s *SwitchButtonText) SelectedOption() string {
	return s.items[s.selected].Text
}
