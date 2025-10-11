package handlers

import (
	authhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers/authHandler"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler lida com as rotas HTTP relacionadas à comunicação entre servidores
type Handler struct {
	gameServer  *service.GameServer
	AuthHandler *authhandler.Authhandler
}

func New(gameServer *service.GameServer) *Handler {
	return &Handler{
		gameServer:  gameServer,
		AuthHandler: authhandler.New(gameServer),
	}
}

// GET /api/v1/info
// Retorna informações do servidor atual
func (h *Handler) GetServerInfo(ctx *gin.Context) {
	info := h.gameServer.GetCurrentServerInfo()
	ctx.JSON(http.StatusOK, info)
}