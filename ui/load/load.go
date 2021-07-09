// The load package contains data structures that are shared by components in the ui package. It is not a dumping ground
// for code you feel might be shared with other components in the future. Before adding code here, ask yourself, can
// the code be isolated in the package you're calling it from? Is it really needed by other packages in the ui package?
// or you're just planning for a use case that might never used.

// todo: fix toast notifications
// todo: clean up NewLoad method

package load

import (
	"image"
	"time"

	"golang.org/x/text/language"

	"gioui.org/io/key"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/wallet"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"golang.org/x/text/message"
)

type DCRUSDTBittrex struct {
	LastTradeRate string
}

type Toast struct {
	text    string
	success bool
	Timer   *time.Timer
}

type Receiver struct {
	InternalLog         chan string
	NotificationsUpdate chan interface{}
	KeyEvents           chan *key.Event
	AcctMixerStatus     chan *wallet.AccountMixer
	SyncedProposal      chan *wallet.Proposal
}

type Icons struct {
	ContentAdd, NavigationCheck, NavigationMore, ActionCheckCircle, ActionInfo, NavigationArrowBack,
	NavigationArrowForward, ActionCheck, ChevronRight, NavigationCancel, NavMoreIcon,
	ImageBrightness1, ContentClear, DropDownIcon, Cached *widget.Icon

	OverviewIcon, OverviewIconInactive, WalletIcon, WalletIconInactive,
	ReceiveIcon, TransactionIcon, TransactionIconInactive, SendIcon, MoreIcon, MoreIconInactive,
	PendingIcon, Logo, RedirectIcon, ConfirmIcon, NewWalletIcon, WalletAlertIcon,
	ImportedAccountIcon, AccountIcon, EditIcon, expandIcon, CopyIcon, mixer, mixerSmall,
	arrowForwardIcon, transactionFingerPrintIcon, SettingsIcon, SecurityIcon, HelpIcon,
	AboutIcon, DebugIcon, VerifyMessageIcon, LocationPinIcon, AlertGray, arrowDownIcon,
	WatchOnlyWalletIcon, currencySwapIcon, SyncingIcon, ProposalIconActive, ProposalIconInactive,
	restore, DocumentationIcon, downloadIcon, timerIcon, TicketIcon, TicketIconInactive, stakeyIcon,
	list, listGridIcon, decredSymbolIcon *widget.Image

	ticketPurchasedIcon,
	TicketImmatureIcon,
	TicketLiveIcon,
	TicketVotedIcon,
	TicketMissedIcon,
	TicketExpiredIcon,
	TicketRevokedIcon,
	TicketUnminedIcon *widget.Image
}

type Load struct {
	Theme *decredmaterial.Theme

	WL       *WalletLoad
	Receiver *Receiver
	Printer  *message.Printer
	Network  string

	Icons          Icons
	Page           *string
	ReturnPage     *string
	DcrUsdtBittrex *DCRUSDTBittrex

	Toast *Toast

	SelectedWallet  *int
	SelectedAccount *int
	SelectedUTXO    map[int]map[int32]map[string]*wallet.UnspentOutput

	ToggleSync       func()
	RefreshWindow    func()
	ShowModal        func(Modal)
	DismissModal     func(Modal)
	ChangeWindowPage func(page Page, keepBackStack bool)
	PopWindowPage    func() bool
	ChangeFragment   func(page Page, id string)
	ChangePage       func(string)
	SetReturnPage    func(string)
}

func NewLoad(th *decredmaterial.Theme, decredIcons map[string]image.Image) *Load {
	ic := Icons{
		ContentAdd:             mustIcon(widget.NewIcon(icons.ContentAdd)),
		NavigationCheck:        mustIcon(widget.NewIcon(icons.NavigationCheck)),
		NavigationMore:         mustIcon(widget.NewIcon(icons.NavigationMoreHoriz)),
		ActionCheckCircle:      mustIcon(widget.NewIcon(icons.ActionCheckCircle)),
		NavigationArrowBack:    mustIcon(widget.NewIcon(icons.NavigationArrowBack)),
		NavigationArrowForward: mustIcon(widget.NewIcon(icons.NavigationArrowForward)),
		ActionInfo:             mustIcon(widget.NewIcon(icons.ActionInfo)),
		ActionCheck:            mustIcon(widget.NewIcon(icons.ActionCheckCircle)),
		NavigationCancel:       mustIcon(widget.NewIcon(icons.NavigationCancel)),
		ImageBrightness1:       mustIcon(widget.NewIcon(icons.ImageBrightness1)),
		ChevronRight:           mustIcon(widget.NewIcon(icons.NavigationChevronRight)),
		ContentClear:           mustIcon(widget.NewIcon(icons.ContentClear)),
		NavMoreIcon:            mustIcon(widget.NewIcon(icons.NavigationMoreHoriz)),
		DropDownIcon:           mustIcon(widget.NewIcon(icons.NavigationArrowDropDown)),
		Cached:                 mustIcon(widget.NewIcon(icons.ActionCached)),

		OverviewIcon:               &widget.Image{Src: paint.NewImageOp(decredIcons["overview"])},
		OverviewIconInactive:       &widget.Image{Src: paint.NewImageOp(decredIcons["overview_inactive"])},
		WalletIconInactive:         &widget.Image{Src: paint.NewImageOp(decredIcons["wallet_inactive"])},
		ReceiveIcon:                &widget.Image{Src: paint.NewImageOp(decredIcons["receive"])},
		TransactionIcon:            &widget.Image{Src: paint.NewImageOp(decredIcons["transaction"])},
		TransactionIconInactive:    &widget.Image{Src: paint.NewImageOp(decredIcons["transaction_inactive"])},
		SendIcon:                   &widget.Image{Src: paint.NewImageOp(decredIcons["send"])},
		MoreIcon:                   &widget.Image{Src: paint.NewImageOp(decredIcons["more"])},
		MoreIconInactive:           &widget.Image{Src: paint.NewImageOp(decredIcons["more_inactive"])},
		Logo:                       &widget.Image{Src: paint.NewImageOp(decredIcons["logo"])},
		ConfirmIcon:                &widget.Image{Src: paint.NewImageOp(decredIcons["confirmed"])},
		PendingIcon:                &widget.Image{Src: paint.NewImageOp(decredIcons["pending"])},
		RedirectIcon:               &widget.Image{Src: paint.NewImageOp(decredIcons["redirect"])},
		NewWalletIcon:              &widget.Image{Src: paint.NewImageOp(decredIcons["addNewWallet"])},
		WalletAlertIcon:            &widget.Image{Src: paint.NewImageOp(decredIcons["walletAlert"])},
		AccountIcon:                &widget.Image{Src: paint.NewImageOp(decredIcons["account"])},
		ImportedAccountIcon:        &widget.Image{Src: paint.NewImageOp(decredIcons["imported_account"])},
		EditIcon:                   &widget.Image{Src: paint.NewImageOp(decredIcons["editIcon"])},
		expandIcon:                 &widget.Image{Src: paint.NewImageOp(decredIcons["expand_icon"])},
		CopyIcon:                   &widget.Image{Src: paint.NewImageOp(decredIcons["copy_icon"])},
		mixer:                      &widget.Image{Src: paint.NewImageOp(decredIcons["mixer"])},
		mixerSmall:                 &widget.Image{Src: paint.NewImageOp(decredIcons["mixer_small"])},
		transactionFingerPrintIcon: &widget.Image{Src: paint.NewImageOp(decredIcons["transaction_fingerprint"])},
		arrowForwardIcon:           &widget.Image{Src: paint.NewImageOp(decredIcons["arrow_forward"])},
		SettingsIcon:               &widget.Image{Src: paint.NewImageOp(decredIcons["settings"])},
		SecurityIcon:               &widget.Image{Src: paint.NewImageOp(decredIcons["security"])},
		HelpIcon:                   &widget.Image{Src: paint.NewImageOp(decredIcons["help_icon"])},
		AboutIcon:                  &widget.Image{Src: paint.NewImageOp(decredIcons["about_icon"])},
		DebugIcon:                  &widget.Image{Src: paint.NewImageOp(decredIcons["debug"])},
		VerifyMessageIcon:          &widget.Image{Src: paint.NewImageOp(decredIcons["verify_message"])},
		LocationPinIcon:            &widget.Image{Src: paint.NewImageOp(decredIcons["location_pin"])},
		AlertGray:                  &widget.Image{Src: paint.NewImageOp(decredIcons["alert_gray"])},
		arrowDownIcon:              &widget.Image{Src: paint.NewImageOp(decredIcons["arrow_down"])},
		WatchOnlyWalletIcon:        &widget.Image{Src: paint.NewImageOp(decredIcons["watch_only_wallet"])},
		currencySwapIcon:           &widget.Image{Src: paint.NewImageOp(decredIcons["swap"])},
		SyncingIcon:                &widget.Image{Src: paint.NewImageOp(decredIcons["syncing"])},
		DocumentationIcon:          &widget.Image{Src: paint.NewImageOp(decredIcons["documentation"])},
		ProposalIconActive:         &widget.Image{Src: paint.NewImageOp(decredIcons["politeiaActive"])},
		ProposalIconInactive:       &widget.Image{Src: paint.NewImageOp(decredIcons["politeiaInactive"])},
		restore:                    &widget.Image{Src: paint.NewImageOp(decredIcons["restore"])},
		downloadIcon:               &widget.Image{Src: paint.NewImageOp(decredIcons["downloadIcon"])},
		timerIcon:                  &widget.Image{Src: paint.NewImageOp(decredIcons["timerIcon"])},
		WalletIcon:                 &widget.Image{Src: paint.NewImageOp(decredIcons["wallet"])},
		TicketIcon:                 &widget.Image{Src: paint.NewImageOp(decredIcons["ticket"])},
		TicketIconInactive:         &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_inactive"])},
		stakeyIcon:                 &widget.Image{Src: paint.NewImageOp(decredIcons["stakey"])},
		ticketPurchasedIcon:        &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_purchased"])},
		TicketImmatureIcon:         &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_immature"])},
		TicketUnminedIcon:          &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_unmined"])},
		TicketLiveIcon:             &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_live"])},
		TicketVotedIcon:            &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_voted"])},
		TicketMissedIcon:           &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_missed"])},
		TicketExpiredIcon:          &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_expired"])},
		TicketRevokedIcon:          &widget.Image{Src: paint.NewImageOp(decredIcons["ticket_revoked"])},
		list:                       &widget.Image{Src: paint.NewImageOp(decredIcons["list"])},
		listGridIcon:               &widget.Image{Src: paint.NewImageOp(decredIcons["list_grid"])},
		decredSymbolIcon:           &widget.Image{Src: paint.NewImageOp(decredIcons["decred_symbol"])},
	}

	wl := &WalletLoad{
		Wallet:         new(wallet.Wallet),
		Account:        new(wallet.Account),
		Info:           new(wallet.MultiWalletInfo),
		SyncStatus:     new(wallet.SyncStatus),
		Transactions:   new(wallet.Transactions),
		UnspentOutputs: new(wallet.UnspentOutputs),
		Tickets:        new(wallet.Tickets),
		VspInfo:        new(wallet.VSP),
		Proposals:      new(wallet.Proposals),

		SelectedProposal: new(dcrlibwallet.Proposal),
	}

	r := &Receiver{
		AcctMixerStatus: make(chan *wallet.AccountMixer),
		SyncedProposal:  make(chan *wallet.Proposal),
	}

	l := &Load{
		Theme:    th,
		Icons:    ic,
		WL:       wl,
		Receiver: r,
		Toast:    &Toast{},

		Printer: message.NewPrinter(language.English),
	}
	fetchExchangeValue(l.DcrUsdtBittrex)

	return l
}

func (l *Load) RefreshTheme() {
	isDarkModeOn := l.WL.Wallet.ReadBoolConfigValueForKey("isDarkModeOn")
	if isDarkModeOn != l.Theme.DarkMode {
		l.Theme.SwitchDarkMode(isDarkModeOn)
	}
}

func (l *Load) CreateToast(text string, success bool) {
	l.Toast = &Toast{
		text:    text,
		success: success,
	}
}

func (l *Load) DestroyToast() {
	l.Toast = nil
}
