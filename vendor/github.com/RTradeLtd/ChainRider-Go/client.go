package chainridergo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// NewClient is used to initialize our ChainRider client
func NewClient(opts *ConfigOpts) (*Client, error) {
	if opts == nil {
		opts = &ConfigOpts{
			APIVersion:      defaultAPIVersion,
			DigitalCurrency: defaultDigitalCurrency,
			Blockchain:      defaultBlockchain,
			Token:           "test",
		}
	}
	urlFormatted := fmt.Sprintf(urlTemplate, opts.APIVersion, opts.DigitalCurrency, opts.Blockchain)
	c := &Client{
		URL:   urlFormatted,
		Token: opts.Token,
		HC:    &http.Client{},
	}
	// generate our payload
	c.GeneratePayload()
	return c, nil
}

// GeneratePayload is used to generate our payload
func (c *Client) GeneratePayload() {
	c.Payload = fmt.Sprintf(payloadTemplate, c.Token)
}

// GetRateLimit is used to get our rate limit information for the current token
func (c *Client) GetRateLimit() (*RateLimitResponse, error) {
	req, err := http.NewRequest("POST", rateLimitURL, strings.NewReader(c.Payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.HC.Do(req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	intf := RateLimitResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}

// GetInformation is used to retrieve  general blockchain information
func (c *Client) GetInformation() (*InformationResponse, error) {
	url := fmt.Sprintf("%s/status?q=getInfo&token=%s", c.URL, c.Token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.HC.Do(req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	intf := InformationResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}

// TransactionByHash is used to get transaction information for a particular hash
func (c *Client) TransactionByHash(txHash string) (*TransactionByHashResponse, error) {
	url := fmt.Sprintf("%s/tx/%s?token=%s", c.URL, txHash, c.Token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.HC.Do(req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	intf := TransactionByHashResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}

// TransactionsForAddress is used to get a list of several transactions for a particular address
func (c *Client) TransactionsForAddress(address string) (*TransactionsForAddressResponse, error) {
	url := fmt.Sprintf("%s/txs?address=%s&token=%s", c.URL, address, c.Token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.HC.Do(req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	intf := TransactionsForAddressResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}

// BalanceForAddress returns the balance for an address in duffs
func (c *Client) BalanceForAddress(address string) (int, error) {
	url := fmt.Sprintf("%s/addr/%s/balance?token=%s", c.URL, address, c.Token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.HC.Do(req)
	if err != nil {
		return 0, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	responseString := string(bodyBytes)
	duffs, err := strconv.ParseInt(responseString, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(duffs), nil
}

func (c *Client) GetBlockByHash(blockHash string) (*BlockByHashResponse, error) {
	url := fmt.Sprintf("%s/block/%s?token=%s", c.URL, blockHash, c.Token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HC.Do(req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	intf := BlockByHashResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}

func (c *Client) GetLastBlockHash() (*LastBlockHashResponse, error) {
	url := fmt.Sprintf("%s/status?q=getLastBlockHash&token=%s", c.URL, c.Token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HC.Do(req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	intf := LastBlockHashResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}

func (c *Client) GetBlockchainSyncStatus() (*BlockchainDataSyncStatusResponse, error) {
	url := fmt.Sprintf("%s/sync?token=%s", c.URL, c.Token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HC.Do(req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	intf := BlockchainDataSyncStatusResponse{}
	if err = json.Unmarshal(bodyBytes, &intf); err != nil {
		return nil, err
	}
	return &intf, nil
}
