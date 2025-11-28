package main

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/cmd/api"
	gameserver "Jogo-de-Cartas-Multiplayer-Distribuido/cmd/game-server"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	contracts "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/service"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/config"
	"fmt"
	"log"
	"os"
)

func main() {

	// Configurações do Gin
	router := config.SetUp()

	// ===== INICIALIZAR BLOCKCHAIN =====
	log.Println("🔗 Inicializando Blockchain...")

	blockchainClient, err := blockchain.NewBlockchainClient(blockchain.Config{

		RPC:        getEnv("RPC_URL", "http://127.0.0.1:7545"),
		PrivateKey: getEnv("PK", ""), // tem que pssar uma chave privada como parameto no executavel
		ChainID:    1337,
	})

	if err != nil {
		log.Printf("  BLOCKCHAIN INDISPONÍVEL: %v", err)
		log.Println("  Continuando SEM blockchain (modo compatibilidade)")
		blockchainClient = nil
	} else {
		log.Println(" Blockchain conectado!")
	}

	var contractService *contracts.ChainService
	if blockchainClient != nil {
		contractService = contracts.New(blockchainClient)
	}

	// ===== INICIALIZAR GAME SERVER (passar blockchain) =====
	if contractService != nil {
		log.Println("  Blockchain: ATIVO")
	} else {
		log.Println("  Blockchain: DESATIVADO")
	}
	gameServer, gameSessionManager, serverInfo, _ := gameserver.SetUpGame(router, contractService)

	handler := handlers.New(gameServer, gameSessionManager)

	// Rotas
	router.StaticFile("/test_trade", "./web/test_trade.html")

	// Inicializar API de comunicação
	api.SetUpApi(router, serverInfo, handler)

	// Logs de inicialização
	log.Printf("🎮 Server %s running on %d", serverInfo.ID, serverInfo.Port)
	log.Printf("📡 WebSocket: ws://localhost:%d/ws", serverInfo.Port)
	log.Printf("🌐 REST API: http://localhost:%d/api/v1", serverInfo.Port)

	log.Println("✅ Server ready!")

	// Iniciar servidor
	port := fmt.Sprintf(":%d", serverInfo.Port)
	router.Run(port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
