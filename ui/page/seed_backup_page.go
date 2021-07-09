package page

import (
	"fmt"

	"image/color"
	"math/rand"
	"strings"
	"time"

	"gioui.org/text"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"
)

const (
	SeedBackupPageID = "SeedBackup"
	infoView         = iota
	seedView
	verifyView
	successView
)

type (
	seedGroup struct {
		selected int
		buttons  []decredmaterial.Button
	}

	viewText struct {
		title       string
		action      string
		steps       string
		instruction string
	}
)

type BackupPage struct {
	theme *decredmaterial.Theme

	*load.Load
	wal  *wallet.Wallet
	info *wallet.MultiWalletInfo

	backButton     decredmaterial.IconButton
	title          decredmaterial.Label
	steps          decredmaterial.Label
	instruction    decredmaterial.Label
	successMessage decredmaterial.Label
	successInfo    decredmaterial.Label
	action         decredmaterial.Button
	checkBoxes     []decredmaterial.CheckBoxStyle
	checkIcon      *widget.Icon

	infoList            *layout.List
	viewList            *layout.List
	seedPhraseListLeft  *layout.List
	seedPhraseListRight *layout.List
	verifyList          *layout.List

	suggestions         []seedGroup
	passwordModal       *decredmaterial.Password
	isPasswordModalOpen bool
	selectedWallet      *int

	seedPhrase     []string
	selectedSeeds  []string
	allSuggestions []string
	active         int
	privpass       []byte
}

func NewBackupPage(l *load.Load) *BackupPage {
	b := &BackupPage{
		theme: l.Theme,
		wal:   l.WL.Wallet,
		info:  l.WL.Info,

		action:         l.Theme.Button(new(widget.Clickable), "View seed phrase"),
		backButton:     l.Theme.PlainIconButton(new(widget.Clickable), l.Icons.NavigationArrowBack),
		title:          l.Theme.H5("Keep in mind"),
		steps:          l.Theme.Body1("Step 1/2"),
		instruction:    l.Theme.H6("Write down all 33 words in the correct order"),
		successMessage: l.Theme.H4("Your seed phrase backup is verified"),
		successInfo:    l.Theme.Body2("Be sure to store your seed phrase backup in a secure location."),
		checkIcon:      l.Icons.ActionCheckCircle,

		active:         infoView,
		selectedSeeds:  make([]string, 0, 33),
		selectedWallet: l.SelectedWallet,
		passwordModal:  l.Theme.Password(),
	}

	b.checkIcon.Color = l.Theme.Color.Success
	b.steps.Color = l.Theme.Color.Hint
	b.successMessage.Alignment = text.Middle
	b.successInfo.Alignment = text.Middle
	b.successInfo.Color = b.theme.Color.Hint

	b.backButton.Color = l.Theme.Color.Hint
	b.backButton.Size = values.MarginPadding30
	b.backButton.Inset = layout.UniformInset(values.MarginPadding0)

	b.action.Background = l.Theme.Color.Hint

	b.checkBoxes = []decredmaterial.CheckBoxStyle{
		l.Theme.CheckBox(new(widget.Bool), "The 33-word seed phrase is EXTREMELY IMPORTANT."),
		l.Theme.CheckBox(new(widget.Bool), "Seed phrase is the only way to restore your wallet."),
		l.Theme.CheckBox(new(widget.Bool), "It is recommended to store your seed phrase in a physical format (e.g. write down on a paper)."),
		l.Theme.CheckBox(new(widget.Bool), "It is highly discouraged to store your seed phrase in any digital format (e.g. screenshot)."),
		l.Theme.CheckBox(new(widget.Bool), "Anyone with your seed phrase can steal your funds. DO NOT show it to anyone."),
	}

	b.instruction.Alignment = text.Middle
	b.allSuggestions = dcrlibwallet.PGPWordList()

	for _, cb := range b.checkBoxes {
		cb.IconColor = l.Theme.Color.Success
		cb.Color = l.Theme.Color.Success
	}

	for i := 0; i < 33; i++ {
		var bg []decredmaterial.Button

		for j := 0; j < 3; j++ {
			bg = append(bg, l.Theme.Button(new(widget.Clickable), ""))
		}
		b.suggestions = append(b.suggestions, seedGroup{selected: -1, buttons: bg})
		b.selectedSeeds = append(b.selectedSeeds, "-")
	}

	b.infoList = &layout.List{Axis: layout.Vertical}
	b.viewList = &layout.List{Axis: layout.Horizontal}
	b.seedPhraseListLeft = &layout.List{Axis: layout.Vertical}
	b.seedPhraseListRight = &layout.List{Axis: layout.Vertical}
	b.verifyList = &layout.List{Axis: layout.Vertical}

	return b
}

func (pg *BackupPage) OnResume() {

}

func (pg *BackupPage) activeButton() {
	pg.action.Background = pg.theme.Color.Primary
	pg.action.Color = pg.theme.Color.InvText
}

func (pg *BackupPage) clearButton() {
	pg.action.Background = color.NRGBA{}
	pg.action.Color = pg.theme.Color.Primary
}

func (pg *BackupPage) Layout(gtx layout.Context) layout.Dimensions {
	dims := pg.theme.Surface(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				pg.action.Background = pg.theme.Color.Hint
				pg.action.Color = pg.theme.Color.InvText
				switch pg.active {
				case infoView:
					if pg.verifyCheckBoxes() {
						pg.activeButton()
					}
					return pg.infoView(gtx)
				case seedView:
					pg.activeButton()
					return pg.seedView(gtx)
				case verifyView:
					if checkSlice(pg.selectedSeeds) {
						pg.activeButton()
					}
					return pg.verifyView(gtx)
				case successView:
					pg.activeButton()
					return pg.successView(gtx)
				default:
					if pg.verifyCheckBoxes() {
						pg.activeButton()
					}
					return pg.infoView(gtx)
				}
			}),
		)
	})

	if pg.isPasswordModalOpen {
		return _modal(gtx, dims, pg.passwordModal.Layout(gtx, pg.confirm, pg.cancel))
	}
	return dims
}

func (pg *BackupPage) pageTitle(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Bottom: values.MarginPadding5, Top: values.MarginPadding20}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return pg.backButton.Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Left: values.MarginPadding10}.Layout(gtx, pg.title.Layout)
					}),
					layout.Rigid(func(gtx C) D {
						if pg.active != infoView {
							return layout.Inset{Left: values.MarginPadding10}.Layout(gtx, pg.steps.Layout)
						}
						return layout.Dimensions{}
					}),
					layout.Rigid(func(gtx C) D {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						if pg.active != infoView {
							return layout.Inset{
								Right: values.MarginPadding30,
								Top:   values.MarginPadding20,
							}.Layout(gtx, pg.instruction.Layout)
						}
						return layout.Dimensions{}
					}),
				)
			}),
		)
	})
}

func (pg *BackupPage) viewTemplate(gtx layout.Context, content layout.Widget) layout.Dimensions {
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	return layout.Inset{Left: values.MarginPadding10, Right: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
		return layout.Stack{}.Layout(gtx,
			layout.Stacked(func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Flexed(1, func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								if pg.active != successView {
									return pg.pageTitle(gtx)
								}
								return layout.Dimensions{}
							}),
							layout.Rigid(content),
						)
					}),
					layout.Rigid(func(gtx C) D {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return layout.Inset{Bottom: values.MarginPadding10}.Layout(gtx, pg.action.Layout)
					}),
				)
			}),
		)
	})
}

func (pg *BackupPage) infoView(gtx layout.Context) layout.Dimensions {
	return pg.viewTemplate(gtx, func(gtx C) D {
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
		return pg.centralize(gtx, func(gtx C) D {
			return pg.infoList.Layout(gtx, len(pg.checkBoxes), func(gtx C, i int) D {
				return layout.Inset{Bottom: values.MarginPadding20}.Layout(gtx, pg.checkBoxes[i].Layout)
			})
		})
	})
}

func (pg *BackupPage) seedView(gtx layout.Context) layout.Dimensions {
	return pg.viewTemplate(gtx, func(gtx C) D {
		return pg.centralize(gtx, func(gtx C) D {
			return pg.viewList.Layout(gtx, 1, func(gtx C, i int) D {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						gtx.Constraints.Max.X = gtx.Constraints.Max.X / 2
						return pg.seedPhraseListLeft.Layout(gtx, len(pg.seedPhrase), func(gtx C, i int) D {
							if i < 17 {
								return pg.seedText(gtx, i)
							}
							return layout.Dimensions{}
						})
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Left: values.MarginPadding30}.Layout(gtx, func(gtx C) D {
							return pg.seedPhraseListRight.Layout(gtx, len(pg.seedPhrase), func(gtx C, i int) D {
								if i > 16 {
									return pg.seedText(gtx, i)
								}
								return layout.Dimensions{}
							})
						})
					}),
				)
			})
		})
	})
}

func (pg *BackupPage) verifyView(gtx layout.Context) layout.Dimensions {
	return pg.viewTemplate(gtx, func(gtx C) D {
		return pg.verifyList.Layout(gtx, len(pg.suggestions), func(gtx C, i int) D {
			s := pg.suggestions[i]
			return layout.Center.Layout(gtx, func(gtx C) D {
				return layout.Inset{Bottom: values.MarginPadding30}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return layout.Inset{
								Left:   values.MarginPadding15,
								Bottom: values.MarginPadding15,
							}.Layout(gtx, pg.theme.H6(fmt.Sprintf("%d. %s", i+1, pg.selectedSeeds[i])).Layout)
						}),
						layout.Rigid(func(gtx C) D {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Flexed(0.3, func(gtx C) D {
									return pg.suggestionButtonGroup(gtx, s, 0)
								}),
								layout.Flexed(0.3, func(gtx C) D {
									return pg.suggestionButtonGroup(gtx, s, 1)
								}),
								layout.Flexed(0.3, func(gtx C) D {
									return pg.suggestionButtonGroup(gtx, s, 2)
								}),
							)
						}),
					)
				})
			})
		})
	})
}

func (pg *BackupPage) successView(gtx layout.Context) layout.Dimensions {
	return pg.viewTemplate(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
				return layout.Center.Layout(gtx, func(gtx C) D {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							gtx.Constraints.Min.X = gtx.Constraints.Max.X
							return layout.Inset{
								Bottom: values.MarginPadding50,
								Right:  values.MarginPadding50,
							}.Layout(gtx, func(gtx C) D {
								return layout.Center.Layout(gtx, func(gtx C) D {
									return layout.UniformInset(values.MarginPadding20).Layout(gtx, func(gtx C) D {
										pg.checkIcon.Color = pg.theme.Color.Success
										return pg.checkIcon.Layout(gtx, unit.Px(float32(150)))
									})
								})
							})
						}),
						layout.Rigid(func(gtx C) D {
							gtx.Constraints.Min.X = gtx.Constraints.Max.X
							return pg.successMessage.Layout(gtx)
						}),
						layout.Rigid(func(gtx C) D {
							gtx.Constraints.Min.X = gtx.Constraints.Max.X
							return pg.successInfo.Layout(gtx)
						}),
					)
				})
			}),
		)
	})
}

func (pg *BackupPage) seedText(gtx layout.Context, index int) layout.Dimensions {
	return layout.Inset{Bottom: values.MarginPadding10, Left: values.MarginPadding20}.Layout(gtx,
		func(gtx C) D {
			seedLabel := pg.theme.H6(fmt.Sprintf("%d.  %s", index+1, pg.seedPhrase[index]))
			seedLabel.Alignment = text.Middle
			return seedLabel.Layout(gtx)
		},
	)
}

func (pg *BackupPage) centralize(gtx layout.Context, content layout.Widget) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			return layout.Center.Layout(gtx, content)
		}),
	)
}

func (pg *BackupPage) suggestionButtonGroup(gtx layout.Context, sg seedGroup, buttonIndex int) layout.Dimensions {
	button := sg.buttons[buttonIndex]
	button.Background = pg.theme.Color.Hint
	button.TextSize = values.TextSize18
	if sg.selected == buttonIndex {
		button.Background = pg.theme.Color.Primary
	}
	return layout.Inset{Right: values.MarginPadding15, Left: values.MarginPadding15}.Layout(gtx, button.Layout)
}

func (pg *BackupPage) verifyCheckBoxes() bool {
	for _, cb := range pg.checkBoxes {
		if !cb.CheckBox.Value {
			return false
		}
	}
	return true
}

func (pg *BackupPage) randomSeeds() []string {
	var randomSeeds []string

	for i := 0; i < 3; i++ {
		random := rand.Intn(len(pg.allSuggestions))
		randomSeeds = append(randomSeeds, pg.allSuggestions[random])
	}
	return randomSeeds
}

func (pg *BackupPage) populateSuggestionSeeds() {
	rand.Seed(time.Now().Unix())

	for k := range pg.seedPhrase {
		seeds := pg.randomSeeds()
		s := pg.suggestions[k]
		for i := range s.buttons {
			s.buttons[i].Text = seeds[i]
		}
		s.buttons[rand.Intn(len(seeds))].Text = pg.seedPhrase[k]
	}
}

func viewTexts(active int) viewText {
	switch active {
	case infoView:
		return viewText{
			title:  "Keep in mind",
			action: "View seed phrase",
		}
	case seedView:
		return viewText{
			title:       "Write down seed phrase",
			action:      "I have written down all 33 words",
			steps:       fmt.Sprintf("Steps %d/2", seedView-1),
			instruction: "Write down all 33 words in the correct order",
		}
	case verifyView:
		return viewText{
			title:       "Verify seed phrase",
			action:      "Verify",
			steps:       fmt.Sprintf("Steps %d/2", verifyView-1),
			instruction: "Select the correct words to verify",
		}
	case successView:
		return viewText{
			action: "Back to Wallets",
		}
	}
	return viewText{}
}

func (pg *BackupPage) updateViewTexts() {
	t := viewTexts(pg.active)
	pg.title.Text = t.title
	pg.action.Text = t.action
	pg.steps.Text = t.steps
	pg.instruction.Text = t.instruction
}

func checkSlice(s []string) bool {
	for _, v := range s {
		if v == "-" {
			return false
		}
	}
	return true
}

func (pg *BackupPage) resetPage() {
	pg.ChangePage(WalletPageID)
	pg.active = infoView
	pg.seedPhrase = []string{}
	pg.selectedSeeds = make([]string, 33)
	pg.privpass = nil
	for _, cb := range pg.checkBoxes {
		cb.CheckBox.Value = false
	}
	for i := range pg.suggestions {
		pg.suggestions[i].selected = -1
	}
	for i := range pg.selectedSeeds {
		pg.selectedSeeds[i] = "-"
	}
	pg.updateViewTexts()
}

func (pg *BackupPage) confirm(password []byte) {
	pg.privpass = password
	s, err := pg.wal.GetWalletSeedPhrase(pg.info.Wallets[*pg.selectedWallet].ID, password)
	if err != nil {
		pg.passwordModal.WithError(err.Error())
		return
	}
	pg.isPasswordModalOpen = false
	pg.seedPhrase = strings.Split(s, " ")
	pg.populateSuggestionSeeds()
	pg.active++
}

func (pg *BackupPage) cancel() {
	pg.isPasswordModalOpen = false
}

func (pg *BackupPage) Handle() {
	if pg.backButton.Button.Clicked() {
		pg.resetPage()
	}

	if pg.action.Button.Clicked() && pg.verifyCheckBoxes() {
		if len(pg.seedPhrase) == 0 {
			pg.isPasswordModalOpen = true
			return
		}
		switch pg.active {
		case verifyView:
			if !checkSlice(pg.selectedSeeds) {
				return
			}
			errMessage := "Failed to verify. Please go through every word and try again."
			s := strings.Join(pg.selectedSeeds, " ")
			if !dcrlibwallet.VerifySeed(s) {
				pg.CreateToast(errMessage, false)
				return
			}

			err := pg.wal.VerifyWalletSeedPhrase(pg.info.Wallets[*pg.selectedWallet].ID, s, pg.privpass)
			if err != nil {
				pg.CreateToast(errMessage, false)
				return
			}
			pg.info.Wallets[*pg.selectedWallet].Seed = nil
			pg.active++
		case successView:
			pg.resetPage()
		default:
			pg.active++
		}
		pg.updateViewTexts()
	}

	for i := range pg.suggestions {
		suggestion := pg.suggestions[i]
		for s := range suggestion.buttons {
			if suggestion.buttons[s].Button.Clicked() {
				pg.suggestions[i].selected = s
				pg.selectedSeeds[i] = suggestion.buttons[s].Text
			}
		}
	}
}

func (pg *BackupPage) OnClose() {}
