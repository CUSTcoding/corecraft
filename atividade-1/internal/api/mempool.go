package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bitcoin-monitor-t1/internal/mempool"
	"bitcoin-monitor-t1/internal/rpc"
)


type MempoolHandlers struct {
	rpc *rpc.Client
}

func NewMempoolHandlers(c *rpc.Client) *MempoolHandlers {
	return &MempoolHandlers{rpc: c}
}






func (h *MempoolHandlers) Summary(c *gin.Context) {
	
	info, err := h.rpc.GetMempoolInfo()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "falha ao obter mempoolinfo: " + err.Error(),
		})
		return
	}

	
	entries, err := h.rpc.GetRawMempoolVerbose()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "falha ao obter rawmempool: " + err.Error(),
		})
		return
	}

	
	summary := mempool.Analyze(entries)

	
	
	if info.Bytes > 0 && summary.TotalVsize == 0 {
		summary.TotalVsize = info.Bytes
	}

	c.JSON(http.StatusOK, summary)
}