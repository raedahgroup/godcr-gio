// util contains functions that don't contain layout code. They could be considered helpers that aren't particularly
// bounded to a page.

package ui

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gioui.org/gesture"
	"gioui.org/widget"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/wallet"
	"golang.org/x/text/message"
)

func mustIcon(ic *widget.Icon, err error) *widget.Icon {
	if err != nil {
		panic(err)
	}
	return ic
}

// getLockWallet returns a list of locked wallets
func getLockedWallets(wallets []*dcrlibwallet.Wallet) []*dcrlibwallet.Wallet {
	var walletsLocked []*dcrlibwallet.Wallet
	for _, wl := range wallets {
		if !wl.HasDiscoveredAccounts && wl.IsLocked() {
			walletsLocked = append(walletsLocked, wl)
		}
	}

	return walletsLocked
}

func formatDateOrTime(timestamp int64) string {
	utcTime := time.Unix(timestamp, 0).UTC()
	if time.Now().UTC().Sub(utcTime).Hours() < 168 {
		return utcTime.Weekday().String()
	}

	t := strings.Split(utcTime.Format(time.UnixDate), " ")
	t2 := t[2]
	if t[2] == "" {
		t2 = t[3]
	}
	return fmt.Sprintf("%s %s", t[1], t2)
}

// createClickGestures returns a slice of click gestures
func createClickGestures(count int) []*gesture.Click {
	var gestures = make([]*gesture.Click, count)
	for i := 0; i < count; i++ {
		gestures[i] = &gesture.Click{}
	}
	return gestures
}

// showBadge loops through a slice of recent transactions and checks if there are transaction from different wallets.
// It returns true if transactions from different wallets exists and false if they don't
func showLabel(recentTransactions []wallet.Transaction) bool {
	var name string
	for _, t := range recentTransactions {
		if name != "" && name != t.WalletName {
			return true
		}
		name = t.WalletName
	}
	return false
}

// breakBalance takes the balance string and returns it in two slices
func breakBalance(p *message.Printer, balance string) (b1, b2 string) {
	var isDecimal = true
	balanceParts := strings.Split(balance, ".")
	if len(balanceParts) == 1 {
		isDecimal = false
		balanceParts = strings.Split(balance, " ")
	}

	b1 = balanceParts[0]
	if bal, err := strconv.Atoi(b1); err == nil {
		b1 = p.Sprint(bal)
	}

	b2 = balanceParts[1]
	if isDecimal {
		b1 = b1 + "." + b2[:2]
		b2 = b2[2:]
		return
	}
	b2 = " " + b2
	return
}

func formatUSDBalance(p *message.Printer, balance float64) string {
	return p.Sprintf("$%.2f", balance)
}

func dcrTOUSD(exchangeRate, dcr float64) float64 {
	return dcr * exchangeRate
}

func usdToDCR(exchangeRate, usd float64) float64 {
	return usd / exchangeRate
}

func goToURL(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Error(err)
	}
}

func computePasswordStrength(pb *decredmaterial.ProgressBarStyle, th *decredmaterial.Theme, editors ...*widget.Editor) {
	password := editors[0]
	strength := dcrlibwallet.ShannonEntropy(password.Text()) / 4.0
	pb.Progress = float32(strength * 100)
	pb.Color = th.Color.Success
}

func ticketStatusIcon(c *pageCommon, ticketStatus string) *struct {
	icon       *widget.Image
	color      color.NRGBA
	background color.NRGBA
} {
	m := map[string]struct {
		icon       *widget.Image
		color      color.NRGBA
		background color.NRGBA
	}{
		"UNMINED": {
			c.icons.ticketUnminedIcon,
			c.theme.Color.DeepBlue,
			c.theme.Color.LightBlue,
		},
		"IMMATURE": {
			c.icons.ticketImmatureIcon,
			c.theme.Color.DeepBlue,
			c.theme.Color.LightBlue,
		},
		"LIVE": {
			c.icons.ticketLiveIcon,
			c.theme.Color.Primary,
			c.theme.Color.LightBlue,
		},
		"VOTED": {
			c.icons.ticketVotedIcon,
			c.theme.Color.Success,
			c.theme.Color.Success2,
		},
		"MISSED": {
			c.icons.ticketMissedIcon,
			c.theme.Color.Gray,
			c.theme.Color.LightGray,
		},
		"EXPIRED": {
			c.icons.ticketExpiredIcon,
			c.theme.Color.Gray,
			c.theme.Color.LightGray,
		},
		"REVOKED": {
			c.icons.ticketRevokedIcon,
			c.theme.Color.Orange,
			c.theme.Color.Orange2,
		},
	}
	st, ok := m[ticketStatus]
	if !ok {
		return nil
	}
	return &st
}

func handleSubmitEvent(editors ...*widget.Editor) bool {
	var submit bool
	for _, editor := range editors {
		for _, e := range editor.Events() {
			if _, ok := e.(widget.SubmitEvent); ok {
				submit = true
			}
		}
	}
	return submit
}

func GetAbsolutePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("error getting executable path: %s", err.Error())
	}

	exSym, err := filepath.EvalSymlinks(ex)
	if err != nil {
		return "", fmt.Errorf("error getting filepath after evaluating sym links")
	}

	return path.Dir(exSym), nil
}
