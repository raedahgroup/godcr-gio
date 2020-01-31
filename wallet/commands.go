package wallet

import (
	"errors"

	"github.com/raedahgroup/dcrlibwallet"
)

var (
	// ErrInvalidArguments is returned when a wallet command is send with invalid arguments.
	ErrInvalidArguments = errors.New("invalid command arguments")

	// ErrNotFound is returned when a wallet command is given that does not exist or is not
	// implemented.
	ErrNotFound = errors.New("command not found or not implemented")

	// ErrNoSuchWallet is returned with the wallet requested by the given id does not exist
	ErrNoSuchWallet = errors.New("no such wallet with id")

	// ErrNoSuchAcct is returned when the given account number cannot be found
	ErrNoSuchAcct = errors.New("no such account")

	// ErrCreateTx is returned when a tx author cannot be created
	ErrCreateTx = errors.New("can not create transaction")
)

func (wal *Wallet) CreateWallet(passphrase string, passtype int32) {
	go func(send chan<- interface{}, passphrase string, passtype int32) {

		wall, err := wal.multi.CreateNewWallet(passphrase, int32(passtype))
		if err != nil {
			send <- err
			return
		}
		send <- &CreatedSeed{
			Seed: wall.Seed,
		}
	}(wal.Send, passphrase, passtype)
}

func (wal *Wallet) RestoreWallet(seed, passphrase string, passtype int32) {
	go func(send chan<- interface{}, seed, passpassphrase string, paspasstype int32) {

		_, err := wal.multi.RestoreWallet(seed, passphrase, int32(passtype))
		if err != nil {
			send <- err
			return
		}

		send <- &Restored{}
	}(wal.Send, seed, passphrase, passtype)
}

func (wal *Wallet) CreateTransaction(walletID int, accountID, confirms int32) {
	go func(send chan<- interface{}, walletID int, acct, confims int32) {
		wallets, err := wal.wallets()
		if err != nil {
			send <- err
			return
		}

		if walletID > len(wallets) || walletID < 0 {
			send <- err
			return
		}

		if _, err := wallets[walletID].GetAccount(acct, confirms); err != nil {
			send <- err
			return
		}

		txAuthor := wallets[walletID].NewUnsignedTx(acct, confirms)
		if txAuthor == nil {
			send <- err
			return
		}

		send <- txAuthor
	}(wal.Send, walletID, accountID, confirms)
}

func (wal *Wallet) GetTransactions(walletID int, offset, limit, txfilter int32) {
	go func(send chan<- interface{}, walletID int, offset, limit, txfilter int32) {
		wallets, err := wal.wallets()
		if err != nil {
			send <- err
			return
		}
		alltxs := make([][]dcrlibwallet.Transaction, len(wallets))
		for i, wall := range wallets {
			txs, err := wall.GetTransactionsRaw(offset, limit, txfilter, true)
			if err != nil {
				send <- err
				return
			}
			alltxs[i] = txs
		}

		send <- &Transactions{
			Txs: alltxs,
		}
	}(wal.Send, walletID, offset, limit, txfilter)
}

func (wal *Wallet) GetMultiWalletInfo(confirms int32) {
	go func(send chan<- interface{}, confims int32) {
		wallets, err := wal.wallets()
		if err != nil {
			send <- err
			return
		}
		var completeTotal int64
		for _, wall := range wallets {
			iter, err := wall.AccountsIterator(confirms)
			if err != nil {
				send <- err
				return
			}
			for acct := iter.Next(); acct != nil; acct = iter.Next() {
				completeTotal += acct.TotalBalance
			}
		}
		best := wal.multi.GetBestBlock()

		send <- &MultiWalletInfo{
			LoadedWallets:   len(wallets),
			TotalBalance:    completeTotal,
			BestBlockHeight: best.Height,
			BestBlockTime:   best.Timestamp,
			Synced:          wal.multi.IsSynced(),
		}
	}(wal.Send, confirms)
}
