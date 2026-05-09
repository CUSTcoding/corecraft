package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"bitcoin-monitor/internal/api"
	"bitcoin-monitor/internal/rpc"
	"bitcoin-monitor/internal/store"
	"bitcoin-monitor/internal/zmqsub"
)

func main() {
	
	if err := godotenv.Load(); err != nil {
		log.Println("[config] Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	rpcHost := getEnv("BTC_RPC_HOST", "http://127.0.0.1:8332")
	rpcUser := getEnv("BTC_RPC_USER", "bitcoinrpc")
	rpcPass := getEnv("BTC_RPC_PASS", "senha_rpc")
	zmqURL  := getEnv("ZMQ_URL", "tcp://127.0.0.1:28332")
	port    := getEnv("PORT", "8000")

	log.Printf("[config] RPC → %s | ZMQ → %s | HTTP :%s", rpcHost, zmqURL, port)

	
	eventStore := store.New(50, 200) // até 50 blocos e 200 txs em memória
	rpcClient  := rpc.NewClient(rpcHost, rpcUser, rpcPass)


	go zmqsub.Start(zmqURL, eventStore)

	
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

	// Serve o frontend estático
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	// Rotas da API
	h := api.NewHandlers(eventStore, rpcClient)
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/events/summary", h.Summary)
		apiGroup.GET("/events/latest", h.Latest)
		apiGroup.GET("/events/state-comparison", h.StateComparison)
		apiGroup.GET("/node/info", h.NodeInfo)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("[server] Iniciando em :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}