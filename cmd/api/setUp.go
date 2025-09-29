package api

import (

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/routes"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// handler <- usecases <- repositories
func SetUpApi(myServerInfo *entities.ServerInfo ,handlers *handlers.CommunicationHandler){


	router := gin.Default()
	routes.SetupRoutes(router,handlers)

	port := fmt.Sprintf(":%d", myServerInfo.Port)
    log.Printf("Server %s starting on %s\n", myServerInfo.ID, port)
    router.Run(port)
}