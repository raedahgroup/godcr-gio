package page

import (
	"fmt"
	"strconv"

	"github.com/planetdecred/godcr/ui/modal"

	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/decred/dcrd/dcrutil/v3"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/values"
)

const AccountDetailsPageID = "AccountDetails"

type AcctDetailsPage struct {
	*load.Load
	wallet  *dcrlibwallet.Wallet
	account *dcrlibwallet.Account

	theme                    *decredmaterial.Theme
	acctDetailsPageContainer layout.List
	backButton               decredmaterial.IconButton
	editAccount              *widget.Clickable

	stakingBalance   int64
	totalBalance     string
	spendable        string
	immatureRewards  string
	lockedByTickets  string
	votingAuthority  string
	immatureStakeGen string
	hdPath           string
	keys             string
}

func NewAcctDetailsPage(l *load.Load, account *dcrlibwallet.Account) *AcctDetailsPage {
	pg := &AcctDetailsPage{
		Load:    l,
		wallet:  l.WL.MultiWallet.WalletWithID(account.WalletID),
		account: account,

		theme: l.Theme,
		acctDetailsPageContainer: layout.List{
			Axis: layout.Vertical,
		},
		backButton:  l.Theme.PlainIconButton(new(widget.Clickable), l.Icons.NavigationArrowBack),
		editAccount: new(widget.Clickable),
	}

	pg.backButton, _ = subpageHeaderButtons(l)

	return pg
}

func (pg *AcctDetailsPage) OnResume() {

	balance := pg.account.Balance

	pg.stakingBalance = balance.ImmatureReward + balance.LockedByTickets + balance.VotingAuthority +
		balance.ImmatureStakeGeneration

	pg.totalBalance = dcrutil.Amount(balance.Total).String()
	pg.spendable = dcrutil.Amount(balance.Spendable).String()
	pg.immatureRewards = dcrutil.Amount(balance.ImmatureReward).String()
	pg.lockedByTickets = dcrutil.Amount(balance.LockedByTickets).String()
	pg.votingAuthority = dcrutil.Amount(balance.VotingAuthority).String()
	pg.immatureStakeGen = dcrutil.Amount(balance.ImmatureStakeGeneration).String()

	pg.hdPath = pg.WL.HDPrefix() + strconv.Itoa(int(pg.account.Number)) + "'"

	ext := pg.account.ExternalKeyCount
	internal := pg.account.InternalKeyCount
	imp := pg.account.ImportedKeyCount
	pg.keys = fmt.Sprintf("%d external, %d internal, %d imported", ext, internal, imp)
}

func (pg *AcctDetailsPage) Layout(gtx layout.Context) layout.Dimensions {
	widgets := []func(gtx C) D{
		func(gtx C) D {
			return pg.accountBalanceLayout(gtx)
		},
		func(gtx C) D {
			m := values.MarginPadding10
			return layout.Inset{Top: m, Bottom: m}.Layout(gtx, func(gtx C) D {
				return pg.theme.Separator().Layout(gtx)
			})
		},
		func(gtx C) D {
			return pg.accountInfoLayout(gtx)
		},
	}

	body := func(gtx C) D {
		sp := SubPage{
			Load:       pg.Load,
			title:      pg.account.Name,
			walletName: pg.wallet.Name,
			backButton: pg.backButton,
			back: func() {
				pg.ChangePage(WalletPageID)
			},
			body: func(gtx C) D {
				return layout.Inset{Bottom: values.MarginPadding7}.Layout(gtx, func(gtx C) D {
					return pg.theme.Card().Layout(gtx, func(gtx C) D {
						return layout.Inset{Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
							return pg.acctDetailsPageContainer.Layout(gtx, len(widgets), func(gtx C, i int) D {
								return layout.Inset{}.Layout(gtx, widgets[i])
							})
						})
					})
				})
			},
			extraItem: pg.editAccount,
			extra: func(gtx C) D {
				return layout.Inset{}.Layout(gtx, func(gtx C) D {
					edit := pg.Icons.EditIcon
					edit.Scale = 1
					return layout.E.Layout(gtx, edit.Layout)
				})
			},
		}
		return sp.Layout(gtx)
	}
	return uniformPadding(gtx, body)
}

func (pg *AcctDetailsPage) accountBalanceLayout(gtx layout.Context) layout.Dimensions {

	return pg.pageSections(gtx, func(gtx C) D {
		accountIcon := pg.Icons.AccountIcon
		if pg.account.Number == MaxInt32 {
			accountIcon = pg.Icons.ImportedAccountIcon
		}
		accountIcon.Scale = 1

		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						m := values.MarginPadding10
						return layout.Inset{
							Right: m,
							Top:   m,
						}.Layout(gtx, accountIcon.Layout)
					}),
					layout.Rigid(func(gtx C) D {
						return pg.acctBalLayout(gtx, "Total Balance", pg.totalBalance, true)
					}),
				)
			}),
			layout.Rigid(func(gtx C) D {
				return pg.acctBalLayout(gtx, "Spendable", pg.spendable, false)
			}),
			layout.Rigid(func(gtx C) D {
				if pg.stakingBalance == 0 {
					return layout.Dimensions{}
				}

				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return pg.acctBalLayout(gtx, "Immature Rewards", pg.immatureRewards, false)
					}),
					layout.Rigid(func(gtx C) D {
						return pg.acctBalLayout(gtx, "Locked By Tickets", pg.lockedByTickets, false)
					}),
					layout.Rigid(func(gtx C) D {
						return pg.acctBalLayout(gtx, "Voting Authority", pg.votingAuthority, false)
					}),
					layout.Rigid(func(gtx C) D {
						return pg.acctBalLayout(gtx, "Immature Stake Gen", pg.immatureStakeGen, false)
					}),
				)
			}),
		)
	})
}

func (pg *AcctDetailsPage) acctBalLayout(gtx layout.Context, balType string, balance string, isTotalBalance bool) layout.Dimensions {

	mainBalance, subBalance := breakBalance(pg.Printer, balance)

	mainLabel := pg.theme.Body1(mainBalance)
	subLabel := pg.theme.Caption(subBalance)
	subLabel.Color = pg.theme.Color.DeepBlue
	marginTop := values.MarginPadding16
	marginLeft := values.MarginPadding35

	if isTotalBalance {
		mainLabel = pg.theme.H4(mainBalance)
		subLabel = pg.theme.Body1(subBalance)
		marginTop = values.MarginPadding0
		marginLeft = values.MarginPadding0
	}
	return layout.Inset{
		Right: values.MarginPadding10,
		Top:   marginTop,
		Left:  marginLeft,
	}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
					layout.Rigid(mainLabel.Layout),
					layout.Rigid(subLabel.Layout),
				)
			}),
			layout.Rigid(func(gtx C) D {
				txt := pg.theme.Body2(balType)
				txt.Color = pg.theme.Color.Gray
				return txt.Layout(gtx)
			}),
		)
	})
}

func (pg *AcctDetailsPage) accountInfoLayout(gtx layout.Context) layout.Dimensions {
	return pg.pageSections(gtx, func(gtx C) D {
		m := values.MarginPadding10
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return pg.acctInfoLayout(gtx, "Account Number", fmt.Sprint(pg.account.Number))
			}),
			layout.Rigid(func(gtx C) D {
				inset := layout.Inset{
					Top:    m,
					Bottom: m,
				}
				return inset.Layout(gtx, func(gtx C) D {
					return pg.acctInfoLayout(gtx, "HD Path", pg.hdPath)
				})
			}),
			layout.Rigid(func(gtx C) D {
				inset := layout.Inset{
					Bottom: m,
				}
				return inset.Layout(gtx, func(gtx C) D {
					return pg.acctInfoLayout(gtx, "Keys", pg.keys)
				})
			}),
		)
	})
}

func (pg *AcctDetailsPage) acctInfoLayout(gtx layout.Context, leftText, rightText string) layout.Dimensions {
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					leftTextLabel := pg.theme.Label(values.TextSize14, leftText)
					leftTextLabel.Color = pg.theme.Color.Gray
					return leftTextLabel.Layout(gtx)
				}),
			)
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, pg.theme.Body1(rightText).Layout)
		}),
	)
}

func (pg *AcctDetailsPage) pageSections(gtx layout.Context, body layout.Widget) layout.Dimensions {
	m := values.MarginPadding20
	mtb := values.MarginPadding5
	return layout.Inset{Left: m, Right: m, Top: mtb, Bottom: mtb}.Layout(gtx, body)
}

func (pg *AcctDetailsPage) Handle() {
	if pg.editAccount.Clicked() {
		textModal := modal.NewTextInputModal(pg.Load).
			Hint("Account name").
			PositiveButton(values.String(values.StrRename), func(newName string, tim *modal.TextInputModal) bool {
				err := pg.wallet.RenameAccount(pg.account.Number, newName)
				if err != nil {
					tim.SetError(err.Error())
					tim.IsLoading = false
					return false
				}

				pg.account.Name = newName
				return true
			})

		textModal.Title("Rename account").
			NegativeButton(values.String(values.StrCancel), func() {})
		textModal.Show()
	}
}

func (pg *AcctDetailsPage) OnClose() {}
