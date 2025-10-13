package gameserver

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/config"
	handlers "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/authHandler"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	con "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/authService"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/discovery"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	websocket "Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub/webSocket"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub/webSocket/topics"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	utils "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/util"
	"fmt"
	"time"
	"log"
	"github.com/gin-gonic/gin"
)




func SetUpServerBaseConfigs() *entities.ServerInfo {
	myServerInfo := &entities.ServerInfo{
		ID:      utils.GetEnv("SERVER_ID", "server-a"),
		Region:  utils.GetEnv("REGION", "us-east-1"),
		Address: utils.GetEnv("SERVER_ADDRESS", "localhost"),
		Port:    utils.GetPortFromEnv("PORT", 8080),
		IsLeader: utils.GetEnvBool("RAFT_BOOTSTRAP",false),
		Status:  "active",
	}
	return myServerInfo
}

func SetUpGame(router *gin.Engine) (*con.GameServer, *entities.ServerInfo, error) {

	// ---------- SETA CONFUGURAÇÕES BASICAS PARA SERVIDOR P2P -----------
	myServerInfo := SetUpServerBaseConfigs()
	
	
	// Inicializa Discovery
	discovery, err := discovery.SetUpDiscovery(myServerInfo)
	if err != nil {
		fmt.Printf("⚠️ Discovery error: %v\n", err)
	}
	log.Printf("%d servidores conhecidos", len(discovery.KnownServers))

	

	// seta interface de cliente do servidor p2p
	apiClient := client.New(5 * time.Second)
	gameserver := con.New(myServerInfo, apiClient, discovery)
	// Inicializa banco de dados e repositório
	db := config.CretaeTable()
	repository := repository.New(&db)

	
	
	// --------- INICIALIZA E INJETA SERVIÇOS DO SERVIDOR P2P -----------
	raftService, _ := raft.InitRaft(repository,myServerInfo, apiClient)
	authService := authService.New(repository, apiClient, discovery.KnownServers, raftService)
	

	gameserver.InitAuth(authService)
	gameserver.InitRaft(raftService)

	// BUSCA LEADER DO RAFT (SE FOR O BOOTSTRAP CRIA O CLUSTER)
	log.Println("⏳ Aguardando eleição de líder...")
	if err := raftService.WaitForLeader(10 * time.Second); err != nil {
		log.Printf("⚠️ Timeout na eleição: %v (pode ser normal se for o primeiro servidor)", err)
	}

	raftService.TryJoinCluster(discovery,myServerInfo)



	// --------------- CONFIGURA PUB SUB E WEBSOCKETS -----------------------
	broker := pubsub.New()
	authHandler := handlers.New(authService, broker)
	wbSocket := websocket.New(broker)

	
	topics.SetUpTopics(*wbSocket, authHandler)

	//  Rota WebSocket (cliente -> servidor)
	router.GET("/ws", func(c *gin.Context) {
		wbSocket.SetWebSocket(c.Writer, c.Request)
	})

	return gameserver, myServerInfo, nil
}