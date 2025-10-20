package handler

import (
	authhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/authHandler"
	matchhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/matchHandler"
	packgehandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/packgeHandler"
)

// Estrutura para agrupar todos os hadler do pub sub
type Handler struct {
	
	AuthHandler *authhandler.AuthTopicHandler
	PackageHandler *packgehandler.PackageTopicHandler
	MatchHandler 	*matchhandler.MatchTopicHandler
}

func New(authHandler *authhandler.AuthTopicHandler, packageHandler *packgehandler.PackageTopicHandler, matchHandler *matchhandler.MatchTopicHandler ) *Handler {
	return &Handler{
		AuthHandler: authHandler,
		PackageHandler: packageHandler,
		MatchHandler: matchHandler,
	}
}