package comunication

// Lida com o recebimento e tratamento de notificações
// enviadas por outros servidores na rede.

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"fmt"
)


func (gs *GameServer) registerNotificationHandlers() {
    gs.NotificationHandlers["Hello_World"] = gs.HandleNotification
  
}

func (gs *GameServer) HandleNotification(notification *entities.NotificationMessage) error {
    fmt.Printf(" %s: Notificação do tipo %s recebida" ,gs.MyInfo.Address , notification.Type)
    fmt.Printf("\n data : %s", notification.Data["msg"])
    gs.SendConfirmNotification(notification.From)
    return nil
}