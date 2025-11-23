package service

import (
	"context"
	"fmt"
	"log"
	"math/big"

	c "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/loader"
	

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)
// Serviço para fazer colsulta a blockchain 
type BlockchainQueryService struct {
	client    *c.BlockchainClient
	contracts *loader.Contracts
}

func NewBlockchainQueryService(client *c.BlockchainClient, contracts *loader.Contracts) *BlockchainQueryService {
	return &BlockchainQueryService{
		client:    client,
		contracts: contracts,
	}
}

// ===== ESTRUTURAS DE RESULTADO =====

type PackageReport struct {
	PackageID   string
	CardIDs     []string
	Opened      bool
	OpenedBy    string
	CreatedAt   uint64
	BlockNumber uint64
}

type CardReport struct {
	TokenID      uint64
	CardID       string
	TemplateID   string
	PackageID    string
	MintedAt     uint64
	CurrentOwner string
}

type PlayerReport struct {
	PlayerID     string
	Address      string
	Balance      string
	BalanceETH   float64
	TotalCards   int
	Cards        []CardReport
	Transactions []TransactionReport
}

type TransactionReport struct {
	TxHash      string
	BlockNumber uint64
	From        string
	To          string
	GasUsed     uint64
	Status      string
}

type FullReport struct {
	Packages []PackageReport
	Players  []PlayerReport
	Summary  Summary
}

type Summary struct {
	TotalPackages       int
	TotalPackagesOpened int
	TotalCards          int
	TotalPlayers        int
}

// ===== CONSULTAS =====

// Gera relatório completo de um pacote
func (qs *BlockchainQueryService) GetPackageReport(ctx context.Context, packageID string) (*PackageReport, error) {
	
	pkg, err := qs.contracts.Package.GetPackage(qs.client.CallOpts(), packageID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar pacote: %w", err)
	}

	cardIDs := make([]string, len(pkg.CardIds))
	copy(cardIDs, pkg.CardIds)

	return &PackageReport{
		PackageID: pkg.Id,
		CardIDs:   cardIDs,
		Opened:    pkg.Opened,
		OpenedBy:  pkg.OpenedBy,
		CreatedAt: pkg.CreatedAt.Uint64(),
	}, nil
}





// Gera relatório completo de uma carta
func (qs *BlockchainQueryService) GetCardReport(ctx context.Context, cardID string) (*CardReport, error) {
	

	// Buscar tokenId
	tokenId, err := qs.contracts.Card.CardIdToTokenId(qs.client.CallOpts(), cardID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar tokenId: %w", err)
	}

	if tokenId.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("carta não encontrada na blockchain")
	}

	// Buscar metadados
	cardToken, err := qs.contracts.Card.GetCardMetadata(
		qs.client.CallOpts(),
		tokenId,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar metadados: %w", err)
	}

	return &CardReport{
		TokenID:      tokenId.Uint64(),
		CardID:       cardToken.CardId,
		TemplateID:   cardToken.TemplateId,
		PackageID:   cardToken.PackageId,
		MintedAt:     cardToken.MintedAt.Uint64(),
		CurrentOwner: cardToken.CurrentOwner.Hex(),
	}, nil
}


// Gera relatório completo de um jogador
func (qs *BlockchainQueryService) GetPlayerReport(ctx context.Context, playerID string, playerAddress string) (*PlayerReport, error) {
	addr := common.HexToAddress(playerAddress)


	balance, err := qs.client.GetBalance(ctx, playerAddress)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar saldo: %w", err)
	}

	tokenIds, err := qs.contracts.Card.GetPlayerCards(qs.client.CallOpts(), addr)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar cartas: %w", err)
	}

	cards := make([]CardReport, 0, len(tokenIds))
	for _, tokenId := range tokenIds {
		cardToken, err := qs.contracts.Card.GetCardMetadata(
			qs.client.CallOpts(),
			tokenId,
		)
		if err != nil {
			log.Printf("⚠️ Erro ao buscar carta %d: %v", tokenId.Uint64(), err)
			continue
		}

		cards = append(cards, CardReport{
			TokenID:      tokenId.Uint64(),
			CardID:     cardToken.CardId,
			TemplateID:   cardToken.TemplateId,
			PackageID:   cardToken.PackageId,
			MintedAt:     cardToken.MintedAt.Uint64(),
			CurrentOwner: cardToken.CurrentOwner.Hex(),
		})
	}

	return &PlayerReport{
		PlayerID:   playerID,
		Address:    playerAddress,
		Balance:    balance.String(),
		BalanceETH: c.WeiToEth(balance),
		TotalCards: len(cards),
		Cards:      cards,
	}, nil
}


// Buscar detalhes de uma transação
func (qs *BlockchainQueryService) GetTransactionDetails(ctx context.Context, txHash string) (*TransactionReport, error) {
	hash := common.HexToHash(txHash)

	// Buscar recibo da transação
	receipt, err := qs.client.Client.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar recibo: %w", err)
	}

	// Buscar transação
	tx, _, err := qs.client.Client.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar transação: %w", err)
	}

	// Extrair sender
	signer := types.NewEIP155Signer(qs.client.ChainID)
	from, err := types.Sender(signer, tx)
	if err != nil {
		return nil, fmt.Errorf("erro ao extrair sender: %w", err)
	}

	status := "Falhou"
	if receipt.Status == 1 {
		status = "Sucesso"
	}

	toAddr := ""
	if tx.To() != nil {
		toAddr = tx.To().Hex()
	}

	return &TransactionReport{
		TxHash:      txHash,
		BlockNumber: receipt.BlockNumber.Uint64(),
		From:        from.Hex(),
		To:          toAddr,
		GasUsed:     receipt.GasUsed,
		Status:      status,
	}, nil
}


// Gera relatório geral do sistema
func (qs *BlockchainQueryService) GetSystemReport(ctx context.Context) (*Summary, error) {
	totalPackages, err := qs.contracts.Package.GetTotalPackages(qs.client.CallOpts())
	if err != nil {
		return nil, err
	}

	totalCards, err := qs.contracts.Card.GetTotalCards(qs.client.CallOpts())
	if err != nil {
		return nil, err
	}

	// Contar pacotes abertos
	opened := 0
	for i := int64(0); i < totalPackages.Int64(); i++ {
		pkg, err := qs.contracts.Package.GetPackageByIndex(qs.client.CallOpts(), big.NewInt(i))
		if err != nil {
			continue
		}
		if pkg.Opened {
			opened++
		}
	}

	return &Summary{
		TotalPackages:       int(totalPackages.Int64()),
		TotalPackagesOpened: opened,
		TotalCards:          int(totalCards.Int64()),
	}, nil
}

