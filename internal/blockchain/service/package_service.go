package service

import (
	c "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/loader"
	
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)
// ===== TYPES =====

type PackageInfo struct {
	PackageID string
	CardIDs   []string
	Opened    bool
	OpenedBy  string
	Timestamp uint64
}




type PackageChainService struct {
	client    *c.BlockchainClient
	contracts *loader.Contracts
}

func NewPackageChainService(client *c.BlockchainClient, contracts *loader.Contracts) *PackageChainService {
	return &PackageChainService{
		client:    client,
		contracts: contracts,
	}
}

// ===== CRIAR PACOTE =====

func (bs *PackageChainService) RegisterPackageCreation(ctx context.Context, packageID string, cardIDs []string,) error {
	auth, err := bs.client.NewTransactor(ctx)
	if err != nil {
		return fmt.Errorf("erro ao criar transactor: %w", err)
	}
	
	tx, err := bs.contracts.Package.CreatePackage(auth, packageID, cardIDs)
	if err != nil {
		return fmt.Errorf("erro ao criar package no blockchain: %w", err)
	}

	log.Printf(" [Blockchain] Package %s registrado. TX: %s", packageID, tx.Hash().Hex())

	// Confirmação assíncrona
	go func() {
		receipt, err := bind.WaitMined(context.Background(), bs.client.Client, tx)
		if err != nil {
			log.Printf(" [Blockchain] Erro ao confirmar TX: %v", err)
			return
		}
		log.Printf(" [Blockchain] Package %s confirmado no bloco %d", 
			packageID, receipt.BlockNumber.Uint64())
	}()

	return nil
}

// ===== ABRIR PACOTE ================
// verifica se jogagor tem saldo suficiente, caso tenha jogador assina a transação 

func (bs *PackageChainService) RegisterPackageOpen(
	ctx context.Context,
	packageID string,
	playerID string,
	playerAddress string,
	playerPrivateKey string,
	templateIDs []string,
) error {

	
	balance, err := bs.client.GetBalance(ctx, playerAddress)
	if err != nil {
		return fmt.Errorf("erro ao verificar saldo: %w", err)
	}

	log.Printf(" [Blockchain] Saldo do jogador %s: %f ETH", 
		playerAddress, c.WeiToEth(balance))

	// Verificar se tem saldo mínimo (0.01 ETH)
	minBalance := c.EthToWei(0.01)
	if balance.Cmp(minBalance) < 0 {
		return fmt.Errorf("saldo insuficiente: jogador tem %f ETH, precisa de pelo menos 0.01 ETH",
			c.WeiToEth(balance))
	}

	//  jogador assina a transação
	auth, err := bs.client.NewTransactorFromPrivateKey(ctx, playerPrivateKey)
	if err != nil {
		return fmt.Errorf("erro ao criar transactor do jogador: %w", err)
	}

	
	playerAddr := common.HexToAddress(playerAddress)

	tx, err := bs.contracts.Package.OpenPackage(
		auth,
		packageID,
		playerID,
		playerAddr,
		templateIDs,
	)
	if err != nil {
		bs.client.ResyncPlayerNonce(ctx, playerAddress)
		return fmt.Errorf("erro ao abrir pacote: %w", err)
	}

	log.Printf(" [Blockchain] Pacote %s aberto pelo jogador %s. TX: %s",
		packageID, playerAddress, tx.Hash().Hex())

	// Aguardar confirmação
	receipt, err := bind.WaitMined(ctx, bs.client.Client, tx)
	if err != nil {
		return fmt.Errorf("erro ao confirmar abertura: %w", err)
	}

	log.Printf(" [Blockchain] Abertura confirmada no bloco %d. Gas usado: %d",
		receipt.BlockNumber.Uint64(), receipt.GasUsed)

	// Log do novo saldo
	newBalance, _ := bs.client.GetBalance(ctx, playerAddress)
	log.Printf(" [Blockchain] Novo saldo do jogador: %f ETH", c.WeiToEth(newBalance))

	return nil
}


// ========================== CONSULTAS =============================

func (bs *PackageChainService) PackageExists(ctx context.Context, packageID string) (bool, error) {
	
	
	exists, err := bs.contracts.Package.PackageExists(bs.client.CallOpts(), packageID)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (bs *PackageChainService) GetPackageInfo(ctx context.Context, packageID string) (*PackageInfo, error) {

	
	pkg, err := bs.contracts.Package.GetPackage(bs.client.CallOpts(),  packageID)
	if err != nil {
		return nil, err
	}

	cardIDs := make([]string, len(pkg.CardIds))
	copy(cardIDs, pkg.CardIds)

	return &PackageInfo{
		PackageID: pkg.Id ,
		CardIDs:   cardIDs,
		Opened:    pkg.Opened,
		OpenedBy:  pkg.OpenedBy,
		Timestamp: pkg.CreatedAt.Uint64(),
	}, nil
}

func (bs *PackageChainService) GetTotalPackages(ctx context.Context) (*big.Int, error) {
	total, err := bs.contracts.Package.GetTotalPackages(bs.client.CallOpts())
	if err != nil {
		return nil, err
	}
	return total, nil
}

