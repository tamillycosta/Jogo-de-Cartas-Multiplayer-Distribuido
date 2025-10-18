package handler

import (
	authhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/authHandler"
	packgehandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/packgeHandler"
)

// Estrutura para agrupar todos os hadler do pub sub
type Handler struct {
	
	AuthHandler *authhandler.AuthTopicHandler
	PackageHandler *packgehandler.PackageTopicHandler
	
}

func New(authHandler *authhandler.AuthTopicHandler, packageHandler *packgehandler.PackageTopicHandler ) *Handler {
	return &Handler{
		AuthHandler: authHandler,
		PackageHandler: packageHandler,
		
	}
}