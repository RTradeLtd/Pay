package bch

import (
	"context"
	"errors"
	"time"

	"github.com/RTradeLtd/config/v2"
	pb "github.com/gcash/bchd/bchrpc/pb"
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
	return c.GetTransaction(ctx, &pb.GetTransactionRequest{Hash: []byte(hash)})
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
		if tx.GetTransaction().GetLockTime() > uint32(height) {
			return nil
		}
		return errors.New(ErrTxNotConfirmedLockTime)
	}
	return errors.New(ErrTxNotConfirmed)
}

// ProcessPaymentTx is used to process a payment transaction
func (c *Client) ProcessPaymentTx(ctx context.Context, totalValue float64, senderAddress, recipientAddress, hash string) error {
	tx, err := c.GetTx(ctx, hash)
	if err != nil {
		return err
	}
	// validate the transaction output value
	if txValue := c.getTotalValueOfTx(tx); txValue != totalValue {
		return errors.New(ErrTxTooLowValue)
	}
	// validate the sender of the inputs
	if err := c.validateSender(tx, senderAddress); err != nil {
		return err
	}
	// validate the recipient of the outputs
	if err := c.validateRecipient(tx, recipientAddress); err != nil {
		return err
	}
	if err := c.IsConfirmed(ctx, tx); err == nil {
		return nil
	}
	// wait for some blocks to pass
	c.pause()
	for {
		tx, err = c.GetTx(ctx, hash)
		if err != nil {
			return err
		}
		if err := c.IsConfirmed(ctx, tx); err == nil {
			return nil
		}
		c.pause()
	}
}

func (c *Client) getTotalValueOfTx(tx *pb.GetTransactionResponse) float64 {
	outputs := tx.GetTransaction().GetOutputs()
	var totalValue int64
	for _, output := range outputs {
		totalValue = totalValue + output.GetValue()
	}
	return float64(totalValue)
}

func (c *Client) validateSender(tx *pb.GetTransactionResponse, senderAddress string) error {
	inputs := tx.GetTransaction().GetInputs()
	for _, input := range inputs {
		if input.GetAddress() != senderAddress {
			return errors.New(ErrInvalidSenderAddress)
		}
	}
	return nil
}

func (c *Client) validateRecipient(tx *pb.GetTransactionResponse, recipientAddress string) error {
	outputs := tx.GetTransaction().GetOutputs()
	for _, output := range outputs {
		if output.GetAddress() != recipientAddress {
			return errors.New(ErrInvalidRecipientAddress)
		}
	}
	return nil
}

func (c *Client) pause() {
	if dev {
		time.Sleep(time.Second * 5)
	} else {
		time.Sleep(time.Minute * 10)
	}
}
