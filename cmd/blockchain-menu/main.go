

package main

import (
	
	"context"
	"fmt"
	"log"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/service"
	s "Jogo-de-Cartas-Multiplayer-Distribuido/blockchain/test/service"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/loader"
	
)

var (
	client         *blockchain.BlockchainClient
	contracts      *loader.Contracts
	queryService   *s.BlockchainQueryService
	matchService 	*service.MatchChainService
	ctx            context.Context
)

func main() {
	fmt.Println(" Conectando à blockchain...")
	
	
	cfg := blockchain.Config{
		RPC:        "http://localhost:7545",
		PrivateKey: "adicione uma chave privada",
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
	matchService  = service.NewMatchChainService(client, contracts)
	
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
			showPackageDetails()
		case "4":
			showPlayerReport()
		case "5":
			showCardHistory()
		case "6":
			showTransactionDetails()
		case "7":
			searchByAddress()
		case "8":
			showRecentActivity()
		case "10":
			showMatchStatistics()
		case "11":
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

