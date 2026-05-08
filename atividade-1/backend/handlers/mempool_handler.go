package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"corecraft-atividade-1/services"
	"corecraft-atividade-1/rpc"
)

func MempoolSummaryHandler(client *rpc.RPCClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := services.GetMempoolSummary(client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	}
}