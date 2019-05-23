package bch

import (
	"context"
	"errors"
	"testing"

	"github.com/RTradeLtd/Pay/log"
	"github.com/RTradeLtd/Pay/mocks"
	"github.com/RTradeLtd/config/v2"
	pb "github.com/gcash/bchd/bchrpc/pb"
)

var (
	url       = "127.0.0.1:5001"
	remoteURL = "192.168.1.225:8335"
	cfgPath   = "../test/config.json"
	txHash    = "011db9a9b0c3e95b4551418643ad995974c4e590e47cba107388d6939dbf97b6"
)

func Test_Integration(t *testing.T) {
	t.Skip("integration")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Services.BchGRPC.URL = remoteURL
	client, err := NewClient(context.Background(), cfg, false)
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
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewClient(context.Background(), cfg, true); err != nil {
		t.Fatal(err)
	}
}

func Test_NewClient_Prod(t *testing.T) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewClient(context.Background(), cfg, false); err != nil {
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
	tx, err := c.GetTx(context.Background(), txHash)
	if err != nil {
		t.Fatal(err)
	}
	if count := c.GetConfirmationCount(tx); count != 6 {
		t.Fatal("bad count returnedx")
	}
}

func Test_GetCurrentBlockHeight(t *testing.T) {
	c, fbc := newMockClient()
	fbc.GetBlockchainInfoReturnsOnCall(0, &pb.GetBlockchainInfoResponse{
		BestHeight: 500,
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
	fbc.GetBlockchainInfoReturnsOnCall(0, &pb.GetBlockchainInfoResponse{
		BestHeight: 511,
	}, nil)
	tx, err := c.GetTx(context.Background(), txHash)
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
			LockTime:      901,
		},
	}, nil)
	fbc.GetBlockchainInfoReturnsOnCall(0, &pb.GetBlockchainInfoResponse{
		BestHeight: 600,
	}, nil)
	tx, err := c.GetTx(context.Background(), txHash)
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
	tx, err := c.GetTx(context.Background(), txHash)
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
			Outputs: []*pb.Transaction_Output{
				{
					Address: "world",
					Value:   100000000,
				},
			},
		},
	}, nil)

	fbc.GetTransactionReturnsOnCall(1, &pb.GetTransactionResponse{
		Transaction: &pb.Transaction{
			Confirmations: 1,
			LockTime:      501,
			Outputs: []*pb.Transaction_Output{
				{
					Address: "world",
					Value:   100000000,
				},
			},
		},
	}, nil)

	fbc.GetTransactionReturnsOnCall(2, &pb.GetTransactionResponse{
		Transaction: &pb.Transaction{
			Confirmations: 5,
			LockTime:      501,
			Outputs: []*pb.Transaction_Output{
				{
					Address: "world",
					Value:   100000000,
				},
			},
		},
	}, nil)

	fbc.GetBlockchainInfoReturnsOnCall(0, &pb.GetBlockchainInfoResponse{
		BestHeight: 999,
	}, nil)

	fbc.GetBlockchainInfoReturnsOnCall(1, &pb.GetBlockchainInfoResponse{
		BestHeight: 999,
	}, nil)

	fbc.GetBlockchainInfoReturnsOnCall(2, &pb.GetBlockchainInfoResponse{
		BestHeight: 999,
	}, nil)
	logger, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.ProcessPaymentTx(context.Background(), logger, 1, txHash, "world"); err != nil {
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
