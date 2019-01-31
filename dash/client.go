package dash

import (
	"errors"
	"time"

	ch "github.com/RTradeLtd/ChainRider-Go/dash"
	"github.com/RTradeLtd/config"
	"go.uber.org/zap"
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
	c := ch.NewClient(opts)
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
func (dc *DashClient) ProcessPayment(opts *ProcessPaymentOpts, l *zap.SugaredLogger) error {
	var (
		toProcessTransactions []ch.ProcessedTxObject
		processedTransactions = make(map[string]bool)
		totalAmountSent       float64
		paymentForwardID      = opts.PaymentForward.PaymentForwardID
	)
	killTime := time.Now().Add(time.Minute * 90)
	if len(opts.PaymentForward.ProcessedTxs) == 0 {
		l.Info("no transactions detected, sleeping for 4 minutes")
		time.Sleep(time.Minute * 4)
	}
	for {
		if time.Now().UnixNano() > killTime.UnixNano() {
			return errors.New("timeout occured while waiting for transaction")
		}
		l.Info("checking for txs to process")
		paymentForward, err := dc.C.GetPaymentForwardByID(paymentForwardID)
		if err != nil {
			return err
		}
		if len(paymentForward.ProcessedTxs) == 0 {
			l.Info("no transactions detected, sleeping for 4 minutes")
			// no processed transactions yet, sleep for 4 minutes
			time.Sleep(time.Minute * 4)
			continue
		}
		l.Info("new transaction(s) detected, ensuring we haven't already processed them")
		// determine which transations we've already processed
		for _, tx := range paymentForward.ProcessedTxs {
			if !processedTransactions[tx.TransactionHash] {
				toProcessTransactions = append(toProcessTransactions, tx)
			}
		}
		if len(toProcessTransactions) == 0 {
			l.Info("all transactions have already been processed, waiting for new ones")
			time.Sleep(time.Minute * 4)
			continue
		}
		// process the actual transactions
		for _, tx := range toProcessTransactions {
			if _, err = dc.ProcessTransaction(tx.TransactionHash, killTime, l); err != nil {
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
			l.Info("total amount sent is the total amount expected, finished processing")
			return nil
		}
		l.Info("funds received, but still less than total amount expected, sleeping for 2 minutes", "amount_received", totalAmountSent)
		// clear to process transactions
		toProcessTransactions = []ch.ProcessedTxObject{}
		// sleep temporarily
		time.Sleep(time.Minute * 4)
		continue
	}
}

// ProcessTransaction is used to process a tx and wait for confirmations
func (dc *DashClient) ProcessTransaction(txHash string, killTime time.Time, logger *zap.SugaredLogger) (*ch.TransactionByHashResponse, error) {
	logger.Info("getting transaction hash to confirm")
	tx, err := dc.C.TransactionByHash(txHash)
	if err != nil {
		return nil, err
	}
	if tx.Confirmations > dc.ConfirmationCount {
		logger.Info("transaction is confirmed, validating lock time and returning")
		return tx, dc.ValidateLockTime(tx.Locktime)
	}
	// determine time to sleep in minutes
	// we multiply by 2 since 1 confirmation means 1 block, for which block time is 2 minutes
	timeToSleep := time.Minute * time.Duration((dc.ConfirmationCount-tx.Confirmations)*2)
	logger.Infof("transaction not yet confirmed, sleeping for %v minutes", timeToSleep.Minutes())
	time.Sleep(timeToSleep)
	for {
		if time.Now().UnixNano() > killTime.UnixNano() {
			return nil, errors.New("timeout occured while waiting for transaction")
		}
		logger.Info("getting transaction hash to confirm")
		tx, err = dc.C.TransactionByHash(txHash)
		if err != nil {
			return nil, err
		}
		if tx.Confirmations > dc.ConfirmationCount {
			logger.Info("transaction confirmed")
			return tx, dc.ValidateLockTime(tx.Locktime)
		}
		timeToSleep := time.Minute * time.Duration((dc.ConfirmationCount-tx.Confirmations)*2)
		logger.Infof("transaction not yet confirmed, sleeping for %v minutes", timeToSleep.Minutes())
		time.Sleep(timeToSleep)
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
