package page

import (
	"image/color"

	"gioui.org/unit"

	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/modal"
	"github.com/planetdecred/godcr/ui/values"
)

// pages that haven't been migrated
// todo: to be removed when the page is migrated but their IDs are need for page navigation
const PagePrivacy = "Privacy"

type (
	C = layout.Context
	D = layout.Dimensions
)

// Container is simply a wrapper for the Inset type. Its purpose is to differentiate the use of an inset as a padding or
// margin, making it easier to visualize the structure of a layout when reading UI code.
type Container struct {
	padding layout.Inset
}

func (c Container) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	return c.padding.Layout(gtx, w)
}

var (
	MaxWidth = unit.Dp(800)
)

type SubPage struct {
	*load.Load
	title        string
	subTitle     string
	walletName   string
	back         func()
	body         layout.Widget
	infoTemplate string
	extraItem    *widget.Clickable
	extra        layout.Widget
	extraText    string
	handleExtra  func()

	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton
}

func subpageHeaderButtons(l *load.Load) (decredmaterial.IconButton, decredmaterial.IconButton) {
	backButton := l.Theme.PlainIconButton(new(widget.Clickable), l.Icons.NavigationArrowBack)
	infoButton := l.Theme.PlainIconButton(new(widget.Clickable), l.Icons.ActionInfo)

	zeroInset := layout.UniformInset(values.MarginPadding0)
	backButton.Color, infoButton.Color = l.Theme.Color.Gray3, l.Theme.Color.Gray3

	m25 := values.MarginPadding25
	backButton.Size, infoButton.Size = m25, m25
	backButton.Inset, infoButton.Inset = zeroInset, zeroInset

	return backButton, infoButton
}

func (sp *SubPage) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: values.MarginPadding15}.Layout(gtx, func(gtx C) D {
				return sp.Header(gtx)
			})
		}),
		layout.Rigid(sp.body),
	)
}

func (sp *SubPage) Header(gtx layout.Context) layout.Dimensions {
	sp.EventHandler()

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Right: values.MarginPadding20}.Layout(gtx, sp.backButton.Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if sp.subTitle == "" {
				return sp.Theme.H6(sp.title).Layout(gtx)
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(sp.Theme.H6(sp.title).Layout),
				layout.Rigid(sp.Theme.Body1(sp.subTitle).Layout),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if sp.walletName != "" {
				return layout.Inset{Left: values.MarginPadding5, Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
					return decredmaterial.Card{
						Color: sp.Theme.Color.Surface,
					}.Layout(gtx, func(gtx C) D {
						return layout.UniformInset(values.MarginPadding2).Layout(gtx, func(gtx C) D {
							walletText := sp.Theme.Caption(sp.walletName)
							walletText.Color = sp.Theme.Color.Gray
							return walletText.Layout(gtx)
						})
					})
				})
			}
			return layout.Dimensions{}
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.E.Layout(gtx, func(gtx C) D {
				if sp.infoTemplate != "" {
					return sp.infoButton.Layout(gtx)
				} else if sp.extraItem != nil {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if sp.extraText != "" {
								return layout.Inset{Right: values.MarginPadding10, Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
									text := sp.Theme.Caption(sp.extraText)
									text.Color = sp.Theme.Color.DeepBlue
									return text.Layout(gtx)
								})
							}
							return layout.Dimensions{}
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return decredmaterial.Clickable(gtx, sp.extraItem, sp.extra)
						}),
					)
				}
				return layout.Dimensions{}
			})
		}),
	)
}

func (sp *SubPage) SplitLayout(gtx layout.Context) layout.Dimensions {
	card := sp.Theme.Card()
	card.Color = color.NRGBA{}
	return card.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D { return sp.Header(gtx) }),
			layout.Rigid(sp.body),
		)
	})
}

func (sp *SubPage) EventHandler() {
	if sp.infoTemplate != "" {
		if sp.infoButton.Button.Clicked() {
			modal.NewInfoModal(sp.Load).
				Title(sp.title).
				SetupWithTemplate(sp.infoTemplate).
				NegativeButton("Got it", func() {}).Show()
		}
	}

	if sp.backButton.Button.Clicked() {
		sp.back()
	}

	if sp.extraItem != nil && sp.extraItem.Clicked() {
		sp.handleExtra()
	}
}

func uniformPadding(gtx layout.Context, body layout.Widget) layout.Dimensions {
	width := gtx.Constraints.Max.X

	padding := values.MarginPadding24

	if (width - 2*gtx.Px(padding)) > gtx.Px(MaxWidth) {
		paddingValue := float32(width-gtx.Px(MaxWidth)) / 2
		padding = unit.Px(paddingValue)
	}

	return layout.Inset{
		Top:    values.MarginPadding24,
		Right:  padding,
		Bottom: values.MarginPadding24,
		Left:   padding,
	}.Layout(gtx, body)
}

// todo: this method will be removed when the new modal implementation is used on the seedbackup page
func _modal(gtx layout.Context, body layout.Dimensions, modal layout.Dimensions) layout.Dimensions {
	dims := layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			return body
		}),
		layout.Stacked(func(gtx C) D {
			return modal
		}),
	)
	return dims
}
