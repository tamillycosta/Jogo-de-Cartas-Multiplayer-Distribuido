package service 

import (
	c "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/loader"

)

// Service responssavel por gerir as ações que interagem com a blockchain 
type ChainService struct {
	client *c.BlockchainClient
	PackageChainService *PackageChainService // serviço para gerir ações relacionadas aos pacotes
}

func New(client *c.BlockchainClient) *ChainService {
	contracts, err := loader.LoadAllContracts(client, "5777") // carrega os contratos gerados pelo abigen
	if err != nil { panic(err) }
	return &ChainService{
		client: client,
		PackageChainService: NewPackageChainService(client,contracts),
	}
}