package bch

import (
	"context"

	pb "github.com/gcash/bchd/bchrpc/pb"
	"google.golang.org/grpc"
)

// Client is used to interface with the BCH blockchain
type Client struct {
	pb.BchrpcClient
}

// NewClient is used to instantiate our new BCH gRPC client
func NewClient(ctx context.Context, url string) (*Client, error) {
	gConn, err := grpc.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	return &Client{pb.NewBchrpcClient(gConn)}, nil
}
