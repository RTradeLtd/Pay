package ethereum

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/RTradeLtd/config/v2"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/onrik/ethrpc"
	ens "github.com/wealdtech/go-ens/v3"
)

const (
	devConfirmationCount  = int(3)
	prodConfirmationCount = int(30)
	dev                   = false
	// TemporalENSName is our official ens name
	TemporalENSName = "ipfstemporal.eth"
)

// Client is our connection to ethereum
type Client struct {
	ETH                    *ethclient.Client
	RPC                    *ethrpc.EthRPC
	Auth                   *bind.TransactOpts
	RTCAddress             string
	PaymentContractAddress string
	ConfirmationCount      int
	// enables locking for write transactions
	// to prevent issues with nonce reuse
	txMux sync.Mutex
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
	case "rpc":
		url := fmt.Sprintf("http://%s:%s", cfg.Ethereum.Connection.RPC.IP, cfg.Ethereum.Connection.RPC.Port)
		eClient, err = ethclient.Dial(url)
		if err != nil {
			return nil, err
		}
		rpcClient = ethrpc.New(url)
	default:
		return nil, errors.New("invalid connection type")
	}
	if dev {
		count = devConfirmationCount
	} else {
		count = prodConfirmationCount
	}
	return &Client{
		ETH:                    eClient,
		RPC:                    rpcClient,
		RTCAddress:             cfg.Ethereum.Contracts.RTCAddress,
		PaymentContractAddress: cfg.Ethereum.Contracts.PaymentContractAddress,
		ConfirmationCount:      count}, nil
}

// SetResolver is used to check if name
// doesnt have a public resolver, set it
func (c *Client) SetResolver(name string) error {
	c.txMux.Lock()
	defer c.txMux.Unlock()
	nm, err := ens.NewName(c.ETH, name)
	if err != nil {
		return err
	}
	resAddr, err := nm.ResolverAddress()
	if err != nil {
		return err
	}
	if resAddr.String() != common.HexToAddress("0x0").String() {
		return nil
	}
	pubResolver, err := ens.PublicResolverAddress(c.ETH)
	if err != nil {
		return err
	}
	c.Auth.GasPrice = c.getGasPrice()
	tx, err := nm.SetResolverAddress(pubResolver, c.Auth)
	if err != nil {
		return err
	}
	if rcpt, err := bind.WaitMined(context.Background(), c.ETH, tx); err != nil {
		return err
	} else if rcpt.Status != 1 {
		return errors.New("tx with incorrect status")
	}
	return nil
}

// RegisterName is used to register an ENS name
func (c *Client) RegisterName(name string) error {
	c.txMux.Lock()
	defer c.txMux.Unlock()
	contract, err := ens.NewName(c.ETH, name)
	if err != nil {
		return err
	}
	// calculate rent cost for 1 year
	cost, err := getRentCost(contract, 31536000)
	if err != nil {
		return err
	}
	c.Auth.GasPrice = c.getGasPrice()
	// start the initial registration step
	tx, secret, err := contract.RegisterStageOne(c.Auth.From, c.Auth)
	if err != nil {
		return err
	}
	if rcpt, err := bind.WaitMined(context.Background(), c.ETH, tx); err != nil {
		return err
	} else if rcpt.Status != 1 {
		return errors.New("tx with incorrect status")
	}
	// get the specified registration step interval period
	sleepDuration, err := contract.RegistrationInterval()
	if err != nil {
		return err
	}
	// sleep for the specified duration plus an additional minute
	time.Sleep(sleepDuration + time.Minute)
	// set rent cost as tx value
	c.Auth.Value = cost
	// defer reset of tx value
	defer func() {
		c.Auth.Value = big.NewInt(0)
	}()
	c.Auth.GasPrice = c.getGasPrice()
	// start the final registration step
	tx, err = contract.RegisterStageTwo(c.Auth.From, secret, c.Auth)
	if err != nil {
		return err
	}
	if rcpt, err := bind.WaitMined(context.Background(), c.ETH, tx); err != nil {
		return err
	} else if rcpt.Status != 1 {
		return errors.New("tx with incorrect status")
	}
	return nil
}

// RegisterSubDomain is used to register a subdomain under
// ipfstemporal.eth, allowing it to be updated with a content hash
func (c *Client) RegisterSubDomain(subName, parentName string) error {
	c.txMux.Lock()
	defer c.txMux.Unlock()
	// create a registry contract handler
	contract, err := ens.NewRegistry(c.ETH)
	if err != nil {
		return err
	}
	c.Auth.GasPrice = c.getGasPrice()
	// create the subdomain, setting the name, and marking us as the owner
	// this ensure we can manage the subdomain
	tx, err := contract.SetSubdomainOwner(
		c.Auth,
		parentName,
		subName,
		c.Auth.From,
	)
	if err != nil {
		return err
	}
	if rcpt, err := bind.WaitMined(context.Background(), c.ETH, tx); err != nil {
		return err
	} else if rcpt.Status != 1 {
		return errors.New("tx with incorrect status")
	}
	pubResolver, err := ens.PublicResolverAddress(c.ETH)
	if err != nil {
		return err
	}
	c.Auth.GasPrice = c.getGasPrice()
	tx, err = contract.SetResolver(
		c.Auth,
		c.GetCombinedName(subName, parentName),
		pubResolver,
	)
	if err != nil {
		return err
	}
	if rcpt, err := bind.WaitMined(context.Background(), c.ETH, tx); err != nil {
		return err
	} else if rcpt.Status != 1 {
		return errors.New("tx with incorrect status")
	}
	return nil
}

// UpdateContentHash is used to update the ipfs content hash
// of a particular *.ipfstemporal.eth subdomain
func (c *Client) UpdateContentHash(subName, parentName, hash string) error {
	c.txMux.Lock()
	defer c.txMux.Unlock()
	resolver, err := ens.NewResolver(c.ETH, c.GetCombinedName(subName, parentName))
	if err != nil {
		return err
	}
	c.Auth.GasPrice = c.getGasPrice()
	tx, err := resolver.SetContenthash(c.Auth, []byte(hash))
	if err != nil {
		return err
	}
	if rcpt, err := bind.WaitMined(context.Background(), c.ETH, tx); err != nil {
		return err
	} else if rcpt.Status != 1 {
		return errors.New("tx with incorrect status")
	}
	return nil
}

// UnlockAccountFromConfig generates a bind transactor opts from temporal config
func (c *Client) UnlockAccountFromConfig(cfg *config.TemporalConfig) error {
	fileBytes, err := ioutil.ReadFile(cfg.Ethereum.Account.KeyFile)
	if err != nil {
		return err
	}
	pk, err := keystore.DecryptKey(fileBytes, cfg.Ethereum.Account.KeyPass)
	if err != nil {
		return err
	}
	c.Auth = bind.NewKeyedTransactor(pk.PrivateKey)
	return nil
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

// ProcessPaymentTx is used to process an ethereum/rtc based
// credit purchase
func (c *Client) ProcessPaymentTx(txHash string) error {
	fmt.Println("getting tx receipt")
	hash := common.HexToHash(txHash)
	tx, pending, err := c.ETH.TransactionByHash(context.Background(), hash)
	if err != nil {
		return err
	}
	fmt.Printf("tx receipt:\n%+v\n", tx)
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
		fmt.Println("current confirmations ", currentConfirmations)
		fmt.Println("confirmations needed ", confirmationsNeeded)
		currentBlock, err = c.RPC.EthBlockNumber()
		if err != nil {
			return err
		}
		// if we get a block that was the same as last, temporarily sleep
		if currentBlock == lastBlockChecked {
			time.Sleep(time.Second * 15)
		}
		lastBlockChecked = currentBlock
		// set current confirmations to difference between current block and confirmed block
		currentConfirmations = currentBlock - confirmedBlock
		if currentConfirmations >= confirmationsNeeded {
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
	if rcpt.Status != TxStatusSuccess {
		return errors.New("transaction status is not 1")
	}
	if len(rcpt.Logs) == 0 {
		return errors.New("no logs were emitted")
	}
	// refetch the transaction receipt, using go-ethereum
	tx, _, err = c.ETH.TransactionByHash(context.Background(), tx.Hash())
	if err != nil {
		return err
	}
	// verify that the destination address, is the RTC contract address
	// we dont want to consider a garbage token transfer to be valid, it MUST
	// be the RTC token
	if tx.To().String() != c.PaymentContractAddress {
		return errors.New("destination address must be the payments contract address")
	}
	// if rcpt.ContractAddress is not empty, then this is a contract transaction,
	// so the contract address should be equal to rtc token address
	if rcpt.ContractAddress != "" {
		if rcpt.ContractAddress != c.RTCAddress {
			return errors.New("token transaction is not rtc")
		}
	}
	fmt.Println("tx confirmed")
	return nil
}

func getRentCost(en *ens.Name, rentSeconds int64) (*big.Int, error) {
	costSec, err := en.RentCost()
	if err != nil {
		return nil, err
	}
	return new(big.Int).Mul(costSec, big.NewInt(rentSeconds)), nil
}

// GetCombinedName is used to return a combined ens name
func (c *Client) GetCombinedName(subName, parentName string) string {
	// ensure that if the first character is not a .
	// we make it a .
	if parentName[0] != '.' {
		parentName = "." + parentName
	}
	// if the subName includes a period, remove it
	if subName[len(subName)-1] == '.' {
		subName = strings.TrimSuffix(subName, ".")
	}
	// return the combined subname and parent name
	return subName + parentName
}

func (c *Client) getGasPrice() *big.Int {
	gprice := new(big.Int)
	gprice.SetString("5000000000", 10)
	return gprice
}
