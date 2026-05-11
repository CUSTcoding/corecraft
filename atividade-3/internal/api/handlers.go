package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bitcoin-monitor-t3/internal/rpc"
	"bitcoin-monitor-t3/internal/store"
)


type NodeHandlers struct {
	node  *rpc.NodeClient
	store *store.EventStore
}

func NewNodeHandlers(node *rpc.NodeClient, s *store.EventStore) *NodeHandlers {
	return &NodeHandlers{node: node, store: s}
}


func (h *NodeHandlers) Info(c *gin.Context) {
	chain, err := h.node.GetBlockchainInfo()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	net, err := h.node.GetNetworkInfo()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	mem, err := h.node.GetMempoolInfo()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"blockchain": chain,
		"network":    net,
		"mempool":    mem,
	})
}


func (h *NodeHandlers) EventSummary(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.GetSummary())
}


func (h *NodeHandlers) EventLatest(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.GetLatest(10, 20))
}


func (h *NodeHandlers) StateComparison(c *gin.Context) {
	best, err := h.node.GetBestBlockHash()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	lastSeen := h.store.LastSeenBlock()
	divergence := lastSeen != "" && lastSeen != best

	note := "✅ ZMQ e RPC estão sincronizados"
	if lastSeen == "" {
		note = "nenhum bloco observado via ZMQ ainda"
	} else if divergence {
		note = "⚠️ ZMQ e RPC estão fora de sincronia"
	}

	c.JSON(http.StatusOK, gin.H{
		"best_block":      best,
		"last_seen_block": lastSeen,
		"divergence":      divergence,
		"note":            note,
	})
}
