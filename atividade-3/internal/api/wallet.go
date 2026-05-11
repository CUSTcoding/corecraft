package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bitcoin-monitor-t3/internal/rpc"
	"bitcoin-monitor-t3/internal/wallet"
)


type WalletHandlers struct {
	node    *rpc.NodeClient
	manager *wallet.Manager
}


func NewWalletHandlers(node *rpc.NodeClient, mgr *wallet.Manager) *WalletHandlers {
	return &WalletHandlers{node: node, manager: mgr}
}


func (h *WalletHandlers) List(c *gin.Context) {
	available, err := h.node.ListWalletDir()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	loaded, err := h.node.ListWallets()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	selected, _ := h.manager.Active()
	c.JSON(http.StatusOK, gin.H{
		"available_wallets": available,
		"loaded_wallets":    loaded,
		"selected_wallet":   selected,
	})
}


func (h *WalletHandlers) Select(c *gin.Context) {
	var body struct {
		Wallet string `json:"wallet" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'wallet' obrigatório"})
		return
	}

	available, err := h.node.ListWalletDir()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	if !sliceContains(available, body.Wallet) {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet não encontrada: " + body.Wallet})
		return
	}
	if err := h.node.LoadWallet(body.Wallet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao carregar wallet: " + err.Error()})
		return
	}

	h.manager.Set(body.Wallet)

	info, err := h.node.ForWallet(body.Wallet).GetWalletInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "wallet selecionada mas erro ao obter info: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"selected_wallet": body.Wallet,
		"wallet_info": gin.H{
			"walletname": info.WalletName,
			"balance":    info.Balance,
			"txcount":    info.TxCount,
		},
	})
}


func (h *WalletHandlers) Status(c *gin.Context) {
	name, err := h.manager.Active()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nenhuma wallet selecionada"})
		return
	}
	wc := h.node.ForWallet(name)
	info, err := wc.GetWalletInfo()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	utxos, err := wc.ListUnspent()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"wallet":  name,
		"balance": info.Balance,
		"utxos":   len(utxos),
	})
}

func sliceContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
