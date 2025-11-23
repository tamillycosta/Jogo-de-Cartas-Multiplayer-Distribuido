package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/loader"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/service"
)
// teste especifico para troca de cartas 
func main() {
	fmt.Println(" Testando transferência de cartas...")

	
	cfg := blockchain.Config{
		RPC:        "http://localhost:7545",
		PrivateKey: "", // inserir quando quiser testar 
		ChainID:    1337,
	}

	client, err := blockchain.NewBlockchainClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	
	contracts, err := loader.LoadAllContracts(client, "1337")
	if err != nil {
		log.Fatal(err)
	}

	cardService := service.NewCardChainService(client, contracts)

	
	ctx := context.Background()
	
	
	tokenID := uint64(6)  // Token que vai transferir
	fromPrivateKey := "CHAVE DE QUEM TEM A CARTA"  // 
	toAddress := "ENDEREÇO DE QUEM VAI RECEBER"        

	fmt.Printf("  Transferindo Token #%d\n", tokenID)
	fmt.Printf("   Para: %s\n", toAddress)


	oldOwner, _ := contracts.Card.OwnerOf(client.CallOpts(), big.NewInt(int64(tokenID)))
	fmt.Printf("   Dono atual: %s\n\n", oldOwner.Hex())

	err = cardService.TransferCard(ctx, tokenID, toAddress, fromPrivateKey)
	if err != nil {
		log.Fatalf(" Erro: %v", err)
	}

	newOwner, _ := contracts.Card.OwnerOf(client.CallOpts(), big.NewInt(int64(tokenID)))
	fmt.Printf("\n Transferência concluída!")
	fmt.Printf("   Novo dono: %s\n", newOwner.Hex())
	fmt.Printf("   Correto? %v\n", newOwner.Hex() == toAddress)
}