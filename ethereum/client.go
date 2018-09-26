package ethereum

import (
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	Connection *ethclient.Client
}
