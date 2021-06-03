package ui

import (
	"fmt"
	"strconv"
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/decred/dcrd/dcrutil"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
)

const PageAccountDetails = "AccountDetails"

const Uint32Size = 32 << (^uint32(0) >> 32 & 1) // 32 or 64
const MaxInt32 = 1<<(Uint32Size-1) - 1

type acctDetailsPage struct {
	common                   pageCommon
	walletName               string
	account                  *dcrlibwallet.Account
	theme                    *decredmaterial.Theme
	acctDetailsPageContainer layout.List
	// backButton               decredmaterial.IconButton
	editAccount *widget.Clickable

	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton
}

func AcctDetailsPage(common pageCommon, account *dcrlibwallet.Account) Page {
	pg := &acctDetailsPage{
		acctDetailsPageContainer: layout.List{
			Axis: layout.Vertical,
		},
		common:      common,
		walletName:  common.multiWallet.WalletWithID(account.WalletID).Name,
		account:     account,
		theme:       common.theme,
		editAccount: new(widget.Clickable),
	}

	pg.backButton, pg.infoButton = common.SubPageHeaderButtons()

	return pg
}

func (pg *acctDetailsPage) pageID() string {
	return PageAccountDetails
}

func (pg *acctDetailsPage) Layout(gtx layout.Context) layout.Dimensions {
	common := pg.common

	widgets := []func(gtx C) D{
		func(gtx C) D {
			return pg.accountBalanceLayout(gtx, &common)
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

	acctName := strings.Title(strings.ToLower(pg.account.Name))
	body := func(gtx C) D {
		page := SubPage{
			title:      acctName,
			walletName: pg.walletName,
			back: func() {
				common.popPage()
			},
			backButton: pg.backButton,
			infoButton: pg.infoButton,
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
					edit := common.icons.editIcon
					edit.Scale = 1
					return layout.E.Layout(gtx, edit.Layout)
				})
			},
			handleExtra: func() {
				go func() {
					common.modalReceiver <- &modalLoad{
						template: RenameAccountTemplate,
						title:    "Rename account",
						confirm: func(name string) {
							go func() {
								wal := pg.common.multiWallet.WalletWithID(pg.account.WalletID)
								err := wal.RenameAccount(pg.account.Number, name)
								if err != nil {
									log.Error("Error renaming account:", err)
									pg.common.modalLoad.setLoading(false)
									if err.Error() == dcrlibwallet.ErrInvalidPassphrase {
										e := values.String(values.StrInvalidPassphrase)
										pg.common.notify(e, false)
										return
									}
									pg.common.notify(err.Error(), false)
								} else {
									pg.account.Name = name
									common.closeModal()
								}
							}()

						},
						confirmText: "Rename",
						cancel:      common.closeModal,
						cancelText:  "Cancel",
					}
				}()
			},
		}
		return common.SubPageLayout(gtx, page)
	}

	return common.UniformPadding(gtx, body)
}

func (pg *acctDetailsPage) accountBalanceLayout(gtx layout.Context, common *pageCommon) layout.Dimensions {
	totalBalance := dcrutil.Amount(pg.account.TotalBalance).String()
	totalBalanceMain, totalBalanceSub := breakBalance(common.printer, totalBalance)

	spendable := dcrutil.Amount(pg.account.Balance.Spendable).String()
	spendableMain, spendableSub := breakBalance(common.printer, spendable)

	immatureRewards := dcrutil.Amount(pg.account.Balance.ImmatureReward).String()
	rewardBalanceMain, rewardBalanceSub := breakBalance(common.printer, immatureRewards)

	lockedByTickets := dcrutil.Amount(pg.account.Balance.LockedByTickets).String()
	lockBalanceMain, lockBalanceSub := breakBalance(common.printer, lockedByTickets)

	votingAuthority := dcrutil.Amount(pg.account.Balance.VotingAuthority).String()
	voteBalanceMain, voteBalanceSub := breakBalance(common.printer, votingAuthority)

	immatureStakeGen := dcrutil.Amount(pg.account.Balance.ImmatureStakeGeneration).String()
	stakeBalanceMain, stakeBalanceSub := breakBalance(common.printer, immatureStakeGen)

	return pg.pageSections(gtx, func(gtx C) D {
		accountIcon := common.icons.accountIcon
		if pg.account.Number == MaxInt32 {
			accountIcon = common.icons.importedAccountIcon
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
						return pg.acctBalLayout(gtx, "Total Balance", totalBalanceMain, totalBalanceSub, true)
					}),
				)
			}),
			layout.Rigid(func(gtx C) D {
				return pg.acctBalLayout(gtx, "Spendable", spendableMain, spendableSub, false)
			}),
			layout.Rigid(func(gtx C) D {
				return pg.acctBalLayout(gtx, "Immature Rewards", rewardBalanceMain, rewardBalanceSub, false)
			}),
			layout.Rigid(func(gtx C) D {
				return pg.acctBalLayout(gtx, "Locked By Tickets", lockBalanceMain, lockBalanceSub, false)
			}),
			layout.Rigid(func(gtx C) D {
				return pg.acctBalLayout(gtx, "Voting Authority", voteBalanceMain, voteBalanceSub, false)
			}),
			layout.Rigid(func(gtx C) D {
				return pg.acctBalLayout(gtx, "Immature Stake Gen", stakeBalanceMain, stakeBalanceSub, false)
			}),
		)
	})
}

func (pg *acctDetailsPage) acctBalLayout(gtx layout.Context, balType string, mainBalance, subBalance string, isFirst bool) layout.Dimensions {
	mainLabel := pg.theme.Body1(mainBalance)
	subLabel := pg.theme.Caption(subBalance)
	subLabel.Color = pg.theme.Color.DeepBlue
	marginTop := values.MarginPadding16
	marginLeft := values.MarginPadding35
	if isFirst {
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

func (pg *acctDetailsPage) accountInfoLayout(gtx layout.Context) layout.Dimensions {
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
					return pg.acctInfoLayout(gtx, "HD Path", pg.hdPath())
				})
			}),
			layout.Rigid(func(gtx C) D {
				inset := layout.Inset{
					Bottom: m,
				}
				return inset.Layout(gtx, func(gtx C) D {
					ext := pg.account.ExternalKeyCount
					internal := pg.account.InternalKeyCount
					imp := pg.account.ImportedKeyCount
					keys := fmt.Sprintf("%d external, %d internal, %d imported", ext, internal, imp)
					return pg.acctInfoLayout(gtx, "Key", keys)
				})
			}),
		)
	})
}

func (pg *acctDetailsPage) hdPath() string {
	return pg.common.HDPrefix() + strconv.Itoa(int(pg.account.Number)) + "'"
}

func (pg *acctDetailsPage) acctInfoLayout(gtx layout.Context, leftText, rightText string) layout.Dimensions {
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

func (pg *acctDetailsPage) pageSections(gtx layout.Context, body layout.Widget) layout.Dimensions {
	m := values.MarginPadding20
	mtb := values.MarginPadding5
	return layout.Inset{Left: m, Right: m, Top: mtb, Bottom: mtb}.Layout(gtx, body)
}

func (pg *acctDetailsPage) handle()  {}
func (pg *acctDetailsPage) onClose() {}
