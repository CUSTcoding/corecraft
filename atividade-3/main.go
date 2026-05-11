package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"bitcoin-monitor-t3/internal/api"
	"bitcoin-monitor-t3/internal/rpc"
	"bitcoin-monitor-t3/internal/store"
	"bitcoin-monitor-t3/internal/wallet"
	"bitcoin-monitor-t3/internal/zmqsub"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[config] .env não encontrado, usando variáveis do sistema")
	}

	rpcHost := env("BTC_RPC_HOST", "http://127.0.0.1:8332")
	rpcUser := env("BTC_RPC_USER", "bitcoinrpc")
	rpcPass := env("BTC_RPC_PASS", "senha_rpc")
	zmqURL  := env("ZMQ_URL", "tcp://127.0.0.1:28332")
	port    := env("PORT", "8000")

	log.Printf("[config] RPC → %s | ZMQ → %s | HTTP :%s", rpcHost, zmqURL, port)

	
	nodeClient  := rpc.NewNodeClient(rpcHost, rpcUser, rpcPass)
	eventStore  := store.New(50, 200)
	walletMgr   := wallet.New("")

	
	go autoSelectWallet(nodeClient, walletMgr)

	
	go zmqsub.Start(zmqURL, eventStore)

	
	nodeH   := api.NewNodeHandlers(nodeClient, eventStore)
	walletH := api.NewWalletHandlers(nodeClient, walletMgr)
	txH     := api.NewTxHandlers(nodeClient, walletMgr)

	
	r := gin.Default()

	r.Use(corsMiddleware())

	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	
	apiG := r.Group("/api")
	{
		apiG.GET("/node/info", nodeH.Info)
		apiG.GET("/events/summary", nodeH.EventSummary)
		apiG.GET("/events/latest", nodeH.EventLatest)
		apiG.GET("/events/state-comparison", nodeH.StateComparison)
	}

	
	r.GET("/wallets", walletH.List)
	r.POST("/wallet/select", walletH.Select)
	r.GET("/wallet/status", walletH.Status)

	
	r.GET("/tx/:txid", txH.Get)

	log.Printf("[server] Iniciando em :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}


func autoSelectWallet(node *rpc.NodeClient, mgr *wallet.Manager) {
	wallets, err := node.ListWallets()
	if err != nil || len(wallets) == 0 {
		return
	}
	mgr.Set(wallets[0])
	log.Printf("[wallet] Auto-selecionada: %s", wallets[0])
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
