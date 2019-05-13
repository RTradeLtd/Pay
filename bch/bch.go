package bch

import (
	"context"
	"errors"

	pb "github.com/gcash/bchd/bchrpc/pb"
	"google.golang.org/grpc"
)

var (
	devConfirmationCount  = 1
	prodConfirmationCount = 3
)

// Client is used to interface with the BCH blockchain
type Client struct {
	pb.BchrpcClient
	confirmationCount int
}

// NewClient is used to instantiate our new BCH gRPC client
func NewClient(ctx context.Context, url string, dev bool) (*Client, error) {
	gConn, err := grpc.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	var confirmationCount int
	if dev {
		confirmationCount = devConfirmationCount
	} else {
		confirmationCount = prodConfirmationCount
	}
	return &Client{pb.NewBchrpcClient(gConn), confirmationCount}, nil
}

// GetTX is used to retrieve a transaction
func (c *Client) GetTX(ctx context.Context, hash string) (*pb.GetTransactionResponse, error) {
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
		return errors.New("tx is not confirmed")
	}
	return errors.New("tx is not confirmed")
}
