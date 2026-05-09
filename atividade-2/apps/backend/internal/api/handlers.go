package api


import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bitcoin-monitor/internal/rpc"
	"bitcoin-monitor/internal/store"
)


type Handlers struct {
	store *store.EventStore
	rpc   *rpc.Client
}


func NewHandlers(s *store.EventStore, r *rpc.Client) *Handlers {
	return &Handlers{store: s, rpc: r}
}


func (h *Handlers) Summary(c *gin.Context) {
	summary := h.store.GetSummary()
	c.JSON(http.StatusOK, summary)
}


func (h *Handlers) Latest(c *gin.Context) {
	latest := h.store.GetLatest(10, 20)
	c.JSON(http.StatusOK, latest)
}


func (h *Handlers) StateComparison(c *gin.Context) {
	// Estado atual via RPC
	bestBlock, err := h.rpc.GetBestBlockHash()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "falha ao consultar RPC: " + err.Error(),
		})
		return
	}

	// Último bloco observado via ZMQ
	lastSeen := h.store.LastSeenBlock()

	divergence := lastSeen != "" && lastSeen != bestBlock

	c.JSON(http.StatusOK, gin.H{
		"best_block":      bestBlock,
		"last_seen_block": lastSeen,
		"divergence":      divergence,
		"note":            divergenceNote(divergence, lastSeen),
	})
}

func divergenceNote(div bool, lastSeen string) string {
	if lastSeen == "" {
		return "nenhum bloco observado via ZMQ ainda"
	}
	if div {
		return "ZMQ e RPC estão fora de sincronia"
	}
	return "ZMQ e RPC estão sincronizados"
}


type NodeInfoResponse struct {
	Blockchain *rpc.BlockchainInfo `json:"blockchain"`
	Network    *rpc.NetworkInfo    `json:"network"`
	Mempool    *rpc.MempoolInfo    `json:"mempool"`
}

func (h *Handlers) NodeInfo(c *gin.Context) {
	chain, err := h.rpc.GetBlockchainInfo()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	net, err := h.rpc.GetNetworkInfo()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	mem, err := h.rpc.GetMempoolInfo()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, NodeInfoResponse{
		Blockchain: chain,
		Network:    net,
		Mempool:    mem,
	})
}