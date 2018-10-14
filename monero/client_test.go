package monero_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Pay/monero"
	"github.com/RTradeLtd/config"
)

const (
	endpoint = "http://192.168.1.236:18081/json_rpc"
)

func TestClient(t *testing.T) {
	cfg := &config.TemporalConfig{
		Endpoints: config.Endpoints{
			MoneroRPC: endpoint,
		},
	}
	mc := monero.GenerateMoneroClient(cfg)
	if count, err := mc.GetBlockCount(); err != nil {
		t.Fatal(err)
	} else {
		fmt.Printf("%+v\n", count)
	}

}
