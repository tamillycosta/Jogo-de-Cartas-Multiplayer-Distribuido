package comunication

import (

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"

	"time"
)

// Faz o enviou de uma notificação para os servidores conhecidos 
func (gs *GameServer) SendNotification(){
	notification := &entities.NotificationMessage{
		From: gs.MyInfo.Address,
		Type: "Hello Wolrd",
		Data: make(map[string]string),
		SentAt: time.Now(),
	}

	// Faz broadcast para servers conhecidos 
	for _ , server := range gs.Discovery.KnownServers{
		gs.ApiClient.SendNotification(server.Address, server.Port, notification)
	}
}

