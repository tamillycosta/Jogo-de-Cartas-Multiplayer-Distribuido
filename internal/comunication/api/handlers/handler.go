package handlers

import (
	authhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers/authHandler"
	matchhandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers/matchHandler"
	packagehandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers/packageHandler"
	rafthandler "Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers/raft_handler"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	gamesession "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/gameSession"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler lida com as rotas HTTP relacionadas à comunicação entre servidores
type Handler struct {
	gameServer  *service.GameServer
	AuthHandler *authhandler.Authhandler
	RaftHandler *rafthandler.RaftHandler
	PackageHandler *packagehandler.PackageHandler
	MatchHandler 	*matchhandler.MatchHandler
}

func New(gameServer *service.GameServer, gameSessionManager *gamesession.GameSessionManager ) *Handler {
	return &Handler{
		gameServer:  gameServer,
		AuthHandler: authhandler.New(gameServer),
		RaftHandler: rafthandler.New(gameServer.Raft),
		PackageHandler: packagehandler.New(gameServer),
		MatchHandler: matchhandler.New(gameSessionManager, gameSessionManager.GetGlobalMacthService()),
	}
}

// GET /api/v1/info
// Retorna informações do servidor atual
func (h *Handler) GetServerInfo(ctx *gin.Context) {
	info := h.gameServer.GetCurrentServerInfo()
	ctx.JSON(http.StatusOK, info)
}