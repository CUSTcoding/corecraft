package rpc


import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client é o cliente JSON-RPC para o Bitcoin Core
type Client struct {
	url      string
	user     string
	password string
	http     *http.Client
}


func NewClient(host, user, password string) *Client {
	if host == "" {
		host = "http://127.0.0.1:8332"
	}
	return &Client{
		url:      host,
		user:     user,
		password: password,
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}

type rpcRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
	ID     string          `json:"id"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) call(method string, params []interface{}) (json.RawMessage, error) {
	req := rpcRequest{
		Jsonrpc: "1.1",
		ID:      "bitcoin-monitor",
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(c.user, c.password)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("rpc request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse rpc response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}


func (c *Client) GetBestBlockHash() (string, error) {
	result, err := c.call("getbestblockhash", nil)
	if err != nil {
		return "", err
	}
	var hash string
	if err := json.Unmarshal(result, &hash); err != nil {
		return "", err
	}
	return hash, nil
}


type BlockchainInfo struct {
	Chain                string  `json:"chain"`
	Blocks               int64   `json:"blocks"`
	Headers              int64   `json:"headers"`
	BestBlockHash        string  `json:"bestblockhash"`
	Difficulty           float64 `json:"difficulty"`
	MedianTime           int64   `json:"mediantime"`
	VerificationProgress float64 `json:"verificationprogress"`
	Pruned               bool    `json:"pruned"`
}


func (c *Client) GetBlockchainInfo() (*BlockchainInfo, error) {
	result, err := c.call("getblockchaininfo", nil)
	if err != nil {
		return nil, err
	}
	var info BlockchainInfo
	if err := json.Unmarshal(result, &info); err != nil {
		return nil, err
	}
	return &info, nil
}


type NetworkInfo struct {
	Version         int    `json:"version"`
	Subversion      string `json:"subversion"`
	ProtocolVersion int    `json:"protocolversion"`
	Connections     int    `json:"connections"`
	RelayFee        float64 `json:"relayfee"`
}


func (c *Client) GetNetworkInfo() (*NetworkInfo, error) {
	result, err := c.call("getnetworkinfo", nil)
	if err != nil {
		return nil, err
	}
	var info NetworkInfo
	if err := json.Unmarshal(result, &info); err != nil {
		return nil, err
	}
	return &info, nil
}


type MempoolInfo struct {
	Loaded  bool    `json:"loaded"`
	Size    int64   `json:"size"`
	Bytes   int64   `json:"bytes"`
	Usage   int64   `json:"usage"`
	MaxMem  int64   `json:"maxmempool"`
	MinFee  float64 `json:"mempoolminfee"`
}


func (c *Client) GetMempoolInfo() (*MempoolInfo, error) {
	result, err := c.call("getmempoolinfo", nil)
	if err != nil {
		return nil, err
	}
	var info MempoolInfo
	if err := json.Unmarshal(result, &info); err != nil {
		return nil, err
	}
	return &info, nil
}