package routes

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers"
	"github.com/gin-gonic/gin"
)

// Rotas da aplicação
func SetupRoutes(router *gin.Engine, handler *handlers.Handler){
	v1 := router.Group("/api/v1")
	{
	  v1.GET("/info")
	  v1.GET("/api/v1/username-available",handler.AuthHandler.UserExists )
	  
	 
	}
	
}