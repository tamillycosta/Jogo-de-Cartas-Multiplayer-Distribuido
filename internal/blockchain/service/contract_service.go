package service 

import (
	c "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/loader"

)

// Service responssavel por gerir as ações que interagem com a blockchain 
type ChainService struct {
	Client *c.BlockchainClient
	PackageChainService *PackageChainService // serviço para gerir ações relacionadas aos pacotes
	CardChainService *CardChainService
}

func New(client *c.BlockchainClient) *ChainService {
	contracts, err := loader.LoadAllContracts(client, "5777") // carrega os contratos gerados pelo abigen
	if err != nil { panic(err) }
	return &ChainService{
		Client: client,
		PackageChainService: NewPackageChainService(client,contracts),
		CardChainService: NewCardChainService(client,contracts),
	}
}