package service

import (
	c "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	"context"
	"fmt"
	"log"
	"math/big" 
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/loader"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// Serviço de alto nível para interagir com blockchain
type PackageChainService struct {
	client *c.BlockchainClient
	contracts *loader.Contracts
}

func NewPackageChainService(client *c.BlockchainClient, contracts *loader.Contracts) *PackageChainService {
	return &PackageChainService{
		client: client,
		contracts: contracts,
	}
}

// ===== PACOTES =====

// Registra criação de pacote no blockchain
func (bs *PackageChainService) RegisterPackageCreation(ctx context.Context, packageID string, cardIDs []string) error {
	auth, err := bs.client.NewTransactor(ctx)
	if err != nil {
		return fmt.Errorf("erro ao criar transactor: %w", err)
	}

	// Converter strings para bytes32 (Solidity)
	packageIDBytes := stringToBytes32(packageID)
	cardIDsBytes := make([][32]byte, len(cardIDs))
	for i, cardID := range cardIDs {
		cardIDsBytes[i] = stringToBytes32(cardID)
	}

	// Chamar contrato
	tx, err := bs.contracts.Package.CreatePackage(auth, packageIDBytes, cardIDsBytes)
	if err != nil {
		return fmt.Errorf("erro ao criar package no blockchain: %w", err)
	}

	log.Printf("📝 [Blockchain] Package %s registrado. TX: %s", packageID, tx.Hash().Hex())

	// Não espera confirmação (async) para não bloquear
	go func() {
		receipt, err := bind.WaitMined(context.Background(), bs.client.Client, tx)
		if err != nil {
			log.Printf("⚠️ [Blockchain] Erro ao confirmar TX: %v", err)
			return
		}
		log.Printf("✅ [Blockchain] Package %s confirmado no bloco %d", 
			packageID, receipt.BlockNumber.Uint64())
	}()

	return nil
}


// ===== CONSULTAS (Leitura - Grátis) =====

// Verifica se pacote existe no blockchain
func (bs *PackageChainService) PackageExists(ctx context.Context, packageID string) (bool, error) {
	packageIDBytes := stringToBytes32(packageID)
	
	exists, err := bs.contracts.Package.PackageExists(bs.client.CallOpts(), packageIDBytes)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// Pega informações do pacote
func (bs *PackageChainService) GetPackageInfo(ctx context.Context, packageID string) (*PackageInfo, error) {
	packageIDBytes := stringToBytes32(packageID)
	
	pkg, err := bs.contracts.Package.GetPackage(bs.client.CallOpts(), packageIDBytes)
	if err != nil {
		return nil, err
	}

	// Converter bytes32 de volta para strings
	cardIDs := make([]string, len(pkg.CardIds))
	for i, cardBytes := range pkg.CardIds {
		cardIDs[i] = bytes32ToString(cardBytes)
	}

	return &PackageInfo{
		PackageID: bytes32ToString(pkg.Id),
		CardIDs:   cardIDs,
		Opened:    pkg.Opened,
		OpenedBy:  bytes32ToString(pkg.OpenedBy),
		Timestamp: pkg.CreatedAt.Uint64(),
	}, nil
}

// Conta total de pacotes no blockchain
func (bs *PackageChainService) GetTotalPackages(ctx context.Context) (*big.Int, error) {
	total, err := bs.contracts.Package.GetTotalPackages(bs.client.CallOpts())
	if err != nil {
		return nil, err
	}
	return total, nil
}

// ===== HELPERS =====

type PackageInfo struct {
	PackageID string
	CardIDs   []string
	Opened    bool
	OpenedBy  string
	Timestamp uint64
}

// Converte string para bytes32 (Solidity)
func stringToBytes32(s string) [32]byte {
	var bytes32 [32]byte
	copy(bytes32[:], []byte(s))
	return bytes32
}

// Converte bytes32 para string
func bytes32ToString(b [32]byte) string {
	// Remove zeros à direita
	n := 0
	for i := 0; i < 32; i++ {
		if b[i] != 0 {
			n = i + 1
		}
	}
	return string(b[:n])
}