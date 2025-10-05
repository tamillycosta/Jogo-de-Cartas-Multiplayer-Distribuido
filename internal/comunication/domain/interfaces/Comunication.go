package interfaces


import "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"

// Representa todos os metodos implementados para comunicação dos servidores 
type Communication interface {
    AskServerInfo(serverAddress string) (*entities.ServerInfo, error)
    SendNotification(serverAddress string, notification *entities.NotificationMessage) error
}