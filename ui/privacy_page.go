package ui

import (
	"fmt"

	"golang.org/x/exp/shiny/materialdesign/icons"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/decred/dcrd/dcrutil/v3"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/modal"
	"github.com/planetdecred/godcr/ui/page"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"
)

const PagePrivacy = "Privacy"

type privacyPage struct {
	wallet                               *dcrlibwallet.Wallet
	theme                                *decredmaterial.Theme
	common                               *pageCommon
	pageContainer                        layout.List
	toggleMixer, allowUnspendUnmixedAcct *widget.Bool
	toPrivacySetup                       decredmaterial.Button
	dangerZoneCollapsible                *decredmaterial.Collapsible
	acctMixerStatus                      *chan *wallet.AccountMixer

	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton
}

func PrivacyPage(common *pageCommon, wallet *dcrlibwallet.Wallet) Page {
	pg := &privacyPage{
		wallet:                  wallet,
		theme:                   common.theme,
		common:                  common,
		pageContainer:           layout.List{Axis: layout.Vertical},
		toggleMixer:             new(widget.Bool),
		allowUnspendUnmixedAcct: new(widget.Bool),
		toPrivacySetup:          common.theme.Button(new(widget.Clickable), "Set up mixer for this wallet"),
		dangerZoneCollapsible:   common.theme.Collapsible(),
		acctMixerStatus:         common.acctMixerStatus,
	}
	pg.toPrivacySetup.Background = pg.theme.Color.Primary
	pg.backButton, pg.infoButton = common.SubPageHeaderButtons()

	return pg
}

func (pg *privacyPage) OnResume() {

}

func (pg *privacyPage) Layout(gtx layout.Context) layout.Dimensions {
	c := pg.common
	d := func(gtx C) D {
		load := SubPage{
			title:      "StakeShuffle",
			walletName: pg.wallet.Name,
			backButton: pg.backButton,
			infoButton: pg.infoButton,
			back: func() {
				c.changePage(page.WalletPageID)
			},
			infoTemplate: modal.PrivacyInfoTemplate,
			body: func(gtx layout.Context) layout.Dimensions {
				if pg.wallet.AccountMixerConfigIsSet() {
					widgets := []func(gtx C) D{
						func(gtx C) D {
							return pg.mixerInfoLayout(gtx, c)
						},
						pg.gutter,
						func(gtx C) D {
							return pg.mixerSettingsLayout(gtx, c)
						},
						pg.gutter,
						func(gtx C) D {
							return pg.dangerZoneLayout(gtx, c)
						},
					}
					return pg.pageContainer.Layout(gtx, len(widgets), func(gtx C, i int) D {
						return widgets[i](gtx)
					})
				}
				return pg.privacyIntroLayout(gtx, c)
			},
		}
		return c.SubPageLayout(gtx, load)
	}
	return c.UniformPadding(gtx, d)
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

	if pg.wallet.IsAccountMixerActive() {
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
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: values.MarginPadding5}.Layout(gtx, subtxt.Layout)
				}),
			)
		}),
	)
}

func (pg *privacyPage) mixersubInfolayout(gtx layout.Context, c *pageCommon) layout.Dimensions {
	txt := pg.theme.Body2("")

	if pg.wallet.IsAccountMixerActive() {
		txt = pg.theme.Body2("The mixer will automatically stop when unmixed balance are fully mixed.")
		txt.Color = c.theme.Color.Gray
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(txt.Layout),
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
							var mixedBalance = "0.00"
							var unmixedBalance = "0.00"
							accounts, _ := pg.wallet.GetAccountsRaw()
							for _, acct := range accounts.Acc {
								if acct.Number == pg.wallet.MixedAccountNumber() {
									mixedBalance = dcrutil.Amount(acct.TotalBalance).String()
								} else if acct.Number == pg.wallet.UnmixedAccountNumber() {
									unmixedBalance = dcrutil.Amount(acct.TotalBalance).String()
								}
							}

							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											txt := c.theme.Label(values.TextSize14, "Unmixed balance")
											txt.Color = c.theme.Color.Gray
											return txt.Layout(gtx)
										}),
										layout.Rigid(func(gtx C) D {
											return c.layoutBalance(gtx, unmixedBalance, true)
										}),
									)
								}),
								layout.Rigid(func(gtx C) D {
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
											return c.layoutBalance(gtx, mixedBalance, true)
										}),
									)
								}),
							)
						})
					})
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx C) D {
							return pg.mixersubInfolayout(gtx, c)
						}),
					)
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

func (pg *privacyPage) Handle() {
	common := pg.common
	if pg.toPrivacySetup.Button.Clicked() {
		go pg.showModalSetupMixerInfo(common)
	}

	if pg.toggleMixer.Changed() {
		if pg.toggleMixer.Value {
			go pg.showModalPasswordStartAccountMixer(common)
		} else {
			go common.multiWallet.StopAccountMixer(pg.wallet.ID)
		}
	}

	select {
	case stt := <-*pg.acctMixerStatus:
		if stt.RunStatus == wallet.MixerStarted {
			common.notify("Start Successfully", true)
		} else {
			common.notify("Stop Successfully", true)
		}
	default:
	}
}

func (pg *privacyPage) showModalSetupMixerInfo(common *pageCommon) {
	info := newInfoModal(common).
		title("Set up mixer by creating two needed accounts").
		body("Each time you receive a payment, a new address is generated to protect your privacy.").
		negativeButton(values.String(values.StrCancel), func() {}).
		positiveButton("Begin setup", func() {
			pg.showModalSetupMixerAcct(common)
		})
	common.showModal(info)
}

func (pg *privacyPage) showModalSetupMixerAcct(common *pageCommon) {
	accounts, _ := pg.wallet.GetAccountsRaw()
	for _, acct := range accounts.Acc {
		if acct.Name == "mixed" || acct.Name == "unmixed" {
			alert := mustIcon(widget.NewIcon(icons.AlertError))
			alert.Color = pg.theme.Color.DeepBlue

			info := newInfoModal(common).
				icon(alert).
				title("Account name is taken").
				body("There are existing accounts named mixed or unmixed. Please change the name to something else for now. You can change them back after the setup.").
				positiveButton("Go back & rename", func() {
					*common.page = page.WalletPageID
				})
			common.showModal(info)
			return
		}
	}

	newPasswordModal(common).
		title("Confirm to create needed accounts").
		negativeButton("Cancel", func() {}).
		positiveButton("Confirm", func(password string, pm *passwordModal) bool {
			go func() {
				err := pg.wallet.CreateMixerAccounts("mixed", "unmixed", password)
				if err != nil {
					pm.setError(err.Error())
					pm.setLoading(false)
					return
				}
				pm.Dismiss()
			}()

			return false
		}).Show()
}

func (pg *privacyPage) showModalPasswordStartAccountMixer(common *pageCommon) {
	newPasswordModal(common).
		title("Confirm to mix account").
		negativeButton("Cancel", func() {}).
		positiveButton("Confirm", func(password string, pm *passwordModal) bool {
			go func() {

				err := common.multiWallet.StartAccountMixer(pg.wallet.ID, password)
				if err != nil {
					pm.setError(err.Error())
					pm.setLoading(false)
					return
				}
				pm.Dismiss()
				common.notify("Start Successfully", true)
			}()

			return false
		}).Show()
}

func (pg *privacyPage) OnClose() {}
