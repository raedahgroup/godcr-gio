package page

import (
	"bytes"
	"image"
	"image/color"
	"path/filepath"
	"time"

	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/modal"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
	qrcode "github.com/yeqown/go-qrcode"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

const ReceivePageID = "Receive"

type ReceivePage struct {
	*load.Load
	multiWallet       *dcrlibwallet.MultiWallet
	pageContainer     layout.List
	isNewAddr, isInfo bool
	currentAddress    string
	qrImage           *image.Image
	newAddr, copy     decredmaterial.Button
	info, more        decredmaterial.IconButton
	card              decredmaterial.Card
	receiveAddress    decredmaterial.Label
	gtx               *layout.Context

	selector *accountSelector

	backdrop   *widget.Clickable
	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton
}

func NewReceivePage(l *load.Load) *ReceivePage {
	pg := &ReceivePage{
		Load:        l,
		multiWallet: l.WL.MultiWallet,
		pageContainer: layout.List{
			Axis: layout.Vertical,
		},
		info:           l.Theme.IconButton(new(widget.Clickable), mustIcon(widget.NewIcon(icons.ActionInfo))),
		copy:           l.Theme.Button(new(widget.Clickable), "Copy"),
		more:           l.Theme.PlainIconButton(new(widget.Clickable), l.Icons.NavMoreIcon),
		newAddr:        l.Theme.Button(new(widget.Clickable), "Generate new address"),
		receiveAddress: l.Theme.Label(values.TextSize20, ""),
		card:           l.Theme.Card(),
		backdrop:       new(widget.Clickable),
	}

	pg.info.Inset, pg.info.Size = layout.UniformInset(values.MarginPadding5), values.MarginPadding20
	pg.copy.Background = color.NRGBA{}
	pg.copy.Color = pg.Theme.Color.Primary
	pg.copy.Inset = layout.Inset{
		Top:    values.MarginPadding19p5,
		Bottom: values.MarginPadding19p5,
		Left:   values.MarginPadding16,
		Right:  values.MarginPadding16,
	}
	pg.more.Color = pg.Theme.Color.Gray3
	pg.more.Inset = layout.UniformInset(values.MarginPadding0)
	pg.newAddr.Inset = layout.Inset{
		Top:    values.MarginPadding20,
		Bottom: values.MarginPadding20,
		Left:   values.MarginPadding16,
		Right:  values.MarginPadding16,
	}
	pg.newAddr.Color = pg.Theme.Color.Text
	pg.newAddr.Background = pg.Theme.Color.Surface
	pg.newAddr.TextSize = values.TextSize16

	pg.backButton, pg.infoButton = subpageHeaderButtons(l)
	pg.backButton.Icon = pg.Icons.ContentClear

	pg.selector = newAccountSelector(pg.Load).
		title("Receiving account").
		accountSelected(func(selectedAccount *dcrlibwallet.Account) {
			selectedWallet := pg.multiWallet.WalletWithID(selectedAccount.WalletID)
			currentAddress, err := selectedWallet.CurrentAddress(selectedAccount.Number)
			if err != nil {
				log.Errorf("Error getting current address: %v", err)
			} else {
				pg.currentAddress = currentAddress
			}

			pg.generateQRForAddress()
		}).
		accountValidator(func(account *dcrlibwallet.Account) bool {

			// Filter out imported account and mixed.
			wal := pg.multiWallet.WalletWithID(account.WalletID)
			if account.Number == MaxInt32 ||
				account.Number == wal.MixedAccountNumber() {
				return false
			}
			return true
		})

	return pg
}

func (pg *ReceivePage) OnResume() {
	pg.selector.selectFirstWalletValidAccount()
}

func (pg *ReceivePage) generateQRForAddress() {
	absoluteWdPath, err := GetAbsolutePath()
	if err != nil {
		log.Error(err.Error())
	}

	opt := qrcode.WithLogoImageFilePNG(filepath.Join(absoluteWdPath, "ui/assets/decredicons/qrcodeSymbol.png"))
	qrCode, err := qrcode.New(pg.currentAddress, opt)
	if err != nil {
		log.Error("Error generating address qrCode: " + err.Error())
		return
	}

	var buff bytes.Buffer
	err = qrCode.SaveTo(&buff)
	if err != nil {
		log.Error(err.Error())
		return
	}

	imgdec, _, err := image.Decode(bytes.NewReader(buff.Bytes()))
	if err != nil {
		log.Error(err.Error())
		return
	}

	pg.qrImage = &imgdec
}

func (pg *ReceivePage) Layout(gtx layout.Context) layout.Dimensions {
	if pg.gtx == nil {
		pg.gtx = &gtx
	}
	pg.pageBackdropLayout(gtx)

	pageContent := []func(gtx C) D{
		func(gtx C) D {
			return pg.pageSections(gtx, func(gtx C) D {
				return pg.selector.Layout(gtx)
			})
		},
		func(gtx C) D {
			return pg.Theme.Separator().Layout(gtx)
		},
		func(gtx C) D {
			return pg.pageSections(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return pg.titleLayout(gtx)
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Center.Layout(gtx, func(gtx C) D {
							return layout.Flex{
								Axis:      layout.Vertical,
								Alignment: layout.Middle,
							}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									if pg.currentAddress != "" {
										return pg.addressLayout(gtx)
									}
									return layout.Dimensions{}
								}),
								layout.Rigid(func(gtx C) D {
									if pg.qrImage == nil {
										return layout.Dimensions{}
									}

									return pg.Theme.ImageIcon(gtx, *pg.qrImage, 360)
								}),
							)
						})
					}),
				)
			})
		},
	}

	dims := uniformPadding(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Bottom: values.MarginPadding16}.Layout(gtx, func(gtx C) D {
					return pg.topNav(gtx)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return pg.Theme.Card().Layout(gtx, func(gtx C) D {
					return pg.pageContainer.Layout(gtx, len(pageContent), func(gtx C, i int) D {
						return pageContent[i](gtx)
					})
				})
			}),
		)
	})

	return dims
}

func (pg *ReceivePage) pageSections(gtx layout.Context, body layout.Widget) layout.Dimensions {
	return pg.Theme.Card().Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return layout.UniformInset(values.MarginPadding16).Layout(gtx, body)
	})
}

// pageBackdropLayout layout of background overlay when the popup button generate new address is show,
// click outside of the generate new address button to hide the button
func (pg *ReceivePage) pageBackdropLayout(gtx layout.Context) {
	if pg.isNewAddr {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
		m := op.Record(gtx.Ops)
		pg.backdrop.Layout(gtx)
		op.Defer(gtx.Ops, m.Stop())
	}
}

func (pg *ReceivePage) topNav(gtx layout.Context) layout.Dimensions {
	m := values.MarginPadding20
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return pg.backButton.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: m}.Layout(gtx, pg.Theme.H6("Receive DCR").Layout)
				}),
			)
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, pg.infoButton.Layout)
		}),
	)
}

func (pg *ReceivePage) titleLayout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			txt := pg.Theme.Body2("Your Address")
			txt.Color = pg.Theme.Color.Gray
			return txt.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					if pg.isNewAddr {
						m := op.Record(gtx.Ops)
						layout.Inset{Top: values.MarginPadding30, Left: unit.Dp(-152)}.Layout(gtx, func(gtx C) D {
							return pg.newAddr.Layout(gtx)
						})
						op.Defer(gtx.Ops, m.Stop())
					}
					return layout.Dimensions{}
				}),
				layout.Rigid(pg.more.Layout),
			)
		}),
	)
}

func (pg *ReceivePage) addressLayout(gtx layout.Context) layout.Dimensions {
	card := decredmaterial.Card{
		Inset: layout.Inset{
			Top:    values.MarginPadding14,
			Bottom: values.MarginPadding16,
		},
		Color: pg.Theme.Color.LightGray,
	}

	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			card.Radius = decredmaterial.CornerRadius{NE: 8, NW: 0, SE: 0, SW: 8}
			return card.Layout(gtx, func(gtx C) D {
				return layout.Inset{
					Top:    values.MarginPadding30,
					Bottom: values.MarginPadding30,
					Left:   values.MarginPadding30,
					Right:  values.MarginPadding30,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{}
				})
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			pg.receiveAddress.Text = pg.currentAddress
			pg.receiveAddress.Color = pg.Theme.Color.DeepBlue
			pg.receiveAddress.Alignment = text.Middle
			pg.receiveAddress.MaxLines = 1
			return card.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return layout.UniformInset(values.MarginPadding16).Layout(gtx, pg.receiveAddress.Layout)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Left: values.MarginPadding1}.Layout(gtx, func(gtx C) D { return layout.Dimensions{} })
		}),
		layout.Rigid(func(gtx C) D {
			card.Radius = decredmaterial.CornerRadius{NE: 0, NW: 8, SE: 8, SW: 0}
			return card.Layout(gtx, pg.copy.Layout)
		}),
	)
}

func (pg *ReceivePage) Handle() {
	gtx := pg.gtx
	if pg.backdrop.Clicked() {
		pg.isNewAddr = false
	}

	if pg.more.Button.Clicked() {
		pg.isNewAddr = !pg.isNewAddr
		if pg.isInfo {
			pg.isInfo = false
		}
	}

	if pg.newAddr.Button.Clicked() {
		newAddr, err := pg.generateNewAddress()
		if err != nil {
			log.Debug("Error generating new address" + err.Error())
			return
		}

		pg.currentAddress = newAddr
		pg.generateQRForAddress()
		pg.isNewAddr = false
	}

	if pg.infoButton.Button.Clicked() {
		info := modal.NewInfoModal(pg.Load).
			Title("Receive DCR").
			Body("Each time you receive a payment, a new address is generated to protect your privacy.").
			PositiveButton("Got it", func() {})
		pg.ShowModal(info)
	}

	if pg.backButton.Button.Clicked() {
		pg.ChangePage(*pg.ReturnPage)
	}

	if pg.copy.Button.Clicked() {

		clipboard.WriteOp{Text: pg.currentAddress}.Add(gtx.Ops)

		pg.copy.Text = "Copied!"
		pg.copy.Color = pg.Theme.Color.Success
		time.AfterFunc(time.Second*3, func() {
			pg.copy.Text = "Copy"
			pg.copy.Color = pg.Theme.Color.Primary
		})
		return
	}
}
func (pg *ReceivePage) generateNewAddress() (string, error) {
	selectedWallet := pg.multiWallet.WalletWithID(pg.selector.selectedAccount.WalletID)

generateAddress:
	newAddr, err := selectedWallet.NextAddress(pg.selector.selectedAccount.Number)
	if err != nil {
		return "", err
	}

	if newAddr == pg.currentAddress {
		goto generateAddress
	}

	return newAddr, nil
}

func (pg *ReceivePage) OnClose() {}
