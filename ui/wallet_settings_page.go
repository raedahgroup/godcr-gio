package ui

import (
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const PageWalletSettings = "WalletSettings"

type walletSettingsPage struct {
	theme          *decredmaterial.Theme
	common         pageCommon
	wal            *dcrlibwallet.Wallet
	walletID       int
	walletName     string
	IsWatchingOnly bool

	changePass, rescan, deleteWallet *widget.Clickable

	notificationW *widget.Bool

	chevronRightIcon *widget.Icon
	backButton       decredmaterial.IconButton
	infoButton       decredmaterial.IconButton
}

func WalletSettingsPage(common pageCommon, walletID int) Page {
	wal := common.multiWallet.WalletWithID(walletID)

	pg := &walletSettingsPage{
		theme:          common.theme,
		common:         common,
		wal:            wal,
		walletID:       walletID,
		walletName:     wal.Name,
		IsWatchingOnly: wal.IsWatchingOnlyWallet(),

		notificationW: new(widget.Bool),

		changePass:   new(widget.Clickable),
		rescan:       new(widget.Clickable),
		deleteWallet: new(widget.Clickable),

		chevronRightIcon: common.icons.chevronRight,
	}

	pg.chevronRightIcon.Color = pg.theme.Color.LightGray
	pg.backButton, pg.infoButton = common.SubPageHeaderButtons()

	return pg
}

func (pg *walletSettingsPage) pageID() string {
	return PageWalletSettings
}

func (pg *walletSettingsPage) Layout(gtx layout.Context) layout.Dimensions {
	common := pg.common

	beep := pg.wal.ReadBoolConfigValueForKey(dcrlibwallet.BeepNewBlocksConfigKey, false)
	pg.notificationW.Value = false
	if beep {
		pg.notificationW.Value = true
	}

	body := func(gtx C) D {
		page := SubPage{
			title:      values.String(values.StrSettings),
			walletName: pg.walletName,
			back: func() {
				common.popPage()
			},
			backButton: pg.backButton,
			infoButton: pg.infoButton,
			body: func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if !pg.IsWatchingOnly {
							return pg.changePassphrase()(gtx)
						}
						return layout.Dimensions{}
					}),
					layout.Rigid(pg.notification()),
					layout.Rigid(pg.debug()),
					layout.Rigid(pg.dangerZone()),
				)
			},
			infoTemplate: "",
		}
		return common.SubPageLayout(gtx, page)
	}

	return common.UniformPadding(gtx, body)
}

func (pg *walletSettingsPage) changePassphrase() layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, values.String(values.StrSpendingPassword), pg.changePass, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(pg.bottomSectionLabel(values.String(values.StrChangeSpendingPass))),
				layout.Flexed(1, func(gtx C) D {
					return layout.E.Layout(gtx, func(gtx C) D {
						return pg.chevronRightIcon.Layout(gtx, values.MarginPadding20)
					})
				}),
			)
		})
	}
}

func (pg *walletSettingsPage) notification() layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, values.String(values.StrNotifications), nil, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(pg.bottomSectionLabel(values.String(values.StrBeepForNewBlocks))),
				layout.Flexed(1, func(gtx C) D {
					return layout.E.Layout(gtx, func(gtx C) D {
						return pg.theme.Switch(pg.notificationW).Layout(gtx)
					})
				}),
			)
		})
	}
}

func (pg *walletSettingsPage) debug() layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, values.String(values.StrDebug), pg.rescan, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(pg.bottomSectionLabel(values.String(values.StrRescanBlockchain))),
				layout.Flexed(1, func(gtx C) D {
					return layout.E.Layout(gtx, func(gtx C) D {
						return pg.chevronRightIcon.Layout(gtx, values.MarginPadding20)
					})
				}),
			)
		})
	}
}

func (pg *walletSettingsPage) dangerZone() layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, values.String(values.StrDangerZone), pg.deleteWallet, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(pg.bottomSectionLabel(values.String(values.StrRemoveWallet))),
				layout.Flexed(1, func(gtx C) D {
					return layout.E.Layout(gtx, func(gtx C) D {
						return pg.chevronRightIcon.Layout(gtx, values.MarginPadding20)
					})
				}),
			)
		})
	}
}

func (pg *walletSettingsPage) pageSections(gtx layout.Context, title string, clickable *widget.Clickable, body layout.Widget) layout.Dimensions {
	dims := func(gtx layout.Context, title string, body layout.Widget) D {
		return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					txt := pg.theme.Body2(title)
					txt.Color = pg.theme.Color.Gray
					return txt.Layout(gtx)
				}),
				layout.Rigid(body),
			)
		})
	}

	return layout.Inset{Bottom: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
		return pg.theme.Card().Layout(gtx, func(gtx C) D {
			if clickable == nil {
				return dims(gtx, title, body)
			}
			return decredmaterial.Clickable(gtx, clickable, func(gtx C) D {
				return dims(gtx, title, body)
			})
		})
	})
}

func (pg *walletSettingsPage) bottomSectionLabel(title string) layout.Widget {
	return func(gtx C) D {
		return pg.theme.Body1(title).Layout(gtx)
	}
}

func (pg *walletSettingsPage) handle() {
	common := pg.common
	for pg.changePass.Clicked() {
		go func() {
			common.modalReceiver <- &modalLoad{
				template: ChangePasswordTemplate,
				title:    values.String(values.StrChangeSpendingPass),
				confirm: func(oldPass, newPass string) {
					// pg.wal.ChangeWalletPassphrase(pg.walletID, oldPass, newPass, pg.errorReceiver)
				},
				confirmText: values.String(values.StrChange),
				cancel:      common.closeModal,
				cancelText:  values.String(values.StrCancel),
			}
		}()
		break
	}

	for pg.rescan.Clicked() {
		go func() {
			common.modalReceiver <- &modalLoad{
				template: RescanWalletTemplate,
				title:    values.String(values.StrRescanBlockchain),
				confirm: func() {
					err := pg.common.multiWallet.RescanBlocks(pg.walletID)
					if err != nil {
						if err.Error() == dcrlibwallet.ErrNotConnected {
							common.notify(values.String(values.StrNotConnected), false)
							return
						}
						common.notify(err.Error(), false)
						return
					}
					msg := values.String(values.StrRescanProgressNotification)
					common.notify(msg, true)
					go func() {
						common.modalReceiver <- &modalLoad{}
					}()
				},
				confirmText: values.String(values.StrRescan),
				cancel:      common.closeModal,
				cancelText:  values.String(values.StrCancel),
			}
		}()
		break
	}

	if pg.notificationW.Changed() {
		pg.wal.SaveUserConfigValue(dcrlibwallet.BeepNewBlocksConfigKey, pg.notificationW.Value)
	}

	for pg.deleteWallet.Clicked() {
		go func() {
			common.modalReceiver <- &modalLoad{
				template: ConfirmRemoveTemplate,
				title:    values.String(values.StrRemoveWallet),
				confirm: func() {
					go func() {
						common.modalReceiver <- &modalLoad{
							template: PasswordTemplate,
							title:    values.String(values.StrConfirmToRemove),
							confirm: func(pass string) {
								go func() {
									err := pg.common.multiWallet.DeleteWallet(pg.walletID, []byte(pass))
									if err != nil {
										common.modalLoad.setLoading(false)
										if err.Error() == dcrlibwallet.ErrInvalidPassphrase {
											e := values.String(values.StrInvalidPassphrase)
											common.notify(e, false)
										}
									} else {
										pg.common.popPage()
									}
								}()
							},
							confirmText: values.String(values.StrConfirm),
							cancel:      common.closeModal,
							cancelText:  values.String(values.StrCancel),
						}
					}()
				},
				confirmText: values.String(values.StrRemove),
				cancel:      common.closeModal,
				cancelText:  values.String(values.StrCancel),
			}
		}()
		break
	}
}

func (pg *walletSettingsPage) onClose() {}
