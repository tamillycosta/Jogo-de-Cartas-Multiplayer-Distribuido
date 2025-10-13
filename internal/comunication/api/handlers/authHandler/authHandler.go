package authhandler

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HANDLER PARA TODOAS AS ROTAS VINCULADAS A AUTENTICAÇÇÃO DE CLIENTES NA APLICAÇÇÃO 
type Authhandler struct {
	gameServer *service.GameServer
}

func New(gameServer *service.GameServer) *Authhandler {
	return &Authhandler{
		gameServer: gameServer,
	}
}

// GET /api/v1/user-exists?username=xyz
// Verifica se o username existe no banco local
// Retorna: {"exists": true/false}
func (ah *Authhandler) UserExists(ctx *gin.Context) {
	username := ctx.Query("username")

	if username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "username query param required"})
		return
	}

	exists := ah.gameServer.Auth.UserExists(username)
	
	ctx.JSON(http.StatusOK, gin.H{
		"exists": exists,
		"username": username,
	})
}


