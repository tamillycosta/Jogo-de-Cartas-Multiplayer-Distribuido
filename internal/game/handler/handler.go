package handler

import (
	inventoryhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/inventoryHandler"
	authhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/authHandler"
	matchhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/matchHandler"
	packgehandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/packgeHandler"
	tradehandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/tradeHandler"
)

// Estrutura para agrupar todos os hadler do pub sub
type Handler struct {
	
	AuthHandler *authhandler.AuthTopicHandler
	PackageHandler *packgehandler.PackageTopicHandler
	MatchHandler 	*matchhandler.MatchTopicHandler
	TradeHandler *tradehandler.TradeTopicHandler
	InventoryHandler *inventoryhandler.InventoryHandler
}

func New(auth *authhandler.AuthTopicHandler, pkg *packgehandler.PackageTopicHandler, match *matchhandler.MatchTopicHandler, trade *tradehandler.TradeTopicHandler, inv *inventoryhandler.InventoryHandler) *Handler {
    return &Handler{
        AuthHandler:    auth,
        PackageHandler: pkg,
        MatchHandler:   match,
        TradeHandler:   trade,
        InventoryHandler: inv,
    }
}