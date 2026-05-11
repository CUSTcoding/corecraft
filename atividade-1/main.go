package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"bitcoin-monitor-t1/internal/api"
	"bitcoin-monitor-t1/internal/rpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[config] .env não encontrado, usando variáveis do sistema")
	}

	rpcURL  := env("BTC_RPC_HOST", "http://127.0.0.1:8332")
	rpcUser := env("BTC_RPC_USER", "bitcoinrpc")
	rpcPass := env("BTC_RPC_PASS", "senha_rpc")
	port    := env("PORT", "8000")

	log.Printf("[config] RPC → %s | HTTP :%s", rpcURL, port)

	
	rpcClient := rpc.NewClient(rpcURL, rpcUser, rpcPass)

	
	mempoolH    := api.NewMempoolHandlers(rpcClient)
	blockchainH := api.NewBlockchainHandlers(rpcClient)

	
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	
	apiG := r.Group("/api")
	{
		apiG.GET("/mempool/summary",   mempoolH.Summary)
		apiG.GET("/blockchain/lag",    blockchainH.Lag)
		apiG.GET("/blockchain/info",   blockchainH.Info)
	}

	log.Printf("[server] Iniciando em :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Erro: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}