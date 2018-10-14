package dash

import (
	"errors"
	"fmt"
	"time"

	ch "github.com/RTradeLtd/ChainRider-Go"
	"github.com/RTradeLtd/config"
)

const (
	devConfirmationCount  = int(3)
	prodConfirmationCount = int(6)
	dev                   = true
)

type DashClient struct {
	C                 *ch.Client
	ConfirmationCount int
}

// GenerateDashClient is used to generate our dash client to process transactions
func GenerateDashClient(cfg *config.TemporalConfig) (*DashClient, error) {
	opts := &ch.ConfigOpts{
		APIVersion:      "v1",
		DigitalCurrency: "dash",
		Blockchain:      "testnet",
		Token:           cfg.APIKeys.ChainRider,
	}
	c, err := ch.NewClient(opts)
	if err != nil {
		return nil, err
	}
	dc := &DashClient{
		C: c,
	}
	if dev {
		dc.ConfirmationCount = devConfirmationCount
	} else {
		dc.ConfirmationCount = prodConfirmationCount
	}
	return dc, nil
}

// ProcessTransaction is used to process a tx and wait for confirmations
func (dc *DashClient) ProcessTransaction(txHash string) error {
	fmt.Println("grabbing transaction")
	tx, err := dc.C.TransactionByHash(txHash)
	if err != nil {
		return err
	}
	if tx.Locktime > 0 {
		return errors.New("lock time must be equal to 0")
	}
	if tx.Confirmations > dc.ConfirmationCount {
		fmt.Println("transaction confirmed")
		return nil
	}
	fmt.Println("sleeping for 2 minutes before querying again ")
	// dash  block time is long, so we can sleep for a bit
	time.Sleep(time.Minute * 2)
	for {
		fmt.Println("grabbing tx")
		tx, err = dc.C.TransactionByHash(txHash)
		if err != nil {
			return err
		}
		if tx.Confirmations > dc.ConfirmationCount {
			fmt.Println("transaction confirmed")
			return nil
		}
		fmt.Println("sleeping for 2 minutes before querying again")
		time.Sleep(time.Minute * 2)
	}
}
