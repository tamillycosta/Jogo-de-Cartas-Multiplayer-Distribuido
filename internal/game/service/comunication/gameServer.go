package comunication

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/interfaces"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/discovery"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"

	"sync"
	"time"
)

// implementa interface game server
var _ interfaces.IGameServer = (*GameServer)(nil)

type NotificationHandler func(*entities.NotificationMessage) error

// modelo para representar um servidor disponivel na rede
type GameServer struct {
	MyInfo          *entities.ServerInfo 
	
	Mu              sync.RWMutex
	StartTime       time.Time
	NotificationHandlers map[string]NotificationHandler 
	// Discovery para conhecer servidores na rede 
	Discovery  *discovery.Discovery
	// Client para comunicar via API
	
	ApiClient       *client.Client
}


func New(myInfo *entities.ServerInfo, apiClient *client.Client, discovery *discovery.Discovery) *GameServer {
	gs := &GameServer{
		MyInfo:        myInfo,
		
		StartTime:     time.Now(),
		Discovery: discovery,
		ApiClient:     apiClient,
		NotificationHandlers: map[string]NotificationHandler{},
		
	}
	gs.registerNotificationHandlers()	// registro de notifições 
	return  gs
}


func (gs *GameServer) GetCurrentServerInfo() *entities.ServerInfo {
	gs.Mu.RLock()
	defer gs.Mu.RUnlock()
	gs.MyInfo.Status = "ACTIVE"
	return gs.MyInfo
}
