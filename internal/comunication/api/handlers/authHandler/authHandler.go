package authhandler

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Authhandler struct {
	gameServer *service.GameServer
}

func New(gameServer *service.GameServer) *Authhandler {
	return &Authhandler{
		gameServer: gameServer,
	}
}

// GET /api/v1/user-exists
// Retorna se o username existe no seu banco de dados para os servidor cliente
func (ah *Authhandler) UserExists(ctx *gin.Context) {
	username := ctx.Query("username")

	if username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "username query param required"})
		return
	}
	isAvailable := ah.gameServer.Auth.UserExists(username)
	ctx.JSON(http.StatusOK, isAvailable)
}
