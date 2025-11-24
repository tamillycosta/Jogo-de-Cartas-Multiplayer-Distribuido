package loader

import (
	client "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/contracts"

	"os"
	"github.com/ethereum/go-ethereum/common"
)

// adicionar no arquivo executavel 
var (
	package_Addr = getEnv("PACKAGE_CONTRACT", "")
	card_Addr = getEnv("CARD_CONTRACT", "")
	match_Addr = getEnv("MATCH_CONTRACT", "")
)




// Os contratos gerados pelo abigen (codigo solidity -> byteCode -> golang)
type Contracts struct {
	Package *blockchain.PackageRegistry
	Card *blockchain.Card
	Match *blockchain.MatchRegistry
	
}

func LoadAllContracts(cli *client.BlockchainClient, networkID string) (*Contracts, error) {
	// Adicionar o endereço  dos contratos no arquivo executavel 
	// acessar pelo truffle console (PackageRegistry.address) ou interface do ganache
	packageAddr := common.HexToAddress(package_Addr) 
	cardAddr := common.HexToAddress(card_Addr)   
	matchAddr := common.HexToAddress(match_Addr)
	
	pkg, err := blockchain.NewPackageRegistry(packageAddr, cli.Client)
	if err != nil {
		return nil, err
	}
	
	card, err := blockchain.NewCard(cardAddr, cli.Client)
	if err != nil {
		return nil, err
	}

	match, err := blockchain.NewMatchRegistry(matchAddr,cli.Client)
	if err != nil {
		return nil, err
	}

	return &Contracts{
		Package: pkg,
		Card:    card,
		Match:   match,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
