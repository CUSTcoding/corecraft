package main

import (
	"github.com/gin-gonic/gin"
	"corecraft-atividade-1/rpc"
	"corecraft-atividade-1/handlers"
	"log"
	"os"
	"github.com/joho/godotenv"
)


func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar .env")
	}

	rpcURL := os.Getenv("RPC_URL")
	rpcUser := os.Getenv("RPC_USER")
	rpcPassword := os.Getenv("RPC_PASSWORD")

	client := &rpc.RPCClient{
		URL:      rpcURL,
		Username: rpcUser,
		Password: rpcPassword,
	}

	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/mempool/summary", handlers.MempoolSummaryHandler(client))
		api.GET("/blockchain/lag", handlers.BlockchainLagHandler(client))
	}

	r.Run(":8000")
}