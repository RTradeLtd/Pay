package dash

import (
	"errors"
	"time"

	ch "github.com/RTradeLtd/ChainRider-Go"
)

const (
	devConfirmationCount  = int(3)
	prodConfirmationCount = int(30)
	dev                   = true
)

type DashClient struct {
	C                 *ch.Client
	ConfirmationCount int
}

// GenerateDashClient is used to generate our dash client to process transactions
func GenerateDashClient(cfg *ch.ConfigOpts) (*DashClient, error) {
	c, err := ch.NewClient(cfg)
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
func (dc *DashClient) ProcessTransaction(txHash string) (bool, error) {
	tx, err := dc.C.TransactionByHash(txHash)
	if err != nil {
		return false, err
	}
	if tx.Locktime > 0 {
		return false, errors.New("lock time must be equal to 0")
	}
	if tx.Confirmations > dc.ConfirmationCount {
		return true, nil
	}
	// dash  block time is long, so we can sleep for a bit
	time.Sleep(time.Minute * 2)
	for {
		tx, err = dc.C.TransactionByHash(txHash)
		if err != nil {
			return false, err
		}
		if tx.Confirmations > dc.ConfirmationCount {
			return true, nil
		}
		time.Sleep(time.Minute * 2)
	}
}
