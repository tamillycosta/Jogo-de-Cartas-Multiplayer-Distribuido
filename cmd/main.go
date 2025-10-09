package main

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/cmd/api"
	gameserver "Jogo-de-Cartas-Multiplayer-Distribuido/cmd/game-server"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/config"
	"fmt"
	"log"
)

func main() {
	
	// seta configurações do gin
	router := config.SetUp()

	// 1. Inicializa Servidor de jogos
	gameServer,serverInfo,_ := gameserver.SetUpGame(router)
	
	handler := handlers.New(gameServer)

	// 2. Inicializa API de comunicação
	api.SetUpApi(router,serverInfo, handler)

	log.Printf("🎮 Server %s running on %d", serverInfo.ID, serverInfo.Port)
    log.Printf("📡 WebSocket: ws://localhost%d/ws", serverInfo.Port)
    log.Printf("🌐 REST API: http://localhost%d/api/v1", serverInfo.Port)
    log.Println("✅ Server ready!")

	// SET DA APLICAÇÃO 
    port := fmt.Sprintf(":%d", serverInfo.Port)
	router.Run(port)


	
}
