package ui

import (
	"fmt"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/decred/dcrd/dcrutil"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const PagePrivacy = "Privacy"

type privacyPage struct {
	theme                                *decredmaterial.Theme
	common                               pageCommon
	pageContainer                        layout.List
	toggleMixer, allowUnspendUnmixedAcct *widget.Bool
	infoBtn                              decredmaterial.IconButton
	toPrivacySetup                       decredmaterial.Button
	dangerZoneCollapsible                *decredmaterial.Collapsible
	backButton                           decredmaterial.IconButton
	errorReceiver                        chan error
	walletID                             int

	walletName     string
	accounts       []*dcrlibwallet.Account
	mixedBalance   string
	unmixedBalance string
}

func PrivacyPage(common pageCommon, walletID int) Page {
	pg := &privacyPage{
		theme:                   common.theme,
		common:                  common,
		pageContainer:           layout.List{Axis: layout.Vertical},
		toggleMixer:             new(widget.Bool),
		allowUnspendUnmixedAcct: new(widget.Bool),
		toPrivacySetup:          common.theme.Button(new(widget.Clickable), "Set up mixer for this wallet"),
		dangerZoneCollapsible:   common.theme.Collapsible(),
		errorReceiver:           make(chan error),
		// acctMixerStatus:         &win.walletAcctMixerStatus, //TODO
		walletID: walletID,

		walletName: common.wallet.WalletWithID(walletID).Name,
	}
	pg.toPrivacySetup.Background = pg.theme.Color.Primary
	pg.infoBtn = common.theme.IconButton(new(widget.Clickable), common.icons.actionInfo)
	pg.infoBtn.Color = common.theme.Color.Gray
	pg.infoBtn.Background = common.theme.Color.Surface
	pg.infoBtn.Inset = layout.UniformInset(values.MarginPadding0)

	pg.backButton, _ = common.SubPageHeaderButtons()

	accounts, err := common.wallet.WalletWithID(pg.walletID).GetAccountsRaw()
	if err != nil {
		log.Error("error getting accounts:", err)
	}
	pg.accounts = accounts.Acc
	pg.updateBalance()

	return pg
}

func (pg *privacyPage) pageID() string {
	return PagePrivacy
}

func (pg *privacyPage) Layout(gtx layout.Context) layout.Dimensions {
	c := pg.common
	d := func(gtx C) D {
		load := SubPage{
			title:      "Privacy",
			walletName: pg.walletName,
			back: func() {
				c.popPage()
			},
			backButton:   pg.backButton,
			infoButton:   pg.infoBtn,
			infoTemplate: PrivacyInfoTemplate,
			body: func(gtx layout.Context) layout.Dimensions {
				if c.wallet.IsAccountMixerConfigSet(pg.walletID) {
					widgets := []func(gtx C) D{
						func(gtx C) D {
							return pg.mixerInfoLayout(gtx, &c)
						},
						pg.gutter,
						func(gtx C) D {
							return pg.mixerSettingsLayout(gtx, &c)
						},
						pg.gutter,
						func(gtx C) D {
							return pg.dangerZoneLayout(gtx, &c)
						},
					}
					return pg.pageContainer.Layout(gtx, len(widgets), func(gtx C, i int) D {
						return widgets[i](gtx)
					})
				}
				return pg.privacyIntroLayout(gtx, &c)
			},
		}
		return c.SubPageLayout(gtx, load)
	}

	return pg.common.UniformPadding(gtx, d)
}

func (pg *privacyPage) privacyIntroLayout(gtx layout.Context, c *pageCommon) layout.Dimensions {
	return pg.theme.Card().Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Flexed(1, func(gtx C) D {
				return layout.Center.Layout(gtx, func(gtx C) D {
					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return layout.Inset{Right: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
										c.icons.transactionFingerPrintIcon.Scale = 1.0
										return c.icons.transactionFingerPrintIcon.Layout(gtx)
									})
								}),
								layout.Rigid(func(gtx C) D {
									c.icons.arrowForwardIcon.Scale = 0.5
									return c.icons.arrowForwardIcon.Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									c.icons.mixerSmall.Scale = 1.0
									return layout.Inset{
										Left:  values.MarginPadding5,
										Right: values.MarginPadding5,
									}.Layout(gtx, c.icons.mixerSmall.Layout)
								}),
								layout.Rigid(func(gtx C) D {
									c.icons.arrowForwardIcon.Scale = 0.5
									return c.icons.arrowForwardIcon.Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									c.icons.transactionIcon.Scale = 1.5
									return layout.Inset{Left: values.MarginPadding10}.Layout(gtx, c.icons.transactionIcon.Layout)
								}),
							)
						}),
						layout.Rigid(func(gtx C) D {
							txt := pg.theme.H6("How does CoinShuffle++ mixer enhance your privacy?")
							txt2 := pg.theme.Body1("CoinShuffle++ mixer can mix your DCRs through CoinJoin transactions.")
							txt3 := pg.theme.Body1("Using mixed DCRs protects you from exposing your financial activities to the public (e.g. how much you own, who pays you).")
							txt.Alignment, txt2.Alignment, txt3.Alignment = text.Middle, text.Middle, text.Middle

							return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(txt.Layout),
								layout.Rigid(func(gtx C) D {
									return layout.Inset{Top: values.MarginPadding10}.Layout(gtx, txt2.Layout)
								}),
								layout.Rigid(func(gtx C) D {
									return layout.Inset{Top: values.MarginPadding10}.Layout(gtx, txt3.Layout)
								}),
							)
						}),
					)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return layout.UniformInset(values.MarginPadding15).Layout(gtx, pg.toPrivacySetup.Layout)
			}),
		)
	})
}

func (pg *privacyPage) mixerInfoStatusTextLayout(gtx layout.Context, c *pageCommon) layout.Dimensions {
	txt := pg.theme.H6("Mixer")
	subtxt := pg.theme.Body2("Ready to mix")
	subtxt.Color = c.theme.Color.Gray
	iconVisibility := false

	if c.wallet.IsAccountMixerActive(pg.walletID) {
		txt.Text = "Mixer is running..."
		subtxt.Text = "Keep this app opened"
		iconVisibility = true
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(txt.Layout),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					if !iconVisibility {
						return layout.Dimensions{}
					}
					c.icons.alertGray.Scale = 1.0
					return c.icons.alertGray.Layout(gtx)
				}),
				layout.Rigid(subtxt.Layout),
			)
		}),
	)
}

func (pg *privacyPage) mixerInfoLayout(gtx layout.Context, c *pageCommon) layout.Dimensions {
	return c.theme.Card().Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							ic := c.icons.mixerSmall
							ic.Scale = 1.0
							return ic.Layout(gtx)
						}),
						layout.Flexed(1, func(gtx C) D {
							return layout.Inset{Left: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
								return pg.mixerInfoStatusTextLayout(gtx, c)
							})
						}),
						layout.Rigid(material.Switch(pg.theme.Base, pg.toggleMixer).Layout),
					)
				}),
				layout.Rigid(pg.gutter),
				layout.Rigid(func(gtx C) D {
					content := c.theme.Card()
					content.Color = c.theme.Color.LightGray
					return content.Layout(gtx, func(gtx C) D {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											txt := c.theme.Label(values.TextSize14, "Unmixed balance")
											txt.Color = c.theme.Color.Gray
											return txt.Layout(gtx)
										}),
										layout.Rigid(func(gtx C) D {
											return c.layoutBalance(gtx, pg.unmixedBalance, true)
										}),
									)
								}),
								layout.Rigid(func(gtx C) D {
									if !c.wallet.IsAccountMixerActive(pg.walletID) {
										return layout.Dimensions{}
									}
									c.icons.arrowDownIcon.Scale = 1.0
									return layout.Center.Layout(gtx, c.icons.arrowDownIcon.Layout)
								}),
								layout.Rigid(func(gtx C) D {
									return layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											t := c.theme.Label(values.TextSize14, "Mixed balance")
											t.Color = c.theme.Color.Gray
											return t.Layout(gtx)
										}),
										layout.Rigid(func(gtx C) D {
											return c.layoutBalance(gtx, pg.mixedBalance, true)
										}),
									)
								}),
							)
						})
					})
				}),
			)
		})
	})
}

func (pg *privacyPage) mixerSettingsLayout(gtx layout.Context, c *pageCommon) layout.Dimensions {
	return c.theme.Card().Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X

		row := func(txt1, txt2 string) D {
			return layout.Inset{
				Left:   values.MarginPadding15,
				Right:  values.MarginPadding15,
				Top:    values.MarginPadding10,
				Bottom: values.MarginPadding10,
			}.Layout(gtx, func(gtx C) D {
				return layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(c.theme.Label(values.TextSize16, txt1).Layout),
					layout.Rigid(c.theme.Body2(txt2).Layout),
				)
			})
		}

		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.UniformInset(values.MarginPadding15).Layout(gtx, c.theme.Body2("Mixer Settings").Layout)
			}),
			layout.Rigid(func(gtx C) D { return row("Mixed account", "mixed") }),
			layout.Rigid(pg.theme.Separator().Layout),
			layout.Rigid(func(gtx C) D { return row("Change account", "unmixed") }),
			layout.Rigid(pg.theme.Separator().Layout),
			layout.Rigid(func(gtx C) D { return row("Account branch", fmt.Sprintf("%d", dcrlibwallet.MixedAccountBranch)) }),
			layout.Rigid(pg.theme.Separator().Layout),
			layout.Rigid(func(gtx C) D { return row("Shuffle server", dcrlibwallet.ShuffleServer) }),
			layout.Rigid(pg.theme.Separator().Layout),
			layout.Rigid(func(gtx C) D { return row("Shuffle port", pg.shufflePortForCurrentNet(c)) }),
		)
	})
}

func (pg *privacyPage) shufflePortForCurrentNet(c *pageCommon) string {
	if c.wallet.Net == "testnet3" {
		return dcrlibwallet.TestnetShufflePort
	}

	return dcrlibwallet.MainnetShufflePort
}

func (pg *privacyPage) dangerZoneLayout(gtx layout.Context, c *pageCommon) layout.Dimensions {
	return c.theme.Card().Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
			return pg.dangerZoneCollapsible.Layout(gtx,
				func(gtx C) D {
					txt := pg.theme.Label(values.MarginPadding15, "Danger Zone")
					txt.Color = c.theme.Color.Gray
					return txt.Layout(gtx)
				},
				func(gtx C) D {
					return layout.Inset{Top: values.MarginPadding15}.Layout(gtx, func(gtx C) D {
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							layout.Flexed(1, c.theme.Label(values.TextSize16, "Allow spending from unmixed accounts").Layout),
							layout.Rigid(material.Switch(pg.theme.Base, pg.allowUnspendUnmixedAcct).Layout),
						)
					})
				},
			)
		})
	})
}

func (pg *privacyPage) gutter(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
		return layout.Dimensions{}
	})
}

// TODO: call on account mixer notifications
func (pg *privacyPage) updateBalance() {
	var mixedBalance dcrutil.Amount
	var unmixedBalance dcrutil.Amount

	wal := pg.common.wallet.WalletWithID(pg.walletID)

	for _, acct := range pg.accounts {
		if acct.Number == wal.ReadInt32ConfigValueForKey(dcrlibwallet.AccountMixerMixedAccount, -1) {
			mixedBalance = dcrutil.Amount(acct.Balance.Spendable)
		} else if acct.Number == wal.ReadInt32ConfigValueForKey(dcrlibwallet.AccountMixerUnmixedAccount, -1) {
			unmixedBalance = dcrutil.Amount(acct.Balance.Spendable)
		}
	}

	pg.mixedBalance = mixedBalance.String()
	pg.unmixedBalance = unmixedBalance.String()
}

func (pg *privacyPage) handle() {
	common := pg.common
	if pg.toPrivacySetup.Button.Clicked() {
		go pg.showModalSetupMixerInfo(&common)
	}

	if pg.toggleMixer.Changed() {
		if pg.toggleMixer.Value {
			go pg.showModalPasswordStartAccountMixer(&common)
		} else {
			common.wallet.StopAccountMixer(pg.walletID, pg.errorReceiver)
		}
	}

	// TODO: redundant
	pg.updateBalance()

	select {
	case err := <-pg.errorReceiver:
		common.modalLoad.setLoading(false)
		common.notify(err.Error(), false)
	// case stt := <-*pg.acctMixerStatus:
	// 	if stt.RunStatus == wallet.MixerStarted {
	// 		common.notify("Start Successfully", true)
	// 		common.closeModal()
	// 	} else {
	// 		common.notify("Stop Successfully", true)
	// 	}
	default:
	}
}

func (pg *privacyPage) showModalSetupMixerInfo(common *pageCommon) {
	common.modalReceiver <- &modalLoad{
		template: SetupMixerInfoTemplate,
		title:    "Set up mixer by creating two needed accounts",
		confirm: func() {
			go pg.showModalSetupMixerAcct(common)
		},
		confirmText: "Begin setup",
		cancel:      common.closeModal,
		cancelText:  "Cancel",
	}
}

func (pg *privacyPage) showModalSetupMixerAcct(common *pageCommon) {
	common.modalReceiver <- &modalLoad{
		template: PasswordTemplate,
		title:    "Confirm to create needed accounts",
		confirm: func(p string) {

			// TODO simplify
			accounts, err := common.wallet.WalletWithID(pg.walletID).GetAccountsRaw()
			if err != nil {
				log.Error("error getting accounts:", err)
				return
			}
			for _, acct := range accounts.Acc {
				if acct.Name == "mixed" || acct.Name == "unmixed" {
					go pg.showModalSetupExistAcct(common)
					return
				}
			}
			common.wallet.SetupAccountMixer(pg.walletID, p, pg.errorReceiver)
		},
		confirmText: "Confirm",
		cancel:      common.closeModal,
		cancelText:  "Cancel",
	}
}

func (pg *privacyPage) showModalSetupExistAcct(common *pageCommon) {
	common.modalReceiver <- &modalLoad{
		template:    ConfirmMixerAcctExistTemplate,
		confirmText: "Go back & rename",
		cancel:      common.closeModal,
		confirm: func() {
			common.closeModal()
			// *common.page = PageWallet TODO
		},
	}
}

func (pg *privacyPage) showModalPasswordStartAccountMixer(common *pageCommon) {
	common.modalReceiver <- &modalLoad{
		template:    PasswordTemplate,
		title:       "Confirm to mix account",
		confirmText: "Confirm",
		cancel: func() {
			common.closeModal()
			pg.toggleMixer.Value = false
		},
		cancelText: "Cancel",
		confirm: func(pass string) {
			common.wallet.StartAccountMixer(pg.walletID, pass, pg.errorReceiver)
		},
	}
}

func (pg *privacyPage) onClose() {}
