package handlers

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/comunication"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"net/http"

	"github.com/gin-gonic/gin"
)

//lida com as rotas HTTP
// relacionadas à comunicação entre servidores de jogo.
// atua como uma camada intermediária entre as requisições HTTP
// e o componente de comunicação (`GameServer`).
type CommunicationHandler struct{
	gameServer *comunication.GameServer

}


func New(gameServer *comunication.GameServer) *CommunicationHandler {
	return &CommunicationHandler{
		gameServer: gameServer,
	}
}


// GET /api/v1/info
//Retorna informações sobre o servidor atual
func (h *CommunicationHandler) GetInfo(ctx *gin.Context) {
	health := h.gameServer.GetCurrentServerInfo()
	ctx.JSON(http.StatusOK, health)
}


// POST /api/v1/notify 
// Recebe uma notificação de outro servidor no formato JSON,
func (h *CommunicationHandler) ReceiveNotification(ctx *gin.Context){
	var notification entities.NotificationMessage
	
	if err := ctx.ShouldBindJSON(&notification); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification"})
		return
	}

	if err := h.gameServer.HandleNotification(&notification); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "notification received"})
}


