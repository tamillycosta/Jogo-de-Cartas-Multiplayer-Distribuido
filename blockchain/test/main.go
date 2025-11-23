package main

import (
	s "Jogo-de-Cartas-Multiplayer-Distribuido/blockchain/test/service"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/loader"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/service"

	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
)


// Gera relatório da blockchain 
func main() {
	// Flags
	packageID := flag.String("package", "", "ID do pacote")
	playerID := flag.String("player", "", "ID do jogador")
	playerAddress := flag.String("address", "", "Endereço do jogador")
	txHash := flag.String("tx", "", "Hash da transação")
	flag.Parse()

	if *packageID == "" || *playerID == "" || *playerAddress == "" {
		fmt.Println("Uso: blockchain-test -package=<id> -player=<id> -address=<0x...> [-tx=<hash>]")
		return
	}

	// Conectar à blockchain
	cfg := blockchain.Config{
		RPC:        "http://localhost:7545",
		PrivateKey: "INSERIR UMA CHAVE PRIVADA",
		ChainID:    1337,
	}

	client, err := blockchain.NewBlockchainClient(cfg)
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}

	
	contracts, err := loader.LoadAllContracts(client, "5777" )
	if err != nil {
		log.Fatalf("Erro ao carregar contratos: %v", err)
	}
	
	
	queryService := s.NewBlockchainQueryService(client, contracts)
	packageService := service.NewPackageChainService(client, contracts)
	cardService := service.NewCardChainService(client, contracts)
	demoService := s.NewBlockchainDemoService(queryService, packageService, cardService)
	ctx := context.Background()
	summary, err := queryService.GetSystemReport(ctx)
	if err != nil {
		log.Printf("Erro: %v", err)
		return
	}
	
	log.Printf(" Total de pacotes na blockchain: %d", summary.TotalPackages)
	
	// Buscar pacotes por índice
	for i := 0; i < summary.TotalPackages; i++ {
		pkg, err := contracts.Package.GetPackageByIndex(
			client.CallOpts(), 
			big.NewInt(int64(i)),
		)
		if err != nil {
			log.Printf("Erro no pacote %d: %v", i, err)
			continue
		}
		
		packageID :=  pkg.Id
		log.Printf("  [%d] ID: %s | Aberto: %v | Cards: %d " ,   
			i, packageID, pkg.Opened, len(pkg.CardIds))
	}

	ctx = context.Background()
	demoService.RunFullDemo(ctx, *packageID, *playerID, *playerAddress, *txHash)
}



