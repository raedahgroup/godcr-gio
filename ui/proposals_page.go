package ui

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/ararog/timeago"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"
)

const PageProposals = "Proposals"

type proposalItem struct {
	btn               *widget.Clickable
	proposal          dcrlibwallet.Proposal
	voteBar           decredmaterial.VoteBar
	infoIcon          *widget.Icon
	stateInfoTooltip  *decredmaterial.Tooltip
	stateTooltipLabel decredmaterial.Label
}

type tab struct {
	title     string
	btn       *widget.Clickable
	category  int32
	proposals []proposalItem
	container *layout.List
}

type tabs struct {
	tabs     []tab
	selected int
}

type proposalsPage struct {
	theme            *decredmaterial.Theme
	common           pageCommon
	wallet           *wallet.Wallet
	proposals        []dcrlibwallet.Proposal
	syncedProposal   chan *wallet.Proposal
	proposalsList    *layout.List
	tabs             tabs
	tabCard          decredmaterial.Card
	itemCard         decredmaterial.Card
	syncCard         decredmaterial.Card
	updatedLabel     decredmaterial.Label
	legendIcon       *widget.Icon
	infoIcon         *widget.Icon
	updatedIcon      *widget.Icon
	syncButton       *widget.Clickable
	startSyncIcon    *widget.Image
	timerIcon        *widget.Image
	isSynced         bool
	proposalsItemSet bool
}

var (
	proposalCategoryTitles = []string{"In discussion", "Voting", "Approved", "Rejected", "Abandoned"}
	proposalCategories     = []int32{
		dcrlibwallet.ProposalCategoryPre,
		dcrlibwallet.ProposalCategoryActive,
		dcrlibwallet.ProposalCategoryApproved,
		dcrlibwallet.ProposalCategoryRejected,
		dcrlibwallet.ProposalCategoryAbandoned,
	}
)

func ProposalsPage(common pageCommon) Page {
	pg := &proposalsPage{
		common:        common,
		theme:         common.theme,
		wallet:        common.wallet,
		proposalsList: &layout.List{},
		tabCard:       common.theme.Card(),
		itemCard:      common.theme.Card(),
		syncCard:      common.theme.Card(),
		legendIcon:    common.icons.imageBrightness1,
		infoIcon:      common.icons.actionInfo,
		// syncedProposal: win.proposal, TODO
		updatedIcon:   common.icons.navigationCheck,
		updatedLabel:  common.theme.Body2("Updated"),
		syncButton:    new(widget.Clickable),
		startSyncIcon: common.icons.restore,
		timerIcon:     common.icons.timerIcon,
	}
	pg.infoIcon.Color = common.theme.Color.Gray
	pg.legendIcon.Color = common.theme.Color.InactiveGray

	pg.updatedIcon.Color = common.theme.Color.Success
	pg.updatedLabel.Color = common.theme.Color.Success

	pg.tabCard.Radius = decredmaterial.CornerRadius{NE: 0, NW: 0, SE: 0, SW: 0}
	pg.syncCard.Radius = decredmaterial.CornerRadius{NE: 0, NW: 0, SE: 0, SW: 0}

	for i := range proposalCategoryTitles {
		pg.tabs.tabs = append(pg.tabs.tabs,
			tab{
				title:     proposalCategoryTitles[i],
				btn:       new(widget.Clickable),
				category:  proposalCategories[i],
				container: &layout.List{Axis: layout.Vertical},
			},
		)
	}

	proposals, err := common.multiWallet.Politeia.GetProposalsRaw(dcrlibwallet.ProposalCategoryAll, 0, 0, true)
	if err != nil {
		log.Error("error fetching proposal:", err)
	} else {
		pg.proposals = proposals
	}

	return pg
}

func (pg *proposalsPage) pageID() string {
	return PageProposals
}

func (pg *proposalsPage) handle() {
	common := pg.common
	for i := range pg.tabs.tabs {
		if pg.tabs.tabs[i].btn.Clicked() {
			pg.tabs.selected = i
		}

		for k := range pg.tabs.tabs[i].proposals {
			for pg.tabs.tabs[i].proposals[k].btn.Clicked() {
				selectedProposal := &pg.tabs.tabs[i].proposals[k].proposal
				common.changePage(ProposalDetailsPage(common, selectedProposal))
			}
		}
	}

	for pg.syncButton.Clicked() {
		pg.wallet.SyncProposals()
		// refresh todo
	}

	select {
	case prop := <-pg.syncedProposal:
		if prop.ProposalStatus == wallet.Synced {
			if !pg.proposalsItemSet {
				pg.initializeProposaltabItems()
			}
			go pg.updateProposalState()
			pg.isSynced = true
		} else if prop.ProposalStatus == wallet.NewProposalFound {
			pg.addDiscoveredProposal(false, *prop.Proposal)
			// refresh todo
		} else if prop.ProposalStatus == wallet.VoteStarted || prop.ProposalStatus == wallet.VoteFinished {
			pg.updateProposalVoteStatus(*prop.Proposal)
			// refresh todo
		}
	default:
	}

	if pg.isSynced {
		time.AfterFunc(time.Second*3, func() {
			pg.isSynced = false
		})
		// refresh todo
	}
}

func (pg *proposalsPage) layoutTabs(gtx C) D {
	width := float32(gtx.Constraints.Max.X-20) / float32(len(pg.tabs.tabs))

	return pg.tabCard.Layout(gtx, func(gtx C) D {
		return layout.Inset{
			Left:  values.MarginPadding12,
			Right: values.MarginPadding12,
		}.Layout(gtx, func(gtx C) D {
			return pg.proposalsList.Layout(gtx, len(pg.tabs.tabs), func(gtx C, i int) D {
				gtx.Constraints.Min.X = int(width)
				return layout.Stack{Alignment: layout.S}.Layout(gtx,
					layout.Stacked(func(gtx C) D {
						return decredmaterial.Clickable(gtx, pg.tabs.tabs[i].btn, func(gtx C) D {
							return layout.UniformInset(values.MarginPadding14).Layout(gtx, func(gtx C) D {
								return layout.Center.Layout(gtx, func(gtx C) D {
									return layout.Flex{}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											lbl := pg.theme.Body1(pg.tabs.tabs[i].title)
											lbl.Color = pg.theme.Color.Gray
											if pg.tabs.selected == i {
												lbl.Color = pg.theme.Color.Primary
											}
											return lbl.Layout(gtx)
										}),
										layout.Rigid(func(gtx C) D {
											return layout.Inset{Left: values.MarginPadding4, Top: values.MarginPadding2}.Layout(gtx, func(gtx C) D {
												c := pg.theme.Card()
												c.Color = pg.theme.Color.LightGray
												r := float32(8.5)
												c.Radius = decredmaterial.CornerRadius{NE: r, NW: r, SE: r, SW: r}
												lbl := pg.theme.Body2(strconv.Itoa(len(pg.tabs.tabs[i].proposals)))
												lbl.Color = pg.theme.Color.Gray
												if pg.tabs.selected == i {
													c.Color = pg.theme.Color.Primary
													lbl.Color = pg.theme.Color.Surface
												}
												return c.Layout(gtx, func(gtx C) D {
													return layout.Inset{
														Left:  values.MarginPadding5,
														Right: values.MarginPadding5,
													}.Layout(gtx, lbl.Layout)
												})
											})
										}),
									)
								})
							})
						})
					}),
					layout.Stacked(func(gtx C) D {
						if pg.tabs.selected != i {
							return D{}
						}
						tabHeight := gtx.Px(unit.Dp(2))
						tabRect := image.Rect(0, 0, int(width), tabHeight)
						paint.FillShape(gtx.Ops, pg.theme.Color.Primary, clip.Rect(tabRect).Op())
						return layout.Dimensions{
							Size: image.Point{X: int(width), Y: tabHeight},
						}
					}),
				)
			})
		})
	})
}

func (pg *proposalsPage) addDiscoveredProposal(first bool, proposal dcrlibwallet.Proposal) {
	for i := range pg.tabs.tabs {
		if pg.tabs.tabs[i].category == proposal.Category {
			item := proposalItem{
				btn:               new(widget.Clickable),
				proposal:          proposal,
				voteBar:           pg.theme.VoteBar(pg.infoIcon, pg.legendIcon),
				infoIcon:          pg.infoIcon,
				stateInfoTooltip:  pg.theme.Tooltip(),
				stateTooltipLabel: pg.theme.Caption(""),
			}
			if first {
				pg.tabs.tabs[i].proposals = append(pg.tabs.tabs[i].proposals, item)
				break
			} else {
				pg.tabs.tabs[i].proposals = append([]proposalItem{item}, pg.tabs.tabs[i].proposals...)
				break
			}
		}
	}
}

// updateProposalVoteStatus is called when voting has either started or ended for a particular proposal
func (pg *proposalsPage) updateProposalVoteStatus(proposal dcrlibwallet.Proposal) {
out:
	for i := range pg.tabs.tabs {
		for k := range pg.tabs.tabs[i].proposals {
			if pg.tabs.tabs[i].proposals[k].proposal.Token == proposal.Token {
				pg.tabs.tabs[i].proposals = append(pg.tabs.tabs[i].proposals[:k], pg.tabs.tabs[i].proposals[k+1:]...)
				break out
			}
		}
	}
	pg.addDiscoveredProposal(false, proposal)
}

func (pg *proposalsPage) updateProposalState() {
	for p := range pg.proposals {
		for i := range pg.tabs.tabs {
			if pg.tabs.tabs[i].category == dcrlibwallet.ProposalCategoryPre || pg.tabs.tabs[i].category == dcrlibwallet.ProposalCategoryActive {
				for k := range pg.tabs.tabs[i].proposals {
					if pg.tabs.tabs[i].proposals[k].proposal.Token == pg.proposals[p].Token {
						if pg.tabs.tabs[i].proposals[k].proposal.VoteStatus != pg.proposals[p].VoteStatus {
							pg.tabs.tabs[i].proposals[k].proposal.VoteStatus = pg.proposals[p].VoteStatus
						}
						if pg.tabs.tabs[i].proposals[k].proposal.YesVotes != pg.proposals[k].YesVotes {
							pg.tabs.tabs[i].proposals[k].proposal.YesVotes = pg.proposals[p].YesVotes
						}
						if pg.tabs.tabs[i].proposals[k].proposal.NoVotes != pg.proposals[k].NoVotes {
							pg.tabs.tabs[i].proposals[k].proposal.NoVotes = pg.proposals[p].NoVotes
						}
					}
				}
			}
		}
	}
}

func (pg *proposalsPage) layoutNoProposalsFound(gtx C) D {
	str := "No " + strings.ToLower(proposalCategoryTitles[pg.tabs.selected]) + " proposals"

	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return layout.Center.Layout(gtx, pg.theme.Body1(str).Layout)
}

func (pg *proposalsPage) layoutAuthorAndDate(gtx C, i int, proposal dcrlibwallet.Proposal) D {
	p := pg.tabs.tabs[pg.tabs.selected]
	grayCol := pg.theme.Color.Gray

	nameLabel := pg.theme.Body2(proposal.Username)
	nameLabel.Color = grayCol

	dotLabel := pg.theme.H4(" . ")
	dotLabel.Color = grayCol

	versionLabel := pg.theme.Body2("Version " + proposal.Version)
	versionLabel.Color = grayCol

	stateLabel := pg.theme.Body2(fmt.Sprintf("%v /2", proposal.VoteStatus))
	stateLabel.Color = grayCol

	timeAgoLabel := pg.theme.Body2(timeAgo(proposal.Timestamp))
	timeAgoLabel.Color = grayCol

	var categoryLabel decredmaterial.Label
	var categoryLabelColor color.NRGBA
	switch proposal.Category {
	case dcrlibwallet.ProposalCategoryApproved:
		categoryLabel = pg.theme.Body2("Approved")
		categoryLabelColor = pg.theme.Color.Success
	case dcrlibwallet.ProposalCategoryActive:
		categoryLabel = pg.theme.Body2("Voting")
		categoryLabelColor = pg.theme.Color.Primary
	case dcrlibwallet.ProposalCategoryRejected:
		categoryLabel = pg.theme.Body2("Rejected")
		categoryLabelColor = pg.theme.Color.Danger
	case dcrlibwallet.ProposalCategoryAbandoned:
		categoryLabel = pg.theme.Body2("Abandoned")
		categoryLabelColor = grayCol
	case dcrlibwallet.ProposalCategoryPre:
		categoryLabel = pg.theme.Body2("In discussion")
		categoryLabelColor = grayCol
	}
	categoryLabel.Color = categoryLabelColor

	return layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(nameLabel.Layout),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: values.MarginPaddingMinus22}.Layout(gtx, dotLabel.Layout)
				}),
				layout.Rigid(versionLabel.Layout),
			)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(categoryLabel.Layout),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: values.MarginPaddingMinus22}.Layout(gtx, dotLabel.Layout)
				}),
				layout.Rigid(func(gtx C) D {
					if p.title == "In discussion" {
						return layout.Flex{}.Layout(gtx,
							layout.Rigid(stateLabel.Layout),
							layout.Rigid(func(gtx C) D {
								rect := image.Rectangle{
									Min: gtx.Constraints.Min,
									Max: gtx.Constraints.Max,
								}
								rect.Max.Y = 20
								pg.layoutInfoTooltip(gtx, i, proposal.VoteStatus, rect)
								return layout.Inset{Left: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
									return p.proposals[i].infoIcon.Layout(gtx, unit.Dp(20))
								})
							}),
						)
					}
					pg.timerIcon.Scale = 1
					return layout.Flex{}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							if p.title == "Voting" {
								return layout.Inset{
									Right: values.MarginPadding4,
									Top:   values.MarginPadding3,
								}.Layout(gtx, pg.timerIcon.Layout)
							}
							return D{}
						}),
						layout.Rigid(timeAgoLabel.Layout),
					)
				}),
			)
		}),
	)
}

func (pg *proposalsPage) layoutInfoTooltip(gtx C, i int, state int32, rect image.Rectangle) {
	proposal := pg.tabs.tabs[pg.tabs.selected].proposals[i]
	inset := layout.Inset{Top: values.MarginPadding20, Left: values.MarginPaddingMinus230}
	proposal.stateInfoTooltip.Layout(gtx, rect, inset, func(gtx C) D {
		proposal.stateTooltipLabel.Color = pg.theme.Color.Gray
		if state == 1 {
			proposal.stateTooltipLabel.Text = "Waiting for author to authorize voting"
		} else if state == 2 {
			proposal.stateTooltipLabel.Text = "Waiting for admin to trigger the start of voting"
		}
		return proposal.stateTooltipLabel.Layout(gtx)
	})
}

func (pg *proposalsPage) layoutTitle(gtx C, proposal dcrlibwallet.Proposal) D {
	lbl := pg.theme.H6(proposal.Name)
	lbl.Font.Weight = text.Bold
	return layout.Inset{Top: values.MarginPadding4}.Layout(gtx, lbl.Layout)
}

func (pg *proposalsPage) layoutProposalVoteBar(gtx C, proposalItem proposalItem) D {
	yes := float32(proposalItem.proposal.YesVotes)
	no := float32(proposalItem.proposal.NoVotes)
	quorumPercent := float32(proposalItem.proposal.QuorumPercentage)
	passPercentage := float32(proposalItem.proposal.PassPercentage)
	eligibleTickets := float32(proposalItem.proposal.EligibleTickets)

	return proposalItem.voteBar.SetParams(yes, no, eligibleTickets, quorumPercent, passPercentage).LayoutWithLegend(gtx)
}

func (pg *proposalsPage) layoutProposalsList(gtx C) D {
	selected := pg.tabs.tabs[pg.tabs.selected]
	wdgs := make([]func(gtx C) D, len(selected.proposals))
	for i := range selected.proposals {
		index := i
		proposalItem := selected.proposals[index]
		pt := values.MarginPadding5
		if index == 0 {
			pt = values.MarginPadding16
		}
		wdgs[index] = func(gtx C) D {
			return layout.Inset{
				Top:    pt,
				Bottom: values.MarginPadding5,
				Left:   values.MarginPadding24,
				Right:  values.MarginPadding24,
			}.Layout(gtx, func(gtx C) D {
				return decredmaterial.Clickable(gtx, selected.proposals[index].btn, func(gtx C) D {
					return pg.itemCard.Layout(gtx, func(gtx C) D {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return layout.UniformInset(values.MarginPadding16).Layout(gtx, func(gtx C) D {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return pg.layoutAuthorAndDate(gtx, index, proposalItem.proposal)
								}),
								layout.Rigid(func(gtx C) D {
									return pg.layoutTitle(gtx, proposalItem.proposal)
								}),
								layout.Rigid(func(gtx C) D {
									if proposalItem.proposal.Category == dcrlibwallet.ProposalCategoryActive ||
										proposalItem.proposal.Category == dcrlibwallet.ProposalCategoryApproved ||
										proposalItem.proposal.Category == dcrlibwallet.ProposalCategoryRejected {
										return pg.layoutProposalVoteBar(gtx, proposalItem)
									}
									return D{}
								}),
							)
						})
					})
				})
			})
		}
	}
	return selected.container.Layout(gtx, len(wdgs), func(gtx C, i int) D {
		return layout.Inset{}.Layout(gtx, wdgs[i])
	})
}

func (pg *proposalsPage) layoutContent(gtx C) D {
	selected := pg.tabs.tabs[pg.tabs.selected]
	if len(selected.proposals) == 0 {
		return pg.layoutNoProposalsFound(gtx)
	}
	return pg.layoutProposalsList(gtx)
}

func (pg *proposalsPage) layoutIsSyncedSection(gtx C) D {
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return pg.updatedIcon.Layout(gtx, values.MarginPadding20)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Left: values.MarginPadding5}.Layout(gtx, pg.updatedLabel.Layout)
		}),
	)
}

func (pg *proposalsPage) layoutIsSyncingSection(gtx C) D {
	txt := pg.theme.Body2("Fetching...")
	txt.Color = pg.theme.Color.Gray
	return txt.Layout(gtx)
}

func (pg *proposalsPage) layoutStartSyncSection(gtx C) D {
	return material.Clickable(gtx, pg.syncButton, func(gtx C) D {
		pg.startSyncIcon.Scale = 0.68
		return pg.startSyncIcon.Layout(gtx)
	})
}

func (pg *proposalsPage) layoutSyncSection(gtx C) D {
	if pg.isSynced {
		return pg.layoutIsSyncedSection(gtx)
	} else if pg.wallet.IsSyncingProposals() {
		return pg.layoutIsSyncingSection(gtx)
	}
	return pg.layoutStartSyncSection(gtx)
}

func (pg *proposalsPage) initializeProposaltabItems() {
	pg.proposalsItemSet = true
	if len(pg.proposals) == 0 {
		pg.wallet.SyncProposals()
		pg.proposalsItemSet = false
	}

	for i := range pg.proposals {
		if i != len(pg.proposals) {
			pg.addDiscoveredProposal(true, pg.proposals[i])
		}
	}
}

func (pg *proposalsPage) Layout(gtx C) D {

	if !pg.proposalsItemSet {
		pg.initializeProposaltabItems()
	}

	border := widget.Border{Color: pg.theme.Color.Gray1, CornerRadius: values.MarginPadding0, Width: values.MarginPadding1}
	borderLayout := func(gtx layout.Context, body layout.Widget) layout.Dimensions {
		return border.Layout(gtx, body)
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{}.Layout(gtx,
				layout.Flexed(1, func(gtx C) D {
					return borderLayout(gtx, pg.layoutTabs)
				}),
				layout.Rigid(func(gtx C) D {
					return borderLayout(gtx, func(gtx C) D {
						return pg.syncCard.Layout(gtx, func(gtx C) D {
							m := values.MarginPadding12
							if pg.isSynced {
								m = values.MarginPadding14
							} else if pg.wallet.IsSyncingProposals() {
								m = values.MarginPadding15
							}
							return layout.UniformInset(m).Layout(gtx, func(gtx C) D {
								return layout.Center.Layout(gtx, func(gtx C) D {
									return pg.layoutSyncSection(gtx)
								})
							})
						})
					})
				}),
			)
		}),
		layout.Flexed(1, pg.layoutContent),
	)
}

func (pg *proposalsPage) onClose() {}

func timeAgo(timestamp int64) string {
	timeAgo, _ := timeago.TimeAgoWithTime(time.Now(), time.Unix(timestamp, 0))
	return timeAgo
}

func truncateString(str string, num int) string {
	bnoden := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		bnoden = str[0:num] + "..."
	}
	return bnoden
}
