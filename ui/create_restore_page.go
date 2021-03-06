package ui

import (
	"fmt"
	"image/color"
	"strings"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

const (
	PageCreateRestore = "CreateRestore"
	numberOfSeeds     = 32
)

type (
	seedEditors struct {
		focusIndex int
		editors    []decredmaterial.RestoreEditor
	}
)

type seedItemMenu struct {
	text   string
	button decredmaterial.Button
}

type createRestore struct {
	common          *pageCommon
	theme           *decredmaterial.Theme
	restoringWallet bool
	keyEvent        chan *key.Event
	seedPhrase      string
	suggestionLimit int
	suggestions     []string
	allSuggestions  []string
	seedClicked     bool
	focused         []int
	seedMenu        []seedItemMenu
	openPopupIndex  int
	selected        int

	closePageBtn     decredmaterial.IconButton
	restoreWalletBtn decredmaterial.Button
	resetSeedFields  decredmaterial.Button

	spendingPassword      decredmaterial.Editor
	walletName            decredmaterial.Editor
	matchSpendingPassword decredmaterial.Editor
	errLabel              decredmaterial.Label
	optionsMenuCard       decredmaterial.Card

	passwordStrength decredmaterial.ProgressBarStyle

	seedEditors seedEditors

	seedList         *layout.List
	restoreContainer layout.List

	createModal     *decredmaterial.Modal
	warningModal    *decredmaterial.Modal
	modalTitleLabel decredmaterial.Label

	alertIcon *widget.Image
}

// Loading lays out the loading widget with a faded background
func CreateRestorePage(common *pageCommon) Page {
	pg := &createRestore{
		common:   common,
		theme:    common.theme,
		keyEvent: common.keyEvents,

		errLabel:              common.theme.Body1(""),
		spendingPassword:      common.theme.EditorPassword(new(widget.Editor), "Spending password"),
		walletName:            common.theme.Editor(new(widget.Editor), "Wallet name"),
		matchSpendingPassword: common.theme.EditorPassword(new(widget.Editor), "Confirm spending password"),
		suggestionLimit:       3,
		createModal:           common.theme.Modal(),
		warningModal:          common.theme.Modal(),
		modalTitleLabel:       common.theme.H6(""),
		passwordStrength:      common.theme.ProgressBar(0),
		openPopupIndex:        -1,
		restoreContainer: layout.List{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		},
	}

	if pg.common.multiWallet.LoadedWalletsCount() == 0 {
		pg.walletName.Editor.SetText("mywallet")
	}

	pg.optionsMenuCard = decredmaterial.Card{Color: pg.theme.Color.Surface}
	pg.optionsMenuCard.Radius = decredmaterial.CornerRadius{NE: 5, NW: 5, SE: 5, SW: 5}

	pg.restoreWalletBtn = common.theme.Button(new(widget.Clickable), "Restore")

	pg.closePageBtn = common.theme.IconButton(new(widget.Clickable), mustIcon(widget.NewIcon(icons.NavigationClose)))
	pg.closePageBtn.Background = color.NRGBA{}
	pg.closePageBtn.Color = common.theme.Color.Hint

	pg.resetSeedFields = common.theme.Button(new(widget.Clickable), "Clear all")
	pg.resetSeedFields.Color = common.theme.Color.Hint
	pg.resetSeedFields.Background = color.NRGBA{}

	pg.alertIcon = common.icons.alertGray
	pg.alertIcon.Scale = 1.0

	pg.restoreWalletBtn.Inset = layout.Inset{
		Top:    values.MarginPadding12,
		Bottom: values.MarginPadding12,
		Right:  values.MarginPadding50,
		Left:   values.MarginPadding50,
	}
	pg.restoreWalletBtn.Background = common.theme.Color.InactiveGray
	pg.restoreWalletBtn.TextSize = values.TextSize16
	pg.errLabel.Color = pg.theme.Color.Danger

	pg.passwordStrength.Color = pg.theme.Color.LightGray

	for i := 0; i <= numberOfSeeds; i++ {
		widgetEditor := new(widget.Editor)
		widgetEditor.SingleLine, widgetEditor.Submit = true, true
		pg.seedEditors.editors = append(pg.seedEditors.editors, common.theme.RestoreEditor(widgetEditor, "", fmt.Sprintf("%d", i+1)))
	}
	pg.seedEditors.focusIndex = -1

	// init suggestion buttons
	pg.initSeedMenu()

	pg.seedList = &layout.List{Axis: layout.Vertical}
	pg.spendingPassword.Editor.SingleLine, pg.matchSpendingPassword.Editor.SingleLine = true, true
	pg.walletName.Editor.SingleLine = true

	pg.allSuggestions = dcrlibwallet.PGPWordList()

	return pg
}

func (pg *createRestore) OnResume() {

}

func (pg *createRestore) Layout(gtx layout.Context) layout.Dimensions {
	pd := values.MarginPadding15
	dims := layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if pg.restoringWallet {
				new(widget.Clickable).Layout(gtx)
				return layout.Inset{Top: pd, Left: pd, Right: pd}.Layout(gtx, pg.processing)
			}
			return pg.restore(gtx)
		}),
	)
	return dims
}

func (pg *createRestore) restore(gtx layout.Context) layout.Dimensions {
	op.TransformOp{}.Add(gtx.Ops)
	paint.Fill(gtx.Ops, pg.theme.Color.LightGray)
	dims := layout.Stack{Alignment: layout.S}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: values.MarginPadding20}.Layout(gtx, func(gtx C) D {
						return layout.W.Layout(gtx, func(gtx C) D {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return layout.Inset{Top: values.MarginPadding6}.Layout(gtx, pg.closePageBtn.Layout)
								}),
								layout.Rigid(func(gtx C) D {
									return layout.Inset{Top: values.MarginPadding16}.Layout(gtx, pg.theme.H6("Restore wallet").Layout)
								}),
							)
						})
					})
				}),
				layout.Rigid(func(gtx C) D {
					m := values.MarginPadding24
					v := values.MarginPadding6
					return Container{padding: layout.Inset{Right: m, Left: m, Top: v, Bottom: m}}.Layout(gtx, func(gtx C) D {
						pageContent := []func(gtx C) D{
							func(gtx C) D {
								return pg.restorePageSections(gtx, "Enter your seed phase", "1/3", pg.enterSeedPhase)
							},
							func(gtx C) D {
								return pg.restorePageSections(gtx, "Create spending password", "2/3", pg.createPasswordPhase)
							},
							func(gtx C) D {
								return pg.restorePageSections(gtx, "Chose a wallet name", "3/3", pg.renameWalletPhase)
							},
						}
						return layout.Inset{Bottom: values.MarginPadding60}.Layout(gtx, func(gtx C) D {
							return pg.restoreContainer.Layout(gtx, len(pageContent), func(gtx C, i int) D {
								return layout.Inset{Bottom: values.MarginPadding4}.Layout(gtx, pageContent[i])
							})
						})
					})
				}),
			)
		}),
		layout.Stacked(func(gtx C) D {
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			return layout.S.Layout(gtx, func(gtx C) D {
				return layout.Inset{Left: values.MarginPadding1}.Layout(gtx, pg.restoreButtonSection)
			})
		}),
	)
	return dims
}

func (pg *createRestore) restoreButtonSection(gtx layout.Context) layout.Dimensions {
	card := pg.theme.Card()
	card.Radius = decredmaterial.CornerRadius{NE: 0, NW: 0, SE: 0, SW: 0}
	return card.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Flexed(1, func(gtx C) D {
				return layout.E.Layout(gtx, func(gtx C) D {
					return layout.Inset{
						Top:    values.MarginPadding16,
						Bottom: values.MarginPadding16,
						Right:  values.MarginPadding16,
					}.Layout(gtx, pg.restoreWalletBtn.Layout)
				})
			}),
		)
	})
}

func (pg *createRestore) enterSeedPhase(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			inset := layout.Inset{
				Right: values.MarginPadding5,
			}
			return layout.Flex{}.Layout(gtx,
				layout.Flexed(1, func(gtx C) D {
					return inset.Layout(gtx, func(gtx C) D {
						return pg.inputsGroup(gtx, pg.seedList, 7, 0)
					})
				}),
				layout.Flexed(1, func(gtx C) D {
					return inset.Layout(gtx, func(gtx C) D {
						return pg.inputsGroup(gtx, pg.seedList, 7, 1)
					})
				}),
				layout.Flexed(1, func(gtx C) D {
					return inset.Layout(gtx, func(gtx C) D {
						return pg.inputsGroup(gtx, pg.seedList, 7, 2)
					})
				}),
				layout.Flexed(1, func(gtx C) D {
					return inset.Layout(gtx, func(gtx C) D {
						return pg.inputsGroup(gtx, pg.seedList, 6, 3)
					})
				}),
				layout.Flexed(1, func(gtx C) D {
					return pg.inputsGroup(gtx, pg.seedList, 6, 4)
				}),
			)
		}),
		layout.Rigid(pg.errLabel.Layout),
		layout.Rigid(pg.resetSeedFields.Layout),
	)

}

func (pg *createRestore) createPasswordPhase(gtx layout.Context) layout.Dimensions {
	phaseContents := []func(gtx C) D{
		func(gtx C) D {
			card := pg.theme.Card()
			card.Color = pg.theme.Color.LightGray
			msg := "This spending password is required to sign transactions. Make sure to use a strong password and keep it safe."
			return card.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return layout.UniformInset(values.MarginPadding16).Layout(gtx, func(gtx C) D {
					return layout.Flex{}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							inset := layout.Inset{
								Right: values.MarginPadding10,
								Top:   values.MarginPadding3,
							}
							return inset.Layout(gtx, pg.alertIcon.Layout)
						}),
						layout.Rigid(pg.theme.Body1(msg).Layout),
					)
				})
			})
		},
		func(gtx C) D {
			return pg.spendingPassword.Layout(gtx)
		},
		func(gtx C) D {
			return pg.passwordStrength.Layout(gtx)
		},
		func(gtx C) D {
			return pg.matchSpendingPassword.Layout(gtx)
		},
	}

	return (&layout.List{Axis: layout.Vertical}).Layout(gtx, len(phaseContents), func(gtx C, i int) D {
		return layout.UniformInset(values.MarginPadding5).Layout(gtx, phaseContents[i])
	})
}

func (pg *createRestore) renameWalletPhase(gtx layout.Context) layout.Dimensions {
	return pg.walletName.Layout(gtx)
}

func (pg *createRestore) restorePageSections(gtx layout.Context, title string, phaseProgress string, body layout.Widget) layout.Dimensions {
	return layout.Inset{Bottom: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
		return pg.theme.Card().Layout(gtx, func(gtx C) D {
			return layout.UniformInset(values.MarginPadding16).Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						m := values.MarginPadding10
						v := values.MarginPadding5
						return Container{padding: layout.Inset{Right: v, Left: v, Bottom: m}}.Layout(gtx, func(gtx C) D {
							gtx.Constraints.Min.X = gtx.Constraints.Max.X
							txt := pg.theme.Body1(title)
							return layout.Flex{
								Axis:    layout.Horizontal,
								Spacing: layout.SpaceBetween,
							}.Layout(gtx,
								layout.Rigid(txt.Layout),
								layout.Rigid(func(gtx C) D {
									border := widget.Border{
										Color:        pg.theme.Color.Gray1,
										CornerRadius: values.MarginPadding14,
										Width:        values.MarginPadding1,
									}
									phase := pg.theme.Body2(phaseProgress)
									return border.Layout(gtx, func(gtx C) D {
										m := values.MarginPadding8
										v := values.MarginPadding5
										return Container{padding: layout.Inset{
											Right:  m,
											Left:   m,
											Top:    v,
											Bottom: v,
										}}.Layout(gtx, phase.Layout)
									})
								}),
							)
						})
					}),
					layout.Rigid(body),
				)
			})
		})
	})
}

func (pg *createRestore) processing(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			message := pg.theme.H3("")
			message.Alignment = text.Middle
			message.Text = "restoring wallet..."
			return layout.Center.Layout(gtx, message.Layout)
		}))
}

func (pg *createRestore) inputsGroup(gtx layout.Context, l *layout.List, len, startIndex int) layout.Dimensions {
	return layout.Stack{Alignment: layout.N}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			return l.Layout(gtx, len, func(gtx C, i int) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Bottom: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
							pg.layoutSeedMenu(gtx, i*5+startIndex)
							return pg.seedEditors.editors[i*5+startIndex].Layout(gtx)
						})
					}),
				)
			})
		}),
	)
}

func (pg *createRestore) onSuggestionSeedsClicked() {
	index := pg.seedEditors.focusIndex
	for i, b := range pg.seedMenu {
		for pg.seedMenu[i].button.Button.Clicked() {
			pg.seedEditors.editors[index].Edit.Editor.SetText(b.text)
			pg.seedEditors.editors[index].Edit.Editor.MoveCaret(len(b.text), 0)
			pg.seedClicked = true
			if index != numberOfSeeds {
				pg.seedEditors.editors[index+1].Edit.Editor.Focus()
			}
		}
	}
}

func diff(a, b []int) []int {
	temp := map[int]int{}
	for _, s := range a {
		temp[s]++
	}
	for _, s := range b {
		temp[s]--
	}

	var f []int
	for s, v := range temp {
		if v != 0 {
			f = append(f, s)
		}
	}
	return f
}

func (pg *createRestore) editorSeedsEventsHandler() {
	var focused []int

	seedEvent := func(i int, text string) {
		if pg.seedClicked {
			pg.seedEditors.focusIndex = -1
			pg.seedClicked = false
		} else {
			pg.seedEditors.focusIndex = i
		}

		if text == "" {
			pg.openPopupIndex = -1
		} else {
			pg.openPopupIndex = i
		}
	}

	for i := 0; i < len(pg.seedEditors.editors); i++ {
		editor := &pg.seedEditors.editors[i]
		text := editor.Edit.Editor.Text()

		if editor.Edit.Editor.Focused() {
			seedEvent(i, text)
			focused = append(focused, i)
		}

		for _, e := range editor.Edit.Editor.Events() {
			switch e.(type) {
			case widget.ChangeEvent:
				seedEvent(i, text)
				pg.resetSeedFields.Color = pg.theme.Color.Primary
				pg.suggestions = pg.suggestionSeeds(text)
				pg.seedMenu = pg.seedMenu[:len(pg.suggestions)]
				for k, s := range pg.suggestions {
					pg.seedMenu[k].text, pg.seedMenu[k].button.Text = s, s
				}
			case widget.SubmitEvent:
				if i != numberOfSeeds {
					pg.seedEditors.editors[i+1].Edit.Editor.Focus()
					pg.selected = 0
				}
			}
		}
	}

	if len(diff(pg.focused, focused)) > 0 {
		pg.seedEditors.focusIndex = -1
	}
	pg.focused = focused
}

func (pg *createRestore) initSeedMenu() {
	for i := 0; i < pg.suggestionLimit; i++ {
		btn := pg.theme.Button(new(widget.Clickable), "")
		btn.Background, btn.Color = color.NRGBA{}, pg.theme.Color.DeepBlue
		pg.seedMenu = append(pg.seedMenu, seedItemMenu{
			text:   "",
			button: btn,
		})
	}
}

func (pg *createRestore) suggestionSeedEffect() {
	for k := range pg.suggestions {
		if pg.selected == k || pg.seedMenu[k].button.Button.Hovered() {
			pg.seedMenu[k].button.Background = pg.theme.Color.LightGray
		} else {
			pg.seedMenu[k].button.Background = color.NRGBA{}
		}
	}
}

func (pg *createRestore) layoutSeedMenu(gtx layout.Context, optionsSeedMenuIndex int) {
	if pg.openPopupIndex != optionsSeedMenuIndex || pg.openPopupIndex != pg.seedEditors.focusIndex {
		return
	}

	inset := layout.Inset{
		Top:  values.MarginPadding35,
		Left: values.MarginPadding0,
	}

	m := op.Record(gtx.Ops)
	inset.Layout(gtx, func(gtx C) D {
		border := widget.Border{Color: pg.theme.Color.LightGray, CornerRadius: values.MarginPadding5, Width: values.MarginPadding2}
		return border.Layout(gtx, func(gtx C) D {
			return pg.optionsMenuCard.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return (&layout.List{Axis: layout.Vertical}).Layout(gtx, len(pg.seedMenu), func(gtx C, i int) D {
					return layout.UniformInset(values.MarginPadding0).Layout(gtx, pg.seedMenu[i].button.Layout)
				})
			})
		})
	})
	op.Defer(gtx.Ops, m.Stop())
}

func (pg createRestore) suggestionSeeds(text string) []string {
	var seeds []string
	if text == "" {
		return seeds
	}

	for _, word := range pg.allSuggestions {
		if strings.HasPrefix(strings.ToLower(word), strings.ToLower(text)) {
			if len(seeds) < pg.suggestionLimit {
				seeds = append(seeds, word)
			}
		}
	}
	return seeds
}

func (pg *createRestore) validateWalletName() string {
	name := pg.walletName.Editor.Text()
	if name == "" {
		pg.errLabel.Text = "wallet name required and cannot be empty"
	}

	return name
}

func (pg *createRestore) validatePassword() string {
	pass := pg.spendingPassword.Editor.Text()
	if pass == "" {
		pg.spendingPassword.HintColor = pg.theme.Color.Danger
		pg.errLabel.Text = "wallet password required and cannot be empty"
		return ""
	}

	return pass
}

func (pg *createRestore) validatePasswords() string {
	pass := pg.validatePassword()
	if pass == "" {
		return ""
	}

	match := pg.matchSpendingPassword.Editor.Text()
	if match == "" {
		pg.matchSpendingPassword.HintColor = pg.theme.Color.Danger
		pg.errLabel.Text = "Enter new wallet password again and it cannot be empty"
		return ""
	}

	if match != pass {
		pg.errLabel.Text = "Passwords does not match"
		return ""
	}

	if !pg.validateSeeds() {
		return ""
	}

	return pass
}

func (pg *createRestore) resetPasswords() {
	pg.spendingPassword.Editor.SetText("")
	pg.matchSpendingPassword.Editor.SetText("")
}

func (pg *createRestore) validateSeeds() bool {
	pg.seedPhrase = ""
	pg.errLabel.Text = ""

	for i, editor := range pg.seedEditors.editors {
		if editor.Edit.Editor.Text() == "" {
			pg.seedEditors.editors[i].Edit.HintColor = pg.theme.Color.Danger
			pg.errLabel.Text = "All seed fields are required"
			return false
		}

		pg.seedPhrase += editor.Edit.Editor.Text() + " "
	}

	if !dcrlibwallet.VerifySeed(pg.seedPhrase) {
		pg.errLabel.Text = "invalid seed phrase"
		return false
	}

	return true
}

func (pg *createRestore) resetSeeds() {
	for i := 0; i < len(pg.seedEditors.editors); i++ {
		pg.seedEditors.editors[i].Edit.Editor.SetText("")
	}
}

func (pg *createRestore) handle() {
	common := pg.common

	for pg.closePageBtn.Button.Clicked() {
		common.popWindowPage()
	}

	if pg.restoreWalletBtn.Button.Clicked() {
		pass := pg.validatePasswords()
		walletName := pg.validateWalletName()
		if !pg.validateSeeds() || pass == "" || walletName == "" {
			return
		}

		go func() {
			pg.restoringWallet = true
			_, err := pg.common.multiWallet.RestoreWallet(walletName, pg.seedPhrase, pass, dcrlibwallet.PassphraseTypePass)
			pg.restoringWallet = false
			if err != nil {
				pg.errLabel.Text = translateErr(err)
				return
			}

			pg.resetSeeds()
			pg.resetPasswords()

			// Go back to wallets page if there's more than one wallet
			// or launch main page.
			if pg.common.multiWallet.LoadedWalletsCount() > 1 {
				pg.common.popWindowPage()
			} else {
				pg.common.wallet.SetupListeners()
				pg.common.changeWindowPage(newMainPage(pg.common), false)
			}
		}()
	}

	if pg.matchSpendingPassword.Editor.Len() > 0 && pg.spendingPassword.Editor.Len() > 0 {
		pg.restoreWalletBtn.Background = pg.theme.Color.Primary
	}

	for pg.resetSeedFields.Button.Clicked() {
		pg.resetSeeds()
		pg.seedEditors.focusIndex = -1
	}

	// handle key events
	select {
	case evt := <-pg.keyEvent:
		if evt.Name == key.NameTab {
			if len(pg.suggestions) == 1 {
				focus := pg.seedEditors.focusIndex
				pg.seedEditors.editors[focus].Edit.Editor.SetText(pg.suggestions[0])
				pg.seedClicked = true
				pg.seedEditors.editors[focus].Edit.Editor.MoveCaret(len(pg.suggestions[0]), -1)
			}
		}
		if evt.Name == key.NameUpArrow && pg.openPopupIndex != -1 && evt.State == key.Press {
			pg.selected--
			if pg.selected < 0 {
				pg.selected = 0
			}
		}
		if evt.Name == key.NameDownArrow && pg.openPopupIndex != -1 && evt.State == key.Press {
			pg.selected++
			if pg.selected >= len(pg.suggestions) {
				pg.selected = len(pg.suggestions) - 1
			}
		}
		if (evt.Name == key.NameReturn || evt.Name == key.NameEnter) && pg.openPopupIndex != -1 && evt.State == key.Press && len(pg.suggestions) != 0 {
			if pg.seedEditors.focusIndex == -1 && len(pg.suggestions) == 1 {
				return
			}
			pg.seedMenu[pg.selected].button.Button.Click()
		}
	default:
	}

	computePasswordStrength(&pg.passwordStrength, common.theme, pg.spendingPassword.Editor)
	pg.editorSeedsEventsHandler()
	pg.onSuggestionSeedsClicked()
	pg.suggestionSeedEffect()
}

func (pg *createRestore) onClose() {}
