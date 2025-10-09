package api

import (

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/routes"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// Configs da api de comunicação 
func SetUpApi(router *gin.Engine,myServerInfo *entities.ServerInfo ,handlers *handlers.Handler){

	// seta as rotas do handler
	routes.SetupRoutes(router,handlers)

	port := fmt.Sprintf(":%d", myServerInfo.Port)
    log.Printf("Server %s starting on %s\n", myServerInfo.ID, port)
   
}