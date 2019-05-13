package bch

import (
	"context"
	"errors"
	"time"

	pb "github.com/gcash/bchd/bchrpc/pb"
	"google.golang.org/grpc"
)

var (
	devConfirmationCount  = 1
	prodConfirmationCount = 3
	dev                   = false
)

// Client is used to interface with the BCH blockchain
type Client struct {
	pb.BchrpcClient
	confirmationCount int
}

// NewClient is used to instantiate our new BCH gRPC client
func NewClient(ctx context.Context, url string, devMode bool) (*Client, error) {
	var opts []grpc.DialOption
	if dev {
		opts = append(opts, grpc.WithInsecure())
	}
	gConn, err := grpc.DialContext(ctx, url, opts...)
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
	resp, err := c.GetBlockInfo(ctx, &pb.GetBlockInfoRequest{})
	if err != nil {
		return -1, err
	}
	return resp.GetInfo().GetHeight(), nil
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
		return errors.New("tx is not confirmed, locktime not passed")
	}
	return errors.New("tx is not confirmed")
}

// ProcessPaymentTx is used to process a payment transaction
func (c *Client) ProcessPaymentTx(ctx context.Context, hash string) error {
	tx, err := c.GetTx(ctx, hash)
	if err != nil {
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

func (c *Client) pause() {
	if dev {
		time.Sleep(time.Second * 5)
	} else {
		time.Sleep(time.Minute * 10)
	}
}
