package ethereum

import (
	"context"
	"errors"
	"time"

	"github.com/RTradeLtd/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/onrik/ethrpc"
)

const (
	devConfirmationCount  = int(3)
	prodConfirmationCount = int(30)
	dev                   = true
)

// Client is our connection to ethereum
type Client struct {
	ETH               *ethclient.Client
	RPC               *ethrpc.EthRPC
	ConfirmationCount int
}

// NewClient is used to generate our Ethereum client wrapper
func NewClient(cfg *config.TemporalConfig, connectionType string) (*Client, error) {
	var (
		err       error
		eClient   *ethclient.Client
		rpcClient *ethrpc.EthRPC
		count     int
	)
	switch connectionType {
	case "infura":
		eClient, err = ethclient.Dial(cfg.Ethereum.Connection.INFURA.URL)
		if err != nil {
			return nil, err
		}
		rpcClient = ethrpc.New(cfg.Ethereum.Connection.INFURA.URL)
	default:
		return nil, errors.New("invalid connection type")
	}
	if dev {
		count = devConfirmationCount
	} else {
		count = prodConfirmationCount
	}
	return &Client{
		ETH:               eClient,
		RPC:               rpcClient,
		ConfirmationCount: count}, nil
}

// ProcessEthPaymentTx is used to process an ethereum payment transaction
func (c *Client) ProcessEthPaymentTx(txHash string) error {
	hash := common.HexToHash(txHash)
	tx, pending, err := c.ETH.TransactionByHash(context.Background(), hash)
	if err != nil {
		return err
	}
	if pending {
		_, err := bind.WaitMined(context.Background(), c.ETH, tx)
		if err != nil {
			return err
		}
	}
	return c.WaitForConfirmations(tx)
}

// WaitForConfirmations is used to wait for enough block confirmations for a tx to be considered valid
func (c *Client) WaitForConfirmations(tx *types.Transaction) error {
	rcpt, err := c.RPC.EthGetTransactionReceipt(tx.Hash().String())
	if err != nil {
		return err
	}
	var (
		currentConfirmations int
		lastBlockChecked     int
		confirmationsNeeded  = c.ConfirmationCount
	)
	confirmedBlock := rcpt.BlockNumber
	currentBlock, err := c.RPC.EthBlockNumber()
	if err != nil {
		return err
	}
	lastBlockChecked = currentBlock
	if currentBlock > confirmedBlock {
		currentConfirmations = currentBlock - confirmedBlock
	}
	if currentConfirmations > confirmationsNeeded {
		return nil
	}
	for currentConfirmations <= confirmationsNeeded {
		currentBlock, err = c.RPC.EthBlockNumber()
		if err != nil {
			return err
		}
		if currentBlock == lastBlockChecked {
			time.Sleep(time.Second * 15)
			continue
		}
		currentConfirmations = currentBlock - confirmedBlock
	}
	rcpt, err = c.RPC.EthGetTransactionReceipt(tx.Hash().String())
	if err != nil {
		return err
	}
	if rcpt.Status != "1" {
		return errors.New("transaction status is not 1")
	}
	return nil
}
