package routes

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers"

	"github.com/gin-gonic/gin"
)

// Rotas da aplicação 
func SetupRoutes(router *gin.Engine, handler *handlers.CommunicationHandler){
	v1 := router.Group("/api/v1")
	{
	  v1.GET("/info", handler.GetInfo)
	  v1.POST("/notify", handler.ReceiveNotification)
	  
	}
}