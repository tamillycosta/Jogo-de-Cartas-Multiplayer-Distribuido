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

// Teste de troca atômica de cartas
func main() {
	fmt.Println("Testando TROCA de cartas (swap)...")

	cfg := blockchain.Config{
		RPC:        "http://localhost:7545",
		PrivateKey: "829e924fdf021ba3dbbc4225edfece9aca04b929d6e75613329ca6f1d31c0bb4", // Chave do jogador 1
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

	// ===== CONFIGURAÇÃO DA TROCA =====
	
	
	player1PrivateKey := "6cf162c3434276400bd01c6858317cae2250a1bd7b5909b54ac145915e2c8d9d"  // sua chave privada
	player1TokenID := uint64(6)     // Carta que VOCÊ vai dar
	
	player2PrivateKey := "6a8191969fd408348c57f5fe337eef69e3f852dde6243d14b12090990babca22"  // Chave privada do outro jogador
	player2TokenID := uint64(1)     // Carta que VOCÊ vai receber
	player2Address := "0x7f9260C27983bA030728A42f44590421d008FC82"            // Endereço do outro jogador

	
	fmt.Println("\n Estado Inicial:")
	
	owner1, _ := contracts.Card.OwnerOf(client.CallOpts(), big.NewInt(int64(player1TokenID)))
	owner2, _ := contracts.Card.OwnerOf(client.CallOpts(), big.NewInt(int64(player2TokenID)))
	
	fmt.Printf("   Token #%d pertence a: %s (Jogador 1)\n", player1TokenID, owner1.Hex())
	fmt.Printf("   Token #%d pertence a: %s (Jogador 2)\n", player2TokenID, owner2.Hex())

	//   Jogador 2 aprova a troca (assina a trasação de consesnso)
	fmt.Println("\n Passo 1: Jogador 2 aprova sua carta para troca...")
	
	err = cardService.ApproveForSwap(ctx, player2TokenID, owner1.Hex(), player2PrivateKey)
	if err != nil {
		log.Fatalf(" Erro ao aprovar carta: %v", err)
	}

	//Jogador 1 executa a troca (ele assina a trasação de troca )
	fmt.Println("\n Passo 2: Executando troca atômica...")
	
	err = cardService.SwapCards(ctx, player1TokenID, player2TokenID, player1PrivateKey)
	if err != nil {
		log.Fatalf(" Erro na troca: %v", err)
	}


	fmt.Println("\n Estado Final:")
	
	newOwner1, _ := contracts.Card.OwnerOf(client.CallOpts(), big.NewInt(int64(player1TokenID)))
	newOwner2, _ := contracts.Card.OwnerOf(client.CallOpts(), big.NewInt(int64(player2TokenID)))
	
	fmt.Printf("   Token #%d agora pertence a: %s\n", player1TokenID, newOwner1.Hex())
	fmt.Printf("   Token #%d agora pertence a: %s\n", player2TokenID, newOwner2.Hex())


	swapCorreto := newOwner1.Hex() == player2Address && newOwner2.Hex() == owner1.Hex()
	
	if swapCorreto {
		fmt.Println("\n TROCA REALIZADA COM SUCESSO!")
	} else {
		fmt.Println("\n Algo deu errado na troca")
	}
}