package ethereum

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	Auth              *bind.TransactOpts
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

// UnlockAccount is used to unlck our main account
func (c *Client) UnlockAccount(keys ...string) error {
	var (
		err  error
		auth *bind.TransactOpts
	)
	if len(keys) > 0 {
		auth, err = bind.NewTransactor(strings.NewReader(keys[0]), keys[1])
	} else {
		return errors.New("config based account unlocked not yet spported")
	}
	if err != nil {
		return err
	}
	c.Auth = auth
	return nil
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
	return c.WaitForConfirmations(tx, true)
}

// WaitForConfirmations is used to wait for enough block confirmations for a tx to be considered valid
func (c *Client) WaitForConfirmations(tx *types.Transaction, ethPayment bool) error {
	fmt.Println("getting tx receipt")
	rcpt, err := c.RPC.EthGetTransactionReceipt(tx.Hash().String())
	if err != nil {
		fmt.Println("failed to get tx receipt")
		return err
	}
	var (
		// current number of confirmations
		currentConfirmations int
		// the last block a check was performed at
		lastBlockChecked int
		// the total number of confirmations needed
		confirmationsNeeded = c.ConfirmationCount
	)
	// set the block the tx was confirmed at
	confirmedBlock := rcpt.BlockNumber
	// get the current block number
	fmt.Println("getting current block number")
	currentBlock, err := c.RPC.EthBlockNumber()
	if err != nil {
		return err
	}
	// set last block checked
	lastBlockChecked = currentBlock
	// check if the current block is greater than the confirmed block
	if currentBlock > confirmedBlock {
		// set current confirmations to difference between current block and confirmed block
		currentConfirmations = currentBlock - confirmedBlock
	}
	fmt.Println("waiting for confirmations")
	// loop until we get the appropriate number of confirmations
	for {
		currentBlock, err = c.RPC.EthBlockNumber()
		if err != nil {
			return err
		}
		// set last block checked
		lastBlockChecked = currentBlock
		// if we get a block that was the same as last, temporarily sleep
		if currentBlock == lastBlockChecked {
			time.Sleep(time.Second * 15)
			continue
		}
		// set current confirmations to difference between current block and confirmed block
		currentConfirmations = currentBlock - confirmedBlock
		if currentConfirmations > confirmationsNeeded {
			break
		}
	}
	fmt.Println("transaction confirmed, refetching tx receipt")
	// get the transaction receipt
	rcpt, err = c.RPC.EthGetTransactionReceipt(tx.Hash().String())
	if err != nil {
		return err
	}
	fmt.Println("verifying transaction status")
	// verify the status of the transaction
	if rcpt.Status != "1" {
		return errors.New("transaction status is not 1")
	}
	fmt.Println("verifying gas usage")
	// if we are confirming an eth payment tx, than we must
	// restrict gas limit to 21K, which is base cost of a tx
	// and the amount of gs required to send ETH. We enforce this
	// check, because confirming contract transactions requires
	// additional validation, which normal eth payments dont
	if ethPayment && rcpt.GasUsed != int(21000) {
		return errors.New("eth payments can only use maximum of 21K gas")
	} else if ethPayment && rcpt.GasUsed == int(21000) {
		fmt.Println("tx confirmed")
		return nil
	}
	// not an eth payment, needs a bit extra processing
	if rcpt.GasUsed <= 21000 {
		return errors.New("bad gas usage, needs to be greater than 21k")
	}
	if len(rcpt.Logs) == 0 {
		return errors.New("no logs were emitted")
	}
	fmt.Println("tx confirmed")
	return nil
}
