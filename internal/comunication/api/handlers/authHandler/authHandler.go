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

// é chamado por outros servidores para verificar se um jogador
// tem uma sessão ativa neste servidor específico.
// GET /api/v1/auth/is-player-logged-in?username=xyz
func (ah *Authhandler) IsPlayerLoggedIn(ctx *gin.Context) {
	username := ctx.Query("username")
	if username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "username query param required"})
		return
	}

	// Usa o SessionManager para verificar se existe uma sessão para este username
	isLoggedIn := ah.gameServer.SessionManager.IsPlayerLoggedInByUsername(username)

	ctx.JSON(http.StatusOK, gin.H{
		"is_logged_in": isLoggedIn,
		"username":     username,
	})
}

// POST /api/v1/auth/create-account
// è rota para servidores que não são lideres 
// requisitarem server lider para criar conta de um jogador 
func (ah *Authhandler) CreateAccount(ctx *gin.Context) {
	var payload struct {
		Username string `json:"username"`
	}
	if err := ctx.BindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	if payload.Username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "username required"})
		return
	}

	err := ah.gameServer.Auth.CreateAccount(payload.Username)
	ctx.JSON(http.StatusOK, gin.H{"error": err})
}
