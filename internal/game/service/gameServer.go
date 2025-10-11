package service

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/interfaces"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/discovery"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	aS "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/authService"
	"sync"
	"time"
)

// Implementa interface game server
var _ interfaces.IGameServer = (*GameServer)(nil)

type GameServer struct {
	MyInfo    *entities.ServerInfo
	Mu        sync.RWMutex
	StartTime time.Time
	Auth      *aS.AuthService
	Discovery *discovery.Discovery
	ApiClient *client.Client
}

func New(myInfo *entities.ServerInfo, apiClient *client.Client, discovery *discovery.Discovery) *GameServer {
	gs := &GameServer{
		MyInfo:     myInfo,
		StartTime:  time.Now(),
		Discovery:  discovery,
		ApiClient:  apiClient,
	}
	
	return gs
}

// Inicializa o AuthService (deve ser chamado após criação do repositório)
func (gs *GameServer) InitAuth(authService *aS.AuthService) {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()
	gs.Auth = authService
}

func (gs *GameServer) GetCurrentServerInfo() *entities.ServerInfo {
	gs.Mu.RLock()
	defer gs.Mu.RUnlock()
	gs.MyInfo.Status = "ACTIVE"
	return gs.MyInfo
}