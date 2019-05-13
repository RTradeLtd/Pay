package bch

import (
	"context"
	"testing"

	"github.com/RTradeLtd/Pay/mocks"
	pb "github.com/gcash/bchd/bchrpc/pb"
)

var (
	url = "127.0.0.1:5001"
)

func Test_NewClient(t *testing.T) {
	if _, err := NewClient(context.Background(), url, true); err != nil {
		t.Fatal(err)
	}
}

func Test_GetTx(t *testing.T) {
	c, fbc := newMockClient()
	fbc.GetTransactionReturnsOnCall(0, &pb.GetTransactionResponse{}, nil)
	if _, err := c.GetTx(context.Background(), "123"); err != nil {
		t.Fatal(err)
	}
}

func newMockClient() (*Client, *mocks.FakeBchrpcClient) {
	fbc := &mocks.FakeBchrpcClient{}
	c := &Client{}
	c.confirmationCount = 3
	c.BchrpcClient = fbc
	return c, fbc
}
