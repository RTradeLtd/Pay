package bch

import (
	"context"
	"errors"
	"testing"

	"github.com/RTradeLtd/Pay/mocks"
	pb "github.com/gcash/bchd/bchrpc/pb"
)

var (
	url       = "127.0.0.1:5001"
	remoteURL = "192.168.1.225:8335"
)

func Test_Integration(t *testing.T) {
	t.Skip("integration")
	client, err := NewClient(context.Background(), Opts{
		CertFile: "./bch.rpc.cert",
		URL:      remoteURL})
	if err != nil {
		t.Fatal(err)
	}
	height, err := client.GetCurrentBlockHeight(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if height == 0 {
		t.Fatal("height is 0")
	}
}

func Test_NewClient_Dev(t *testing.T) {
	if _, err := NewClient(context.Background(), Opts{URL: url, Dev: true}); err != nil {
		t.Fatal(err)
	}
}

func Test_NewClient_Prod(t *testing.T) {
	if _, err := NewClient(context.Background(), Opts{
		CertFile: "./bch.rpc.cert",
		URL:      url}); err != nil {
		t.Fatal(err)
	}
}

func Test_GetTx(t *testing.T) {
	c, fbc := newMockClient()
	fbc.GetTransactionReturnsOnCall(0, &pb.GetTransactionResponse{}, nil)
	if _, err := c.GetTx(context.Background(), "123"); err != nil {
		t.Fatal(err)
	}
	fbc.GetTransactionReturnsOnCall(1, &pb.GetTransactionResponse{}, errors.New("hello"))
	if _, err := c.GetTx(context.Background(), "123"); err == nil {
		t.Fatal("error expected")
	}
}

func Test_GetConfirmationCount(t *testing.T) {
	c, fbc := newMockClient()
	fbc.GetTransactionReturnsOnCall(0, &pb.GetTransactionResponse{
		Transaction: &pb.Transaction{
			Confirmations: 6,
		},
	}, nil)
	tx, err := c.GetTx(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if count := c.GetConfirmationCount(tx); count != 6 {
		t.Fatal("bad count returnedx")
	}
}

func Test_GetCurrentBlockHeight(t *testing.T) {
	c, fbc := newMockClient()
	fbc.GetBlockInfoReturnsOnCall(0, &pb.GetBlockInfoResponse{
		Info: &pb.BlockInfo{Height: 500},
	}, nil)
	height, err := c.GetCurrentBlockHeight(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if height != 500 {
		t.Fatal("bad height recovered")
	}
}

func Test_IsConfirmed_Success(t *testing.T) {
	c, fbc := newMockClient()
	fbc.GetTransactionReturnsOnCall(0, &pb.GetTransactionResponse{
		Transaction: &pb.Transaction{
			Confirmations: 6,
			LockTime:      501,
		},
	}, nil)
	fbc.GetBlockInfoReturnsOnCall(0, &pb.GetBlockInfoResponse{
		Info: &pb.BlockInfo{Height: 499},
	}, nil)
	tx, err := c.GetTx(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.IsConfirmed(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
}

func Test_IsConfirmed_Fail_Locktime(t *testing.T) {
	c, fbc := newMockClient()
	fbc.GetTransactionReturnsOnCall(0, &pb.GetTransactionResponse{
		Transaction: &pb.Transaction{
			Confirmations: 6,
			LockTime:      501,
		},
	}, nil)
	fbc.GetBlockInfoReturnsOnCall(0, &pb.GetBlockInfoResponse{
		Info: &pb.BlockInfo{Height: 600},
	}, nil)
	tx, err := c.GetTx(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.IsConfirmed(context.Background(), tx); err == nil {
		t.Fatal("error expected")
	} else if err.Error() != "tx is not confirmed, locktime not passed" {
		t.Fatal("wrong error message returned")
	}
}

func Test_IsConfirmed_NoConfirmations(t *testing.T) {
	c, fbc := newMockClient()
	fbc.GetTransactionReturnsOnCall(0, &pb.GetTransactionResponse{
		Transaction: &pb.Transaction{
			Confirmations: 0,
			LockTime:      501,
		},
	}, nil)
	fbc.GetBlockInfoReturnsOnCall(0, &pb.GetBlockInfoResponse{
		Info: &pb.BlockInfo{Height: 600},
	}, nil)
	tx, err := c.GetTx(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.IsConfirmed(context.Background(), tx); err == nil {
		t.Fatal("error expected")
	} else if err.Error() != "tx is not confirmed" {
		t.Fatal("wrong error message returned")
	}
}

func Test_ProcessPaymentTx(t *testing.T) {
	c, fbc := newMockClient()

	fbc.GetTransactionReturnsOnCall(0, &pb.GetTransactionResponse{
		Transaction: &pb.Transaction{
			Confirmations: 0,
			LockTime:      501,
		},
	}, nil)

	fbc.GetTransactionReturnsOnCall(1, &pb.GetTransactionResponse{
		Transaction: &pb.Transaction{
			Confirmations: 1,
			LockTime:      501,
		},
	}, nil)

	fbc.GetTransactionReturnsOnCall(2, &pb.GetTransactionResponse{
		Transaction: &pb.Transaction{
			Confirmations: 5,
			LockTime:      501,
		},
	}, nil)

	fbc.GetBlockInfoReturnsOnCall(0, &pb.GetBlockInfoResponse{
		Info: &pb.BlockInfo{Height: 499},
	}, nil)

	fbc.GetBlockInfoReturnsOnCall(1, &pb.GetBlockInfoResponse{
		Info: &pb.BlockInfo{Height: 499},
	}, nil)

	fbc.GetBlockInfoReturnsOnCall(2, &pb.GetBlockInfoResponse{
		Info: &pb.BlockInfo{Height: 499},
	}, nil)

	if err := c.ProcessPaymentTx(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
}

func newMockClient() (*Client, *mocks.FakeBchrpcClient) {
	dev = true
	fbc := &mocks.FakeBchrpcClient{}
	c := &Client{}
	c.confirmationCount = 3
	c.BchrpcClient = fbc
	return c, fbc
}
