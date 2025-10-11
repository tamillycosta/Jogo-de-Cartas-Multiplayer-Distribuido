package routes

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers"
	"github.com/gin-gonic/gin"
)

// Rotas da aplicação
func SetupRoutes(router *gin.Engine, handler *handlers.Handler){
	v1 := router.Group("/api/v1")
	{
	  // Rota para verificar info do servidor
		v1.GET("/info", handler.GetServerInfo)
		
		// Rotas de autenticação (P2P)
		v1.GET("/user-exists", handler.AuthHandler.UserExists)
		v1.POST("/propagate-user", handler.AuthHandler.PropagateUser)
	}
	  
	 
	}
	
