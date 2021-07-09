package ui

import (
	"sort"
	"strings"
	"time"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/wallet"

	"gioui.org/layout"
	"gioui.org/text"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const PageTicketsActivity = "TicketsActivity"

type ticketsActivityPage struct {
	th           *decredmaterial.Theme
	tickets      **wallet.Tickets
	ticketsList  layout.List
	filterSorter int

	orderDropDown      *decredmaterial.DropDown
	ticketTypeDropDown *decredmaterial.DropDown
	walletDropDown     *decredmaterial.DropDown
	common             *pageCommon

	wallets []*dcrlibwallet.Wallet

	backButton decredmaterial.IconButton
}

func TicketActivityPage(c *pageCommon) Page {
	pg := &ticketsActivityPage{
		th:          c.theme,
		common:      c,
		tickets:     c.walletTickets,
		ticketsList: layout.List{Axis: layout.Vertical},
		wallets:     c.multiWallet.AllWallets(),
	}
	pg.orderDropDown = createOrderDropDown(c)
	pg.ticketTypeDropDown = c.theme.DropDown([]decredmaterial.DropDownItem{
		{Text: "All"},
		{Text: "Unmined"},
		{Text: "Immature"},
		{Text: "Live"},
		{Text: "Voted"},
		{Text: "Missed"},
		{Text: "Expired"},
		{Text: "Revoked"},
	}, 1)

	pg.backButton, _ = c.SubPageHeaderButtons()

	return pg
}

func (pg *ticketsActivityPage) OnResume() {

}

func (pg *ticketsActivityPage) Layout(gtx layout.Context) layout.Dimensions {
	c := pg.common
	c.createOrUpdateWalletDropDown(&pg.walletDropDown, pg.wallets)
	body := func(gtx C) D {
		page := SubPage{
			title:      "Ticket activity",
			backButton: pg.backButton,
			back: func() {
				c.changePage(PageTickets)
			},
			body: func(gtx C) D {
				walletID := pg.wallets[pg.walletDropDown.SelectedIndex()].ID
				tickets := (*pg.tickets).Confirmed[walletID]
				return layout.Stack{Alignment: layout.N}.Layout(gtx,
					layout.Expanded(func(gtx C) D {
						return layout.Inset{Top: values.MarginPadding60}.Layout(gtx, func(gtx C) D {
							return c.theme.Card().Layout(gtx, func(gtx C) D {
								gtx.Constraints.Min = gtx.Constraints.Max
								if pg.ticketTypeDropDown.SelectedIndex()-1 != -1 {
									tickets = filterTickets(tickets, func(ticketStatus string) bool {
										return ticketStatus == strings.ToUpper(pg.ticketTypeDropDown.Selected())
									})
								}

								if len(tickets) == 0 {
									txt := c.theme.Body1("No tickets yet")
									txt.Color = c.theme.Color.Gray2
									txt.Alignment = text.Middle
									return layout.Inset{Top: values.MarginPadding15}.Layout(gtx, func(gtx C) D { return txt.Layout(gtx) })
								}
								return layout.UniformInset(values.MarginPadding16).Layout(gtx, func(gtx C) D {
									return pg.ticketsList.Layout(gtx, len(tickets), func(gtx C, index int) D {
										return ticketActivityRow(gtx, c, tickets[index], index)
									})
								})
							})
						})
					}),
					layout.Stacked(func(gtx C) D {
						return pg.dropDowns(gtx)
					}),
				)
			},
		}
		return c.SubPageLayout(gtx, page)
	}

	return c.UniformPadding(gtx, body)
}

func (pg *ticketsActivityPage) dropDowns(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return pg.walletDropDown.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Left: values.MarginPadding5,
					}.Layout(gtx, func(gtx C) D {
						return pg.ticketTypeDropDown.Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Left: values.MarginPadding5,
					}.Layout(gtx, func(gtx C) D {
						return pg.orderDropDown.Layout(gtx)
					})
				}),
			)
		}),
	)
}

func filterTickets(tickets []wallet.Ticket, f func(string) bool) []wallet.Ticket {
	t := make([]wallet.Ticket, 0)
	for _, v := range tickets {
		if f(v.Info.Status) {
			t = append(t, v)
		}
	}
	return t
}

func (pg *ticketsActivityPage) Handle() {

	sortSelection := pg.orderDropDown.SelectedIndex()
	if pg.filterSorter != sortSelection {
		pg.filterSorter = sortSelection
		newestFirst := pg.filterSorter == 0
		for _, wal := range pg.wallets {
			tickets := (*pg.tickets).Confirmed[wal.ID]
			sort.SliceStable(tickets, func(i, j int) bool {
				backTime := time.Unix(tickets[j].Info.Ticket.Timestamp, 0)
				frontTime := time.Unix(tickets[i].Info.Ticket.Timestamp, 0)
				if newestFirst {
					return backTime.Before(frontTime)
				}
				return frontTime.Before(backTime)
			})
		}
	}

}

func (pg *ticketsActivityPage) OnClose() {}
