package bch

import (
	"context"
	"errors"
	"time"

	"github.com/RTradeLtd/config/v2"
	pb "github.com/gcash/bchd/bchrpc/pb"
	chainhash "github.com/gcash/bchd/chaincfg/chainhash"
	"github.com/gcash/bchutil"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	devConfirmationCount  = 1
	prodConfirmationCount = 3
	dev                   = false
	// ErrTxNotConfirmedLockTime is an error used to indicate
	// that a transaction a transaction was not confirmed
	// because the locktime is greater than current block height
	ErrTxNotConfirmedLockTime = "tx is not confirmed, locktime not passed"
	// ErrTxNotConfirmed is a general error to indicate that
	// a transaction is not yet confirmed
	ErrTxNotConfirmed = "tx is not confirmed"
	// ErrTxTooLowValue is an error used to indicate that the
	// total value of a transaction does not match the expected value
	ErrTxTooLowValue = "value of transaction does not match expected total value"
	// ErrInvalidSenderAddress is an error used to indicate that
	// a sender for one input does not match the expected
	// sender for all inputs
	ErrInvalidSenderAddress = "invalid sender address detected"
	// ErrInvalidRecipientAddress is an error used to indicate that
	// a recipient for one output does not match the expected
	// recipient for all outputs
	ErrInvalidRecipientAddress = "invalid recipient address detected"
)

// Client is used to interface with the BCH blockchain
type Client struct {
	pb.BchrpcClient
	confirmationCount int
}

// Opts is used to configure our BCH gRPC connection
type Opts struct {
	KeyFile  string
	CertFile string
	URL      string
	Dev      bool
}

// NewClient is used to instantiate our new BCH gRPC client
func NewClient(ctx context.Context, cfg *config.TemporalConfig, devMode bool) (*Client, error) {
	var dialOpts []grpc.DialOption
	if devMode {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	} else {
		creds, err := credentials.NewClientTLSFromFile(cfg.Services.BchGRPC.CertFile, "")
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	}
	gConn, err := grpc.DialContext(ctx, cfg.Services.BchGRPC.URL, dialOpts...)
	if err != nil {
		return nil, err
	}
	var confirmationCount int
	if devMode {
		confirmationCount = devConfirmationCount
	} else {
		confirmationCount = prodConfirmationCount
	}
	dev = devMode
	return &Client{pb.NewBchrpcClient(gConn), confirmationCount}, nil
}

// GetTx is used to retrieve a transaction
func (c *Client) GetTx(ctx context.Context, hash string) (*pb.GetTransactionResponse, error) {
	hsh, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, err
	}
	return c.GetTransaction(ctx, &pb.GetTransactionRequest{Hash: hsh[:]})
}

// GetConfirmationCount is used to get the number of confirmations for a particular tx
func (c *Client) GetConfirmationCount(tx *pb.GetTransactionResponse) int32 {
	return tx.GetTransaction().GetConfirmations()
}

// GetCurrentBlockHeight is used to retrieve the current height, aka block number
func (c *Client) GetCurrentBlockHeight(ctx context.Context) (int32, error) {
	resp, err := c.GetBlockchainInfo(ctx, &pb.GetBlockchainInfoRequest{})
	if err != nil {
		return -1, err
	}
	return resp.GetBestHeight(), nil
}

// IsConfirmed is used to check if a transaction is confirmed
func (c *Client) IsConfirmed(ctx context.Context, tx *pb.GetTransactionResponse) error {
	if c.GetConfirmationCount(tx) > int32(c.confirmationCount) {
		height, err := c.GetCurrentBlockHeight(ctx)
		if err != nil {
			return err
		}
		// ensure that the lock time is less than or equal to current height
		if tx.GetTransaction().GetLockTime() <= uint32(height) {
			return nil
		}
		return errors.New(ErrTxNotConfirmedLockTime)
	}
	return errors.New(ErrTxNotConfirmed)
}

// ProcessPaymentTx is used to process a payment transaction
func (c *Client) ProcessPaymentTx(ctx context.Context, l *zap.SugaredLogger, expectedValue float64, hash, depositAddress string) error {
	l.Info("getting tx from blockchain")
	tx, err := c.GetTx(ctx, hash)
	if err != nil {
		return err
	}
	l.Info("validating value of tx")
	// validate the transaction output value
	// this will ensure that we only examine outputs
	// that match the depositAddress
	txValue := c.getTotalValueOfTx(tx, depositAddress)
	if txValue < expectedValue {
		return errors.New(ErrTxTooLowValue)
	}
	l.Info("checking if tx is confirmed")
	// check whether the tx is confirmed
	if err := c.IsConfirmed(ctx, tx); err == nil {
		l.Info("tx confirmed")
		return nil
	}
	l.Info("tx not confirmed, waiting for confirmations")
	// dissallow processing times longer than 3 hours
	killTime := time.Now().Add(time.Hour * 3)
	// wait for some blocks to pass
	c.pause()
	for {
		if time.Now().UnixNano() > killTime.UnixNano() {
			return errors.New("timeout occured while waiting for transaction to confirm")
		}
		l.Info("checking if tx is confirmed")
		// refetch tx from blockchain
		tx, err = c.GetTx(ctx, hash)
		if err != nil {
			return err
		}
		// check whether tx is confirmed
		if err := c.IsConfirmed(ctx, tx); err == nil {
			l.Info("tx confirmed")
			return nil
		}
		l.Info("tx not confirmed, waiting for confirmations")
		c.pause()
	}
}

func (c *Client) getTotalValueOfTx(tx *pb.GetTransactionResponse, depositAddress string) float64 {
	// total value measured in satoshis
	var totalValue int64
	// since BCH is a UTXO based blockchain, we must validate only the outputs
	// since we don't particularly care about the inputs. For each output
	// we find, we must first validate the address. If the address matches
	// our deposit address, then we add the value of that output to our total value
	for _, output := range tx.GetTransaction().GetOutputs() {
		// ensure we only account for outputs
		// whose recipient is our deposit address
		if output.GetAddress() != depositAddress {
			continue
		}
		totalValue = totalValue + output.GetValue()
	}
	amt := bchutil.Amount(totalValue)
	// convert the total value from satoshi's into BCH.
	return amt.ToBCH()
}

func (c *Client) pause() {
	if dev {
		time.Sleep(time.Second * 5)
	} else {
		time.Sleep(time.Minute * 10)
	}
}
