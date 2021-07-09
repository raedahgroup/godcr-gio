package ui

import (
	"encoding/json"
	"image"
	"image/color"
	"net/http"
	"sort"

	"golang.org/x/exp/shiny/materialdesign/icons"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/load"
	"github.com/planetdecred/godcr/ui/page"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"
)

type (
	Page  = load.Page
	Modal = load.Modal
)

type pageIcons struct {
	contentAdd, navigationCheck, navigationMore, actionCheckCircle, actionInfo, navigationArrowBack,
	navigationArrowForward, actionCheck, chevronRight, navigationCancel, navMoreIcon,
	imageBrightness1, contentClear, dropDownIcon, cached, contentRemove *widget.Icon

	overviewIcon, overviewIconInactive, walletIcon, walletIconInactive,
	receiveIcon, transactionIcon, transactionIconInactive, sendIcon, moreIcon, moreIconInactive,
	pendingIcon, logo, redirectIcon, confirmIcon, newWalletIcon, walletAlertIcon,
	importedAccountIcon, accountIcon, editIcon, expandIcon, copyIcon, mixer, mixerSmall,
	arrowForwardIcon, transactionFingerPrintIcon, settingsIcon, securityIcon, helpIcon,
	aboutIcon, debugIcon, verifyMessageIcon, locationPinIcon, alertGray, arrowDownIcon,
	watchOnlyWalletIcon, currencySwapIcon, syncingIcon, proposalIconActive, proposalIconInactive,
	restore, documentationIcon, downloadIcon, timerIcon, ticketIcon, ticketIconInactive, stakeyIcon,
	list, listGridIcon, decredSymbolIcon *widget.Image

	ticketPurchasedIcon,
	ticketImmatureIcon,
	ticketLiveIcon,
	ticketVotedIcon,
	ticketMissedIcon,
	ticketExpiredIcon,
	ticketRevokedIcon,
	ticketUnminedIcon *widget.Image
}

type navHandler struct {
	clickable     *widget.Clickable
	image         *widget.Image
	imageInactive *widget.Image
	page          string
}

type walletAccount struct {
	evt          *gesture.Click
	walletIndex  int
	accountIndex int
	accountName  string
	totalBalance string
	spendable    string
	number       int32
}

type DCRUSDTBittrex struct {
	LastTradeRate string
}

type pageCommon struct {
	printer             *message.Printer
	multiWallet         *dcrlibwallet.MultiWallet
	network             string
	notificationsUpdate chan interface{}
	wallet              *wallet.Wallet
	walletAccount       **wallet.Account
	info                *wallet.MultiWalletInfo
	selectedWallet      *int
	selectedAccount     *int
	theme               *decredmaterial.Theme
	icons               pageIcons
	page                *string
	returnPage          *string
	dcrUsdtBittrex      DCRUSDTBittrex
	navTab              *decredmaterial.Tabs
	keyEvents           chan *key.Event
	toast               **toast
	states              *states
	internalLog         *chan string
	walletSyncStatus    *wallet.SyncStatus
	walletTransactions  **wallet.Transactions
	acctMixerStatus     *chan *wallet.AccountMixer
	txAuthor            *dcrlibwallet.TxAuthor
	broadcastResult     *wallet.Broadcast
	signatureResult     **wallet.Signature
	walletTickets       **wallet.Tickets
	vspInfo             **wallet.VSP
	unspentOutputs      **wallet.UnspentOutputs
	showModal           func(Modal)
	dismissModal        func(Modal)
	toggleSync          func()

	testButton decredmaterial.Button

	selectedUTXO map[int]map[int32]map[string]*wallet.UnspentOutput

	refreshWindow    func()
	changeWindowPage func(Page, bool)
	popWindowPage    func() bool
	changePage       func(string)
	setReturnPage    func(string)
	changeFragment   func(Page, string)
}

type (
	C = layout.Context
	D = layout.Dimensions
)

func (win *Window) newPageCommon(decredIcons map[string]image.Image) *pageCommon {
	ic := pageIcons{
		contentAdd:             mustIcon(widget.NewIcon(icons.ContentAdd)),
		navigationCheck:        mustIcon(widget.NewIcon(icons.NavigationCheck)),
		navigationMore:         mustIcon(widget.NewIcon(icons.NavigationMoreHoriz)),
		actionCheckCircle:      mustIcon(widget.NewIcon(icons.ActionCheckCircle)),
		navigationArrowBack:    mustIcon(widget.NewIcon(icons.NavigationArrowBack)),
		navigationArrowForward: mustIcon(widget.NewIcon(icons.NavigationArrowForward)),
		actionInfo:             mustIcon(widget.NewIcon(icons.ActionInfo)),
		actionCheck:            mustIcon(widget.NewIcon(icons.ActionCheckCircle)),
		navigationCancel:       mustIcon(widget.NewIcon(icons.NavigationCancel)),
		imageBrightness1:       mustIcon(widget.NewIcon(icons.ImageBrightness1)),
		chevronRight:           mustIcon(widget.NewIcon(icons.NavigationChevronRight)),
		contentClear:           mustIcon(widget.NewIcon(icons.ContentClear)),
		navMoreIcon:            mustIcon(widget.NewIcon(icons.NavigationMoreHoriz)),
		dropDownIcon:           mustIcon(widget.NewIcon(icons.NavigationArrowDropDown)),
		cached:                 mustIcon(widget.NewIcon(icons.ActionCached)),
		contentRemove:          mustIcon(widget.NewIcon(icons.ContentRemove)),

		overviewIcon:               &widget.Image{Src: paint.NewImageOp(decredIcons["overview"])},
		overviewIconInactive:       &widget.Image{Src: paint.NewImageOp(decredIcons["overview_inactive"])},
		walletIconInactive:         &widget.Image{Src: paint.NewImageOp(decredIcons["wallet_inactive"])},
		receiveIcon:                &widget.Image{Src: paint.NewImageOp(decredIcons["receive"])},
		transactionIcon:            &widget.Image{Src: paint.NewImageOp(decredIcons["transaction"])},
		transactionIconInactive:    &widget.Image{Src: paint.NewImageOp(decredIcons["transaction_inactive"])},
		sendIcon:                   &widget.Image{Src: paint.NewImageOp(decredIcons["send"])},
		moreIcon:                   &widget.Image{Src: paint.NewImageOp(decredIcons["more"])},
		moreIconInactive:           &widget.Image{Src: paint.NewImageOp(decredIcons["more_inactive"])},
		logo:                       &widget.Image{Src: paint.NewImageOp(decredIcons["logo"])},
		confirmIcon:                &widget.Image{Src: paint.NewImageOp(decredIcons["confirmed"])},
		pendingIcon:                &widget.Image{Src: paint.NewImageOp(decredIcons["pending"])},
		redirectIcon:               &widget.Image{Src: paint.NewImageOp(decredIcons["redirect"])},
		newWalletIcon:              &widget.Image{Src: paint.NewImageOp(decredIcons["addNewWallet"])},
		walletAlertIcon:            &widget.Image{Src: paint.NewImageOp(decredIcons["walletAlert"])},
		accountIcon:                &widget.Image{Src: paint.NewImageOp(decredIcons["account"])},
		importedAccountIcon:        &widget.Image{Src: paint.NewImageOp(decredIcons["imported_account"])},
		editIcon:                   &widget.Image{Src: paint.NewImageOp(decredIcons["editIcon"])},
		expandIcon:                 &widget.Image{Src: paint.NewImageOp(decredIcons["expand_icon"])},
		copyIcon:                   &widget.Image{Src: paint.NewImageOp(decredIcons["copy_icon"])},
		mixer:                      &widget.Image{Src: paint.NewImageOp(decredIcons["mixer"])},
		mixerSmall:                 &widget.Image{Src: paint.NewImageOp(decredIcons["mixer_small"])},
		transactionFingerPrintIcon: &widget.Image{Src: paint.NewImageOp(decredIcons["transaction_fingerprint"])},
		arrowForwardIcon:           &widget.Image{Src: paint.NewImageOp(decredIcons["arrow_forward"])},
		settingsIcon:               &widget.Image{Src: paint.NewImageOp(decredIcons["settings"])},
		securityIcon:               &widget.Image{Src: paint.NewImageOp(decredIcons["security"])},
		helpIcon:                   &widget.Image{Src: paint.NewImageOp(decredIcons["help_icon"])},
		aboutIcon:                  &widget.Image{Src: paint.NewImageOp(decredIcons["about_icon"])},
		debugIcon:                  &widget.Image{Src: paint.NewImageOp(decredIcons["debug"])},
		verifyMessageIcon:          &widget.Image{Src: paint.NewImageOp(decredIcons["verify_message"])},
		locationPinIcon:            &widget.Image{Src: paint.NewImageOp(decredIcons["location_pin"])},
		alertGray:                  &widget.Image{Src: paint.NewImageOp(decredIcons["alert_gray"])},
		arrowDownIcon:              &widget.Image{Src: paint.NewImageOp(decredIcons["arrow_down"])},
		watchOnlyWalletIcon:        &widget.Image{Src: paint.NewImageOp(decredIcons["watch_only_wallet"])},
		currencySwapIcon:           &widget.Image{Src: paint.NewImageOp(decredIcons["swap"])},
		syncingIcon:                &widget.Image{Src: paint.NewImageOp(decredIcons["syncing"])},
		documentationIcon:          &widget.Image{Src: paint.NewImageOp(decredIcons["documentation"])},
		proposalIconActive:         &widget.Image{Src: paint.NewImageOp(decredIcons["politeiaActive"])},
		proposalIconInactive:       &widget.Image{Src: paint.NewImageOp(decredIcons["politeiaInactive"])},
		restore:                    &widget.Image{Src: paint.NewImageOp(decredIcons["restore"])},
		downloadIcon:               &widget.Image{Src: paint.NewImageOp(decredIcons["downloadIcon"])},
		timerIcon:                  &widget.Image{Src: paint.NewImageOp(decredIcons["timerIcon"])},
		walletIcon:                 &widget.Image{Src: paint.NewImageOp(decredIcons["wallet"])},
		ticketIcon:                 &widget.Image{Src: paint.NewImageOp(decredIcons["ticket"])},
		ticketIconInactive:         &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_inactive"])},
		stakeyIcon:                 &widget.Image{Src: paint.NewImageOp(decredIcons["stakey"])},
		ticketPurchasedIcon:        &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_purchased"])},
		ticketImmatureIcon:         &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_immature"])},
		ticketUnminedIcon:          &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_unmined"])},
		ticketLiveIcon:             &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_live"])},
		ticketVotedIcon:            &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_voted"])},
		ticketMissedIcon:           &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_missed"])},
		ticketExpiredIcon:          &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_expired"])},
		ticketRevokedIcon:          &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_revoked"])},
		list:                       &widget.Image{Src: paint.NewImageOp(decredIcons["list"])},
		listGridIcon:               &widget.Image{Src: paint.NewImageOp(decredIcons["list_grid"])},
		decredSymbolIcon:           &widget.Image{Src: paint.NewImageOp(decredIcons["decred_symbol"])},
	}

	common := &pageCommon{
		printer:             message.NewPrinter(language.English),
		multiWallet:         win.wallet.GetMultiWallet(),
		network:             win.wallet.Net,
		notificationsUpdate: make(chan interface{}, 10),
		wallet:              win.wallet,
		walletAccount:       &win.walletAccount,
		info:                win.walletInfo,
		selectedWallet:      &win.selected,
		selectedAccount:     &win.selectedAccount,
		theme:               win.theme,
		keyEvents:           win.keyEvents,
		states:              &win.states,
		icons:               ic,
		walletSyncStatus:    win.walletSyncStatus,
		walletTransactions:  &win.walletTransactions,
		// walletTransaction:  &win.walletTransaction,
		acctMixerStatus:  &win.walletAcctMixerStatus,
		txAuthor:         &win.txAuthor,
		broadcastResult:  &win.broadcastResult,
		signatureResult:  &win.signatureResult,
		walletTickets:    &win.walletTickets,
		vspInfo:          &win.vspInfo,
		unspentOutputs:   &win.walletUnspentOutputs,
		showModal:        win.showModal,
		dismissModal:     win.dismissModal,
		changeWindowPage: win.changePage,
		popWindowPage:    win.popPage,
		refreshWindow:    win.refreshWindow,

		selectedUTXO: make(map[int]map[int32]map[string]*wallet.UnspentOutput),
		toast:        &win.toast,
		internalLog:  &win.internalLog,
	}

	if common.fetchExchangeValue(&common.dcrUsdtBittrex) != nil {
		log.Info("Error fetching exchange value")
	}

	return common
}

func loadPages(common *pageCommon, l *load.Load) map[string]Page {

	common.testButton = common.theme.Button(new(widget.Clickable), "test button")

	pages := make(map[string]Page)

	pages[page.MorePageID] = page.NewMorePage(l)
	pages[page.ReceivePageID] = page.NewReceivePage(l)
	pages[PageSend] = SendPage(common)
	pages[page.VerifyMessagePageID] = page.NewVerifyMessagePage(l)
	pages[page.SeedBackupPageID] = page.NewBackupPage(l)
	pages[page.SettingsPageID] = page.NewSettingsPage(l)
	pages[page.SecurityToolsPageID] = page.NewSecurityToolsPage(l)
	pages[page.DebugPageID] = page.NewDebugPage(l)
	pages[page.LogPageID] = page.NewLogPage(l)
	pages[page.StatisticsPageID] = page.NewStatPage(l)
	pages[page.AboutPageID] = page.NewAboutPage(l)
	pages[page.HelpPageID] = page.NewHelpPage(l)
	pages[PageUTXO] = UTXOPage(common)
	pages[PageTickets] = TicketPage(common)
	pages[page.ValidateAddressPageID] = page.NewValidateAddressPage(l)
	pages[PageTicketsList] = TicketPageList(common)
	pages[PageTicketsActivity] = TicketActivityPage(common)

	return pages
}

func (common *pageCommon) refreshTheme() {
	isDarkModeOn := common.wallet.ReadBoolConfigValueForKey("isDarkModeOn")
	if isDarkModeOn != common.theme.DarkMode {
		common.theme.SwitchDarkMode(isDarkModeOn)
	}
}

func (common *pageCommon) fetchExchangeValue(target interface{}) error {
	url := "https://api.bittrex.com/v3/markets/DCR-USDT/ticker"
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(target)
	if err != nil {
		log.Error(err)
	}

	return nil
}

func (common *pageCommon) notify(text string, success bool) {
	*common.toast = &toast{
		text:    text,
		success: success,
	}
}

func (common *pageCommon) sortedWalletList() []*dcrlibwallet.Wallet {
	wallets := common.multiWallet.AllWallets()

	sort.Slice(wallets, func(i, j int) bool {
		return wallets[i].ID < wallets[j].ID
	})

	return wallets
}

func (common *pageCommon) HDPrefix() string {
	switch common.network {
	case "testnet3": // should use a constant
		return dcrlibwallet.TestnetHDPath
	case "mainnet":
		return dcrlibwallet.MainnetHDPath
	default:
		return ""
	}
}

// Container is simply a wrapper for the Inset type. Its purpose is to differentiate the use of an inset as a padding or
// margin, making it easier to visualize the structure of a layout when reading UI code.
type Container struct {
	padding layout.Inset
}

func (c Container) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	return c.padding.Layout(gtx, w)
}

var (
	MaxWidth = unit.Dp(800)
)

func (common *pageCommon) UniformPadding(gtx layout.Context, body layout.Widget) layout.Dimensions {
	width := gtx.Constraints.Max.X

	padding := values.MarginPadding24

	if (width - 2*gtx.Px(padding)) > gtx.Px(MaxWidth) {
		paddingValue := float32(width-gtx.Px(MaxWidth)) / 2
		padding = unit.Px(paddingValue)
	}

	return layout.Inset{
		Top:    values.MarginPadding24,
		Right:  padding,
		Bottom: values.MarginPadding24,
		Left:   padding,
	}.Layout(gtx, body)
}

func (common *pageCommon) SubPageHeaderButtons() (decredmaterial.IconButton, decredmaterial.IconButton) {
	backButton := common.theme.PlainIconButton(new(widget.Clickable), common.icons.navigationArrowBack)
	infoButton := common.theme.PlainIconButton(new(widget.Clickable), common.icons.actionInfo)

	zeroInset := layout.UniformInset(values.MarginPadding0)
	backButton.Color, infoButton.Color = common.theme.Color.Gray3, common.theme.Color.Gray3

	m25 := values.MarginPadding25
	backButton.Size, infoButton.Size = m25, m25
	backButton.Inset, infoButton.Inset = zeroInset, zeroInset

	return backButton, infoButton
}

type SubPage struct {
	title        string
	subTitle     string
	walletName   string
	back         func()
	body         layout.Widget
	infoTemplate string
	extraItem    *widget.Clickable
	extra        layout.Widget
	extraText    string
	handleExtra  func()

	backButton decredmaterial.IconButton
	infoButton decredmaterial.IconButton
}

func (common *pageCommon) SubPageLayout(gtx layout.Context, sp SubPage) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: values.MarginPadding15}.Layout(gtx, func(gtx C) D {
				return common.subpageHeader(gtx, sp)
			})
		}),
		layout.Rigid(sp.body),
	)
}

func (common *pageCommon) subpageHeader(gtx layout.Context, sp SubPage) layout.Dimensions {
	common.subpageEventHandler(sp)

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Right: values.MarginPadding20}.Layout(gtx, sp.backButton.Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if sp.subTitle == "" {
				return common.theme.H6(sp.title).Layout(gtx)
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(common.theme.H6(sp.title).Layout),
				layout.Rigid(common.theme.Body1(sp.subTitle).Layout),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if sp.walletName != "" {
				return layout.Inset{Left: values.MarginPadding5, Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
					return decredmaterial.Card{
						Color: common.theme.Color.Surface,
					}.Layout(gtx, func(gtx C) D {
						return layout.UniformInset(values.MarginPadding2).Layout(gtx, func(gtx C) D {
							walletText := common.theme.Caption(sp.walletName)
							walletText.Color = common.theme.Color.Gray
							return walletText.Layout(gtx)
						})
					})
				})
			}
			return layout.Dimensions{}
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.E.Layout(gtx, func(gtx C) D {
				if sp.infoTemplate != "" {
					return sp.infoButton.Layout(gtx)
				} else if sp.extraItem != nil {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if sp.extraText != "" {
								return layout.Inset{Right: values.MarginPadding10, Top: values.MarginPadding5}.Layout(gtx, func(gtx C) D {
									text := common.theme.Caption(sp.extraText)
									text.Color = common.theme.Color.DeepBlue
									return text.Layout(gtx)
								})
							}
							return layout.Dimensions{}
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return decredmaterial.Clickable(gtx, sp.extraItem, sp.extra)
						}),
					)
				}
				return layout.Dimensions{}
			})
		}),
	)
}

func (common *pageCommon) SubpageSplitLayout(gtx layout.Context, sp SubPage) layout.Dimensions {
	card := common.theme.Card()
	card.Color = color.NRGBA{}
	return card.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D { return common.subpageHeader(gtx, sp) }),
			layout.Rigid(sp.body),
		)
	})
}

func (common *pageCommon) subpageEventHandler(sp SubPage) {
	if sp.infoTemplate != "" {
		if sp.infoButton.Button.Clicked() {
			newInfoModal(common).
				title(sp.title).
				setupWithTemplate(sp.infoTemplate).
				negativeButton("Got it", func() {}).Show()
		}
	}

	if sp.backButton.Button.Clicked() {
		sp.back()
	}

	if sp.extraItem != nil && sp.extraItem.Clicked() {
		sp.handleExtra()
	}
}
