package interfaces

import "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"


// interface mínima só para quem precisa de informações básicas do servidor
type ServerInfoProvider interface {
    GetCurrentServerInfo() *entities.ServerInfo
}


// interface para envio de notificações
type NotificationSender interface {
    SendNotification()
    SendConfirmNotification(serverAddress string)
}

// interface para tratamento de notificações recebidas
type NotificationHandler interface {
    HandleNotification(notification *entities.NotificationMessage) error
}

// representa o contrato completo que um servidor de jogo deve seguir
// para participar da rede
type IGameServer interface {
    ServerInfoProvider
    NotificationSender
    NotificationHandler
}
