// SPDX-License-Identifier: Unlicense OR MIT

package decredmaterial

import (
	"image"
	"image/color"

	"gioui.org/f32"
	// "gioui.org/internal/f32color"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	// "gioui.org/widget/material"
)

type Switch struct {
	Enabled    color.NRGBA
	Disabled   color.NRGBA
	thumbColor color.NRGBA
	Switch     *widget.Bool
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

func (t *Theme) Switch(swtch *widget.Bool) *Switch {
	sw := &Switch{
		Switch: swtch,
	}
	sw.Enabled, sw.Disabled, sw.thumbColor = t.Color.Primary, t.Color.InactiveGray, t.Color.Surface
	return sw
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

func (s *Switch) Layout(gtx layout.Context) layout.Dimensions {
	trackWidth := gtx.Px(unit.Dp(32))
	trackHeight := gtx.Px(unit.Dp(20))
	thumbSize := gtx.Px(unit.Dp(18))
	trackOff := float32(thumbSize-trackHeight) * .5

	// Draw track.
	stack := op.Save(gtx.Ops)
	trackCorner := float32(trackHeight) / 2
	trackRect := f32.Rectangle{Max: f32.Point{
		X: float32(trackWidth),
		Y: float32(trackHeight),
	}}
	col := s.Disabled
	if s.Switch.Value {
		col = s.Enabled
	}

	trackColor := col
	op.Offset(f32.Point{Y: trackOff}).Add(gtx.Ops)
	clip.UniformRRect(trackRect, trackCorner).Add(gtx.Ops)
	paint.ColorOp{Color: trackColor}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	stack.Load()

	// Compute thumb offset and color.
	stack = op.Save(gtx.Ops)
	if s.Switch.Value {
		off := trackWidth - thumbSize
		op.Offset(f32.Point{X: float32(off)}).Add(gtx.Ops)
	}

	thumbRadius := float32(thumbSize) / 2

	// Draw thumb shadow, a translucent disc slightly larger than the
	// thumb itself.
	// Center shadow horizontally and slightly adjust its Y.
	paint.FillShape(gtx.Ops, col,
		clip.Circle{
			Center: f32.Point{X: thumbRadius, Y: thumbRadius + .25},
			Radius: thumbRadius + 1,
		}.Op(gtx.Ops))

	// Draw thumb.
	paint.FillShape(gtx.Ops, s.thumbColor,
		clip.Circle{
			Center: f32.Point{X: thumbRadius, Y: thumbRadius},
			Radius: thumbRadius,
		}.Op(gtx.Ops))

	// Set up click area.
	stack = op.Save(gtx.Ops)
	clickSize := gtx.Px(unit.Dp(40))
	clickOff := f32.Point{
		X: (float32(trackWidth) - float32(clickSize)) * .5,
		Y: (float32(trackHeight)-float32(clickSize))*.5 + trackOff,
	}
	op.Offset(clickOff).Add(gtx.Ops)
	sz := image.Pt(clickSize, clickSize)
	pointer.Ellipse(image.Rectangle{Max: sz}).Add(gtx.Ops)
	gtx.Constraints.Min = sz
	s.Switch.Layout(gtx)
	stack.Load()

	dims := image.Point{X: trackWidth, Y: thumbSize}
	return layout.Dimensions{Size: dims}
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

func (s *SwitchButtonText) SelectedOption() string {
	return s.items[s.selected].Text
}
