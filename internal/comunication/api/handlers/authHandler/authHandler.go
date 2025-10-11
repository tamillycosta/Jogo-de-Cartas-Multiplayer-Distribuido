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


// POST /api/v1/propagate-user
// Recebe propagação de usuário de outro servidor
// Body: {"user_id": "uuid", "username": "nome"}
func (ah *Authhandler) PropagateUser(ctx *gin.Context) {
	var payload struct {
		UserID   string `json:"user_id" binding:"required"`
		Username string `json:"username" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "details": err.Error()})
		return
	}

	err := ah.gameServer.Auth.ReceiveUserPropagation(payload.UserID, payload.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "user propagated successfully",
		"user_id": payload.UserID,
		"username": payload.Username,
	})
}