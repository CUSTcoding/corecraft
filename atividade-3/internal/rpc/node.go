package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)



type NodeClient struct {
	baseURL  string
	user     string
	password string
	http     *http.Client
}

func NewNodeClient(baseURL, user, password string) *NodeClient {
	return &NodeClient{
		baseURL:  baseURL,
		user:     user,
		password: password,
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}


func (c *NodeClient) ForWallet(name string) *WalletClient {
	return &WalletClient{
		url:     fmt.Sprintf("%s/wallet/%s", c.baseURL, name),
		user:    c.user,
		pass:    c.password,
		httpCli: c.http,
	}
}



type rpcReq struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type rpcResp struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcErr         `json:"error"`
}

type rpcErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func call(httpClient *http.Client, url, user, password, method string, params any) (json.RawMessage, error) {
	if params == nil {
		params = []any{}
	}
	body, _ := json.Marshal(rpcReq{Jsonrpc: "1.1", ID: "btc-monitor", Method: method, Params: params})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(user, password)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rpc: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var r rpcResp
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("rpc parse: %w", err)
	}
	if r.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", r.Error.Code, r.Error.Message)
	}
	return r.Result, nil
}

func (c *NodeClient) call(method string, params any) (json.RawMessage, error) {
	return call(c.http, c.baseURL, c.user, c.password, method, params)
}



type BlockchainInfo struct {
	Chain                string  `json:"chain"`
	Blocks               int64   `json:"blocks"`
	Headers              int64   `json:"headers"`
	BestBlockHash        string  `json:"bestblockhash"`
	Difficulty           float64 `json:"difficulty"`
	VerificationProgress float64 `json:"verificationprogress"`
}

func (c *NodeClient) GetBlockchainInfo() (*BlockchainInfo, error) {
	raw, err := c.call("getblockchaininfo", nil)
	if err != nil {
		return nil, err
	}
	var v BlockchainInfo
	return &v, json.Unmarshal(raw, &v)
}

func (c *NodeClient) GetBestBlockHash() (string, error) {
	raw, err := c.call("getbestblockhash", nil)
	if err != nil {
		return "", err
	}
	var s string
	return s, json.Unmarshal(raw, &s)
}

type NetworkInfo struct {
	Version     int    `json:"version"`
	Subversion  string `json:"subversion"`
	Connections int    `json:"connections"`
}

func (c *NodeClient) GetNetworkInfo() (*NetworkInfo, error) {
	raw, err := c.call("getnetworkinfo", nil)
	if err != nil {
		return nil, err
	}
	var v NetworkInfo
	return &v, json.Unmarshal(raw, &v)
}

type MempoolInfo struct {
	Size  int64   `json:"size"`
	Bytes int64   `json:"bytes"`
	Usage int64   `json:"usage"`
}

func (c *NodeClient) GetMempoolInfo() (*MempoolInfo, error) {
	raw, err := c.call("getmempoolinfo", nil)
	if err != nil {
		return nil, err
	}
	var v MempoolInfo
	return &v, json.Unmarshal(raw, &v)
}


func (c *NodeClient) SendRawTransaction(hex string) (string, error) {
	raw, err := c.call("sendrawtransaction", []any{hex})
	if err != nil {
		return "", err
	}
	var txid string
	return txid, json.Unmarshal(raw, &txid)
}


type MempoolEntry struct {
	Time    int64   `json:"time"`
	Fee     float64 `json:"fee"`
	VSize   int64   `json:"vsize"`
	Height  int64   `json:"height"`
}

func (c *NodeClient) GetMempoolEntry(txid string) (*MempoolEntry, error) {
	raw, err := c.call("getmempoolentry", []any{txid})
	if err != nil {
		return nil, err
	}
	var v MempoolEntry
	return &v, json.Unmarshal(raw, &v)
}




func (c *NodeClient) ListWalletDir() ([]string, error) {
	raw, err := c.call("listwalletdir", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Wallets []struct {
			Name string `json:"name"`
		} `json:"wallets"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	names := make([]string, len(resp.Wallets))
	for i, w := range resp.Wallets {
		names[i] = w.Name
	}
	return names, nil
}


func (c *NodeClient) ListWallets() ([]string, error) {
	raw, err := c.call("listwallets", nil)
	if err != nil {
		return nil, err
	}
	var v []string
	return v, json.Unmarshal(raw, &v)
}


func (c *NodeClient) LoadWallet(name string) error {
	_, err := c.call("loadwallet", []any{name})
	if err != nil {
		
		
		if rpcAlreadyLoaded(err) {
			return nil
		}
		return err
	}
	return nil
}

func rpcAlreadyLoaded(err error) bool {
	return err != nil && (contains(err.Error(), "-35") || contains(err.Error(), "already loaded"))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
