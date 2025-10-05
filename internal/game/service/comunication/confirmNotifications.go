package comunication

// Faz feadback para as menssagens enviadas dos servers

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"time"
)

// retorna uma confirmação de recebimento de menssagem 
func (gs *GameServer) SendConfirmNotification(serverAddress string){
	notification := &entities.NotificationMessage{
		From: gs.MyInfo.Address,
		Type: "Message Recive",
		Data: make(map[string]string),
		SentAt: time.Now(),
	}
	for _ , server := range gs.Discovery.KnownServers{
		if(server.Address == serverAddress){
			gs.ApiClient.SendNotification(server.Address, server.Port,notification)
		}
		
	}
}

