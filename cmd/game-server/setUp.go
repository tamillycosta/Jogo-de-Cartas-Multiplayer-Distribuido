package gameserver

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	con "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/discovery"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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

func SetUpGame() (*con.GameServer, *entities.ServerInfo , error) {
	// Configuração do servidor
	myServerInfo := &entities.ServerInfo{
		ID:      getEnv("SERVER_ID", "server-a"),
		Region:  getEnv("REGION", "us-east-1"),
		Address: getEnv("SERVER_ADDRESS", "server-a"),
		Port:    getPortFromEnv("PORT", 8080),
		Status:  "active",
	}

	// Porta para o gossip protocol (memberlist)
	gossipPort := getPortFromEnv("GOSSIP_PORT", 7947)

	// Seed servers
	seedServersEnv := getEnv("SEED_SERVERS", "")
	var seedServers []string
	if seedServersEnv != "" {
		seedServers = strings.Split(seedServersEnv, ",")
	}

	// Cria discovery
	disc, err := discovery.New(myServerInfo, gossipPort, seedServers)
	if err != nil {
		return nil, nil,   fmt.Errorf("erro ao criar discovery: %w", err)
	}
	
	// interface api cliente 
	apiClient := client.New(5 * time.Second)
	// servidor do jogo 
	gameserver := con.New(myServerInfo, apiClient, disc)

	return  gameserver, myServerInfo,  nil
}