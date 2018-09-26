package ethereum_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal_Payment-ETH/ethereum"
	"github.com/RTradeLtd/config"
)

func TestEth_NewClient(t *testing.T) {
	cfg, err := config.LoadConfig("/home/solidity/config.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ethereum.NewClient(cfg, "infura"); err != nil {
		t.Fatal(err)
	}

}
