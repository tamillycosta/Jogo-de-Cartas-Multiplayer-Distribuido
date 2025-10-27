package service

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/interfaces"
	aS "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/authService"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/discovery"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/packageService"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	seedService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/seed"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/session"
	tradeService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/tradeService"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"sync"
	"time"
)

// Implementa interface game server
var _ interfaces.IGameServer = (*GameServer)(nil)

type GameServer struct {
	MyInfo    *entities.ServerInfo
	Mu        sync.RWMutex
	StartTime time.Time
	Discovery *discovery.Discovery
	Auth      *aS.AuthService
	Package *packageService.PackageService
	Raft	  *raft.RaftService
	Trade     *tradeService.TradeService
	ApiClient *client.Client

	SessionManager *session.SessionManager
	Seeds *seedService.SeedService
}

func New(myInfo *entities.ServerInfo, apiClient *client.Client, discovery *discovery.Discovery) *GameServer {
	gs := &GameServer{
		MyInfo:         myInfo,
		StartTime:      time.Now(),
		Discovery:      discovery,
		ApiClient:      apiClient,
		SessionManager: session.New(),
	}

	return gs
}

// Inicializa o AuthService (deve ser chamado após criação do repositório)
func (gs *GameServer) InitAuth(authService *aS.AuthService) {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()
	gs.Auth = authService
}

func (gs *GameServer) InitRaft(raftService *raft.RaftService) {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()
	gs.Raft = raftService
}

func (gs *GameServer) InitPackageSystem(packageService *packageService.PackageService) {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()
	gs.Package = packageService
}

func (gs *GameServer) InitSeeds(seedService *seedService.SeedService) {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()
	gs.Seeds = seedService
}

func (gs *GameServer) InitTrade(tradeService *tradeService.TradeService){
	gs.Mu.Lock()
	defer  gs.Mu.Unlock()
	gs.Trade = tradeService
}

func (gs *GameServer) GetCurrentServerInfo() *entities.ServerInfo {
	gs.Mu.RLock()
	defer gs.Mu.RUnlock()
	gs.MyInfo.Status = "ACTIVE"
	return gs.MyInfo
}
