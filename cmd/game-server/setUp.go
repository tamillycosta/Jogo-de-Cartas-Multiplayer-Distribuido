package gameserver

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/authHandler"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/config"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	websocket "Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub/webSocket"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub/webSocket/topics"

	con "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	"github.com/gin-gonic/gin"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/authService"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/discovery"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	utils "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/util"
	"fmt"
	"time"
)

func SetUpServerBaseConfigs() (*entities.ServerInfo){
	// Configuração do servidor
	myServerInfo := &entities.ServerInfo{
		ID:      utils.GetEnv("SERVER_ID", "server-a"),
		Region:  utils.GetEnv("REGION", "us-east-1"),
		Address: utils.GetEnv("SERVER_ADDRESS", "server-a"),
		Port:    utils.GetPortFromEnv("PORT", 8080),
		Status:  "active",
	}
	return myServerInfo
}




func SetUpGame(router *gin.Engine) (*con.GameServer, *entities.ServerInfo , error) {
	myServerInfo := SetUpServerBaseConfigs()
	discovery , err := discovery.SetUpDiscovery(myServerInfo)
	if(err!= nil){
		fmt.Printf("%v",err)
	}

	// interface cliente do servidor de jogo (server to server)
	apiClient := client.New(5 * time.Second)
	// servidor do jogo 
	gameserver := con.New(myServerInfo, apiClient, discovery)


	// INJEÇÃO DE DEPENDENCIA
	db := config.CretaeTable()
	// DEPOIS MUDAR PARA INTERFACES
	repository := repository.New(&db)
	authService := authService.New(repository,apiClient,discovery.KnownServers)
	
	broker := pubsub.New()
	authHandler := handlers.New(authService,broker)
	wbSocket := websocket.New(broker)

	// servidor do jogo 
	topics.SetUpTopics(*wbSocket,authHandler)

	// rota WebSocket
	// cliete _-> servidor
	router.GET("/ws", func(c *gin.Context) {
        wbSocket.SetWebSocket(c.Writer, c.Request)
    }) 
	return  gameserver, myServerInfo,  nil
}