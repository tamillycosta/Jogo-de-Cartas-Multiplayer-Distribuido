package packagehandler


import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	"net/http"
	"log"
	"github.com/gin-gonic/gin"
)

// HANDLER PARA TODOAS AS ROTAS VINCULADAS A AUTENTICAÇÇÃO DE CLIENTES NA APLICAÇÇÃO 
type PackageHandler struct {
	gameServer *service.GameServer
}

func New(gameServer *service.GameServer) *PackageHandler {
	return &PackageHandler{
		gameServer: gameServer,
	}
}



// POST /api/v1/package/open-package
// è rota para servidores que não são lideres 
// requisitarem server lider para abrir um pacote conta de um jogador 

func (p *PackageHandler) OpenPackage(ctx *gin.Context) {
	var payload struct {
		PlayerId string `json:"player_id"`
	}

	if err := ctx.BindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	if payload.PlayerId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "player_id required"})
		return
	}

	log.Printf("[PackageHandler] Requisição de abertura de pacote: player=%s", payload.PlayerId)

	
	err := p.gameServer.Package.OpenPackage(payload.PlayerId)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	log.Printf("[PackageHandler] Pacote aberto com sucesso para player %s", payload.PlayerId)
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "pacote aberto com sucesso",
	})
}