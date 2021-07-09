package ui

import (
	"os"

	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/page"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const Start = "start_page"

type startPage struct {
	*pageCommon
	load *load.Load

	loading bool

	decredSymbol  *widget.Image
	networkType   decredmaterial.Label
	loadStatus    decredmaterial.Label
	welcomeText   decredmaterial.Label
	createButton  decredmaterial.Button
	restoreButton decredmaterial.Button
}

func newStartPage(common *pageCommon, l *load.Load) Page {
	sp := &startPage{
		pageCommon: common,
		load:       l,

		loading: true,

		decredSymbol: common.icons.decredSymbolIcon,
		networkType:  common.theme.Label(values.TextSize20, "testnet"),
		loadStatus:   common.theme.Label(values.TextSize20, "Loading"),
		welcomeText:  common.theme.Label(values.TextSize24, "Welcome to Decred Wallet, a secure & open-source mobile wallet."),

		createButton:  common.theme.Button(new(widget.Clickable), "Create a new wallet"),
		restoreButton: common.theme.Button(new(widget.Clickable), "Restore an existing wallet"),
	}

	sp.decredSymbol.Scale = 0.5

	sp.networkType.Color = sp.theme.Color.DeepBlue
	sp.networkType.Font.Weight = text.Medium

	sp.loadStatus.Color = sp.theme.Color.DeepBlue
	sp.welcomeText.Color = sp.theme.Color.DeepBlue

	return sp
}

func (sp *startPage) OnResume() {
	sp.wallet.InitMultiWallet()
	sp.multiWallet = sp.wallet.GetMultiWallet()
	sp.load.WL.MultiWallet = sp.wallet.GetMultiWallet()

	// refresh theme now that config is available
	sp.refreshTheme()

	if sp.multiWallet.LoadedWalletsCount() > 0 {
		sp.loadStatus.Text = "Opening wallets"

		if sp.multiWallet.IsStartupSecuritySet() {
			sp.unlock()
		} else {
			go sp.openWallets("")
		}

	} else {
		sp.loading = false
	}
}

func (sp *startPage) unlock() {
	newPasswordModal(sp.pageCommon).
		title("Unlock with passphrase").
		negativeButton("Exit", func() {
			sp.multiWallet.Shutdown()
			os.Exit(0)
		}).
		positiveButton("Unlock", func(password string, m *passwordModal) bool {
			go func() {
				err := sp.openWallets(password)
				if err != nil {
					m.setError(translateErr(err))
					m.setLoading(false)
					return
				}

				m.Dismiss()
			}()
			return false
		}).Show()
}

func (sp *startPage) openWallets(passphrase string) error {
	err := sp.multiWallet.OpenWallets([]byte(passphrase))
	if err != nil {
		log.Info("Error opening wallet:", err)
		// show err dialog
		return err
	}

	sp.proceedToMainPage()
	return nil
}

func (sp *startPage) proceedToMainPage() {
	sp.wallet.SetupListeners()
	sp.changeWindowPage(newMainPage(sp.pageCommon, sp.load), false)
}

func (sp *startPage) Handle() {
	for sp.createButton.Button.Clicked() {
		newCreatePasswordModal(sp.pageCommon).
			title("Create new wallet").
			passwordCreated(func(_, password string, m *createPasswordModal) bool {
				go func() {
					_, err := sp.multiWallet.CreateNewWallet("mywallet", password, dcrlibwallet.PassphraseTypePass)
					if err != nil {
						m.setError(err.Error())
						m.setLoading(false)
						return
					}
					m.Dismiss()

					sp.proceedToMainPage()
				}()
				return false
			}).Show()
	}

	for sp.restoreButton.Button.Clicked() {
		sp.changeWindowPage(page.NewCreateRestorePage(sp.load), true)
	}
}

func (sp *startPage) OnClose() {}

func (sp *startPage) Layout(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min = gtx.Constraints.Max // use maximum height & width
	return layout.Stack{Alignment: layout.N}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return sp.loadingSection(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if sp.loading {
						return layout.Dimensions{}
					}

					return sp.buttonSection(gtx)
				}),
			)
		}),
	)
}

func (sp *startPage) loadingSection(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X // use maximum width
	if sp.loading {
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	} else {
		gtx.Constraints.Min.Y = (gtx.Constraints.Max.Y * 75) / 100 // use 75% of view height
	}

	return layout.Stack{Alignment: layout.Center}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle, Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, sp.decredSymbol.Layout)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return sp.networkType.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if sp.loading {
						return layout.Inset{Top: values.MarginPadding24}.Layout(gtx, sp.loadStatus.Layout)
					}

					return layout.Inset{Top: values.MarginPadding24}.Layout(gtx, sp.welcomeText.Layout)
				}),
			)
		}),
	)
}

func (sp *startPage) buttonSection(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X              // use maximum width
	gtx.Constraints.Min.Y = (gtx.Constraints.Max.Y * 25) / 100 // use 25% of view height
	return layout.Stack{Alignment: layout.S}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle, Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: values.MarginPadding24, Left: values.MarginPadding24, Right: values.MarginPadding24}.Layout(gtx, func(gtx C) D {
								gtx.Constraints.Min.X = gtx.Constraints.Max.X
								return sp.createButton.Layout(gtx)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: values.MarginPadding24, Bottom: values.MarginPadding24, Left: values.MarginPadding24, Right: values.MarginPadding24}.Layout(gtx, func(gtx C) D {
								gtx.Constraints.Min.X = gtx.Constraints.Max.X
								return sp.restoreButton.Layout(gtx)
							})
						}),
					)
				}),
			)
		}),
	)
}
