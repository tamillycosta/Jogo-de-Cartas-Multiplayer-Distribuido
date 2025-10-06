package handlers

import (
	authhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers/authHandler"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	

	
)

//lida com as rotas HTTP
// relacionadas à comunicação entre servidores de jogo.
// atua como uma camada intermediária entre as requisições HTTP
// e o componente de comunicação (`GameServer`).
type Handler struct{
	gameServer *service.GameServer
	authHandler *authhandler.Authhandler
}


func New(gameServer *service.GameServer) *Handler {
	return &Handler{
		gameServer: gameServer,
		authHandler: authhandler.New(gameServer),
	}
}





// POST /api/v1/notify 
// Recebe uma notificação de outro servidor no formato JSON,


