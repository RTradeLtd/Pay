package ethereum

import (
	"errors"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Client is our connection to ethereum
type Client struct {
	Connection        *ethclient.Client
	ConfirmationCount uint64
}

// NewClient is used to generate our Ethereum client wrapper
func NewClient(cfg *config.TemporalConfig, connectionType string) (*Client, error) {
	var (
		err    error
		client *ethclient.Client
	)
	switch connectionType {
	case "infura":
		client, err = ethclient.Dial(cfg.Ethereum.Connection.INFURA.URL)
	default:
		return nil, errors.New("invalid connection type")
	}
	if err != nil {
		return nil, err
	}
	return &Client{Connection: client, ConfirmationCount: 12}, nil
}

/*
func (c *Client) ProcessEthPaymentTx(txHash string) error {
	hash := common.HexToHash(txHash)
	tx, pending, err := c.Connection.TransactionByHash(context.Background(), hash)
	if err != nil {
		return err
	}
}

func (c *Client) processPendingTransaction(tx *types.Transaction) error {
	var (
		err                  error
		currentConfirmations uint64 = 0
		confirmationsNeeded         = c.ConfirmationCount
	)
	currentBlock, err := c.Connection.BlockByNumber(context.Background(), nil)
	if err != nil {
		return err
	}
}*/
