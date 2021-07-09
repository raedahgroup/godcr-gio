package page

import (
	"gioui.org/layout"

	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/values"
)

const LicensePageID = "License"

const license = `ISC License

Copyright (c) 2018-2021, Raedah Group

Permission to use, copy, modify, and/or distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.`

type LicensePage struct {
	*load.Load
	pageContainer layout.List

	backButton decredmaterial.IconButton
}

func NewLicensePage(l *load.Load) *LicensePage {
	pg := &LicensePage{
		Load:          l,
		pageContainer: layout.List{Axis: layout.Vertical},
	}
	pg.backButton, _ = subpageHeaderButtons(l)

	return pg
}

func (pg *LicensePage) OnResume() {}

//main page layout
func (pg *LicensePage) Layout(gtx layout.Context) layout.Dimensions {
	d := func(gtx C) D {
		sp := SubPage{
			Load:       pg.Load,
			title:      "License",
			backButton: pg.backButton,
			back: func() {
				pg.ChangePage(AboutPageID)
			},
			body: func(gtx C) D {
				return pg.Theme.Card().Layout(gtx, func(gtx C) D {
					return layout.UniformInset(values.MarginPadding25).Layout(gtx, func(gtx C) D {
						licenseText := pg.Theme.Body1(license)
						licenseText.Color = pg.Theme.Color.Gray
						return layout.Inset{Bottom: values.MarginPadding20}.Layout(gtx, licenseText.Layout)
					})
				})
			},
		}
		return sp.Layout(gtx)
	}
	return uniformPadding(gtx, d)
}

func (pg *LicensePage) Handle() {}

func (pg *LicensePage) OnClose() {}
