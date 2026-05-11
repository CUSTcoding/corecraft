package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"bitcoin-monitor-t3/internal/rpc"
	"bitcoin-monitor-t3/internal/tx"
	"bitcoin-monitor-t3/internal/wallet"
)


type TxHandlers struct {
	node    *rpc.NodeClient
	manager *wallet.Manager
	
	
	broadcastLog map[string]time.Time
}


func NewTxHandlers(node *rpc.NodeClient, mgr *wallet.Manager) *TxHandlers {
	return &TxHandlers{
		node:         node,
		manager:      mgr,
		broadcastLog: make(map[string]time.Time),
	}
}


func (h *TxHandlers) RecordBroadcast(txid string) {
	h.broadcastLog[txid] = time.Now()
}



func (h *TxHandlers) Get(c *gin.Context) {
	txid := c.Param("txid")
	if txid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "txid obrigatório"})
		return
	}

	walletName, err := h.manager.Active()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nenhuma wallet selecionada"})
		return
	}

	
	wc := h.node.ForWallet(walletName)
	walletTx, walletErr := wc.GetTransaction(txid)
	if walletErr != nil {
		
		walletTx = nil
	}

	
	memEntry, _ := h.node.GetMempoolEntry(txid)

	
	broadcastAt := h.broadcastLog[txid] 
	interp := tx.Interpret(txid, walletName, walletTx, memEntry, broadcastAt)

	c.JSON(http.StatusOK, interp)
}
