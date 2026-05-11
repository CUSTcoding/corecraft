package rpc

import (
	"encoding/json"
	"net/http"
)



type WalletClient struct {
	url     string
	user    string
	pass    string
	httpCli *http.Client
}

func (c *WalletClient) call(method string, params any) (json.RawMessage, error) {
	return call(c.httpCli, c.url, c.user, c.pass, method, params)
}



type WalletInfo struct {
	WalletName  string  `json:"walletname"`
	Balance     float64 `json:"balance"`
	TxCount     int     `json:"txcount"`
	KeyPoolSize int     `json:"keypoolsize"`
}

func (c *WalletClient) GetWalletInfo() (*WalletInfo, error) {
	raw, err := c.call("getwalletinfo", nil)
	if err != nil {
		return nil, err
	}
	var v WalletInfo
	return &v, json.Unmarshal(raw, &v)
}



type UTXO struct {
	Txid          string  `json:"txid"`
	Vout          int     `json:"vout"`
	Address       string  `json:"address"`
	Amount        float64 `json:"amount"`
	Confirmations int64   `json:"confirmations"`
	Spendable     bool    `json:"spendable"`
}

func (c *WalletClient) ListUnspent() ([]UTXO, error) {
	raw, err := c.call("listunspent", nil)
	if err != nil {
		return nil, err
	}
	var v []UTXO
	return v, json.Unmarshal(raw, &v)
}



type WalletTx struct {
	Txid          string  `json:"txid"`
	Confirmations int64   `json:"confirmations"`
	BlockHash     string  `json:"blockhash"`
	BlockTime     int64   `json:"blocktime"`
	Time          int64   `json:"time"`
	TimeReceived  int64   `json:"timereceived"`
	Amount        float64 `json:"amount"`
	Fee           float64 `json:"fee"`
}

func (c *WalletClient) GetTransaction(txid string) (*WalletTx, error) {
	raw, err := c.call("gettransaction", []any{txid})
	if err != nil {
		return nil, err
	}
	var v WalletTx
	return &v, json.Unmarshal(raw, &v)
}

func (c *WalletClient) GetRawChangeAddress() (string, error) {
	raw, err := c.call("getrawchangeaddress", nil)
	if err != nil {
		return "", err
	}
	var addr string
	return addr, json.Unmarshal(raw, &addr)
}

func (c *WalletClient) SignRawTransactionWithWallet(hexTx string) (string, bool, error) {
	raw, err := c.call("signrawtransactionwithwallet", []any{hexTx})
	if err != nil {
		return "", false, err
	}
	var resp struct {
		Hex      string `json:"hex"`
		Complete bool   `json:"complete"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", false, err
	}
	return resp.Hex, resp.Complete, nil
}
