package ui

import (
	"fmt"
	"time"

	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/renderers"
	"github.com/planetdecred/godcr/ui/values"
)

const (
	PageProposalDetails = "ProposalDetails"
)

type proposalItemWidgets struct {
	widgets    []layout.Widget
	clickables map[string]*widget.Clickable
}

type proposalDetails struct {
	theme              *decredmaterial.Theme
	loadingDescription bool
	common             *pageCommon
	proposal           dcrlibwallet.Proposal
	descriptionCard    decredmaterial.Card
	proposalItems      map[string]proposalItemWidgets
	descriptionList    *layout.List
	redirectIcon       *widget.Image
	voteBar            decredmaterial.VoteBar
	rejectedIcon       *widget.Icon
	downloadIcon       *widget.Image
	timerIcon          *widget.Image
	successIcon        *widget.Icon
	vote               decredmaterial.Button
	backButton         decredmaterial.IconButton
}

func ProposalDetailsPage(common *pageCommon, proposal dcrlibwallet.Proposal) Page {
	pg := &proposalDetails{
		theme:              common.theme,
		loadingDescription: false,
		common:             common,
		proposal:           proposal,
		descriptionCard:    common.theme.Card(),
		descriptionList:    &layout.List{Axis: layout.Vertical},
		redirectIcon:       common.icons.redirectIcon,
		downloadIcon:       common.icons.downloadIcon,
		voteBar:            common.theme.VoteBar(common.icons.actionInfo, common.icons.imageBrightness1),
		proposalItems:      make(map[string]proposalItemWidgets),
		rejectedIcon:       common.icons.navigationCancel,
		successIcon:        common.icons.actionCheckCircle,
		timerIcon:          common.icons.timerIcon,
	}

	pg.downloadIcon.Scale = 1
	pg.backButton, _ = common.SubPageHeaderButtons()

	pg.vote = common.theme.Button(new(widget.Clickable), "Vote")
	pg.vote.TextSize = values.TextSize14
	pg.vote.Background = common.theme.Color.Primary
	pg.vote.Color = common.theme.Color.Surface
	pg.vote.CornerRadius = values.MarginPadding8
	pg.vote.Inset = layout.Inset{
		Top:    values.MarginPadding8,
		Bottom: values.MarginPadding8,
		Left:   values.MarginPadding12,
		Right:  values.MarginPadding12,
	}

	return pg
}

func (pg *proposalDetails) OnResume() {

}

func (pg *proposalDetails) Handle() {
	for token := range pg.proposalItems {
		for location, clickable := range pg.proposalItems[token].clickables {
			if clickable.Clicked() {
				goToURL(location)
			}
		}
	}
}

func (pg *proposalDetails) layoutProposalVoteBar(gtx C) D {
	proposal := pg.proposal

	yes := float32(proposal.YesVotes)
	no := float32(proposal.NoVotes)
	quorumPercent := float32(proposal.QuorumPercentage)
	passPercentage := float32(proposal.PassPercentage)
	eligibleTickets := float32(proposal.EligibleTickets)

	return pg.voteBar.SetParams(yes, no, eligibleTickets, quorumPercent, passPercentage).LayoutWithLegend(gtx)
}

func (pg *proposalDetails) layoutProposalVoteAction(gtx C) D {
	proposal := pg.proposal
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	txt := pg.theme.Label(values.TextSize14, fmt.Sprintf("%d eligible tickets", proposal.EligibleTickets))
	return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(pg.vote.Layout),
		layout.Rigid(txt.Layout),
	)
}

func (pg *proposalDetails) layoutInDiscussionState(gtx C) D {
	stateText1 := "Waiting for author to authorize voting"
	stateText2 := "Waiting for admin to trigger the start of voting"

	proposal := pg.proposal

	c := func(gtx layout.Context, val int32, info string) layout.Dimensions {
		return layout.Flex{}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				if proposal.VoteStatus == val || proposal.VoteStatus < val {
					c := pg.theme.Card()
					c.Color = pg.theme.Color.Primary

					r := float32(9.5)
					c.Radius = decredmaterial.CornerRadius{NE: r, NW: r, SE: r, SW: r}
					lbl := pg.theme.Body1(fmt.Sprint(val))
					lbl.Color = pg.theme.Color.Surface
					if proposal.VoteStatus < val {
						c.Color = pg.theme.Color.LightGray
						lbl.Color = pg.theme.Color.Hint
					}
					return c.Layout(gtx, func(gtx C) D {
						m := values.MarginPadding6
						return layout.Inset{Left: m, Right: m}.Layout(gtx, lbl.Layout)
					})
				}
				icon := pg.successIcon
				icon.Color = pg.theme.Color.Primary
				return layout.Inset{
					Left:   values.MarginPaddingMinus2,
					Right:  values.MarginPaddingMinus2,
					Bottom: values.MarginPaddingMinus2,
				}.Layout(gtx, func(gtx C) D {
					return icon.Layout(gtx, values.MarginPadding24)
				})
			}),
			layout.Rigid(func(gtx C) D {
				col := pg.theme.Color.Primary
				txt := info + "..."
				if proposal.VoteStatus != val {
					txt = info
					col = pg.theme.Color.Hint
					if proposal.VoteStatus > 1 {
						col = pg.theme.Color.DeepBlue
					}
				}
				lbl := pg.theme.Body1(txt)
				lbl.Color = col
				return layout.Inset{Left: values.MarginPadding16}.Layout(gtx, lbl.Layout)
			}),
		)
	}

	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return c(gtx, 1, stateText1)
		}),
		layout.Rigid(func(gtx C) D {
			height, width := gtx.Px(values.MarginPadding26), gtx.Px(values.MarginPadding4)
			line := pg.theme.Line(height, width)
			if proposal.VoteStatus > 1 {
				line.Color = pg.theme.Color.Primary
			} else {
				line.Color = pg.theme.Color.Gray1
			}
			return layout.Inset{Left: values.MarginPadding8}.Layout(gtx, line.Layout)
		}),
		layout.Rigid(func(gtx C) D {
			return c(gtx, 2, stateText2)
		}),
	)
}

func (pg *proposalDetails) layoutNormalTitle(gtx C) D {
	var label decredmaterial.Label
	var icon *widget.Icon
	proposal := pg.proposal
	switch proposal.Category {
	case dcrlibwallet.ProposalCategoryApproved:
		label = pg.theme.Body2("Approved")
		icon = pg.successIcon
		icon.Color = pg.theme.Color.Success
	case dcrlibwallet.ProposalCategoryRejected:
		label = pg.theme.Body2("Rejected")
		icon = pg.rejectedIcon
		icon.Color = pg.theme.Color.Danger
	case dcrlibwallet.ProposalCategoryAbandoned:
		label = pg.theme.Body2("Abandoned")
	case dcrlibwallet.ProposalCategoryActive:
		label = pg.theme.Body2("Voting in progress...")
	}
	timeagoLabel := pg.theme.Body2(timeAgo(proposal.Timestamp))

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					if icon == nil {
						return D{}
					}
					return icon.Layout(gtx, unit.Dp(20))
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Left: values.MarginPadding5}.Layout(gtx, label.Layout)
				}),
				layout.Flexed(1, func(gtx C) D {
					return layout.E.Layout(gtx, func(gtx C) D {
						return layout.Flex{}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								if proposal.Category == dcrlibwallet.ProposalCategoryActive {
									pg.timerIcon.Scale = 1
									return layout.Inset{
										Right: values.MarginPadding4,
										Top:   values.MarginPadding3,
									}.Layout(gtx, pg.timerIcon.Layout)
								}
								return D{}
							}),
							layout.Rigid(timeagoLabel.Layout),
						)
					})
				}),
			)
		}),
		layout.Rigid(pg.lineSeparator(layout.Inset{Top: values.MarginPadding10, Bottom: values.MarginPadding10})),
		layout.Rigid(pg.layoutProposalVoteBar),
		layout.Rigid(func(gtx C) D {
			if proposal.Category != dcrlibwallet.ProposalCategoryActive {
				return D{}
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(pg.lineSeparator(layout.Inset{Top: values.MarginPadding10, Bottom: values.MarginPadding10})),
				layout.Rigid(pg.layoutProposalVoteAction),
			)
		}),
	)
}

func (pg *proposalDetails) layoutTitle(gtx C) D {
	proposal := pg.proposal

	return pg.descriptionCard.Layout(gtx, func(gtx C) D {
		return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
			if proposal.Category == dcrlibwallet.ProposalCategoryPre {
				return pg.layoutInDiscussionState(gtx)
			}
			return pg.layoutNormalTitle(gtx)
		})
	})
}

func (pg *proposalDetails) layoutDescription(gtx C) D {
	grayCol := pg.theme.Color.Gray
	proposal := pg.proposal

	dotLabel := pg.theme.H4(" . ")
	dotLabel.Color = grayCol

	userLabel := pg.theme.Body2(proposal.Username)
	userLabel.Color = grayCol

	versionLabel := pg.theme.Body2("Version " + proposal.Version)
	versionLabel.Color = grayCol

	publishedLabel := pg.theme.Body2("Published " + timeAgo(proposal.PublishedAt))
	publishedLabel.Color = grayCol

	updatedLabel := pg.theme.Body2("Updated " + timeAgo(proposal.Timestamp))
	updatedLabel.Color = grayCol

	w := []layout.Widget{
		func(gtx C) D {
			lbl := pg.theme.H5(proposal.Name)
			lbl.Font.Weight = text.Bold
			return lbl.Layout(gtx)
		},
		pg.lineSeparator(layout.Inset{Top: values.MarginPadding16, Bottom: values.MarginPadding16}),
		func(gtx C) D {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(userLabel.Layout),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: values.MarginPaddingMinus22}.Layout(gtx, dotLabel.Layout)
				}),
				layout.Rigid(publishedLabel.Layout),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: values.MarginPaddingMinus22}.Layout(gtx, dotLabel.Layout)
				}),
				layout.Rigid(versionLabel.Layout),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: values.MarginPaddingMinus22}.Layout(gtx, dotLabel.Layout)
				}),
				layout.Rigid(updatedLabel.Layout),
			)
		},
		pg.lineSeparator(layout.Inset{Top: values.MarginPadding16, Bottom: values.MarginPadding16}),
	}

	_, ok := pg.proposalItems[proposal.Token]
	if ok {
		w = append(w, pg.proposalItems[proposal.Token].widgets...)
	} else {
		th := material.NewTheme(gofont.Collection())
		loading := func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, layout.Flexed(1, func(gtx C) D {
				return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
					return layout.Center.Layout(gtx, material.Loader(th).Layout)
				})
			}))
		}

		w = append(w, loading)
	}

	return pg.descriptionCard.Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return layout.UniformInset(values.MarginPadding16).Layout(gtx, func(gtx C) D {
			return pg.descriptionList.Layout(gtx, len(w), func(gtx C, i int) D {
				return layout.UniformInset(unit.Dp(0)).Layout(gtx, w[i])
			})
		})
	})
}

func (pg *proposalDetails) layoutRedirect(text string, icon *widget.Image, btn *widget.Clickable) layout.Widget {
	return func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(pg.lineSeparator(layout.Inset{Top: values.MarginPadding12, Bottom: values.MarginPadding12})),
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						txt := pg.theme.Body1(text)
						txt.Color = pg.theme.Color.DeepBlue
						return txt.Layout(gtx)
					}),
					layout.Rigid(func(gtx C) D {
						return decredmaterial.Clickable(gtx, btn, func(gtx C) D {
							return layout.Inset{}.Layout(gtx, func(gtx C) D {
								return layout.E.Layout(gtx, icon.Layout)
							})
						})
					}),
				)
			}),
		)
	}
}

func (pg *proposalDetails) lineSeparator(inset layout.Inset) layout.Widget {
	return func(gtx C) D {
		return inset.Layout(gtx, pg.theme.Separator().Layout)
	}
}

func (pg *proposalDetails) Layout(gtx C) D {
	common := pg.common

	proposal := pg.proposal
	_, ok := pg.proposalItems[proposal.Token]
	if !ok && !pg.loadingDescription {
		pg.loadingDescription = true
		go func() {
			var proposalDescription string
			if proposal.IndexFile != "" && proposal.IndexFileVersion == proposal.Version {
				proposalDescription = proposal.IndexFile
			} else {
				var err error
				proposalDescription, err = common.wallet.FetchProposalDescription(proposal.Token)
				if err != nil {
					log.Infof("Error loading proposal description: %v", err)
					time.Sleep(7 * time.Second)
					pg.loadingDescription = false
					return
				}
			}

			r := renderers.RenderMarkdown(gtx, pg.theme, proposalDescription)
			proposalWidgets, proposalClickables := r.Layout()
			pg.proposalItems[proposal.Token] = proposalItemWidgets{
				widgets:    proposalWidgets,
				clickables: proposalClickables,
			}
			pg.loadingDescription = false
		}()
	}

	body := func(gtx C) D {
		page := SubPage{
			title:      truncateString(proposal.Name, 40),
			backButton: pg.backButton,
			back: func() {
				common.changePage(PageProposals)
				pg.descriptionList.Position.First, pg.descriptionList.Position.Offset = 0, 0
			},
			body: func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Bottom: values.MarginPadding10}.Layout(gtx, pg.layoutTitle)
					}),
					layout.Rigid(pg.layoutDescription),
				)
			},
			extra: func(gtx C) D {
				return layout.Inset{}.Layout(gtx, func(gtx C) D {
					pg.redirectIcon.Scale = 1
					return layout.E.Layout(gtx, pg.redirectIcon.Layout)
				})
			},
		}
		return common.SubPageLayout(gtx, page)
	}
	return common.UniformPadding(gtx, body)
}

func (pg *proposalDetails) OnClose() {}
