package wallet

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrutil/v3"
	"github.com/planetdecred/dcrlibwallet"
)

// CreateWallet creates a new wallet with the given parameters.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) CreateWallet(name, passphrase string, errChan chan error) {
	go func() {
		var resp Response
		wall, err := wal.multi.CreateNewWallet(name, passphrase, dcrlibwallet.PassphraseTypePass)
		sendErr := func(err error) {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(MultiWalletError{
				Message: "Could not create wallet",
				Err:     err,
			})
		}
		if err != nil {
			sendErr(err)
			return
		}
		seeds, err := wall.DecryptSeed([]byte(passphrase))
		if err != nil {
			sendErr(err)
			return
		}

		resp.Resp = CreatedSeed{
			Seed: seeds,
		}
		wal.Send <- resp
	}()
}

// RestoreWallet restores a wallet with the given parameters.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) RestoreWallet(seed, passphrase string, errChan chan error) {
	go func() {
		var resp Response
		_, err := wal.multi.RestoreWallet("wallet", seed, passphrase, dcrlibwallet.PassphraseTypePass)
		if err != nil {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(MultiWalletError{
				Message: "Could not restore wallet",
				Err:     err,
			})
			return
		}
		resp.Resp = Restored{}
		wal.Send <- resp
	}()
}

// DeleteWallet deletes a wallet.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) DeleteWallet(walletID int, passphrase []byte, errChan chan error) {
	log.Debug("Deleting Wallet")
	go func() {
		var resp Response
		log.Debugf("Wallet %d: %+v", walletID, wal.multi.WalletWithID(walletID))
		err := wal.multi.DeleteWallet(walletID, passphrase)
		if err != nil {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(InternalWalletError{
				Message:  "Could not delete wallet",
				Affected: []int{walletID},
				Err:      err,
			})
			return
		}
		resp.Resp = DeletedWallet{
			ID: walletID,
		}
		wal.Send <- resp
	}()
}

// AddAccount adds an account to a wallet.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) AddAccount(walletID int, name string, pass []byte, errChan chan error, onCreate func(*dcrlibwallet.Account)) {
	go func() {
		var resp Response
		wall := wal.multi.WalletWithID(walletID)
		if wall == nil {
			go func() {
				errChan <- ErrIDNotExist
			}()
			resp.Err = ErrIDNotExist
			wal.Send <- Response{
				Resp: AddedAccount{},
				Err:  ErrIDNotExist,
			}
			return
		}

		id, err := wall.CreateNewAccount(name, pass)
		if err != nil {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(InternalWalletError{
				Message:  "Could not create account",
				Affected: []int{walletID},
				Err:      err,
			})
			return
		}

		acct, err := wall.GetAccount(id)
		if err != nil {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(InternalWalletError{
				Message:  "Could not fetch newly created account",
				Affected: []int{walletID},
				Err:      err,
			})
			return
		}
		onCreate(acct)

		resp.Resp = AddedAccount{
			ID: id,
		}
		wal.Send <- resp
	}()
}

// CreateTransaction creates a TxAuthor with the given parameters.
// The created TxAuthor will have to have a destination added before broadcasting.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) CreateTransaction(walletID int, accountID int32, errChan chan error) {
	go func() {
		var resp Response
		wall := wal.multi.WalletWithID(walletID)
		_, err := wall.GetAccount(accountID)
		if err != nil {
			errChan <- err
			return
		}

		txAuthor := wal.multi.NewUnsignedTx(wall, accountID)
		if txAuthor == nil {
			errChan <- err
			return
		}
		resp.Resp = txAuthor
		wal.Send <- resp
	}()
}

// transactionStatus accepts the bestBlockHeight, transactionBlockHeight returns a transaction status
// which could be confirmed/pending and confirmations count
func transactionStatus(bestBlockHeight, txnBlockHeight int32) (string, int32) {
	confirmations := bestBlockHeight - txnBlockHeight + 1
	if txnBlockHeight != -1 && confirmations > dcrlibwallet.DefaultRequiredConfirmations {
		return "confirmed", confirmations
	}
	return "pending", confirmations
}

// BroadcastTransaction broadcasts the transaction built with txAuthor to the network.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) BroadcastTransaction(txAuthor *dcrlibwallet.TxAuthor, passphrase []byte, errChan chan error) {
	go func() {
		var resp Response

		txHash, err := txAuthor.Broadcast(passphrase)
		if err != nil {
			errChan <- fmt.Errorf("error broadcasting transaction: %s", err.Error())
			return
		}

		hash, err := chainhash.NewHash(txHash)
		if err != nil {
			errChan <- fmt.Errorf("error parsing successful transaction hash: %s", err.Error())
			return
		}

		resp.Resp = &Broadcast{
			TxHash: hash.String(),
		}
		wal.Send <- resp
	}()
}

func (wal *Wallet) GetTransactions(offset, limit, txfilter int32, newestFirst bool) ([]dcrlibwallet.Transaction, error) {
	transactionsJSON, err := wal.multi.GetTransactions(offset, limit, txfilter, true)
	if err != nil {
		return nil, err
	}

	var transactions []dcrlibwallet.Transaction
	err = json.Unmarshal([]byte(transactionsJSON), &transactions)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetAllTransactions collects a per-wallet slice of transactions fitting the parameters.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) GetAllTransactions(offset, limit, txfilter int32) {
	go func() {
		var resp Response
		wallets, err := wal.wallets()
		if err != nil {
			resp.Err = err
			wal.Send <- resp
			return
		}

		var recentTxs []Transaction

		transactions := make(map[int][]Transaction)
		ticketTxs := make(map[int][]Transaction)
		bestBestBlock := wal.multi.GetBestBlock()
		totalTxn := 0

		for _, wall := range wallets {
			txs, err := wall.GetTransactionsRaw(offset, limit, txfilter, true)
			if err != nil {
				resp.Err = err
				wal.Send <- resp
				return
			}
			for _, txnRaw := range txs {
				totalTxn++
				status, confirmations := transactionStatus(bestBestBlock.Height, txnRaw.BlockHeight)
				txn := Transaction{
					Txn:           txnRaw,
					Status:        status,
					Balance:       dcrutil.Amount(txnRaw.Amount).String(),
					WalletName:    wall.Name,
					Confirmations: confirmations,
					DateTime:      dcrlibwallet.ExtractDateOrTime(txnRaw.Timestamp),
				}
				recentTxs = append(recentTxs, txn)
				if txn.Txn.Type == dcrlibwallet.TxTypeTicketPurchase {
					ticketTxs[wall.ID] = append(ticketTxs[wall.ID], txn)
				}
				transactions[txnRaw.WalletID] = append(transactions[txnRaw.WalletID], txn)
			}
		}

		sort.SliceStable(recentTxs, func(i, j int) bool {
			backTime := time.Unix(recentTxs[j].Txn.Timestamp, 0)
			frontTime := time.Unix(recentTxs[i].Txn.Timestamp, 0)
			return backTime.Before(frontTime)
		})

		recentTxsLimit := 5
		if len(recentTxs) > recentTxsLimit {
			recentTxs = recentTxs[:recentTxsLimit]
		}

		resp.Resp = &Transactions{
			Total:   totalTxn,
			Txs:     transactions,
			Recent:  recentTxs,
			Tickets: ticketTxs,
		}
		wal.Send <- resp
	}()
}

// GetTransaction get transaction information by wallet ID and transaction hash
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) GetTransaction(walletID int, txnHash string) {
	go func() {
		var resp Response
		wall := wal.multi.WalletWithID(walletID)

		hash, err := chainhash.NewHashFromStr(txnHash)
		if err != nil {
			resp.Err = err
			wal.Send <- resp
			return
		}

		txn, err := wall.GetTransactionRaw(hash[:])
		if err != nil {
			resp.Err = err
			wal.Send <- resp
			return
		}
		bestBestBlock := wal.multi.GetBestBlock()
		status, confirmations := transactionStatus(bestBestBlock.Height, txn.BlockHeight)
		acct, err := wall.GetAccount(txn.Inputs[0].AccountNumber)
		if err != nil {
			resp.Err = err
			wal.Send <- resp
			return
		}
		resp.Resp = &Transaction{
			Txn:           *txn,
			Status:        status,
			Balance:       dcrutil.Amount(txn.Amount).String(),
			WalletName:    wall.Name,
			Confirmations: confirmations,
			DateTime:      dcrlibwallet.ExtractDateOrTime(txn.Timestamp),
			AccountName:   acct.Name,
		}
		wal.Send <- resp
	}()
}

// WalletSyncStatus returns the sync status of a single wallet
func walletSyncStatus(isWaiting bool, walletBestBlock, bestBlockHeight int32) string {
	if isWaiting {
		return "waiting for other wallets"
	}
	if walletBestBlock < bestBlockHeight {
		return "syncing..."
	}

	return "synced"
}

func (wal *Wallet) SignMessage(walletID int, passphrase []byte, address, message string, errChan chan error) {
	go func() {
		var resp Response

		wall := wal.multi.WalletWithID(walletID)
		if wall == nil {
			resp.Err = ErrIDNotExist
			wal.Send <- resp
			return
		}

		signedMessageBytes, err := wall.SignMessage(passphrase, address, message)
		if err != nil {
			go func() {
				errChan <- err
			}()
			wal.Send <- resp
			return
		}

		resp.Resp = &Signature{
			Signature: base64.StdEncoding.EncodeToString(signedMessageBytes),
		}

		wal.Send <- resp
	}()
}

// RenameWallet renames the wallet identified by walletID.
func (wal *Wallet) RenameWallet(walletID int, name string, errChan chan error) {
	go func() {
		var resp Response
		err := wal.multi.RenameWallet(walletID, name)
		if err != nil {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(MultiWalletError{
				Message: "Could not rename wallet",
				Err:     err,
			})
			return
		}
		resp.Resp = Renamed{
			ID: walletID,
		}
		wal.Send <- resp
	}()
}

// ImportWatchOnlyWallet imports a watch only wallet with the given parameters.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) ImportWatchOnlyWallet(name, extendedPublicKey string) error {
	var g errgroup.Group
	g.Go(func() error {
		_, err := wal.multi.CreateWatchOnlyWallet(name, extendedPublicKey)
		if err != nil {
			return fmt.Errorf("error importing watch only wallet: %s", err.Error())
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

// ChangeWalletPassphrase changes the spending passphrase of the wallet identified by walletID.
func (wal *Wallet) ChangeWalletPassphrase(walletID int, oldPrivatePassphrase, newPrivatePassphrase string, errChan chan error) {
	go func() {
		var resp Response
		err := wal.multi.ChangePrivatePassphraseForWallet(walletID, []byte(oldPrivatePassphrase), []byte(newPrivatePassphrase), dcrlibwallet.PassphraseTypePass)
		if err != nil {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(InternalWalletError{
				Message:  "Could not change password",
				Affected: []int{walletID},
				Err:      err,
			})
			return
		}

		resp.Resp = &ChangePassword{
			ID: walletID,
		}
	}()
}

func (wal *Wallet) OpenWallets(passphrase string, errChan chan error) {
	go func() {
		var resp Response
		err := wal.multi.OpenWallets([]byte(passphrase))
		if err != nil {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(MultiWalletError{
				Message: "Could not open wallets",
				Err:     err,
			})
			return
		}

		resp.Resp = OpenWallet{}
		wal.Send <- resp
	}()
}

func (wal *Wallet) SetStartupPassphrase(passphrase string, errChan chan error) {
	go func() {
		var resp Response
		err := wal.multi.SetStartupPassphrase([]byte(passphrase), dcrlibwallet.PassphraseTypePass)
		if err != nil {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(MultiWalletError{
				Message: "Could not set up startup passphrase",
				Err:     err,
			})
			return
		}
		resp.Resp = &StartupPassphrase{
			Msg: "Startup password set",
		}
		wal.Send <- resp
	}()
}

func (wal *Wallet) ChangeStartupPassphrase(oldPrivatePassphrase, newPrivatePassphrase string, errChan chan error) {
	go func() {
		var resp Response
		err := wal.multi.ChangeStartupPassphrase([]byte(oldPrivatePassphrase), []byte(newPrivatePassphrase), dcrlibwallet.PassphraseTypePass)
		if err != nil {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(MultiWalletError{
				Message: "Could not change startup passphrase",
				Err:     err,
			})
			return
		}

		resp.Resp = &StartupPassphrase{
			Msg: "Startup password changed",
		}
		wal.Send <- resp
	}()
}

func (wal *Wallet) RemoveStartupPassphrase(passphrase string, errChan chan error) {
	go func() {
		var resp Response
		err := wal.multi.RemoveStartupPassphrase([]byte(passphrase))
		if err != nil {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(MultiWalletError{
				Message: "Could not remove startup passphrase",
				Err:     err,
			})
			return
		}
		resp.Resp = &StartupPassphrase{
			Msg: "Startup password removed",
		}
		wal.Send <- resp
	}()
}

// IsStartupSecuritySet checks if start up password is set
func (wal *Wallet) IsStartupSecuritySet() bool {
	return wal.multi.IsStartupSecuritySet()
}

// RenameAccount renames the acct of wallet with id walletID.
func (wal *Wallet) RenameAccount(walletID int, acct int32, name string, errChan chan<- error) {
	go func() {
		var resp Response
		wall := wal.multi.WalletWithID(walletID)
		if wall == nil {
			go func() {
				errChan <- ErrIDNotExist
			}()
			resp.Err = ErrIDNotExist
			wal.Send <- Response{
				Resp: UpdatedAccount{},
				Err:  ErrIDNotExist,
			}
			return
		}

		err := wall.RenameAccount(acct, name)
		if err != nil {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(InternalWalletError{
				Message:  "Could not rename account",
				Affected: []int{walletID},
				Err:      err,
			})
			return
		}
		resp.Resp = UpdatedAccount{
			ID: acct,
		}
		wal.Send <- resp
	}()
}

func (wal *Wallet) GetAllProposals() ([]dcrlibwallet.Proposal, error) {
	return wal.multi.Politeia.GetProposalsRaw(dcrlibwallet.ProposalCategoryAll, 0, 0, true)
}

func (wal *Wallet) FetchProposalDescription(token string) (string, error) {
	return wal.multi.Politeia.FetchProposalDescription(token)
}

func (wal *Wallet) UnlockWallet(walletID int, passphrase []byte) error {
	return wal.multi.UnlockWallet(walletID, passphrase)
}

// CurrentAddress returns the next address for the specified wallet account.
func (wal *Wallet) CurrentAddress(walletID int, accountID int32) (string, error) {
	wall := wal.multi.WalletWithID(walletID)
	if wall == nil {
		return "", ErrIDNotExist
	}
	return wall.CurrentAddress(accountID)
}

// NextAddress returns the next address for the specified wallet account.
func (wal *Wallet) NextAddress(walletID int, accountID int32) (string, error) {
	wall := wal.multi.WalletWithID(walletID)
	if wall == nil {
		return "", ErrIDNotExist
	}
	return wall.NextAddress(accountID)
}

// IsAddressValid checks if the given address is valid for the multiwallet network
func (wal *Wallet) IsAddressValid(address string) (bool, error) {
	return wal.multi.IsAddressValid(address), nil
}

// HaveAddress checks if the given address is valid for the wallet
func (wal *Wallet) HaveAddress(walletID int, address string) (bool, error) {
	wall := wal.multi.WalletWithID(walletID)
	if wall == nil {
		return false, ErrIDNotExist
	}
	return wall.HaveAddress(address), nil
}

// VerifyMessage checks if the given message matches the signature for the address.
func (wal *Wallet) VerifyMessage(address string, message string, signature string) (bool, error) {
	return wal.multi.VerifyMessage(address, message, signature)
}

// StartSync starts the multiwallet SPV sync
func (wal *Wallet) StartSync() error {
	return wal.multi.SpvSync()
}

func (wal *Wallet) IsConnectedToDecredWallet() bool {
	return wal.multi.IsConnectedToDecredNetwork()
}

func (wal *Wallet) IsSynced() bool {
	return wal.multi.IsSynced()
}

func (wal *Wallet) IsSyncing() bool {
	return wal.multi.IsSyncing()
}

func (wal *Wallet) BestBlock() *dcrlibwallet.BlockInfo {
	return wal.multi.GetBestBlock()
}

// RescanBlocks rescans the multiwallet
func (wal *Wallet) RescanBlocks(walletID int) error {
	return wal.multi.RescanBlocks(walletID)
}

// CancelSync cancels the SPV sync
func (wal *Wallet) CancelSync() {
	go wal.multi.CancelSync()
}

func (wal *Wallet) SyncProposals() {
	go wal.multi.Politeia.Sync()
}

func (wal *Wallet) IsSyncingProposals() bool {
	return wal.multi.Politeia.IsSyncing()
}

func (wal *Wallet) GetWalletSeedPhrase(walletID int, password []byte) (string, error) {
	return wal.multi.WalletWithID(walletID).DecryptSeed(password)
}

func (wal *Wallet) VerifyWalletSeedPhrase(walletID int, seedPhrase string, privpass []byte) error {
	_, err := wal.multi.VerifySeedForWallet(walletID, seedPhrase, privpass)
	return err
}

func (wal *Wallet) SaveConfigValueForKey(key string, value interface{}) {
	wal.multi.SaveUserConfigValue(key, value)
}

func (wal *Wallet) ReadBoolConfigValueForKey(key string) bool {
	return wal.multi.ReadBoolConfigValueForKey(key, false)
}

func (wal *Wallet) ReadStringConfigValueForKey(key string) string {
	return wal.multi.ReadStringConfigValueForKey(key)
}

func (wal *Wallet) RemoveUserConfigValueForKey(key string) {
	wal.multi.DeleteUserConfigValueForKey(key)
}

func (wal *Wallet) LoadedWalletsCount() int32 {
	return wal.multi.LoadedWalletsCount()
}

func calculateDaysBehind(lastHeaderTime int64) string {
	diff := time.Since(time.Unix(lastHeaderTime, 0))
	daysBehind := int(math.Round(diff.Hours() / 24))
	if daysBehind < 1 {
		return "<1 day"
	} else if daysBehind == 1 {
		return "1 day"
	} else {
		return fmt.Sprintf("%d days", daysBehind)
	}
}

// SecondsToDays takes time in seconds and returns its string equivalent in the format ddhhmm.
func SecondsToDays(totalTimeLeft int64) string {
	q, r := divMod(totalTimeLeft, 24*60*60)
	timeLeft := time.Duration(r) * time.Second
	if q > 0 {
		return fmt.Sprintf("%dd%s", q, timeLeft.String())
	}
	return timeLeft.String()
}

// GetRawBalance gets the balance in int64, formats it and returns a string while also leaving out the "DCR" suffix
func GetRawBalance(balance int64, AmountUnit int) string {
	return strconv.FormatFloat(float64(balance)/math.Pow10(AmountUnit+8), 'f', -(AmountUnit + 8), 64)
}

// divMod divides a numerator by a denominator and returns its quotient and remainder.
func divMod(numerator, denominator int64) (quotient, remainder int64) {
	quotient = numerator / denominator // integer division, decimals are truncated
	remainder = numerator % denominator
	return
}

// AllUnspentOutputs get all unspent outputs by walletID and acct
func (wal *Wallet) AllUnspentOutputs(walletID int, acct int32) {
	go func() {
		var resp Response
		wall := wal.multi.WalletWithID(walletID)
		if wall == nil {
			resp.Err = ErrIDNotExist
			wal.Send <- Response{
				Resp: UnspentOutputs{},
				Err:  ErrIDNotExist,
			}
			return
		}
		utxos, err := wall.UnspentOutputs(acct)
		if err != nil {
			resp.Err = err
			wal.Send <- ResponseError(InternalWalletError{
				Message:  "Could not get unspent outputs",
				Affected: []int{walletID, int(acct)},
				Err:      err,
			})
			return
		}

		var list []*UnspentOutput
		for _, utxo := range utxos {
			item := UnspentOutput{
				UTXO:     *utxo,
				Amount:   dcrutil.Amount(utxo.Amount).String(),
				DateTime: dcrlibwallet.ExtractDateOrTime(utxo.ReceiveTime),
			}
			list = append(list, &item)
		}
		resp.Resp = &UnspentOutputs{
			List: list,
		}
		wal.Send <- resp
	}()
}

// IsAccountMixerConfigSet check the wallet have account mixer config set
func (wal *Wallet) IsAccountMixerConfigSet(walletID int) bool {
	wall := wal.multi.WalletWithID(walletID)
	if wall == nil {
		return false
	}
	return wall.ReadBoolConfigValueForKey(dcrlibwallet.AccountMixerConfigSet, false)
}

// SetupAccountMixer setup account mixer with the given parameters.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) SetupAccountMixer(walletID int, walletPassphrase string, errChan chan error) {
	go func() {
		var resp Response

		wall := wal.multi.WalletWithID(walletID)
		if wall == nil {
			resp.Err = ErrIDNotExist
			wal.Send <- Response{
				Resp: SetupAccountMixer{},
				Err:  ErrIDNotExist,
			}
			return
		}

		var err error
		var mixedAcctNumber int32
		var unmixedAcctNumber int32
		mixedAcct := "mixed"
		unmixedAcct := "unmixed"

		sendErr := func(err error) {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(InternalWalletError{
				Message:  "Could not set account mixer",
				Err:      err,
				Affected: []int{walletID},
			})
		}

		if !wall.HasAccount(mixedAcct) {
			mixedAcctNumber, err = wall.CreateNewAccount(mixedAcct, []byte(walletPassphrase))
			if err != nil {
				sendErr(err)
				return
			}
		} else {
			mixedAcctNumber, err = wall.AccountNumber(mixedAcct)
			if err != nil {
				sendErr(err)
				return
			}
		}

		if !wall.HasAccount(unmixedAcct) {
			unmixedAcctNumber, err = wall.CreateNewAccount(unmixedAcct, []byte(walletPassphrase))
			if err != nil {
				sendErr(err)
				return
			}
		} else {
			unmixedAcctNumber, err = wall.AccountNumber(unmixedAcct)
			if err != nil {
				sendErr(err)
				return
			}
		}

		err = wall.SetAccountMixerConfig(mixedAcctNumber, unmixedAcctNumber, walletPassphrase)
		if err != nil {
			sendErr(err)
			return
		}

		resp.Resp = SetupAccountMixer{}
		wal.Send <- resp
	}()
}

// TicketPrice get ticket price
func (wal *Wallet) TicketPrice() (int64, string) {
	pr, err := wal.multi.WalletsIterator().Next().TicketPrice()
	if err != nil {
		log.Error(err)
		return 0, ""
	}
	return pr.TicketPrice, dcrutil.Amount(pr.TicketPrice).String()
}

func (wal *Wallet) NewVSPD(host string, walletID int, accountID int32) (*dcrlibwallet.VSPD, error) {
	if host == "" {
		return nil, fmt.Errorf("Host is required")
	}
	wall := wal.multi.WalletWithID(walletID)
	if wall == nil {
		return nil, ErrIDNotExist
	}
	vspd := wal.multi.NewVSPD(host, walletID, accountID)
	if vspd == nil {
		return nil, fmt.Errorf("Something wrong when creating new VSPD")
	}
	return vspd, nil
}

// PurchaseTicket buy a ticket with given parameters
func (wal *Wallet) PurchaseTicket(walletID int, accountID int32, tickets uint32, passphrase []byte, vspd *dcrlibwallet.VSPD, errChan chan error) {
	go func() {
		var resp Response
		wall := wal.multi.WalletWithID(walletID)
		if wall == nil {
			go func() {
				errChan <- ErrIDNotExist
			}()
			return
		}

		_, err := vspd.GetInfo()
		if err != nil {
			go func() {
				errChan <- err
			}()
			return
		}

		request := &dcrlibwallet.PurchaseTicketsRequest{
			Account:               uint32(accountID),
			Passphrase:            passphrase,
			NumTickets:            tickets,
			Expiry:                uint32(wal.multi.GetBestBlock().Height) + 256,
			RequiredConfirmations: dcrlibwallet.DefaultRequiredConfirmations,
		}

		hashes, err := wall.PurchaseTickets(request, "")
		if err != nil {
			go func() {
				errChan <- err
			}()
			return
		}

		for _, hash := range hashes {
			r, err := vspd.GetVSPFeeAddress(hash, passphrase)
			if err != nil {
				go func() {
					errChan <- err
				}()
				return
			}

			transactionResponse, err := vspd.CreateTicketFeeTx(r.FeeAmount, hash, r.FeeAddress, passphrase)
			if err != nil {
				go func() {
					errChan <- err
				}()
				return
			}
			_, err = vspd.PayVSPFee(transactionResponse, hash, "", passphrase)
			if err != nil {
				go func() {
					errChan <- err
				}()
				return
			}
		}

		go func() {
			errChan <- nil
		}()

		resp.Resp = &TicketPurchase{}
		wal.Send <- resp
	}()
}

// GetAllTickets collects a per-wallet slice of tickets fitting the parameters.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) GetAllTickets() (*Tickets, error) {
	wallets, err := wal.wallets()
	if err != nil {
		return nil, err
	}

	var liveRecentTickets []Ticket
	var recentActivity []Ticket

	tickets := make(map[int][]Ticket)
	unconfirmedTickets := make(map[int][]UnconfirmedPurchase)

	stackingRecordCounter := []struct {
		Status string
		Count  int
	}{
		{"UNMINED", 0},
		{"IMMATURE", 0},
		{"LIVE", 0},
		{"VOTED", 0},
		{"MISSED", 0},
		{"EXPIRED", 0},
		{"REVOKED", 0},
	}

	liveCounter := []struct {
		Status string
		Count  int
	}{
		{"UNMINED", 0},
		{"IMMATURE", 0},
		{"LIVE", 0},
	}

	for _, wall := range wallets {
		ticketsInfo, err := wall.GetTicketsForBlockHeightRange(0, wall.GetBestBlock(), math.MaxInt32)
		if err != nil {
			return nil, err
		}

		for _, tinfo := range ticketsInfo {
			if tinfo.Status == "UNKNOWN" {
				continue
			}

			var amount dcrutil.Amount
			for _, output := range tinfo.Ticket.MyOutputs {
				amount += output.Amount
			}
			info := Ticket{
				Info:       *tinfo,
				DateTime:   time.Unix(tinfo.Ticket.Timestamp, 0).Format("Jan 2, 2006 03:04:05 PM"),
				MonthDay:   time.Unix(tinfo.Ticket.Timestamp, 0).Format("Jan 2"),
				DaysBehind: calculateDaysBehind(tinfo.Ticket.Timestamp),
				Amount:     amount.String(),
				Fee:        tinfo.Ticket.Fee.String(),
				WalletName: wall.Name,
			}
			tickets[wall.ID] = append(tickets[wall.ID], info)

			for i := range liveCounter {
				if liveCounter[i].Status == tinfo.Status {
					liveCounter[i].Count++
				}
			}

			if tinfo.Status == "UNMINED" || tinfo.Status == "IMMATURE" || tinfo.Status == "LIVE" {
				liveRecentTickets = append(liveRecentTickets, info)
			}

			recentActivity = append(recentActivity, info)

			for i := range stackingRecordCounter {
				if stackingRecordCounter[i].Status == tinfo.Status {
					stackingRecordCounter[i].Count++
				}
			}
		}

		sort.SliceStable(tickets[wall.ID], func(i, j int) bool {
			backTime := time.Unix(tickets[wall.ID][j].Info.Ticket.Timestamp, 0)
			frontTime := time.Unix(tickets[wall.ID][i].Info.Ticket.Timestamp, 0)
			return backTime.Before(frontTime)
		})

		unconfirmedTicketPurchases, err := getUnconfirmedPurchases(wall, tickets[wall.ID])
		if err != nil {
			return nil, err
		}
		unconfirmedTickets[wall.ID] = unconfirmedTicketPurchases
	}

	sort.SliceStable(liveRecentTickets, func(i, j int) bool {
		backTime := time.Unix(liveRecentTickets[j].Info.Ticket.Timestamp, 0)
		frontTime := time.Unix(liveRecentTickets[i].Info.Ticket.Timestamp, 0)
		return backTime.Before(frontTime)
	})

	recentLimit := 5
	if len(liveRecentTickets) > recentLimit {
		liveRecentTickets = liveRecentTickets[:recentLimit]
	}

	sort.SliceStable(recentActivity, func(i, j int) bool {
		backTime := time.Unix(recentActivity[j].Info.Ticket.Timestamp, 0)
		frontTime := time.Unix(recentActivity[i].Info.Ticket.Timestamp, 0)
		return backTime.Before(frontTime)
	})

	if len(recentActivity) > recentLimit {
		recentActivity = recentActivity[:recentLimit]
	}

	return &Tickets{
		Confirmed:             tickets,
		Unconfirmed:           unconfirmedTickets,
		RecentActivity:        recentActivity,
		StackingRecordCounter: stackingRecordCounter,
		LiveRecent:            liveRecentTickets,
		LiveCounter:           liveCounter,
	}, nil
}

func getUnconfirmedPurchases(wall dcrlibwallet.Wallet, tickets []Ticket) (unconfirmed []UnconfirmedPurchase, err error) {
	contains := func(slice []Ticket, item string) bool {
		set := make(map[string]struct{}, len(slice))
		for _, s := range slice {
			set[s.Info.Ticket.Hash.String()] = struct{}{}
		}

		_, ok := set[item]
		return ok
	}

	txs, err := wall.GetTransactionsRaw(0, 0, dcrlibwallet.TxFilterAll, true)
	if err != nil {
		return
	}

	ticketTxs := make(map[int][]dcrlibwallet.Transaction)
	for _, txn := range txs {
		if txn.Type == dcrlibwallet.TxTypeTicketPurchase {
			ticketTxs[wall.ID] = append(ticketTxs[wall.ID], txn)
		}
	}

	if len(tickets) == len(ticketTxs) {
		return
	}

	for _, txn := range ticketTxs[wall.ID] {
		var amount int64
		for _, output := range txn.Outputs {
			amount += output.Amount
		}

		if !contains(tickets, txn.Hash) {
			unconfirmed = append(unconfirmed, UnconfirmedPurchase{
				Hash:        txn.Hash,
				Status:      "UNCONFIRMED",
				DateTime:    dcrlibwallet.ExtractDateOrTime(txn.Timestamp),
				BlockHeight: txn.BlockHeight,
				Amount:      dcrutil.Amount(amount).String(),
			})
		}
	}

	return
}

func (wal *Wallet) StartAccountMixer(walletID int, walletPassphrase string, errChan chan error) {
	err := wal.multi.StartAccountMixer(walletID, walletPassphrase)
	if err != nil {
		go func() {
			errChan <- err
		}()
	}
}

func (wal *Wallet) StopAccountMixer(walletID int, errChan chan error) {
	err := wal.multi.StopAccountMixer(walletID)
	if err != nil {
		go func() {
			errChan <- err
		}()
	}
}

func (wal *Wallet) IsAccountMixerActive(walletID int) bool {
	wall := wal.multi.WalletWithID(walletID)
	if wall == nil {
		return false
	}
	return wall.IsAccountMixerActive()
}

func (wal *Wallet) AllWallets() []*dcrlibwallet.Wallet {
	return wal.multi.AllWallets()
}

func (wal *Wallet) WalletWithID(walletID int) *dcrlibwallet.Wallet {
	return wal.multi.WalletWithID(walletID)
}

func (wal *Wallet) ReadMixerConfigValueForKey(key string, walletID int) int32 {
	wallet := wal.multi.WalletWithID(walletID)
	if wallet != nil {
		return wallet.ReadInt32ConfigValueForKey(key, -1)
	}
	return 0
}

func (wal *Wallet) AddVSP(host string, errChan chan error) {
	// wal.multi.DeleteUserConfigValueForKey(dcrlibwallet.VSPHostConfigKey)
	go func() {
		var resp Response
		var valueOut struct {
			Remember string
			List     []string
		}

		wal.multi.ReadUserConfigValue(dcrlibwallet.VSPHostConfigKey, &valueOut)

		for _, v := range valueOut.List {
			if v == host {
				go func() {
					errChan <- fmt.Errorf("Existing host %s", host)
				}()
				return
			}
		}

		info, err := GetVSPInfo(host)
		if err != nil {
			go func() {
				errChan <- err
			}()
			resp.Err = err
			wal.Send <- ResponseError(MultiWalletError{
				Message: "Could not create vsp",
				Err:     err,
			})
			return
		}

		if info.Network != wal.Net {
			go func() {
				errChan <- fmt.Errorf("Invalid net %s", info.Network)
			}()
			return
		}

		valueOut.List = append(valueOut.List, host)
		wal.multi.SaveUserConfigValue(dcrlibwallet.VSPHostConfigKey, valueOut)
		resp.Resp = &VSPInfo{
			Host: host,
			Info: info,
		}
		wal.Send <- resp
	}()
}

func (wal *Wallet) GetAllVSP() {
	go func() {
		var valueOut struct {
			Remember string
			List     []string
		}

		wal.multi.ReadUserConfigValue(dcrlibwallet.VSPHostConfigKey, &valueOut)
		var loadedVSP []VSPInfo

		for _, host := range valueOut.List {
			v, err := GetVSPInfo(host)
			if err == nil {
				loadedVSP = append(loadedVSP, VSPInfo{
					Host: host,
					Info: v,
				})
			}
		}

		l, _ := GetInitVSPInfo("https://api.decred.org/?c=vsp")
		for h, v := range l {
			if strings.Contains(wal.Net, v.Network) {
				loadedVSP = append(loadedVSP, VSPInfo{
					Host: fmt.Sprintf("https://%s", h),
					Info: v,
				})
			}
		}

		var resp Response
		resp.Resp = &VSP{
			List: loadedVSP,
		}
		wal.Send <- resp
	}()
}

func (wal *Wallet) RememberVSP(host string) {
	var valueOut struct {
		Remember string
		List     []string
	}
	wal.multi.ReadUserConfigValue(dcrlibwallet.VSPHostConfigKey, &valueOut)
	valueOut.Remember = host
	wal.multi.SaveUserConfigValue(dcrlibwallet.VSPHostConfigKey, valueOut)
}

func (wal *Wallet) GetRememberVSP() string {
	var valueOut struct {
		Remember string
	}
	wal.multi.ReadUserConfigValue(dcrlibwallet.VSPHostConfigKey, &valueOut)

	return valueOut.Remember
}

// getVSPInfo returns the information of the specified VSP base URL
func GetVSPInfo(url string) (*dcrlibwallet.GetVspInfoResponse, error) {
	rq := new(http.Client)
	resp, err := rq.Get((url + "/api/v3/vspinfo"))

	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response from server: %v", string(b))
	}

	var vspInfoResponse dcrlibwallet.GetVspInfoResponse
	err = json.Unmarshal(b, &vspInfoResponse)
	if err != nil {
		return nil, err
	}

	err = validateVSPServerSignature(resp, vspInfoResponse.PubKey, b)
	if err != nil {
		return nil, err
	}
	return &vspInfoResponse, nil
}

// getInitVSPInfo returns the list information of the VSP
func GetInitVSPInfo(url string) (map[string]*dcrlibwallet.GetVspInfoResponse, error) {
	rq := new(http.Client)
	resp, err := rq.Get((url))
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response from server: %v", string(b))
	}

	var vspInfoResponse map[string]*dcrlibwallet.GetVspInfoResponse
	err = json.Unmarshal(b, &vspInfoResponse)
	if err != nil {
		return nil, err
	}

	return vspInfoResponse, nil
}

func validateVSPServerSignature(resp *http.Response, pubKey, body []byte) error {
	sigStr := resp.Header.Get("VSP-Server-Signature")
	sig, err := base64.StdEncoding.DecodeString(sigStr)
	if err != nil {
		return fmt.Errorf("error validating VSP signature: %v", err)
	}

	if !ed25519.Verify(pubKey, body, sig) {
		return errors.New("bad signature from VSP")
	}

	return nil
}
