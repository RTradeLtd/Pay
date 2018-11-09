package dash

import (
	"errors"
	"fmt"
	"time"

	ch "github.com/RTradeLtd/ChainRider-Go/dash"
	"github.com/RTradeLtd/config"
)

const (
	devConfirmationCount  = int(3)
	prodConfirmationCount = int(6)
	dev                   = true
)

// DashClient is our connection to the dash blockchain via chainrider api
type DashClient struct {
	C                 *ch.Client
	ConfirmationCount int
}

// ProcessPaymentOpts are parameters needed to validate a payment
type ProcessPaymentOpts struct {
	Number         int64
	ChargeAmount   float64
	PaymentForward *ch.GetPaymentForwardByIDResponse
}

// GenerateDashClient is used to generate our dash client to process transactions
func GenerateDashClient(cfg *config.TemporalConfig) (*DashClient, error) {
	opts := &ch.ConfigOpts{
		APIVersion:      "v1",
		DigitalCurrency: "dash",
		Blockchain:      "testnet",
		Token:           cfg.APIKeys.ChainRider,
	}
	if !dev {
		opts.Blockchain = "main"
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

// ProcessPayment is used to process a dash based payment
func (dc *DashClient) ProcessPayment(opts *ProcessPaymentOpts) error {
	var (
		toProcessTransactions []ch.ProcessedTxObject
		processedTransactions = make(map[string]bool)
		totalAmountSent       float64
		paymentForwardID      = opts.PaymentForward.PaymentForwardID
	)
	if len(opts.PaymentForward.ProcessedTxs) == 0 {
		fmt.Println("no transactions detected, sleeping")
		// no processed transactions yet, sleep for 2 minutes and then check again
		time.Sleep(time.Minute * 2)
	}
	for {
		fmt.Println("checking for txs to process")
		paymentForward, err := dc.C.GetPaymentForwardByID(paymentForwardID)
		if err != nil {
			return err
		}
		if len(paymentForward.ProcessedTxs) == 0 {
			fmt.Println("no transactions detected, sleeping")
			// no processed transactions yet, sleep for 2 minutes
			time.Sleep(time.Minute * 2)
			continue
		}
		fmt.Println("new transaction(s) detected")
		// determine which transations we've already processed
		for _, tx := range paymentForward.ProcessedTxs {
			if !processedTransactions[tx.TransactionHash] {
				toProcessTransactions = append(toProcessTransactions, tx)
			}
		}
		if len(toProcessTransactions) == 0 {
			fmt.Println("no new transactions to process, sleeping")
			time.Sleep(time.Minute * 2)
			continue
		}
		// process the actual transactions
		for _, tx := range toProcessTransactions {
			if _, err = dc.ProcessTransaction(tx.TransactionHash); err != nil {
				return err
			}
			txValueFloat := ch.DuffsToDash(float64(int64(tx.ReceivedAmountDuffs)))
			totalAmountSent = totalAmountSent + txValueFloat
			// get the value of the transaction and add it to the total amount
			// set the transaction being processed to true in order to avoid reprocessing
			processedTransactions[tx.TransactionHash] = true
		}
		// if they have paid enough, quit processing
		if totalAmountSent >= opts.ChargeAmount {
			fmt.Println("total charge amount reached, processing finished")
			return nil
		}
		fmt.Println("still need more funds, sleeping for 2 minutes")
		// clear to process transactions
		toProcessTransactions = []ch.ProcessedTxObject{}
		// sleep temporarily
		time.Sleep(time.Minute * 2)
		continue
	}
}

// ProcessTransaction is used to process a tx and wait for confirmations
func (dc *DashClient) ProcessTransaction(txHash string) (*ch.TransactionByHashResponse, error) {
	fmt.Println("grabbing transaction")
	tx, err := dc.C.TransactionByHash(txHash)
	if err != nil {
		return nil, err
	}
	if tx.Confirmations > dc.ConfirmationCount {
		fmt.Println("transaction confirmed")
		return tx, dc.ValidateLockTime(tx.Locktime)
	}
	fmt.Println("sleeping for 2 minutes before querying again ")
	// dash  block time is long, so we can sleep for a bit
	time.Sleep(time.Minute * 2)
	for {
		fmt.Println("grabbing tx")
		tx, err = dc.C.TransactionByHash(txHash)
		if err != nil {
			return nil, err
		}
		if tx.Confirmations > dc.ConfirmationCount {
			fmt.Println("transaction confirmed")
			return tx, dc.ValidateLockTime(tx.Locktime)
		}
		fmt.Println("current confirmation count ", tx.Confirmations)
		fmt.Println("sleeping for 2 minutes before querying again")
		time.Sleep(time.Minute * 2)
	}
}

// ValidateLockTime is used to validate the given lock time compared to the current block height
func (dc *DashClient) ValidateLockTime(locktime int) error {
	blockHash, err := dc.C.GetLastBlockHash()
	if err != nil {
		return err
	}
	block, err := dc.C.GetBlockByHash(blockHash.LastBlockHash)
	if err != nil {
		return err
	}
	if locktime > block.Height {
		return errors.New("locktime is greater than block height")
	}
	return nil
}
