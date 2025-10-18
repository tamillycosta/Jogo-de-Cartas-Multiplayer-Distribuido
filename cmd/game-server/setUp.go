package gameserver

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/config"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler"
	authhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/authHandler"

	packgehandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/packgeHandler"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	con "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/authService"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/discovery"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/packageService"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	seedService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/seed"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	websocket "Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub/webSocket"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub/webSocket/topics"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	utils "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/util"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func SetUpServerBaseConfigs() *entities.ServerInfo {
	myServerInfo := &entities.ServerInfo{
		ID:       utils.GetEnv("SERVER_ID", "server-a"),
		Region:   utils.GetEnv("REGION", "us-east-1"),
		Address:  utils.GetEnv("SERVER_ADDRESS", "localhost"),
		Port:     utils.GetPortFromEnv("PORT", 8080),
		IsLeader: utils.GetEnvBool("RAFT_BOOTSTRAP", false),
		Status:   "active",
	}
	return myServerInfo
}

func SetUpGame(router *gin.Engine) (*con.GameServer, *entities.ServerInfo, error) {

	// ---------- SETA CONFUGURAÇÕES BASICAS PARA SERVIDOR P2P -----------
	myServerInfo := SetUpServerBaseConfigs()

	// Inicializa Discovery
	discovery, err := discovery.SetUpDiscovery(myServerInfo)
	if err != nil {
		fmt.Printf("Discovery error: %v\n", err)
	}
	log.Printf("%d servidores conhecidos", len(discovery.KnownServers))

	apiClient := client.New(5 * time.Second)
	gameserver := con.New(myServerInfo, apiClient, discovery)

	// Inicializa banco de dados e repositório
	db := config.CretaeTable()
	playerRepo := repository.NewPlayerRepository(&db)
	packageRepo := repository.NewPackageRepository(&db)
	cardRepo := repository.NewCardRepository(&db)

	// --------- INICIALIZA E INJETA SERVIÇOS DO SERVIDOR P2P -----------
	raftService, _ := raft.InitRaft(playerRepo, packageRepo, cardRepo, myServerInfo, apiClient)
	authService := authService.New(playerRepo, apiClient, discovery.KnownServers, raftService, gameserver.SessionManager)
	pkgService := packageService.New(packageRepo, cardRepo, raftService, gameserver.SessionManager)
	seedSvc := seedService.New(raftService, pkgService)

	gameserver.InitAuth(authService)
	gameserver.InitRaft(raftService)
	gameserver.InitPackageSystem(pkgService)
	gameserver.InitSeeds(seedSvc)

	// BUSCA LEADER DO RAFT (SE FOR O BOOTSTRAP CRIA O CLUSTER)
	log.Println("Aguardando eleição de líder...")
	if err := raftService.WaitForLeader(10 * time.Second); err != nil {
		log.Printf("Timeout na eleição: %v (pode ser normal se for o primeiro servidor)", err)
	}

	raftService.TryJoinCluster(discovery, myServerInfo)

	//  Levanta os seeds do banco de dados apos ter um lider (apenas lider pode subir os seeds)
	seedSvc.Init(packageRepo.GetAll())

	// --------------- CONFIGURA PUB SUB E WEBSOCKETS -----------------------
	broker := pubsub.New()
	authHandler := authhandler.New(authService, broker)
	packgehandler := packgehandler.New(pkgService, broker)

	// injeta todos os handlers da aplicação para o pub sub
	handler := handler.New(authHandler, packgehandler)
	wbSocket := websocket.New(broker, gameserver.SessionManager)
	topics.SetUpTopics(*wbSocket, handler)

	//  Rota WebSocket (cliente -> servidor)
	router.GET("/ws", func(c *gin.Context) {
		wbSocket.SetWebSocket(c.Writer, c.Request)
	})

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if raftService.IsLeader() {
		log.Println("👑 Status: LÍDER")
	} else {
		log.Printf("👤 Status: FOLLOWER (Líder: %s)", raftService.GetLeaderID())
	}

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return gameserver, myServerInfo, nil
}
