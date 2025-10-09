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

// implementa interface game server
var _ interfaces.IGameServer = (*GameServer)(nil)



// modelo para representar um servidor disponivel na rede
type GameServer struct {
	MyInfo          *entities.ServerInfo 
	Mu              sync.RWMutex
	StartTime       time.Time
	Auth 			*aS.AuthService
	// Componente reponssavel por conhecer servidores na rede 
	Discovery  *discovery.Discovery
	//  Componente reponssavel por criar interface Client para servidor
	ApiClient       *client.Client
}


func New(myInfo *entities.ServerInfo, apiClient *client.Client, discovery *discovery.Discovery) *GameServer {
	gs := &GameServer{
		MyInfo:        myInfo,
		
		StartTime:     time.Now(),
		Discovery: discovery,
		ApiClient:     apiClient,
		
	}
	
	return  gs
}


func (gs *GameServer) GetCurrentServerInfo() *entities.ServerInfo {
	gs.Mu.RLock()
	defer gs.Mu.RUnlock()
	gs.MyInfo.Status = "ACTIVE"
	return gs.MyInfo
}
