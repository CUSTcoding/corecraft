package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)


type Client struct {
	url  string
	user string
	pass string
	http *http.Client
}

func NewClient(url, user, pass string) *Client {
	return &Client{
		url:  url,
		user: user,
		pass: pass,
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

type rpcRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}


func (c *Client) Call(method string, params any) (json.RawMessage, error) {
	if params == nil {
		params = []any{}
	}

	body, err := json.Marshal(rpcRequest{
		Jsonrpc: "1.1",
		ID:      "btc-monitor",
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.pass)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rpc: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r rpcResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("rpc parse: %w", err)
	}
	if r.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", r.Error.Code, r.Error.Message)
	}

	return r.Result, nil
}




type BlockchainInfo struct {
	Chain                string  `json:"chain"`
	Blocks               int64   `json:"blocks"`
	Headers              int64   `json:"headers"`
	BestBlockHash        string  `json:"bestblockhash"`
	Difficulty           float64 `json:"difficulty"`
	VerificationProgress float64 `json:"verificationprogress"`
	MedianTime           int64   `json:"mediantime"`
}

func (c *Client) GetBlockchainInfo() (*BlockchainInfo, error) {
	raw, err := c.Call("getblockchaininfo", nil)
	if err != nil {
		return nil, err
	}
	var v BlockchainInfo
	return &v, json.Unmarshal(raw, &v)
}


type MempoolInfo struct {
	Loaded      bool    `json:"loaded"`
	Size        int64   `json:"size"`
	Bytes       int64   `json:"bytes"`
	Usage       int64   `json:"usage"`
	MaxMempool  int64   `json:"maxmempool"`
	MinFeeRate  float64 `json:"mempoolminfee"`
}

func (c *Client) GetMempoolInfo() (*MempoolInfo, error) {
	raw, err := c.Call("getmempoolinfo", nil)
	if err != nil {
		return nil, err
	}
	var v MempoolInfo
	return &v, json.Unmarshal(raw, &v)
}



type RawMempoolEntry struct {
	VSize   int64   `json:"vsize"`
	Weight  int64   `json:"weight"`
	Fee     float64 `json:"fee"`    
	ModFee  float64 `json:"modifiedfee"`
	Time    int64   `json:"time"`
	Height  int64   `json:"height"`
}



func (c *Client) GetRawMempoolVerbose() (map[string]RawMempoolEntry, error) {
	raw, err := c.Call("getrawmempool", []any{true})
	if err != nil {
		return nil, err
	}
	var v map[string]RawMempoolEntry
	return v, json.Unmarshal(raw, &v)
}