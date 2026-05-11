package tx

import (
	"time"

	"bitcoin-monitor-t3/internal/rpc"
)


type Status string

const (
	StatusBroadcast Status = "broadcast"  
	StatusMempool   Status = "mempool"    
	StatusConfirmed Status = "confirmed"  
	StatusUnknown   Status = "unknown"    
)


const MempoolSlowThreshold = 2 * time.Minute


type Interpretation struct {
	Txid          string  `json:"txid"`
	Wallet        string  `json:"wallet"`
	Status        Status  `json:"status"`
	Confirmed     bool    `json:"confirmed"`
	Confirmations int64   `json:"confirmations"`
	BlockHash     *string `json:"block_hash"`
	AgeSeconds    int64   `json:"age_seconds"`
	Message       string  `json:"message"`
	Warning       string  `json:"warning,omitempty"`
}









func Interpret(
	txid string,
	walletName string,
	walletTx *rpc.WalletTx,
	memEntry *rpc.MempoolEntry,
	broadcastAt time.Time,
) Interpretation {
	result := Interpretation{
		Txid:   txid,
		Wallet: walletName,
	}

	
	if walletTx == nil && memEntry == nil {
		result.Status = StatusUnknown
		result.Warning = "Transação não localizada na wallet selecionada."
		return result
	}

	
	if walletTx != nil && walletTx.Confirmations > 0 {
		blockHash := walletTx.BlockHash
		result.Status = StatusConfirmed
		result.Confirmed = true
		result.Confirmations = walletTx.Confirmations
		result.BlockHash = &blockHash
		result.Message = "Transação confirmada em bloco."

		if walletTx.BlockTime > 0 {
			result.AgeSeconds = time.Now().Unix() - walletTx.BlockTime
		}
		return result
	}

	
	if memEntry != nil {
		result.Status = StatusMempool
		result.Confirmed = false
		result.Confirmations = 0
		result.Message = "Transação aceita na mempool, aguardando inclusão em bloco."

		
		entryTime := time.Unix(memEntry.Time, 0)
		age := time.Since(entryTime)
		result.AgeSeconds = int64(age.Seconds())

		if age > MempoolSlowThreshold {
			result.Warning = "Transação está na mempool há mais de 2 minutos."
		}
		return result
	}

	
	result.Status = StatusBroadcast
	result.Message = "Transação enviada ao node, aguardando aceitação na mempool."

	if !broadcastAt.IsZero() {
		result.AgeSeconds = int64(time.Since(broadcastAt).Seconds())
	}

	return result
}
