package ui

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
	"sync"
	"time"

	"gioui.org/font/gofont"
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
	proposal     dcrlibwallet.Proposal
	voteBar      decredmaterial.VoteBar
	tooltip      *decredmaterial.Tooltip
	tooltipLabel decredmaterial.Label
}

type proposalsPage struct {
	*pageCommon
	pageClosing           chan bool
	proposalMu            sync.Mutex
	proposalItems         []proposalItem
	selectedProposal      **dcrlibwallet.Proposal
	proposalCount         []int
	selectedCategoryIndex int
	categoryList          *decredmaterial.ClickableList
	proposalsList         *decredmaterial.ClickableList
	tabCard               decredmaterial.Card
	itemCard              decredmaterial.Card
	syncCard              decredmaterial.Card
	updatedLabel          decredmaterial.Label
	legendIcon            *widget.Icon
	infoIcon              *widget.Icon
	updatedIcon           *widget.Icon
	syncButton            *widget.Clickable
	startSyncIcon         *widget.Image
	timerIcon             *widget.Image

	showSyncedCompleted bool
	isSyncing           bool
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

func ProposalsPage(common *pageCommon) Page {
	pg := &proposalsPage{
		pageCommon:            common,
		pageClosing:           make(chan bool, 1),
		selectedCategoryIndex: -1,
		categoryList:          common.theme.NewClickableList(layout.Horizontal),
		proposalsList:         common.theme.NewClickableList(layout.Vertical),
		tabCard:               common.theme.Card(),
		itemCard:              common.theme.Card(),
		syncCard:              common.theme.Card(),
		legendIcon:            common.icons.imageBrightness1,
		infoIcon:              common.icons.actionInfo,
		updatedIcon:           common.icons.navigationCheck,
		updatedLabel:          common.theme.Body2("Updated"),
		syncButton:            new(widget.Clickable),
		startSyncIcon:         common.icons.restore,
		timerIcon:             common.icons.timerIcon,
	}
	pg.infoIcon.Color = common.theme.Color.Gray
	pg.legendIcon.Color = common.theme.Color.InactiveGray

	pg.updatedIcon.Color = common.theme.Color.Success
	pg.updatedLabel.Color = common.theme.Color.Success

	pg.tabCard.Radius = decredmaterial.CornerRadius{NE: 0, NW: 0, SE: 0, SW: 0}
	pg.syncCard.Radius = decredmaterial.CornerRadius{NE: 0, NW: 0, SE: 0, SW: 0}

	pg.proposalsList.DividerHeight = values.MarginPadding8

	return pg
}

func (pg *proposalsPage) OnResume() {
	pg.listenForSyncNotifications()

	pg.proposalMu.Lock()
	selectedCategory := pg.selectedCategoryIndex
	pg.proposalMu.Unlock()
	if selectedCategory == -1 {
		pg.countProposals()
		pg.loadProposals(0)
	}

	pg.isSyncing = pg.multiWallet.Politeia.IsSyncing()
}

func (pg *proposalsPage) countProposals() {
	proposalCount := make([]int, len(proposalCategories))
	for i, category := range proposalCategories {
		count, err := pg.multiWallet.Politeia.Count(category)
		if err == nil {
			proposalCount[i] = int(count)
		}
	}

	pg.proposalMu.Lock()
	pg.proposalCount = proposalCount
	pg.proposalMu.Unlock()
}

func (pg *proposalsPage) loadProposals(category int) {
	proposals, err := pg.multiWallet.Politeia.GetProposalsRaw(proposalCategories[category], 0, 0, true)
	if err != nil {
		log.Error("Error loading proposals:", err)
		pg.proposalMu.Lock()
		pg.proposalItems = make([]proposalItem, 0)
		pg.proposalMu.Unlock()
	} else {
		proposalItems := make([]proposalItem, len(proposals))
		for i := 0; i < len(proposals); i++ {
			proposal := proposals[i]
			item := proposalItem{
				proposal: proposals[i],
				voteBar:  pg.theme.VoteBar(pg.infoIcon, pg.legendIcon),
			}

			if proposal.Category == dcrlibwallet.ProposalCategoryPre {
				tooltipLabel := pg.theme.Caption("")
				tooltipLabel.Color = pg.theme.Color.Gray
				if proposal.VoteStatus == 1 {
					tooltipLabel.Text = "Waiting for author to authorize voting"
				} else if proposal.VoteStatus == 2 {
					tooltipLabel.Text = "Waiting for admin to trigger the start of voting"
				}

				item.tooltip = pg.theme.Tooltip()
				item.tooltipLabel = tooltipLabel
			}

			proposalItems[i] = item
		}
		pg.proposalMu.Lock()
		pg.selectedCategoryIndex = category
		pg.proposalItems = proposalItems
		pg.proposalMu.Unlock()
	}
}

func (pg *proposalsPage) Handle() {

	if clicked, selectedItem := pg.categoryList.ItemClicked(); clicked {
		go pg.loadProposals(selectedItem)
	}

	if clicked, selectedItem := pg.proposalsList.ItemClicked(); clicked {
		pg.proposalMu.Lock()
		selectedProposal := pg.proposalItems[selectedItem].proposal
		pg.proposalMu.Unlock()

		pg.changeFragment(ProposalDetailsPage(pg.pageCommon, selectedProposal), PageProposalDetails)
	}

	for pg.syncButton.Clicked() {
		pg.isSyncing = true
		go pg.multiWallet.Politeia.Sync(dcrlibwallet.PoliteiaMainnetHost)
	}

	if pg.showSyncedCompleted {
		time.AfterFunc(time.Second*3, func() {
			pg.showSyncedCompleted = false
		})
	}
}

func (pg *proposalsPage) layoutTabs(gtx C) D {
	width := float32(gtx.Constraints.Max.X-20) / float32(len(proposalCategoryTitles))
	pg.proposalMu.Lock()
	selectedCategory := pg.selectedCategoryIndex
	pg.proposalMu.Unlock()

	return pg.tabCard.Layout(gtx, func(gtx C) D {
		return layout.Inset{
			Left:  values.MarginPadding12,
			Right: values.MarginPadding12,
		}.Layout(gtx, func(gtx C) D {
			return pg.categoryList.Layout(gtx, len(proposalCategoryTitles), func(gtx C, i int) D {
				gtx.Constraints.Min.X = int(width)
				return layout.Stack{Alignment: layout.S}.Layout(gtx,
					layout.Stacked(func(gtx C) D {
						return layout.UniformInset(values.MarginPadding14).Layout(gtx, func(gtx C) D {
							return layout.Center.Layout(gtx, func(gtx C) D {
								return layout.Flex{}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										lbl := pg.theme.Body1(proposalCategoryTitles[i])
										lbl.Color = pg.theme.Color.Gray
										if selectedCategory == i {
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
											lbl := pg.theme.Body2(strconv.Itoa(pg.proposalCount[i]))
											lbl.Color = pg.theme.Color.Gray
											if selectedCategory == i {
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
					}),
					layout.Stacked(func(gtx C) D {
						if selectedCategory != i {
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

func (pg *proposalsPage) layoutNoProposalsFound(gtx C) D {
	pg.proposalMu.Lock()
	selectedCategory := pg.selectedCategoryIndex
	pg.proposalMu.Unlock()
	str := fmt.Sprintf("No %s proposals", strings.ToLower(proposalCategoryTitles[selectedCategory]))

	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return layout.Center.Layout(gtx, pg.theme.Body1(str).Layout)
}

func (pg *proposalsPage) layoutAuthorAndDate(gtx C, item proposalItem) D {
	proposal := item.proposal
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

	pg.proposalMu.Lock()
	selectedCategory := pg.selectedCategoryIndex
	pg.proposalMu.Unlock()

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
					if proposalCategories[selectedCategory] == dcrlibwallet.ProposalCategoryPre {
						return layout.Flex{}.Layout(gtx,
							layout.Rigid(stateLabel.Layout),
							layout.Rigid(func(gtx C) D {
								rect := image.Rectangle{
									Min: gtx.Constraints.Min,
									Max: gtx.Constraints.Max,
								}
								rect.Max.Y = 20
								pg.layoutInfoTooltip(gtx, rect, item)
								return layout.Inset{Left: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
									return pg.infoIcon.Layout(gtx, unit.Dp(20))
								})
							}),
						)
					}
					pg.timerIcon.Scale = 1
					return layout.Flex{}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							if proposalCategories[selectedCategory] == dcrlibwallet.ProposalCategoryActive {
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

func (pg *proposalsPage) layoutInfoTooltip(gtx C, rect image.Rectangle, item proposalItem) {
	inset := layout.Inset{Top: values.MarginPadding20, Left: values.MarginPaddingMinus230}
	item.tooltip.Layout(gtx, rect, inset, func(gtx C) D {
		return item.tooltipLabel.Layout(gtx)
	})
}

func (pg *proposalsPage) layoutTitle(gtx C, proposal dcrlibwallet.Proposal) D {
	lbl := pg.theme.H6(proposal.Name)
	lbl.Font.Weight = text.Bold
	return layout.Inset{Top: values.MarginPadding4}.Layout(gtx, lbl.Layout)
}

func (pg *proposalsPage) layoutProposalVoteBar(gtx C, item proposalItem) D {
	proposal := item.proposal
	yes := float32(proposal.YesVotes)
	no := float32(proposal.NoVotes)
	quorumPercent := float32(proposal.QuorumPercentage)
	passPercentage := float32(proposal.PassPercentage)
	eligibleTickets := float32(proposal.EligibleTickets)

	return item.voteBar.SetParams(yes, no, eligibleTickets, quorumPercent, passPercentage).LayoutWithLegend(gtx)
}

func (pg *proposalsPage) layoutProposalsList(gtx C) D {
	pg.proposalMu.Lock()
	proposalItems := pg.proposalItems
	pg.proposalMu.Unlock()
	return pg.proposalsList.Layout(gtx, len(proposalItems), func(gtx C, i int) D {
		return layout.Inset{}.Layout(gtx, func(gtx C) D {
			return layout.Inset{
				Top:    values.MarginPadding2,
				Bottom: values.MarginPadding2,
				Left:   values.MarginPadding2,
				Right:  values.MarginPadding2,
			}.Layout(gtx, func(gtx C) D {
				return pg.itemCard.Layout(gtx, func(gtx C) D {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return layout.UniformInset(values.MarginPadding16).Layout(gtx, func(gtx C) D {
						item := proposalItems[i]
						proposal := item.proposal
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								return pg.layoutAuthorAndDate(gtx, item)
							}),
							layout.Rigid(func(gtx C) D {
								return pg.layoutTitle(gtx, proposal)
							}),
							layout.Rigid(func(gtx C) D {
								if proposal.Category == dcrlibwallet.ProposalCategoryActive ||
									proposal.Category == dcrlibwallet.ProposalCategoryApproved ||
									proposal.Category == dcrlibwallet.ProposalCategoryRejected {
									return pg.layoutProposalVoteBar(gtx, item)
								}
								return D{}
							}),
						)
					})
				})
			})
		})
	})
}

func (pg *proposalsPage) layoutContent(gtx C) D {
	pg.proposalMu.Lock()
	proposalItems := pg.proposalItems
	pg.proposalMu.Unlock()
	if len(proposalItems) == 0 {
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
	th := material.NewTheme(gofont.Collection())
	gtx.Constraints.Min.X = gtx.Px(unit.Dp(20))
	loader := material.Loader(th)
	loader.Color = pg.theme.Color.Gray
	return loader.Layout(gtx)
}

func (pg *proposalsPage) layoutStartSyncSection(gtx C) D {
	return material.Clickable(gtx, pg.syncButton, func(gtx C) D {
		pg.startSyncIcon.Scale = 0.68
		return pg.startSyncIcon.Layout(gtx)
	})
}

func (pg *proposalsPage) layoutSyncSection(gtx C) D {
	if pg.showSyncedCompleted {
		return pg.layoutIsSyncedSection(gtx)
	} else if pg.multiWallet.Politeia.IsSyncing() {
		return pg.layoutIsSyncingSection(gtx)
	}
	return pg.layoutStartSyncSection(gtx)
}

func (pg *proposalsPage) Layout(gtx C) D {

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
							if pg.showSyncedCompleted || pg.isSyncing {
								m = values.MarginPadding14
							}
							return layout.UniformInset(m).Layout(gtx, func(gtx C) D {
								return layout.Center.Layout(gtx, pg.layoutSyncSection)
							})
						})
					})
				}),
			)
		}),
		layout.Flexed(1, func(gtx C) D {
			return pg.UniformPadding(gtx, pg.layoutContent)
		}),
	)
}

func (pg *proposalsPage) listenForSyncNotifications() {
	go func() {
		for {
			var notification interface{}

			select {
			case notification = <-pg.notificationsUpdate:
			case <-pg.pageClosing:
				return
			}

			switch n := notification.(type) {
			case wallet.Proposal:
				if n.ProposalStatus == wallet.Synced {
					pg.isSyncing = false
					pg.showSyncedCompleted = true

					pg.proposalMu.Lock()
					selectedCategory := pg.selectedCategoryIndex
					pg.proposalMu.Unlock()
					if selectedCategory != -1 {
						pg.countProposals()
						pg.loadProposals(selectedCategory)
					}
				}
			}
		}
	}()
}

func (pg *proposalsPage) OnClose() {
	pg.pageClosing <- true
}

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
