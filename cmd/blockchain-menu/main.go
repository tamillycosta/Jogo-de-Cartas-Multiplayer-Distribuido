package main

import (
	s "Jogo-de-Cartas-Multiplayer-Distribuido/blockchain/test/service"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/loader"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/service"
	"context"
	"fmt"
	"log"
)

var (
	client       *blockchain.BlockchainClient
	contracts    *loader.Contracts
	queryService *s.BlockchainQueryService
	matchService *service.MatchChainService
	ctx          context.Context
)

// Menu interação para consulta das operações na chain
func main() {
	fmt.Println(" Conectando à blockchain...")

	cfg := blockchain.Config{
		RPC:        "http://localhost:7545",
		PrivateKey: "e485d098507f54e7733a205420dfddbe58db035fa577fc294ebd14db90767a52",
		ChainID:    1337,
	}

	var err error
	client, err = blockchain.NewBlockchainClient(cfg)
	if err != nil {
		log.Fatalf(" Erro ao conectar: %v", err)
	}

	contracts, err = loader.LoadAllContracts(client, "1337")
	if err != nil {
		log.Fatalf(" Erro ao carregar contratos: %v", err)
	}

	// Inicializar serviços
	queryService = s.NewBlockchainQueryService(client, contracts)
	matchService = service.NewMatchChainService(client, contracts)

	ctx = context.Background()

	fmt.Println(" Conectado com sucesso!")

	// Loop do menu
	for {
		showMenu()
		option := readInput("Escolha uma opção: ")

		clearScreen()

		switch option {
		case "1":
			showSystemReport()
		case "2":
			listAllPackages()
		case "3":
			showRecentActivity()
		case "4":
			showPackageDetails()
		case "5":
			showPlayerReport()
		case "6":
			showCardHistory()
		case "7":
			showTransactionDetails()
		case "8":
			showMatchStatistics()
		case "9":
			showPlayerMatchStats()
		case "0":
			fmt.Println(" Até logo!")
			return
		default:
			fmt.Println(" Opção inválida!")
		}

		waitForEnter()
	}
}
