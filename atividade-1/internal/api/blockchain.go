package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bitcoin-monitor-t1/internal/rpc"
)


type BlockchainHandlers struct {
	rpc *rpc.Client
}

func NewBlockchainHandlers(c *rpc.Client) *BlockchainHandlers {
	return &BlockchainHandlers{rpc: c}
}






func (h *BlockchainHandlers) Lag(c *gin.Context) {
	info, err := h.rpc.GetBlockchainInfo()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "falha ao obter blockchaininfo: " + err.Error(),
		})
		return
	}

	lag := info.Headers - info.Blocks
	synced := lag == 0

	
	status := interpretSyncStatus(lag, info.VerificationProgress)

	c.JSON(http.StatusOK, gin.H{
		"blocks":               info.Blocks,
		"headers":              info.Headers,
		"lag":                  lag,
		"synced":               synced,
		"verification_progress": roundPct(info.VerificationProgress),
		"best_block_hash":      info.BestBlockHash,
		"chain":                info.Chain,
		"status":               status,
	})
}




func (h *BlockchainHandlers) Info(c *gin.Context) {
	info, err := h.rpc.GetBlockchainInfo()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}


func interpretSyncStatus(lag int64, progress float64) string {
	switch {
	case lag == 0:
		return "✅ Node sincronizado"
	case lag < 10:
		return "🔄 Quase sincronizado — poucos blocos restantes"
	case lag < 144: 
		return "⏳ Sincronizando — menos de 1 dia atrás"
	case lag < 2016: 
		return "⚠️ Sincronização parcial — dias de atraso"
	default:
		return "🚨 Node muito desatualizado — sincronização necessária"
	}
}

func roundPct(v float64) float64 {
	return float64(int(v*10000)) / 100
}