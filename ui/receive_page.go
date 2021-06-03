package ui

import (
	"bytes"
	"image"
	"image/color"
	"path/filepath"
	"time"

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

const PageReceive = "Receive"

type receivePage struct {
	common            pageCommon
	pageContainer     layout.List
	theme             *decredmaterial.Theme
	isNewAddr, isInfo bool
	newAddr, copy     decredmaterial.Button
	info, more        decredmaterial.IconButton
	card              decredmaterial.Card
	receiveAddress    decredmaterial.Label
	selector          *accountSelector
	gtx               *layout.Context

	backdrop   *widget.Clickable
	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton

	currentAddress string
}

func ReceivePage(common pageCommon) Page {
	page := &receivePage{
		pageContainer: layout.List{
			Axis: layout.Vertical,
		},
		common:         common,
		theme:          common.theme,
		info:           common.theme.IconButton(new(widget.Clickable), mustIcon(widget.NewIcon(icons.ActionInfo))),
		copy:           common.theme.Button(new(widget.Clickable), "Copy"),
		more:           common.theme.PlainIconButton(new(widget.Clickable), common.icons.navMoreIcon),
		newAddr:        common.theme.Button(new(widget.Clickable), "Generate new address"),
		receiveAddress: common.theme.Label(values.TextSize20, ""),
		card:           common.theme.Card(),
		backdrop:       new(widget.Clickable),
		backButton:     common.theme.PlainIconButton(new(widget.Clickable), common.icons.navigationArrowBack),
		infoButton:     common.theme.PlainIconButton(new(widget.Clickable), common.icons.actionInfo),
	}

	page.info.Inset, page.info.Size = layout.UniformInset(values.MarginPadding5), values.MarginPadding20
	page.copy.Background = color.NRGBA{}
	page.copy.Color = common.theme.Color.Primary
	page.copy.Inset = layout.Inset{
		Top:    values.MarginPadding19p5,
		Bottom: values.MarginPadding19p5,
		Left:   values.MarginPadding16,
		Right:  values.MarginPadding16,
	}
	page.more.Color = common.theme.Color.Gray3
	page.more.Inset = layout.UniformInset(values.MarginPadding0)
	page.newAddr.Inset = layout.Inset{
		Top:    values.MarginPadding20,
		Bottom: values.MarginPadding20,
		Left:   values.MarginPadding16,
		Right:  values.MarginPadding16,
	}
	page.newAddr.Color = common.theme.Color.Text
	page.newAddr.Background = common.theme.Color.Surface
	page.newAddr.TextSize = values.TextSize16

	zeroInset := layout.UniformInset(values.MarginPadding0)
	page.backButton.Color, page.infoButton.Color = common.theme.Color.Gray3, common.theme.Color.Gray3

	m25 := values.MarginPadding25
	page.backButton.Size, page.infoButton.Size = m25, m25
	page.backButton.Inset, page.infoButton.Inset = zeroInset, zeroInset

	page.selector = newAccountSelector(common).
		title("Receiving account").
		accountSelected(func(selectedAccount *dcrlibwallet.Account) {
			selectedWallet := page.common.multiWallet.WalletWithID(selectedAccount.WalletID)
			currentAddress, err := selectedWallet.CurrentAddress(selectedAccount.Number)
			if err != nil {
				log.Errorf("Error getting current address: %v", err)
			} else {
				page.currentAddress = currentAddress
			}
		}).
		accountValidator(func(account *dcrlibwallet.Account) bool {

			// Filter out imported account and mixed.

			wal := page.common.multiWallet.WalletWithID(account.WalletID)
			mixedAccountNumber := wal.ReadInt32ConfigValueForKey(dcrlibwallet.AccountMixerMixedAccount, -1)
			if account.Number == MaxInt32 ||
				account.Number == mixedAccountNumber {
				return false
			}
			return true
		})

	return page
}

func (pg *receivePage) pageID() string {
	return PageReceive
}

func (pg *receivePage) Layout(gtx layout.Context) layout.Dimensions {
	common := pg.common
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
			return pg.theme.Separator().Layout(gtx)
		},
		func(gtx C) D {
			return pg.pageSections(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return pg.titleLayout(gtx, common)
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Center.Layout(gtx, func(gtx C) D {
							return layout.Flex{
								Axis:      layout.Vertical,
								Alignment: layout.Middle,
							}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return pg.addressLayout(gtx, common)
								}),
								layout.Rigid(func(gtx C) D {
									return pg.addressQRCodeLayout(gtx)
								}),
							)
						})
					}),
				)
			})
		},
	}

	return common.UniformPadding(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Bottom: values.MarginPadding16}.Layout(gtx, func(gtx C) D {
					return pg.topNav(gtx, common)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return common.theme.Card().Layout(gtx, func(gtx C) D {
					return pg.pageContainer.Layout(gtx, len(pageContent), func(gtx C, i int) D {
						return pageContent[i](gtx)
					})
				})
			}),
		)
	})
}

func (pg *receivePage) pageSections(gtx layout.Context, body layout.Widget) layout.Dimensions {
	return pg.theme.Card().Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return layout.UniformInset(values.MarginPadding16).Layout(gtx, body)
	})
}

// pageBackdropLayout layout of background overlay when the popup button generate new address is show,
// click outside of the generate new address button to hide the button
func (pg *receivePage) pageBackdropLayout(gtx layout.Context) {
	if pg.isNewAddr {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
		m := op.Record(gtx.Ops)
		pg.backdrop.Layout(gtx)
		op.Defer(gtx.Ops, m.Stop())
	}
}

func (pg *receivePage) topNav(gtx layout.Context, common pageCommon) layout.Dimensions {
	m := values.MarginPadding20
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					pg.backButton.Icon = common.icons.contentClear
					return pg.backButton.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: m}.Layout(gtx, pg.theme.H6("Receive DCR").Layout)
				}),
			)
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, pg.info.Layout)
		}),
	)
}

func (pg *receivePage) titleLayout(gtx layout.Context, common pageCommon) layout.Dimensions {
	return layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			txt := common.theme.Body2("Your Address")
			txt.Color = pg.theme.Color.Gray
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

func (pg *receivePage) addressLayout(gtx layout.Context, c pageCommon) layout.Dimensions {
	card := decredmaterial.Card{
		Inset: layout.Inset{
			Top:    values.MarginPadding14,
			Bottom: values.MarginPadding16,
		},
		Color: c.theme.Color.LightGray,
	}

	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			pg.receiveAddress.Text = pg.currentAddress
			pg.receiveAddress.Color = pg.theme.Color.DeepBlue
			pg.receiveAddress.Alignment = text.Middle
			pg.receiveAddress.MaxLines = 1
			card.Radius = decredmaterial.CornerRadius{NE: 8, NW: 0, SE: 0, SW: 8}
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

func (pg *receivePage) addressQRCodeLayout(gtx layout.Context) layout.Dimensions {
	absoluteWdPath, err := GetAbsolutePath()
	if err != nil {
		log.Error(err.Error())
	}

	opt := qrcode.WithLogoImageFilePNG(filepath.Join(absoluteWdPath, "ui/assets/decredicons/qrcodeSymbol.png"))
	qrCode, err := qrcode.New(pg.currentAddress, opt)
	if err != nil {
		log.Error("Error generating address qrCode: " + err.Error())
		return layout.Dimensions{}
	}

	var buff bytes.Buffer
	err = qrCode.SaveTo(&buff)
	if err != nil {
		log.Error(err.Error())
		return layout.Dimensions{}
	}
	imgdec, _, err := image.Decode(bytes.NewReader(buff.Bytes()))
	if err != nil {
		log.Error(err.Error())
		return layout.Dimensions{}
	}

	return pg.theme.ImageIcon(gtx, imgdec, 360)
}

func (pg *receivePage) handle() {

	pg.selector.handle()

	common := pg.common
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
		pg.isNewAddr = false
	}

	if pg.infoButton.Button.Clicked() {
		go func() {
			common.modalReceiver <- &modalLoad{
				template:   ReceiveInfoTemplate,
				title:      "Receive DCR",
				cancel:     common.closeModal,
				cancelText: "Got it",
			}
		}()
	}

	if pg.backButton.Button.Clicked() {
		common.popPage()
	}

	if pg.copy.Button.Clicked() {

		clipboard.WriteOp{Text: pg.currentAddress}.Add(gtx.Ops)

		pg.copy.Text = "Copied!"
		pg.copy.Color = common.theme.Color.Success
		time.AfterFunc(time.Second*3, func() {
			pg.copy.Text = "Copy"
			pg.copy.Color = common.theme.Color.Primary
		})
		return
	}
}

func (pg *receivePage) generateNewAddress() (string, error) {
	selectedWallet := pg.common.multiWallet.WalletWithID(pg.selector.selectedAccount.WalletID)

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

func (pg *receivePage) onClose() {}
