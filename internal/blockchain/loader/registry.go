package loader

import (
	"github.com/ethereum/go-ethereum/common"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/contracts"
	client "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
)
// Os contratos gerados pelo abigen (codigo solidity -> byteCode -> golang)
type Contracts struct {
	Package *blockchain.PackageRegistry
	
}

func LoadAllContracts(cli *client.BlockchainClient, networkID string) (*Contracts, error) {
	// carrega o endereço do contrato atravez do loader
	packageAddrStr, _ := LoadContractAddress("build/contracts/PackageRegistry.json", networkID)
	// card
	// player 
	pkg, _ := blockchain.NewPackageRegistry(common.HexToAddress(packageAddrStr), cli.Client)
	
	return &Contracts{
		Package: pkg,
		
	}, nil
}
