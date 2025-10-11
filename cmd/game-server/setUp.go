package gameserver

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/config"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	websocket "Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub/webSocket"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub/webSocket/topics"
	con "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/authService"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/discovery"
	handlers "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/authHandler"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	utils "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/util"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func SetUpServerBaseConfigs() *entities.ServerInfo {
	myServerInfo := &entities.ServerInfo{
		ID:      utils.GetEnv("SERVER_ID", "server-a"),
		Region:  utils.GetEnv("REGION", "us-east-1"),
		Address: utils.GetEnv("SERVER_ADDRESS", "localhost"),
		Port:    utils.GetPortFromEnv("PORT", 8080),
		Status:  "active",
	}
	return myServerInfo
}

func SetUpGame(router *gin.Engine) (*con.GameServer, *entities.ServerInfo, error) {
	myServerInfo := SetUpServerBaseConfigs()
	
	// 1. Inicializa Discovery
	discovery, err := discovery.SetUpDiscovery(myServerInfo)
	if err != nil {
		fmt.Printf("⚠️ Discovery error: %v\n", err)
	}

	// 2. Cria cliente API
	apiClient := client.New(5 * time.Second)
	
	// 3. Cria GameServer
	gameserver := con.New(myServerInfo, apiClient, discovery)

	// 4. Inicializa banco de dados e repositório
	db := config.CretaeTable()
	repository := repository.New(&db)

	// 5. Cria AuthService com referência aos servidores conhecidos
	authService := authService.New(repository, apiClient, discovery.KnownServers)
	
	// 6. Injeta AuthService no GameServer
	gameserver.InitAuth(authService)

	// 7. Configura pub/sub e WebSocket
	broker := pubsub.New()
	authHandler := handlers.New(authService, broker)
	wbSocket := websocket.New(broker)

	// 8. Configura tópicos
	topics.SetUpTopics(*wbSocket, authHandler)

	// 9. Rota WebSocket (cliente -> servidor)
	router.GET("/ws", func(c *gin.Context) {
		wbSocket.SetWebSocket(c.Writer, c.Request)
	})

	return gameserver, myServerInfo, nil
}