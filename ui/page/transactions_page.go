package page

import (
	"image"
	"time"

	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"
)

const Transactions = "Transactions"

type transactionWdg struct {
	statusIcon           *widget.Image
	direction            *widget.Image
	time, status, wallet decredmaterial.Label
}

type TransactionsPage struct {
	*load.Load
	pageClosing  chan bool
	container    layout.Flex
	txsList      layout.List
	toTxnDetails []*gesture.Click
	separator    decredmaterial.Line
	theme        *decredmaterial.Theme

	orderDropDown  *decredmaterial.DropDown
	txTypeDropDown *decredmaterial.DropDown
	walletDropDown *decredmaterial.DropDown

	transactions []dcrlibwallet.Transaction
	wallets      []*dcrlibwallet.Wallet
}

func NewTransactionsPage(l *load.Load) *TransactionsPage {
	pg := &TransactionsPage{
		Load:        l,
		pageClosing: make(chan bool, 1),
		container:   layout.Flex{Axis: layout.Vertical},
		txsList:     layout.List{Axis: layout.Vertical},
		separator:   l.Theme.Separator(),
		theme:       l.Theme,
	}

	pg.orderDropDown = createOrderDropDown(l)
	pg.txTypeDropDown = l.Theme.DropDown([]decredmaterial.DropDownItem{
		{
			Text: values.String(values.StrAll),
		},
		{
			Text: values.String(values.StrSent),
		},
		{
			Text: values.String(values.StrReceived),
		},
		{
			Text: values.String(values.StrYourself),
		},
		{
			Text: values.String(values.StrStaking),
		},
	}, 1)

	return pg
}

func (pg *TransactionsPage) OnResume() {
	pg.wallets = pg.WL.SortedWalletList()
	createOrUpdateWalletDropDown(pg.Load, &pg.walletDropDown, pg.wallets)
	pg.listenForTxNotifications()
	pg.loadTransactions()
}

func (pg *TransactionsPage) loadTransactions() {
	selectedWallet := pg.wallets[pg.walletDropDown.SelectedIndex()]
	newestFirst := pg.orderDropDown.SelectedIndex() == 0

	txFilter := dcrlibwallet.TxFilterAll
	switch pg.txTypeDropDown.SelectedIndex() {
	case 1:
		txFilter = dcrlibwallet.TxFilterSent
	case 2:
		txFilter = dcrlibwallet.TxFilterReceived
	case 3:
		txFilter = dcrlibwallet.TxFilterTransferred
	case 4:
		txFilter = dcrlibwallet.TxFilterStaking
	}

	wallTxs, err := selectedWallet.GetTransactionsRaw(0, 0, txFilter, newestFirst) //TODO
	if err != nil {
		log.Error("Error loading transactions:", err)
	} else {
		pg.transactions = wallTxs
	}
}

func (pg *TransactionsPage) Layout(gtx layout.Context) layout.Dimensions {
	container := func(gtx C) D {
		wallTxs := pg.transactions
		return layout.Stack{Alignment: layout.N}.Layout(gtx,
			layout.Expanded(func(gtx C) D {
				return layout.Inset{
					Top: values.MarginPadding60,
				}.Layout(gtx, func(gtx C) D {
					return pg.Theme.Card().Layout(gtx, func(gtx C) D {
						padding := values.MarginPadding16
						return Container{layout.Inset{Bottom: padding, Left: padding}}.Layout(gtx,
							func(gtx C) D {
								// return "No transactions yet" text if there are no transactions
								if len(wallTxs) == 0 {
									gtx.Constraints.Min.X = gtx.Constraints.Max.X
									txt := pg.Theme.Body1(values.String(values.StrNoTransactionsYet))
									txt.Color = pg.Theme.Color.Gray2
									return txt.Layout(gtx)
								}

								// update transaction row click gesture when the length of the click gesture slice and
								// transactions list are different.
								if len(wallTxs) != len(pg.toTxnDetails) {
									pg.toTxnDetails = createClickGestures(len(wallTxs))
								}

								return pg.txsList.Layout(gtx, len(wallTxs), func(gtx C, index int) D {
									click := pg.toTxnDetails[index]
									pointer.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Add(gtx.Ops)
									click.Add(gtx.Ops)
									pg.goToTxnDetails(click.Events(gtx), &wallTxs[index])
									var row = TransactionRow{
										transaction: wallTxs[index],
										index:       index,
										showBadge:   false,
									}
									return transactionRow(gtx, pg.Load, row)
								})
							})
					})
				})
			}),
			layout.Stacked(pg.dropDowns),
		)
	}
	return uniformPadding(gtx, container)
}

func (pg *TransactionsPage) dropDowns(gtx layout.Context) layout.Dimensions {
	return layout.Inset{
		Bottom: values.MarginPadding10,
	}.Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Rigid(pg.walletDropDown.Layout),
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Inset{
							Left: values.MarginPadding5,
						}.Layout(gtx, pg.txTypeDropDown.Layout)
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{
							Left: values.MarginPadding5,
						}.Layout(gtx, pg.orderDropDown.Layout)
					}),
				)
			}),
		)
	})
}

func (pg *TransactionsPage) Handle() {
	for pg.txTypeDropDown.Changed() {
		pg.loadTransactions()
	}

	for pg.orderDropDown.Changed() {
		pg.loadTransactions()
	}

	for pg.walletDropDown.Changed() {
		pg.loadTransactions()
	}
}

func (pg *TransactionsPage) goToTxnDetails(events []gesture.ClickEvent, txn *dcrlibwallet.Transaction) {
	for _, e := range events {
		if e.Type == gesture.TypeClick {
			pg.SetReturnPage(Transactions)
			pg.ChangeFragment(NewTransactionDetailsPage(pg.Load, txn), TransactionDetailsPageID)
		}
	}
}

func (pg *TransactionsPage) listenForTxNotifications() {
	go func() {
		for {
			var notification interface{}

			select {
			case notification = <-pg.Receiver.NotificationsUpdate:
			case <-pg.pageClosing:
				return
			}

			switch n := notification.(type) {
			case wallet.NewTransaction:
				selectedWallet := pg.wallets[pg.walletDropDown.SelectedIndex()]
				if selectedWallet.ID == n.Transaction.WalletID {
					pg.loadTransactions()
					pg.RefreshWindow()
				}
			}
		}
	}()
}

func (pg *TransactionsPage) OnClose() {
	pg.pageClosing <- true
}

func initTxnWidgets(l *load.Load, transaction *dcrlibwallet.Transaction) transactionWdg {

	var txn transactionWdg
	t := time.Unix(transaction.Timestamp, 0).UTC()
	txn.time = l.Theme.Body1(t.Format(time.UnixDate))
	txn.status = l.Theme.Body1("")
	txn.wallet = l.Theme.Body2(l.WL.MultiWallet.WalletWithID(transaction.WalletID).Name)

	if txConfirmations(l, *transaction) > 1 {
		txn.status.Text = formatDateOrTime(transaction.Timestamp)
		txn.statusIcon = l.Icons.ConfirmIcon
	} else {
		txn.status.Text = "pending"
		txn.status.Color = l.Theme.Color.Gray
		txn.statusIcon = l.Icons.PendingIcon
	}

	if transaction.Direction == dcrlibwallet.TxDirectionSent {
		txn.direction = l.Icons.SendIcon
	} else {
		txn.direction = l.Icons.ReceiveIcon
	}

	return txn
}
