package wallet_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/planetdecred/godcr/wallet"
)

const (
	testnet = "testnet3"
)

func getTestDir() string {
	now := time.Now().UTC().Unix()
	testDir := fmt.Sprintf(".godcr_test_%d", now)
	_, err := os.Stat(testDir)
	i := 1
	for err == nil {
		testDir = fmt.Sprintf(".godcr_test_%d_%d", now, i)
		_, err = os.Stat(testDir)
		i++
	}
	err = os.Mkdir(testDir, os.ModePerm)
	Expect(err).To(BeNil())
	return testDir
}

var (
	wal *Wallet
)
var _ = BeforeSuite(func() {
	var err error
	wal, err = NewWallet(getTestDir(), testnet, make(chan Response), 2)
	Expect(err).To(BeNil())
	wal.LoadWallets()
	resp := <-wal.Send
	Expect(resp.Resp).To(BeAssignableToTypeOf(LoadedWallets{}))
	tempChan := make(chan error)
	wal.CreateWallet("123", "password", tempChan)
	go func() {
		err := <-tempChan
		Expect(err).To(BeNil())
	}()
	resp = <-wal.Send
	Expect(resp.Resp).To(BeAssignableToTypeOf(CreatedSeed{}))
})

var _ = AfterSuite(func() {
	wal.Shutdown()
})

var _ = Describe("Wallet", func() {
	It("can rename a wallet", func() {
		tempChan := make(chan error)
		wal.RenameWallet(1, "random", tempChan)
		go func() {
			err := <-tempChan
			Expect(err).To(BeNil())
		}()
		resp := <-wal.Send
		Expect(resp.Resp).To(BeAssignableToTypeOf(Renamed{}))
	})
	It("can get the current address", func() {
		addr, err := wal.CurrentAddress(1, 0)
		Expect(err).To(BeNil())
		Expect(wal.IsAddressValid(addr)).To(Equal(true))
		addr2, err := wal.CurrentAddress(1, 0)
		Expect(err).To(BeNil())
		Expect(addr).To(Equal(addr2))
	})
	It("can create a new address", func() {
		addr, err := wal.NextAddress(1, 0)
		Expect(err).To(BeNil())
		Expect(wal.IsAddressValid(addr)).To(Equal(true))
	})
})
