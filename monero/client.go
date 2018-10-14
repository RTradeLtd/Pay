package monero

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/RTradeLtd/config"
)

const (
	devConfirmationCount  = int(3)
	prodConfirmationCount = int(10)
	dev                   = true
)

// MoneroClient is our connection to the monero blockchain
type MoneroClient struct {
	DaemonURL         string
	WalletURL         string
	HC                *http.Client
	ConfirmationCount int
}

// GenerateMoneroClient is used to generate our monero connection
func GenerateMoneroClient(cfg *config.TemporalConfig) *MoneroClient {
	mc := &MoneroClient{
		DaemonURL: cfg.Endpoints.MoneroRPC,
		HC:        &http.Client{Timeout: time.Minute * 1},
	}
	if dev {
		mc.ConfirmationCount = devConfirmationCount
	} else {
		mc.ConfirmationCount = prodConfirmationCount
	}
	return mc
}

//'{"jsonrpc":"2.0","id":"0","method":"get_block_count"}'
func (mc *MoneroClient) FormatPayload(id, method string) string {
	return fmt.Sprintf(
		"{\n  \"jsonrpc\": \"2.0\",\n  \"id\": \"%s\",\n  \"method\": \"%s\"\n}",
		id, method,
	)
}

// GetBlockCount is used to get the current block count
func (mc *MoneroClient) GetBlockCount() (*BlockCountResponse, error) {
	payload := mc.FormatPayload("0", "get_block_count")
	req, err := http.NewRequest("POST", mc.DaemonURL, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := mc.HC.Do(req)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Status)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	intf := BlockCountResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}
