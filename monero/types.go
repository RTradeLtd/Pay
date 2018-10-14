package monero

// BlockCountResponse is a resposnse from the get block count rpc call
type BlockCountResponse struct {
	ID      string `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Result  struct {
		Count  int    `json:"count"`
		Status string `json:"status"`
	} `json:"result"`
}
