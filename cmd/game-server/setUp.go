package gameserver 

import (

	
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	con "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/comunication"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"time"
	"os"
	"strconv"
	
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}


func getPortFromEnv(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        port, err := strconv.Atoi(value)
        if err == nil {
            return port
        }
    }
    return defaultValue
}



func SetUpGame() (*con.GameServer, *entities.ServerInfo) {
	// 1. Configuração do servidor
	myServerInfo := &entities.ServerInfo{
		ID:      getEnv("SERVER_ID", "server-a"),
		Region:  getEnv("REGION", "us-east-1"),
		Address: getEnv("SERVER_ADDRESS", "localhost"),
		Port:    getPortFromEnv("PORT", 8080),
		Status:  "active",
	}
	
	
	apiClient := client.New(5 * time.Second)

	// instancia na main? 
	gameserver := con.New(myServerInfo,apiClient)

	//handlers := handlers.New(gameserver)


	if myServerInfo.ID == "server-a" {
        gameserver.AddKnownServer(&entities.ServerInfo{
            ID: "server-b",
            Address: "localhost",
            Port: 8081,
        })
    } else if myServerInfo.ID == "server-b" {
        gameserver.AddKnownServer(&entities.ServerInfo{
            ID: "server-a",
            Address: "localhost",
            Port: 8080,
        })
    }
	
	return  gameserver, myServerInfo

	


}

