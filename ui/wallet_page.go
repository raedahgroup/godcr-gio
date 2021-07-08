package ui

import (
	"image"
	"image/color"

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

type walletListItem struct {
	wal      *dcrlibwallet.Wallet
	accounts []*dcrlibwallet.Account

	totalBalance string
	optionsMenu  []menuItem
	accountsList *decredmaterial.ClickableList

	// normal wallets
	collapsible *decredmaterial.CollapsibleWithOption

	// watch only
	moreButton decredmaterial.IconButton
}

type menuItem struct {
	text     string
	id       string
	button   *widget.Clickable
	action   func(*pageCommon)
	separate bool
}

type walletPage struct {
	multiWallet *dcrlibwallet.MultiWallet
	listItems   []*walletListItem

	common *pageCommon
	theme  *decredmaterial.Theme

	walletIcon               *widget.Image
	accountIcon              *widget.Image
	walletAlertIcon          *widget.Image
	container, walletsList   layout.List
	watchWalletsList         *decredmaterial.ClickableList
	addAcct                  []*gesture.Click
	backupAcct               []*gesture.Click
	iconButton               decredmaterial.IconButton
	card                     decredmaterial.Card
	backdrop                 *widget.Clickable
	optionsMenuCard          decredmaterial.Card
	addWalletMenu            []menuItem
	openPopupIndex           int
	openAddWalletPopupButton *widget.Clickable
	isAddWalletMenuOpen      bool
	watchOnlyWalletLabel     decredmaterial.Label
	watchOnlyWalletIcon      *widget.Image
	shadowBox                *decredmaterial.Shadow
	separator                decredmaterial.Line
	addAcctIcon              *widget.Icon
	backupAcctIcon           *widget.Icon
}

func WalletPage(common *pageCommon) Page {
	pg := &walletPage{
		common:                   common,
		multiWallet:              common.multiWallet,
		container:                layout.List{Axis: layout.Vertical},
		walletsList:              layout.List{Axis: layout.Vertical},
		watchWalletsList:         common.theme.NewClickableList(layout.Vertical),
		theme:                    common.theme,
		card:                     common.theme.Card(),
		backdrop:                 new(widget.Clickable),
		openAddWalletPopupButton: new(widget.Clickable),
		openPopupIndex:           -1,
		shadowBox:                common.theme.Shadow(),
		separator:                common.theme.Separator(),
		addAcctIcon:              common.icons.contentAdd,
		backupAcctIcon:           common.icons.navigationArrowForward,
	}

	pg.separator.Color = common.theme.Color.Background
	pg.addAcctIcon.Color = common.theme.Color.Text

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

	pg.initializeFloatingMenu()
	pg.watchOnlyWalletIcon = common.icons.watchOnlyWalletIcon

	pg.addAcct = make([]*gesture.Click, 0)
	pg.backupAcct = make([]*gesture.Click, 0)

	return pg
}

func (pg *walletPage) OnResume() {
	wallets := pg.common.sortedWalletList()

	pg.listItems = make([]*walletListItem, 0)
	for _, wal := range wallets {
		accountsResult, err := wal.GetAccountsRaw()
		if err != nil {
			continue
		}

		var totalBalance int64
		for _, acc := range accountsResult.Acc {
			totalBalance += acc.TotalBalance
		}

		listItem := &walletListItem{
			wal:      wal,
			accounts: accountsResult.Acc,

			totalBalance: dcrutil.Amount(totalBalance).String(),
			optionsMenu:  pg.getWalletMenu(wal),
			accountsList: pg.theme.NewClickableList(layout.Vertical),
		}

		if wal.IsWatchingOnlyWallet() {
			moreBtn := decredmaterial.IconButton{
				IconButtonStyle: material.IconButtonStyle{
					Button:     new(widget.Clickable),
					Icon:       pg.common.icons.navigationMore,
					Size:       values.MarginPadding25,
					Background: color.NRGBA{},
					Color:      pg.theme.Color.Text,
					Inset:      layout.UniformInset(values.MarginPadding0),
				},
			}
			listItem.moreButton = moreBtn
		} else {
			listItem.collapsible = pg.theme.CollapsibleWithOption()
		}

		pg.listItems = append(pg.listItems, listItem)
	}
}

func (pg *walletPage) initializeFloatingMenu() {
	pg.addWalletMenu = []menuItem{
		{
			text:   values.String(values.StrCreateANewWallet),
			button: new(widget.Clickable),
			action: pg.showAddWalletModal,
		},
		{
			text:   values.String(values.StrImportExistingWallet),
			button: new(widget.Clickable),
			action: func(common *pageCommon) {
				common.changeWindowPage(CreateRestorePage(common), true)
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
	if wal.IsWatchingOnlyWallet() {
		return pg.getWatchOnlyWalletMenu(wal)
	}

	return []menuItem{
		{
			text:   values.String(values.StrSignMessage),
			button: new(widget.Clickable),
			id:     PageSignMessage,
		},
		{
			text:     values.String(values.StrViewProperty),
			button:   new(widget.Clickable),
			separate: true,
			action: func(common *pageCommon) {
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
			action: func(common *pageCommon) {
				textModal := newTextInputModal(common).
					hint("Wallet name").
					positiveButton(values.String(values.StrRename), func(newName string, tim *textInputModal) bool {
						// todo handle error
						pg.multiWallet.RenameWallet(wal.ID, newName)
						return true
					})

				textModal.title(values.String(values.StrRenameWalletSheetTitle)).
					negativeButton(values.String(values.StrCancel), func() {})
				textModal.Show()
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
			action: func(common *pageCommon) {
				textModal := newTextInputModal(common).
					hint("Wallet name").
					positiveButton(values.String(values.StrRename), func(newName string, tim *textInputModal) bool {
						//TODO
						pg.multiWallet.RenameWallet(wal.ID, newName)
						return true
					})

				textModal.title(values.String(values.StrRenameWalletSheetTitle)).
					negativeButton(values.String(values.StrCancel), func() {})
				textModal.Show()
			},
		},
	}
}

func (pg *walletPage) showAddWalletModal(common *pageCommon) {
	newCreatePasswordModal(common).
		title("Create new wallet").
		enableName(true).
		passwordCreated(func(walletName, password string, m *createPasswordModal) bool {
			go func() {
				_, err := pg.multiWallet.CreateNewWallet(walletName, password, dcrlibwallet.PassphraseTypePass)
				if err != nil {
					m.setError(err.Error())
					m.setLoading(false)
					return
				}
				m.Dismiss()
			}()
			return false
		}).Show()
}

func (pg *walletPage) showImportWatchOnlyWalletModal(common *pageCommon) {
	newCreateWatchOnlyModal(common).
		watchOnlyCreated(func(walletName, extPubKey string, m *createWatchOnlyModal) bool {
			go func() {
				_, err := pg.multiWallet.CreateWatchOnlyWallet(walletName, extPubKey)
				if err != nil {
					common.notify(err.Error(), false)
					m.setError(err.Error())
					m.setLoading(false)
				} else {
					// pg.wallet.GetMultiWalletInfo() TODO
					common.notify(values.String(values.StrWatchOnlyWalletImported), true)
					m.Dismiss()
				}
			}()
			return false
		}).Show()
}

// Layout lays out the widgets for the main wallets pg.
func (pg *walletPage) Layout(gtx layout.Context) layout.Dimensions {
	common := pg.common

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
				return pg.backdrop.Layout(gtx)
			}
			return D{}
		}),
	)
}

func (pg *walletPage) layoutOptionsMenu(gtx layout.Context, optionsMenuIndex int, listItem *walletListItem) {
	if pg.openPopupIndex != optionsMenuIndex {
		return
	}

	var leftInset float32
	if listItem.wal.IsWatchingOnlyWallet() {
		leftInset = -35
	} else {
		leftInset = -120
	}

	inset := layout.Inset{
		Top:  unit.Dp(30),
		Left: unit.Dp(leftInset),
	}
	menu := listItem.optionsMenu

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

func (pg *walletPage) walletSection(gtx layout.Context, common *pageCommon) layout.Dimensions {
	return pg.walletsList.Layout(gtx, len(pg.listItems), func(gtx C, i int) D {
		listItem := pg.listItems[i]
		if listItem.wal.IsWatchingOnlyWallet() {
			return D{}
		}

		pg.updateAcctFuncButtons(listItem.accounts)

		collapsibleMore := func(gtx C) {
			pg.layoutOptionsMenu(gtx, i, listItem)
		}

		collapsibleHeader := func(gtx C) D {
			return pg.layoutCollapsibleHeader(gtx, listItem)
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
						return listItem.accountsList.Layout(gtx, len(listItem.accounts), func(gtx C, x int) D {
							return pg.walletAccountsLayout(gtx, listItem.accounts[x])
						})
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Top: values.MarginPadding10, Bottom: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									click := pg.addAcct[i]
									pointer.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Add(gtx.Ops)
									click.Add(gtx.Ops)
									pg.openAddAcctDialog(gtx, i, click)
									return layout.Inset{
										Right: values.MarginPadding10,
										Left:  values.MarginPadding38,
									}.Layout(gtx, func(gtx C) D {
										return pg.addAcctIcon.Layout(gtx, values.MarginPadding25)
									})
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
				return listItem.collapsible.Layout(gtx, collapsibleHeader, collapsibleBody, collapsibleMore)
			}))

			if len(listItem.wal.EncryptedSeed) > 0 {
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
									return pg.backupSeedNotification(gtx, i)
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
	hasWatchOnly := false
	for _, listItem := range pg.listItems {
		if listItem.wal.IsWatchingOnlyWallet() {
			hasWatchOnly = true
			break
		}
	}
	if !hasWatchOnly {
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
	return pg.watchWalletsList.Layout(gtx, len(pg.listItems), func(gtx C, i int) D {
		listItem := pg.listItems[i]
		if !listItem.wal.IsWatchingOnlyWallet() {
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
							pg.watchOnlyWalletIcon.Scale = 1.0
							return inset.Layout(gtx, pg.watchOnlyWalletIcon.Layout)
						}),
						layout.Rigid(pg.theme.Body2(listItem.wal.Name).Layout),
						layout.Flexed(1, func(gtx C) D {
							return layout.E.Layout(gtx, func(gtx C) D {
								return layout.Flex{}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										return pg.theme.Body2(listItem.totalBalance).Layout(gtx)
									}),
									layout.Rigid(func(gtx C) D {
										pg.layoutOptionsMenu(gtx, i, listItem)
										return layout.Inset{Top: unit.Dp(-3)}.Layout(gtx, listItem.moreButton.Layout)
									}),
								)
							})
						}),
					)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: values.MarginPadding10, Left: values.MarginPadding38, Right: values.MarginPaddingMinus10}.Layout(gtx, func(gtx C) D {
						if i == len(pg.listItems)-1 {
							return D{}
						}
						return pg.theme.Separator().Layout(gtx)
					})
				}),
			)
		})
	})
}

func (pg *walletPage) layoutCollapsibleHeader(gtx layout.Context, listItem *walletListItem) D {
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Right: values.MarginPadding10,
			}.Layout(gtx, pg.walletIcon.Layout)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return pg.theme.Body1(listItem.wal.Name).Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					if len(listItem.wal.EncryptedSeed) > 0 {
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
				balanceLabel := pg.theme.Body1(listItem.totalBalance)
				balanceLabel.Color = pg.theme.Color.Gray
				return layout.Inset{Right: values.MarginPadding5}.Layout(gtx, balanceLabel.Layout)
			})
		}),
	)
}

func (pg *walletPage) tableLayout(gtx layout.Context, leftLabel, rightLabel decredmaterial.Label) layout.Dimensions {
	m := values.MarginPadding0

	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			inset := layout.Inset{
				Top: m,
			}
			return inset.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return leftLabel.Layout(gtx)
					}),
				)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, rightLabel.Layout)
		}),
	)
}

func (pg *walletPage) walletAccountsLayout(gtx layout.Context, account *dcrlibwallet.Account) layout.Dimensions {
	common := pg.common

	pg.accountIcon = common.icons.accountIcon
	if account.Number == MaxInt32 {
		pg.accountIcon = common.icons.importedAccountIcon
		if account.TotalBalance == 0 {
			return D{}
		}
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
											return pg.theme.H6(account.Name).Layout(gtx)
										}),
										layout.Flexed(1, func(gtx C) D {
											return layout.E.Layout(gtx, func(gtx C) D {
												totalBal := dcrutil.Amount(account.Balance.Spendable).String()
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
									spendableLabel := pg.theme.Body2(values.String(values.StrLabelSpendable))
									spendableLabel.Color = pg.theme.Color.Gray

									spendableBal := dcrutil.Amount(account.Balance.Spendable).String()
									spendableBalLabel := pg.theme.Body2(spendableBal)
									spendableBalLabel.Color = pg.theme.Color.Gray
									return pg.tableLayout(gtx, spendableLabel, spendableBalLabel)
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

func (pg *walletPage) backupSeedNotification(gtx layout.Context, index int) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	textColour := pg.theme.Color.InvText
	click := pg.backupAcct[index]
	pointer.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Add(gtx.Ops)
	click.Add(gtx.Ops)
	pg.openAcctBackup(gtx, click)
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
				pg.backupAcctIcon.Color = textColour
				return inset.Layout(gtx, func(gtx C) D {
					return layout.E.Layout(gtx, func(gtx C) D {
						return pg.backupAcctIcon.Layout(gtx, values.MarginPadding20)
					})
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

func (pg *walletPage) layoutAddWalletSection(gtx layout.Context, common *pageCommon) layout.Dimensions {
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

func (pg *walletPage) updateAcctFuncButtons(walAcct []*dcrlibwallet.Account) {
	if len(walAcct) != len(pg.addAcct) {
		for i := 0; i < len(walAcct); i++ {
			pg.addAcct = append(pg.addAcct, &gesture.Click{})
			pg.backupAcct = append(pg.backupAcct, &gesture.Click{})
		}
	}
}

func (pg *walletPage) openAddAcctDialog(gtx layout.Context, index int, click *gesture.Click) {
	for _, e := range click.Events(gtx) {
		if e.Type == gesture.TypeClick {
			listItem := pg.listItems[index]
			walletID := listItem.wal.ID

			textModal := newTextInputModal(pg.common).
				hint("Account name").
				positiveButton(values.String(values.StrCreate), func(accountName string, tim *textInputModal) bool {
					if accountName != "" {
						newPasswordModal(pg.common).
							title(values.String(values.StrCreateNewAccount)).
							hint("Spending password").
							negativeButton(values.String(values.StrCancel), func() {}).
							positiveButton(values.String(values.StrConfirm), func(password string, pm *passwordModal) bool {
								go func() {

									wal := pg.multiWallet.WalletWithID(walletID)
									wal.CreateNewAccount(accountName, []byte(password)) // TODO
									pm.Dismiss()
								}()

								return false
							}).Show()
					}
					return true
				})

			textModal.title(values.String(values.StrCreateNewAccount)).
				negativeButton(values.String(values.StrCancel), func() {})
			textModal.Show()
		}
	}
}

func (pg *walletPage) openAcctBackup(gtx layout.Context, click *gesture.Click) {
	for _, e := range click.Events(gtx) {
		if e.Type == gesture.TypeClick {
			pg.common.changePage(PageSeedBackup)
		}
	}
}

func (pg *walletPage) closePopups() {
	pg.openPopupIndex = -1
	pg.isAddWalletMenuOpen = false
}

func (pg *walletPage) openPopup(index int) {
	if pg.openPopupIndex >= 0 {
		if pg.openPopupIndex == index {
			pg.closePopups()
			return
		}
		pg.closePopups()
	}

	pg.openPopupIndex = index
}

func (pg *walletPage) handle() {
	common := pg.common

	for pg.backdrop.Clicked() {
		pg.closePopups()
	}

	if ok, selectedItem := pg.watchWalletsList.ItemClicked(); ok {
		listItem := pg.listItems[selectedItem]
		// TODO: find default account using number
		pg.common.changeFragment(AcctDetailsPage(common, listItem.accounts[0]), PageAccountDetails)
	}

	for index, listItem := range pg.listItems {
		*common.selectedWallet = index

		if ok, selectedItem := listItem.accountsList.ItemClicked(); ok {
			pg.common.changeFragment(AcctDetailsPage(common, listItem.accounts[selectedItem]), PageAccountDetails)
		}

		if listItem.wal.IsWatchingOnlyWallet() {
			for listItem.moreButton.Button.Clicked() {
				pg.openPopup(index)
			}
		} else {
			for listItem.collapsible.MoreTriggered() {
				pg.openPopup(index)
			}
		}

		for _, menu := range listItem.optionsMenu {
			if menu.button.Clicked() {
				switch menu.id {
				case PageSignMessage:
					common.changeFragment(SignMessagePage(common, listItem.wal), PageSignMessage)
				case PagePrivacy:
					common.changeFragment(PrivacyPage(common, listItem.wal), PagePrivacy)
				case PageSettings:
					common.changeFragment(WalletSettingsPage(common, listItem.wal), PageWalletSettings)
				default:
					menu.action(common)
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
}

func (pg *walletPage) onClose() {
	pg.closePopups()
}
