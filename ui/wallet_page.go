package ui

import (
	"image"
	"image/color"
	"strings"

	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/decred/dcrd/dcrutil"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const PageWallet = "Wallets"

type menuItem struct {
	text     string
	id       string
	button   *widget.Clickable
	action   func(pageCommon)
	separate bool
}

type collapsible struct {
	collapsible   *decredmaterial.CollapsibleWithOption
	addAcctBtn    decredmaterial.IconButton
	backupAcctBtn decredmaterial.IconButton
}

type walletPage struct {
	common                                     pageCommon
	multiWallet                                *dcrlibwallet.MultiWallet
	wallets                                    []*dcrlibwallet.Wallet
	accounts                                   map[int][]*dcrlibwallet.Account // key = wallet id
	theme                                      *decredmaterial.Theme
	walletIcon                                 *widget.Image
	accountIcon                                *widget.Image
	walletAlertIcon                            *widget.Image
	container, accountsList, walletsList, list layout.List
	collapsibles                               map[int]collapsible
	toAcctDetails                              []*gesture.Click
	iconButton                                 decredmaterial.IconButton
	errorReceiver                              chan error
	card                                       decredmaterial.Card
	backdrops                                  []*widget.Clickable
	backdropList                               *layout.List
	optionsMenuCard                            decredmaterial.Card
	optionsMenu                                map[int][]menuItem
	addWalletMenu                              []menuItem
	watchOnlyWalletMenu                        map[int][]menuItem
	openPopupIndex                             int
	openAddWalletPopupButton                   *widget.Clickable
	isAddWalletMenuOpen                        bool
	watchOnlyWalletLabel                       decredmaterial.Label
	watchOnlyWalletIcon                        *widget.Image
	watchOnlyWalletMoreButtons                 map[int]decredmaterial.IconButton
	shadowBox                                  *decredmaterial.Shadow
	separator                                  decredmaterial.Line
}

func WalletPage(common pageCommon) Page {
	pg := &walletPage{
		common:                   common,
		multiWallet:              common.multiWallet,
		container:                layout.List{Axis: layout.Vertical},
		accountsList:             layout.List{Axis: layout.Vertical},
		walletsList:              layout.List{Axis: layout.Vertical},
		list:                     layout.List{Axis: layout.Vertical},
		theme:                    common.theme,
		wallets:                  common.multiWallet.AllWallets(),
		card:                     common.theme.Card(),
		backdropList:             &layout.List{Axis: layout.Vertical},
		errorReceiver:            make(chan error),
		openAddWalletPopupButton: new(widget.Clickable),
		openPopupIndex:           -1,
		shadowBox:                common.theme.Shadow(),
		separator:                common.theme.Separator(),
	}

	for i := 0; i < 4; i++ {
		pg.backdrops = append(pg.backdrops, new(widget.Clickable))
	}

	pg.separator.Color = common.theme.Color.Background

	pg.collapsibles = make(map[int]collapsible)
	pg.watchOnlyWalletMoreButtons = make(map[int]decredmaterial.IconButton)

	pg.watchOnlyWalletLabel = pg.theme.Body1(values.String(values.StrWatchOnlyWallets))
	pg.watchOnlyWalletLabel.Color = pg.theme.Color.Gray

	pg.iconButton = decredmaterial.IconButton{
		IconButtonStyle: material.IconButtonStyle{
			Size:       unit.Dp(25),
			Background: color.NRGBA{},
			Color:      pg.theme.Color.Text,
			Inset:      layout.UniformInset(unit.Dp(0)),
		},
	}

	pg.optionsMenuCard = decredmaterial.Card{Color: pg.theme.Color.Surface}
	pg.optionsMenuCard.Radius = decredmaterial.CornerRadius{NE: 5, NW: 5, SE: 5, SW: 5}

	pg.walletIcon = common.icons.walletIcon
	pg.walletIcon.Scale = 1

	pg.walletAlertIcon = common.icons.walletAlertIcon
	pg.walletAlertIcon.Scale = 1

	pg.initializeWalletMenu()
	pg.watchOnlyWalletIcon = common.icons.watchOnlyWalletIcon

	pg.toAcctDetails = make([]*gesture.Click, 0)

	pg.accounts = make(map[int][]*dcrlibwallet.Account)
	for _, wal := range pg.wallets {
		pg.loadAccounts(wal)
	}

	return pg
}

func (pg *walletPage) pageID() string {
	return PageWallet
}

func (pg *walletPage) loadAccounts(wal *dcrlibwallet.Wallet) error {
	accountsResult, err := wal.GetAccountsRaw()
	if err != nil {
		return err
	}

	pg.accounts[wal.ID] = accountsResult.Acc

	return nil
}

func (pg *walletPage) initializeWalletMenu() {
	///todo: should be improved,
	pg.optionsMenu = make(map[int][]menuItem)
	pg.watchOnlyWalletMenu = make(map[int][]menuItem)
	for _, wallet := range pg.wallets {
		if !wallet.IsWatchingOnlyWallet() {
			pg.optionsMenu[wallet.ID] = pg.getWalletMenu(wallet)
		}

		if wallet.IsWatchingOnlyWallet() {
			pg.watchOnlyWalletMenu[wallet.ID] = pg.getWatchOnlyWalletMenu(wallet)
		}

	}
	pg.addWalletMenu = []menuItem{
		{
			text:   values.String(values.StrCreateANewWallet),
			button: new(widget.Clickable),
			action: pg.showAddWalletModal,
		},
		{
			text:   values.String(values.StrImportExistingWallet),
			button: new(widget.Clickable),
			action: func(common pageCommon) {
				common.changePage(CreateRestorePage(common)) //TODO
			},
		},
		{
			text:   values.String(values.StrImportWatchingOnlyWallet),
			button: new(widget.Clickable),
			action: pg.showImportWatchOnlyWalletModal,
		},
	}

}

func (pg *walletPage) getWalletMenu(wal *dcrlibwallet.Wallet) []menuItem {
	return []menuItem{
		{
			text:   values.String(values.StrSignMessage),
			button: new(widget.Clickable),
			id:     PageSignMessage,
		},
		{
			text:     values.String(values.StrVerifyMessage),
			button:   new(widget.Clickable),
			separate: true,
			action: func(common pageCommon) {
				common.changePage(VerifyMessagePage(common))
			},
		},
		{
			text:     values.String(values.StrViewProperty),
			button:   new(widget.Clickable),
			separate: true,
			action: func(common pageCommon) {
				common.changePage(HelpPage(common))
			},
		},
		{
			text:     values.String(values.StrStakeShuffle),
			button:   new(widget.Clickable),
			separate: true,
			id:       PagePrivacy,
		},
		{
			text:   values.String(values.StrRename),
			button: new(widget.Clickable),
			action: func(common pageCommon) {
				go func() {
					common.modalReceiver <- &modalLoad{
						template: RenameWalletTemplate,
						title:    values.String(values.StrRenameWalletSheetTitle),
						confirm: func(name string) {
							common.wallet.RenameWallet(wal.ID, name, pg.errorReceiver) //TODO
						},
						confirmText: values.String(values.StrRename),
						cancel:      common.closeModal,
						cancelText:  values.String(values.StrCancel),
					}
				}()
			},
		},
		{
			text:   values.String(values.StrSettings),
			button: new(widget.Clickable),
			id:     PageSettings,
		},
	}
}

func (pg *walletPage) getWatchOnlyWalletMenu(wal *dcrlibwallet.Wallet) []menuItem {
	return []menuItem{
		{
			text:   values.String(values.StrSettings),
			button: new(widget.Clickable),
			id:     PageSettings,
		},
		{
			text:   values.String(values.StrRename),
			button: new(widget.Clickable),
			action: func(common pageCommon) {
				go func() {
					common.modalReceiver <- &modalLoad{
						template: RenameWalletTemplate,
						title:    values.String(values.StrRenameWalletSheetTitle),
						confirm: func(name string) {
							// id := common.info.Wallets[*common.selectedWallet].ID
							common.wallet.RenameWallet(wal.ID, name, pg.errorReceiver) //TODO
						},
						confirmText: values.String(values.StrRename),
						cancel:      common.closeModal,
						cancelText:  values.String(values.StrCancel),
					}
				}()
			},
		},
	}
}

func (pg *walletPage) showAddWalletModal(common pageCommon) {
	go func() {
		common.modalReceiver <- &modalLoad{
			template: CreateWalletTemplate,
			title:    values.String(values.StrCreate),
			confirm: func(name string, passphrase string) {
				go func() {
					wal, err := pg.multiWallet.CreateNewWallet(name, passphrase, dcrlibwallet.PassphraseTypePass)
					if err != nil {
						pg.handleError(err)
					} else {
						pg.loadAccounts(wal)
						pg.optionsMenu[wal.ID] = pg.getWalletMenu(wal)
						pg.wallets = append(pg.wallets, wal)

						common.closeModal()
						common.notify("wallet created", true)
					}
				}()

			},
			confirmText: values.String(values.StrCreate),
			cancel:      common.closeModal,
			cancelText:  values.String(values.StrCancel),
		}
	}()
}

func (pg *walletPage) showImportWatchOnlyWalletModal(common pageCommon) {
	go func() {
		common.modalReceiver <- &modalLoad{
			template: ImportWatchOnlyWalletTemplate,
			title:    values.String(values.StrImportWatchingOnlyWallet),
			confirm: func(name, extendedPubKey string) {
				go func() {
					wal, err := pg.multiWallet.CreateWatchOnlyWallet(name, extendedPubKey)
					if err != nil {
						pg.handleError(err)
					} else {

						//load accounts before adding new wallet
						pg.loadAccounts(wal)
						pg.watchOnlyWalletMenu[wal.ID] = pg.getWatchOnlyWalletMenu(wal)
						pg.wallets = append(pg.wallets, wal)

						common.closeModal()
						common.notify(values.String(values.StrWatchOnlyWalletImported), true)
					}
				}()
			},
			confirmText: values.String(values.StrImport),
			cancel:      common.closeModal,
			cancelText:  values.String(values.StrCancel),
		}
	}()
}

// Layout lays out the widgets for the main wallets pg.
func (pg *walletPage) Layout(gtx layout.Context) layout.Dimensions {
	common := pg.common

	for index := 0; index < len(pg.wallets); index++ {
		if pg.wallets[index].IsWatchingOnlyWallet() {
			if _, ok := pg.watchOnlyWalletMoreButtons[index]; !ok {
				pg.watchOnlyWalletMoreButtons[index] = decredmaterial.IconButton{
					IconButtonStyle: material.IconButtonStyle{
						Button:     new(widget.Clickable),
						Icon:       common.icons.navigationMore,
						Size:       values.MarginPadding25,
						Background: color.NRGBA{},
						Color:      common.theme.Color.Text,
						Inset:      layout.UniformInset(values.MarginPadding0),
					},
				}
			}
		} else {
			if _, ok := pg.collapsibles[index]; !ok {
				addAcctBtn := common.theme.IconButton(new(widget.Clickable), common.icons.contentAdd)
				addAcctBtn.Inset = layout.UniformInset(values.MarginPadding0)
				addAcctBtn.Size = values.MarginPadding25
				addAcctBtn.Background = color.NRGBA{}
				addAcctBtn.Color = common.theme.Color.Text

				backupBtn := common.theme.PlainIconButton(new(widget.Clickable), common.icons.navigationArrowForward)
				backupBtn.Color = common.theme.Color.Surface
				backupBtn.Inset = layout.UniformInset(values.MarginPadding0)
				backupBtn.Size = values.MarginPadding20

				pg.collapsibles[index] = collapsible{
					collapsible:   pg.theme.CollapsibleWithOption(),
					addAcctBtn:    addAcctBtn,
					backupAcctBtn: backupBtn,
				}
			}

		}
	}

	pageContent := []func(gtx C) D{
		func(gtx C) D {
			return pg.walletSection(gtx, common)
		},
		func(gtx C) D {
			return pg.watchOnlyWalletSection(gtx)
		},
	}

	body := func(gtx C) D {
		return layout.Stack{Alignment: layout.SE}.Layout(gtx,
			layout.Expanded(func(gtx C) D {
				return pg.container.Layout(gtx, len(pageContent), func(gtx C, i int) D {
					dims := layout.UniformInset(values.MarginPadding5).Layout(gtx, pageContent[i])
					if pg.isAddWalletMenuOpen || pg.openPopupIndex != -1 {
						dims.Size.Y += 60
					}
					return dims
				})
			}),
			layout.Stacked(func(gtx C) D {
				return pg.layoutAddWalletSection(gtx, common)
			}),
		)
	}

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			return common.UniformPadding(gtx, body)
		}),
		layout.Expanded(func(gtx C) D {
			if pg.isAddWalletMenuOpen || pg.openPopupIndex != -1 {
				halfHeight := gtx.Constraints.Max.Y / 2
				return pg.container.Layout(gtx, len(pg.backdrops), func(gtx C, i int) D {
					gtx.Constraints.Min.Y = halfHeight
					return pg.backdrops[i].Layout(gtx)
				})
			}
			return D{}
		}),
	)
}

func (pg *walletPage) layoutOptionsMenu(gtx layout.Context, optionsMenuIndex, walletID int, isWatchOnlyWalletMenu bool) {
	if pg.openPopupIndex != optionsMenuIndex {
		return
	}

	var menu []menuItem
	var leftInset float32
	if isWatchOnlyWalletMenu {
		menu = pg.watchOnlyWalletMenu[walletID]
		leftInset = -35
	} else {
		menu = pg.optionsMenu[walletID]
		leftInset = -120
	}

	inset := layout.Inset{
		Top:  unit.Dp(30),
		Left: unit.Dp(leftInset),
	}

	m := op.Record(gtx.Ops)
	inset.Layout(gtx, func(gtx C) D {
		width := unit.Value{U: unit.UnitDp, V: 150}
		gtx.Constraints.Max.X = gtx.Px(width)
		return pg.shadowBox.Layout(gtx, func(gtx C) D {
			return pg.optionsMenuCard.Layout(gtx, func(gtx C) D {
				return (&layout.List{Axis: layout.Vertical}).Layout(gtx, len(menu), func(gtx C, i int) D {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return material.Clickable(gtx, menu[i].button, func(gtx C) D {
								m10 := values.MarginPadding10
								return layout.Inset{Top: m10, Bottom: m10, Left: m10, Right: m10}.Layout(gtx, func(gtx C) D {
									gtx.Constraints.Min.X = gtx.Constraints.Max.X
									return pg.theme.Body1(menu[i].text).Layout(gtx)
								})
							})
						}),
						layout.Rigid(func(gtx C) D {
							if menu[i].separate {
								m := values.MarginPadding5
								return layout.Inset{Top: m, Bottom: m}.Layout(gtx, pg.separator.Layout)
							}
							return D{}
						}),
					)
				})
			})
		})
	})
	op.Defer(gtx.Ops, m.Stop())
}

func (pg *walletPage) walletSection(gtx layout.Context, common pageCommon) layout.Dimensions {
	return pg.walletsList.Layout(gtx, len(pg.wallets), func(gtx C, i int) D {
		if pg.wallets[i].IsWatchingOnlyWallet() {
			return D{}
		}

		accounts := pg.accounts[pg.wallets[i].ID]
		pg.updateAcctDetailsButtons(accounts)

		collapsibleMore := func(gtx C) {
			pg.layoutOptionsMenu(gtx, i, pg.wallets[i].ID, false)
		}

		collapsibleHeader := func(gtx C) D {
			return pg.layoutCollapsibleHeader(gtx, pg.wallets[i])
		}

		collapsibleBody := func(gtx C) D {
			return layout.UniformInset(values.MarginPadding5).Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				gtx.Constraints.Min.Y = 100

				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Inset{
							Left:  values.MarginPadding38,
							Right: values.MarginPadding10,
						}.Layout(gtx, pg.theme.Separator().Layout)
					}),
					layout.Rigid(func(gtx C) D {
						return pg.accountsList.Layout(gtx, len(accounts), func(gtx C, x int) D {
							click := pg.toAcctDetails[x]
							pointer.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Add(gtx.Ops)
							click.Add(gtx.Ops)
							pg.goToAcctDetails(gtx, common, accounts[x], click)
							return pg.walletAccountsLayout(gtx, accounts[x].Name, dcrutil.Amount(accounts[x].TotalBalance).String(), dcrutil.Amount(accounts[x].Balance.Spendable).String(), common)
						})
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Top: values.MarginPadding10, Bottom: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return layout.Inset{
										Right: values.MarginPadding10,
										Left:  values.MarginPadding38,
									}.Layout(gtx, pg.collapsibles[i].addAcctBtn.Layout)
								}),
								layout.Rigid(func(gtx C) D {
									txt := pg.theme.H6(values.String(values.StrAddNewAccount))
									txt.Color = common.theme.Color.Gray
									return txt.Layout(gtx)
								}),
							)
						})
					}),
				)
			})
		}

		return layout.Inset{Bottom: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
			var children []layout.FlexChild
			children = append(children, layout.Rigid(func(gtx C) D {
				return pg.collapsibles[i].collapsible.Layout(gtx, collapsibleHeader, collapsibleBody, collapsibleMore)
			}))

			if len(pg.wallets[i].EncryptedSeed) > 0 {
				children = append(children, layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: unit.Dp(-10)}.Layout(gtx, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								blankLine := common.theme.Line(10, gtx.Constraints.Max.X)
								blankLine.Color = common.theme.Color.Surface
								return blankLine.Layout(gtx)
							}),
							layout.Rigid(func(gtx C) D {
								pg.card.Color = pg.theme.Color.Danger
								pg.card.Radius = decredmaterial.CornerRadius{SW: 10, SE: 10}
								return pg.card.Layout(gtx, func(gtx C) D {
									return pg.backupSeedNotification(gtx, common, i)
								})
							}),
						)
					})
				}))
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
		})
	})
}

func (pg *walletPage) watchOnlyWalletSection(gtx layout.Context) layout.Dimensions {
	watchOnlyWalletCount := 0
	for _, wal := range pg.wallets {
		if wal.IsWatchingOnlyWallet() {
			watchOnlyWalletCount++
		}
	}
	if watchOnlyWalletCount == 0 {
		return D{}
	}
	card := pg.card
	card.Color = pg.theme.Color.Surface
	card.Radius = decredmaterial.CornerRadius{NE: 10, NW: 10, SE: 10, SW: 10}

	return card.Layout(gtx, func(gtx C) D {
		m := values.MarginPadding20
		return layout.Inset{Top: m, Left: m}.Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(pg.watchOnlyWalletLabel.Layout),
				layout.Rigid(func(gtx C) D {
					m := values.MarginPadding10
					return layout.Inset{Top: m, Bottom: m}.Layout(gtx, pg.theme.Separator().Layout)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Right: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
						return pg.layoutWatchOnlyWallets(gtx)
					})
				}),
			)
		})
	})
}

func (pg *walletPage) layoutWatchOnlyWallets(gtx layout.Context) D {
	return (&layout.List{Axis: layout.Vertical}).Layout(gtx, len(pg.wallets), func(gtx C, i int) D {
		if !pg.wallets[i].IsWatchingOnlyWallet() {
			return D{}
		}
		m := values.MarginPadding10
		return layout.Inset{Top: m, Bottom: m}.Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Flex{}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							inset := layout.Inset{
								Right: values.MarginPadding10,
							}
							return inset.Layout(gtx, func(gtx C) D {
								pg.watchOnlyWalletIcon.Scale = 1.0
								return pg.watchOnlyWalletIcon.Layout(gtx)
							})
						}),
						layout.Rigid(func(gtx C) D {
							return pg.theme.Body2(pg.wallets[i].Name).Layout(gtx)
						}),
						layout.Rigid(pg.theme.Body2(pg.wallets[i].Name).Layout),
						layout.Flexed(1, func(gtx C) D {
							return layout.E.Layout(gtx, func(gtx C) D {
								return layout.Flex{}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										totalBalanceText, _ := pg.common.totalWalletBalance(pg.wallets[i].ID)
										return pg.theme.Body2(totalBalanceText.String()).Layout(gtx) //TODO get this value in handle
									}),
									layout.Rigid(func(gtx C) D {
										pg.layoutOptionsMenu(gtx, i, pg.wallets[i].ID, true)
										return layout.Inset{Top: unit.Dp(-3)}.Layout(gtx, pg.watchOnlyWalletMoreButtons[i].Layout)
									}),
								)
							})
						}),
					)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: values.MarginPadding10, Left: values.MarginPadding38, Right: values.MarginPaddingMinus10}.Layout(gtx, func(gtx C) D {
						if i == len(pg.wallets)-1 {
							return D{}
						}
						return pg.theme.Separator().Layout(gtx)
					})
				}),
			)
		})
	})
}

func (pg *walletPage) layoutCollapsibleHeader(gtx layout.Context, wallet *dcrlibwallet.Wallet) D {
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Right: values.MarginPadding10,
			}.Layout(gtx, pg.walletIcon.Layout)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return pg.theme.Body1(strings.Title(strings.ToLower(wallet.Name))).Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					if len(wallet.EncryptedSeed) > 0 {
						txt := pg.theme.Caption(values.String(values.StrNotBackedUp))
						txt.Color = pg.theme.Color.Danger
						return txt.Layout(gtx)
					}
					return D{}
				}),
			)
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				totalBalance, _ := pg.common.totalWalletBalance(wallet.ID)
				balanceLabel := pg.theme.Body1(totalBalance.String()) // TODO
				balanceLabel.Color = pg.theme.Color.Gray
				return layout.Inset{Right: values.MarginPadding5}.Layout(gtx, balanceLabel.Layout)
			})
		}),
	)
}

func (pg *walletPage) tableLayout(gtx layout.Context, leftLabel, rightLabel decredmaterial.Label, isIcon bool, seed int) layout.Dimensions {
	m := values.MarginPadding0
	if seed > 0 {
		m = values.MarginPaddingMinus5
	}

	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if isIcon {
				inset := layout.Inset{
					Right: values.MarginPadding10,
				}
				return inset.Layout(gtx, pg.walletIcon.Layout)
			}
			return layout.Dimensions{}
		}),
		layout.Rigid(func(gtx C) D {
			inset := layout.Inset{
				Top: m,
			}
			return inset.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return leftLabel.Layout(gtx)
					}),
					layout.Rigid(func(gtx C) D {
						if isIcon {
							if seed > 0 {
								txt := pg.theme.Caption(values.String(values.StrNotBackedUp))
								txt.Color = pg.theme.Color.Danger
								inset := layout.Inset{
									Bottom: values.MarginPadding5,
								}
								return inset.Layout(gtx, txt.Layout)
							}
						}
						return layout.Dimensions{}
					}),
				)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, rightLabel.Layout)
		}),
	)
}

func (pg *walletPage) walletAccountsLayout(gtx layout.Context, name, totalBal, spendableBal string, common pageCommon) layout.Dimensions {
	pg.accountIcon = common.icons.accountIcon
	if name == "imported" {
		pg.accountIcon = common.icons.importedAccountIcon
	}
	pg.accountIcon.Scale = 1.0

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			inset := layout.Inset{
				Top:    values.MarginPadding10,
				Left:   values.MarginPadding38,
				Bottom: values.MarginPadding20,
			}
			return inset.Layout(gtx, func(gtx C) D {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						inset := layout.Inset{
							Right: values.MarginPadding10,
							Top:   values.MarginPadding13,
						}
						return inset.Layout(gtx, pg.accountIcon.Layout)
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								inset := layout.Inset{
									Right: values.MarginPadding10,
								}
								return inset.Layout(gtx, func(gtx C) D {
									return layout.Flex{}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											acctName := strings.Title(strings.ToLower(name))
											return pg.theme.H6(acctName).Layout(gtx)
										}),
										layout.Flexed(1, func(gtx C) D {
											return layout.E.Layout(gtx, func(gtx C) D {
												return common.layoutBalance(gtx, totalBal, true)
											})
										}),
									)
								})
							}),
							layout.Rigid(func(gtx C) D {
								inset := layout.Inset{
									Right: values.MarginPadding10,
								}
								return inset.Layout(gtx, func(gtx C) D {
									spendibleLabel := pg.theme.Body2(values.String(values.StrLabelSpendable))
									spendibleLabel.Color = pg.theme.Color.Gray
									spendibleBalLabel := pg.theme.Body2(spendableBal)
									spendibleBalLabel.Color = pg.theme.Color.Gray
									return pg.tableLayout(gtx, spendibleLabel, spendibleBalLabel, false, 0)
								})
							}),
						)
					}),
				)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Left:  values.MarginPadding70,
				Right: values.MarginPadding10,
			}.Layout(gtx, pg.theme.Separator().Layout)
		}),
	)
}

func (pg *walletPage) backupSeedNotification(gtx layout.Context, common pageCommon, i int) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	textColour := common.theme.Color.InvText
	return layout.UniformInset(values.MarginPadding10).Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return pg.walletAlertIcon.Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				inset := layout.Inset{
					Left: values.MarginPadding10,
				}
				return inset.Layout(gtx, func(gtx C) D {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							txt := pg.theme.Body2(values.String(values.StrBackupSeedPhrase))
							txt.Color = textColour
							return txt.Layout(gtx)
						}),
						layout.Rigid(func(gtx C) D {
							txt := pg.theme.Caption(values.String(values.StrVerifySeedInfo))
							txt.Color = textColour
							return txt.Layout(gtx)
						}),
					)
				})
			}),
			layout.Flexed(1, func(gtx C) D {
				inset := layout.Inset{
					Top: values.MarginPadding5,
				}
				return inset.Layout(gtx, func(gtx C) D {
					return layout.E.Layout(gtx, pg.collapsibles[i].backupAcctBtn.Layout)
				})
			}),
		)
	})
}

func (pg *walletPage) layoutAddWalletMenu(gtx layout.Context) layout.Dimensions {
	inset := layout.Inset{
		Top:  unit.Dp(-100),
		Left: unit.Dp(-130),
	}

	return inset.Layout(gtx, func(gtx C) D {
		border := widget.Border{Color: pg.theme.Color.LightGray, CornerRadius: unit.Dp(5), Width: unit.Dp(2)}
		return border.Layout(gtx, func(gtx C) D {
			return pg.optionsMenuCard.Layout(gtx, func(gtx C) D {
				return (&layout.List{Axis: layout.Vertical}).Layout(gtx, len(pg.addWalletMenu), func(gtx C, i int) D {
					return material.Clickable(gtx, pg.addWalletMenu[i].button, func(gtx C) D {
						return layout.UniformInset(unit.Dp(10)).Layout(gtx, pg.theme.Body2(pg.addWalletMenu[i].text).Layout)
					})
				})
			})
		})
	})
}

func (pg *walletPage) layoutAddWalletSection(gtx layout.Context, common pageCommon) layout.Dimensions {
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	return layout.SE.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				if pg.isAddWalletMenuOpen {
					m := op.Record(gtx.Ops)
					pg.layoutAddWalletMenu(gtx)
					op.Defer(gtx.Ops, m.Stop())
				}
				return D{}
			}),
			layout.Rigid(func(gtx C) D {
				icon := common.icons.newWalletIcon
				sz := gtx.Constraints.Max.X
				icon.Scale = float32(sz) / float32(gtx.Px(unit.Dp(float32(sz))))
				return decredmaterial.Clickable(gtx, pg.openAddWalletPopupButton, icon.Layout)
			}),
		)
	})
}

func (pg *walletPage) updateAcctDetailsButtons(walAcct []*dcrlibwallet.Account) {
	if len(walAcct) != len(pg.toAcctDetails) {
		for i := 0; i < len(walAcct); i++ {
			pg.toAcctDetails = append(pg.toAcctDetails, &gesture.Click{})
		}
	}
}

func (pg *walletPage) goToAcctDetails(gtx layout.Context, common pageCommon, acct *dcrlibwallet.Account, click *gesture.Click) {
	for _, e := range click.Events(gtx) {
		if e.Type == gesture.TypeClick {
			common.changePage(AcctDetailsPage(common, acct))
		}
	}
}

func (pg *walletPage) isCollapsibleMenuOpen() bool {
	return pg.openPopupIndex > -1
}

func (pg *walletPage) closePopups() {
	pg.openPopupIndex = -1
	pg.isAddWalletMenuOpen = false
}

func (pg *walletPage) openPopup(common pageCommon, index int) {
	*common.selectedWallet = index
	pg.openPopupIndex = index
}

func (pg *walletPage) handle() {
	common := pg.common

	for index := range pg.backdrops {
		for pg.backdrops[index].Clicked() {
			pg.closePopups()
		}
	}

	for index := range pg.collapsibles {
		for pg.collapsibles[index].collapsible.MoreTriggered() {
			if pg.isCollapsibleMenuOpen() {
				if pg.openPopupIndex == index {
					pg.closePopups()
				} else {
					pg.closePopups()
					pg.openPopup(common, index)
				}
			} else {
				pg.openPopup(common, index)
			}
		}

		for pg.collapsibles[index].addAcctBtn.Button.Clicked() {
			wal := pg.wallets[index]
			go func() {
				common.modalReceiver <- &modalLoad{
					template: CreateAccountTemplate,
					title:    values.String(values.StrCreateNewAccount),
					confirm: func(name string, passphrase string) {
						go func() {
							_, err := wal.CreateNewAccount(name, []byte(passphrase))
							if err != nil {
								pg.handleError(err)
							} else {
								common.closeModal()
								pg.loadAccounts(wal) //Todo: should actually insert not reload
							}
						}()
					},
					confirmText: values.String(values.StrCreate),
					cancel:      common.closeModal,
					cancelText:  values.String(values.StrCancel),
				}
			}()
			break
		}

		for pg.collapsibles[index].backupAcctBtn.Button.Clicked() {
			common.changePage(BackupPage(common, pg.wallets[index].ID))
		}
	}

	for walletIndex, button := range pg.watchOnlyWalletMoreButtons {
		for button.Button.Clicked() {
			*common.selectedWallet = walletIndex
			pg.openPopupIndex = walletIndex
		}
	}

	for walletID, optionsMenu := range pg.optionsMenu {
		for i := range optionsMenu {
			if optionsMenu[i].button.Clicked() {
				switch optionsMenu[i].id {
				case PageSignMessage:
					common.changePage(SignMessagePage(common, walletID))
				case PagePrivacy:
					common.changePage(PrivacyPage(common, walletID))
				case PageSettings:
					common.changePage(WalletSettingsPage(common, walletID))
				default:
					optionsMenu[i].action(common)
				}

				pg.openPopupIndex = -1
			}
		}

	}

	for walletID, optionsMenu := range pg.watchOnlyWalletMenu {
		for i := range optionsMenu {
			if optionsMenu[i].button.Clicked() {
				switch optionsMenu[i].id {
				case PageSettings:
					common.changePage(WalletSettingsPage(common, walletID))
				default:
					optionsMenu[i].action(common)
				}

				pg.openPopupIndex = -1
			}
		}

	}

	for index := range pg.addWalletMenu {
		for pg.addWalletMenu[index].button.Clicked() {
			pg.isAddWalletMenuOpen = false
			pg.addWalletMenu[index].action(common)
		}
	}

	for pg.openAddWalletPopupButton.Clicked() {
		pg.isAddWalletMenuOpen = !pg.isAddWalletMenuOpen
	}

	select {
	case err := <-pg.errorReceiver:
		common.modalLoad.setLoading(false)
		if err.Error() == dcrlibwallet.ErrInvalidPassphrase {
			e := values.String(values.StrInvalidPassphrase)
			common.notify(e, false)
			return
		}
		common.notify(err.Error(), false)
	default:
	}
}

func (pg *walletPage) handleError(err error) {
	pg.common.modalLoad.setLoading(false)
	if err.Error() == dcrlibwallet.ErrInvalidPassphrase {
		e := values.String(values.StrInvalidPassphrase)
		pg.common.notify(e, false)
		return
	}
	pg.common.notify(err.Error(), false)
}

func (pg *walletPage) onClose() {
	pg.closePopups()
}
