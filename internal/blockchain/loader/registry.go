package loader

import (
	"github.com/ethereum/go-ethereum/common"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/contracts"
	client "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	
)
// Os contratos gerados pelo abigen (codigo solidity -> byteCode -> golang)
type Contracts struct {
	Package *blockchain.PackageRegistry
	Card *blockchain.Card
	
}

func LoadAllContracts(cli *client.BlockchainClient, networkID string) (*Contracts, error) {
	// Adicionar o endereço real dos contratos criados que serão apresentados na interface do ganache
	// ou acessar pelo truffle console (PackageRegistry.address)
	packageAddr := common.HexToAddress("0x5b1869D9A4C187F2EAa108f3062412ecf0526b24") 
	cardAddr := common.HexToAddress("0xe78A0F7E598Cc8b0Bb87894B0F60dD2a88d6a8Ab")   
	
	pkg, err := blockchain.NewPackageRegistry(packageAddr, cli.Client)
	if err != nil {
		return nil, err
	}
	
	card, err := blockchain.NewCard(cardAddr, cli.Client)
	if err != nil {
		return nil, err
	}
	
	return &Contracts{
		Package: pkg,
		Card:    card,
	}, nil
}
