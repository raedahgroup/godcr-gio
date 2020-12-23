package ui

import (
	//"strconv"
	"encoding/base64"
	"fmt"
	"strings"
	//"regexp"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const PageProposalDetails = "proposaldetails"

type ProposalPage struct {
	theme    *decredmaterial.Theme
	proposal **dcrlibwallet.Proposal
	line     *decredmaterial.Line
	pageLinks map[string]layout.Dimensions

	backButton decredmaterial.IconButton
	legendIcon *widget.Icon
	container  *layout.List
}

func (win *Window) ProposalPage(common pageCommon) layout.Widget {
	pg := ProposalPage{
		theme:      common.theme,
		proposal:   &win.selectedProposal,
		backButton: common.theme.PlainIconButton(new(widget.Clickable), common.icons.navigationArrowBack),
		legendIcon: common.icons.imageBrightness1,
		line:       common.theme.Line(),
		container:  &layout.List{Axis: layout.Vertical},
	}
	pg.backButton.Color = common.theme.Color.Hint
	pg.backButton.Size = values.MarginPadding30
	pg.line.Color = pg.theme.Color.Hint

	return func(gtx C) D {
		pg.Handle(common)
		return pg.Layout(gtx, common)
	}
}

func (pg *ProposalPage) Handle(common pageCommon) {
	for pg.backButton.Button.Clicked() {
		*common.page = PageProposals
	}
}

func (pg *ProposalPage) Layout(gtx layout.Context, c pageCommon) layout.Dimensions {
	return c.Layout(gtx, func(gtx C) D {
		return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.W.Layout(gtx, func(gtx C) D {
						return pg.backButton.Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx C) D {
					return pg.layoutProposalDescription(gtx)
				}),
			)
		})
	})
}

func (pg *ProposalPage) layoutProposalDescription(gtx layout.Context) layout.Dimensions {
	proposal := *pg.proposal

	w := []layout.Widget{
		func(gtx C) D {
			return pg.layoutProposalHeader(gtx, false)
		},
		func(gtx C) D {
			return pg.layoutProposalDetailsSubHeader(gtx)
		},
		func(gtx C) D {
			category := proposal.Category
			if category == dcrlibwallet.ProposalCategoryApproved || category == dcrlibwallet.ProposalCategoryActive || category == dcrlibwallet.ProposalCategoryRejected {
				return layout.Inset{
					Top:    unit.Dp(8),
					Bottom: unit.Dp(8),
				}.Layout(gtx, func(gtx C) D {
					yes, no := calculateVotes(proposal.VoteSummary.OptionsResult)
					return pg.theme.VoteBar(yes, no).LayoutWithLegend(gtx, pg.legendIcon)
				})
			}
			return layout.Dimensions{}
		},
		func(gtx C) D {
			return layout.Inset{
				Top:    unit.Dp(12),
				Bottom: unit.Dp(12),
			}.Layout(gtx, func(gtx C) D {
				pg.line.Width = gtx.Constraints.Max.X
				return pg.line.Layout(gtx)
			})
		},
	}

	wordRows := pg.getProposalDescriptionTextParts()
	for i := range wordRows {
		index := i 
		w = append(w, func(gtx C) D {
			return decredmaterial.GridWrap{
				Axis: layout.Horizontal,
				Alignment: layout.Baseline,
			}.Layout(gtx, len(wordRows[index]), func(gtx layout.Context, k int) layout.Dimensions {
				return pg.layoutWord(gtx, wordRows[index][k], index)
			})
		})
	}

	return pg.container.Layout(gtx, len(w), func(gtx C, i int) D {
		return layout.UniformInset(unit.Dp(0)).Layout(gtx, w[i])
	})
}

func (pg *ProposalPage) checkIfLink(word string) (bool, string, string) {
	if strings.HasPrefix(word, "[") && strings.Contains(word, "]") {
		link :=  getStringInBetween(word, "(", ")")
		title := getStringInBetween(word, "[", "]")
		
		return true, title, link 
	}
	return false, "", ""
}


func (pg *ProposalPage) layoutWord(gtx layout.Context, word string, index int) layout.Dimensions {
	if isLink, _, _ := pg.checkIfLink(word); isLink {
		return pg.theme.Body2(" [] ").Layout(gtx)
	}
	
	lbl := pg.theme.Body2
	if index == 0 || index % 2 == 0 {
		lbl = pg.theme.H6
	}
	
	return lbl(word + " ").Layout(gtx)
}

func (pg *ProposalPage) getProposalDescriptionTextParts() [][]string {
	proposal := *pg.proposal

	var description string
	for i := range proposal.Files {
		if proposal.Files[i].Name == "index.md" {
			descBytes, _ := base64.StdEncoding.DecodeString(proposal.Files[i].Payload)
			description = string(descBytes)
			break
		}
	}

	rows := strings.FieldsFunc(description, func(c rune) bool {
		return c == '*' || c == '#'
	})
	
	var rowWords [][]string
	for i := range rows {
		if i == 0 { // row at index 0 is proposal title. Omit it
			continue
		}
		
		words := pg.splitRowWords(rows[i])			
		rowWords = append(rowWords, words)
	}
	return rowWords
}

func (pg *ProposalPage) splitRowWords(row string) []string {
	isInsideBrackets := false 
	return strings.FieldsFunc(row, func(r rune) bool {
		if r == '[' {
			isInsideBrackets = !isInsideBrackets
		}
		if r == ']' || r == ')' {
			isInsideBrackets = false
		}
		return !isInsideBrackets && r == ' '
	})
}

func (pg *ProposalPage) layoutProposalHeader(gtx layout.Context, truncateTitle bool) layout.Dimensions {
	proposal := *pg.proposal

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(0.55, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return getTitleLabel(pg.theme, proposal.Name).Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return getSubtitleLabel(pg.theme, proposal.CensorshipRecord.Token).Layout(gtx)
				}),
			)
		}),
		layout.Flexed(0.45, func(gtx C) D {
			if proposal.Category == dcrlibwallet.ProposalCategoryPre || proposal.Category == dcrlibwallet.ProposalCategoryAbandoned {
				return layout.E.Layout(gtx, func(gtx C) D {
					return getSubtitleLabel(pg.theme, fmt.Sprintf("Last updated %s", timeAgo(proposal.Timestamp))).Layout(gtx)
				})
			}

			return layout.Dimensions{}
		}),
	)
}

func (pg *ProposalPage) layoutProposalDetailsSubHeader(gtx layout.Context) layout.Dimensions {
	proposal := *pg.proposal

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return pg.layoutProposalDetailsSubHeaderRow(gtx, "Created by:", proposal.Username)
		}),
		layout.Rigid(func(gtx C) D {
			return pg.layoutProposalDetailsSubHeaderRow(gtx, "Version:", proposal.Version)
		}),
		layout.Rigid(func(gtx C) D {
			return pg.layoutProposalDetailsSubHeaderRow(gtx, "Last updated:", timeAgo(proposal.Timestamp))
		}),
	)
}

func (pg *ProposalPage) layoutProposalDetailsSubHeaderRow(gtx layout.Context, leftText, rightText string) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(0.03, func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				return getSubtitleLabel(pg.theme, leftText).Layout(gtx)
			})
		}),
		layout.Flexed(0.2, func(gtx C) D {
			return layout.Inset{
				Left: unit.Dp(4),
			}.Layout(gtx, func(gtx C) D {
				return getTitleLabel(pg.theme, rightText).Layout(gtx)
			})
		}),
	)
}

// GetStringInBetween Returns empty string if no start string found
func getStringInBetween(str string, start string, end string) (result string) {
    s := strings.Index(str, start)
    if s == -1 {
        return
    }
    s += len(start)
    e := strings.Index(str[s:], end)
    if e == -1 {
        return
    }
    return str[s:e+1]
}

