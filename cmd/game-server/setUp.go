package gameserver

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/config"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler"
	authhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/authHandler"
	matchhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/matchHandler"
	contracts "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/service"
	packgehandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/packgeHandler"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	con "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/authService"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/discovery"
	gamesession "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/gameSession"
	
	inventoryhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/inventoryHandler"
	matchglobal "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/matchMacking/match_global"
	matchlocal "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/matchMacking/match_local"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/packageService"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	seedService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/seed"
	tradeService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/tradeService"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	websocket "Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub/webSocket"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub/webSocket/topics"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	tradehandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/tradeHandler"
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
		IsLeader: utils.GetEnvBool("RAFT_BOOTSTRAP", true),
		Status:   "active",
	}
	return myServerInfo
}

func SetUpGame(router *gin.Engine, contractsService *contracts.ChainService) (*con.GameServer,*gamesession.GameSessionManager,*entities.ServerInfo, error) {

	// ---------- SETA CONFUGURAÇÕES BASICAS PARA SERVIDOR P2P -----------
	myServerInfo := SetUpServerBaseConfigs()

	// Inicializa Discovery
	discovery, err := discovery.SetUpDiscovery(myServerInfo)
	if err != nil {
		fmt.Printf("Discovery error: %v\n", err)
	}
	log.Printf("%d servidores conhecidos", len(discovery.KnownServers))

	apiClient := client.New()
	gameserver := con.New(myServerInfo, apiClient, discovery)

	// Inicializa banco de dados e repositório
	db := config.CretaeTable()
	playerRepo := repository.NewPlayerRepository(&db)
	packageRepo := repository.NewPackageRepository(&db)
	cardRepo := repository.NewCardRepository(&db)

	// MatchState
	

	// --------- INICIALIZA E INJETA SERVIÇOS DO SERVIDOR P2P -----------
	
	raftService, _ := raft.InitRaft(playerRepo, packageRepo, cardRepo, myServerInfo, apiClient)
	authService := authService.New(playerRepo, apiClient, discovery.KnownServers, raftService, gameserver.SessionManager, contractsService.Client)
	pkgService := packageService.New(packageRepo, cardRepo, playerRepo,apiClient, raftService, gameserver.SessionManager, contractsService)
	tradeSvc := tradeService.New(apiClient, raftService, gameserver.SessionManager, contractsService, playerRepo)

	seedSvc := seedService.New(raftService, pkgService)

	gameserver.InitAuth(authService)
	gameserver.InitRaft(raftService)
	gameserver.InitPackageSystem(pkgService)
	gameserver.InitSeeds(seedSvc)
	gameserver.InitTrade(tradeSvc)

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
	inventoryHandler := inventoryhandler.New(cardRepo, playerRepo, broker)

	tradeHandler := tradehandler.New(tradeSvc, broker, gameserver.SessionManager, playerRepo)
	localMatchmaking := matchlocal.New(myServerInfo.ID, raftService,cardRepo)
	globalMatchmaking := matchglobal.New(raftService,raftService.Fsm, apiClient, myServerInfo.Address)
	gameSessionManager := gamesession.New(playerRepo,cardRepo,gameserver.SessionManager,localMatchmaking,globalMatchmaking,broker,myServerInfo.Address,apiClient,raftService,contractsService)
	matchHandler := matchhandler.New(localMatchmaking, gameSessionManager, gameserver.SessionManager, broker)

	// injeta todos os handlers da aplicação para o pub sub
	handler := handler.New(authHandler, packgehandler, matchHandler, tradeHandler, inventoryHandler)
	wbSocket := websocket.New(broker, gameserver.SessionManager, gameSessionManager)
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

	return gameserver,gameSessionManager, myServerInfo, nil
}
