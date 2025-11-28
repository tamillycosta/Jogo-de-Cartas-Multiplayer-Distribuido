package service

import (
	"context"
	"fmt"
	"log"
	"math/big"

	c "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/loader"
	

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Serviço para fazer colsulta a blockchain
type BlockchainQueryService struct {
	client    *c.BlockchainClient
	Contracts *loader.Contracts
}



func NewBlockchainQueryService(client *c.BlockchainClient, Contracts *loader.Contracts) *BlockchainQueryService {
	return &BlockchainQueryService{
		client:    client,
		Contracts: Contracts,
	}
}

// ===== ESTRUTURAS DE RESULTADO =====

type Transaction struct {
	Hash        string
	From        string
	To          string
	Value       string
	BlockNumber uint64
	Timestamp   uint64
	Status      string
	GasUsed     uint64
	Type        string // "CardMinted", "CardTransferred", "PackageCreated", etc
	Details     string
}


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


type TransferEvent struct {
	TokenID      uint64
	CardID       string
	From         string
	To           string
	BlockNumber  uint64
	TxHash       string
	Timestamp    uint64
}

type CardHistoryReport struct {
	CardID       string
	TokenID      uint64
	TemplateID   string
	PackageID    string
	MintedAt     uint64
	CurrentOwner string
	Transfers    []TransferEvent
}



// ===== CONSULTAS =====


// GetAllTransactions retorna todas as transações de um range de blocos
func (ts *BlockchainQueryService) GetAllTransactions(ctx context.Context, fromBlock, toBlock uint64) ([]Transaction, error) {
	var transactions []Transaction

	// Se toBlock for 0, pega o último bloco
	if toBlock == 0 {
		header, err := ts.client.Client.HeaderByNumber(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("erro ao obter último bloco: %w", err)
		}
		toBlock = header.Number.Uint64()
	}

	fmt.Printf("🔍 Buscando transações dos blocos %d até %d...\n", fromBlock, toBlock)

	for blockNum := fromBlock; blockNum <= toBlock; blockNum++ {
		block, err := ts.client.Client.BlockByNumber(ctx, big.NewInt(int64(blockNum)))
		if err != nil {
			continue
		}

		for _, tx := range block.Transactions() {
			receipt, err := ts.client.Client.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				continue
			}


			txInfo := Transaction{
				Hash:        tx.Hash().Hex(),
			
				Value:       tx.Value().String(),
				BlockNumber: blockNum,
				Timestamp:   block.Time(),
				GasUsed:     receipt.GasUsed,
				Status:      "Success",
			}

			if tx.To() != nil {
				txInfo.To = tx.To().Hex()
			}

			if receipt.Status == 0 {
				txInfo.Status = "Failed"
			}

			
			transactions = append(transactions, txInfo)
		}
	}

	return transactions, nil
}


// Gera relatório completo de um pacote
func (qs *BlockchainQueryService) GetPackageReport(ctx context.Context, packageID string) (*PackageReport, error) {
	
	pkg, err := qs.Contracts.Package.GetPackage(qs.client.CallOpts(), packageID)
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
	tokenId, err := qs.Contracts.Card.CardIdToTokenId(qs.client.CallOpts(), cardID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar tokenId: %w", err)
	}

	if tokenId.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("carta não encontrada na blockchain")
	}

	// Buscar metadados
	cardToken, err := qs.Contracts.Card.GetCardMetadata(
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

	tokenIds, err := qs.Contracts.Card.GetPlayerCards(qs.client.CallOpts(), addr)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar cartas: %w", err)
	}

	cards := make([]CardReport, 0, len(tokenIds))
	for _, tokenId := range tokenIds {
		cardToken, err := qs.Contracts.Card.GetCardMetadata(
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
	totalPackages, err := qs.Contracts.Package.GetTotalPackages(qs.client.CallOpts())
	if err != nil {
		return nil, err
	}

	totalCards, err := qs.Contracts.Card.GetTotalCards(qs.client.CallOpts())
	if err != nil {
		return nil, err
	}

	// Contar pacotes abertos
	opened := 0
	for i := int64(0); i < totalPackages.Int64(); i++ {
		pkg, err := qs.Contracts.Package.GetPackageByIndex(qs.client.CallOpts(), big.NewInt(i))
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


// retorna o histórico completo de uma carta incluindo todas as transferências
func (qs *BlockchainQueryService) GetCardHistory(ctx context.Context, cardID string) (*CardHistoryReport, error) {
	//  Busca os metadados da carta
	tokenId, err := qs.Contracts.Card.CardIdToTokenId(qs.client.CallOpts(), cardID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar tokenId: %w", err)
	}

	if tokenId.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("carta não encontrada na blockchain")
	}

	cardToken, err := qs.Contracts.Card.GetCardMetadata(
		qs.client.CallOpts(),
		tokenId,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar metadados: %w", err)
	}

	//Buscar evento de mint (primeira transferência)
	mintEvents, err := qs.Contracts.Card.FilterCardMinted(
		&bind.FilterOpts{
			Start:   0,
			End:     nil,
			Context: ctx,
		},
		nil, nil,nil,
	)
	if err != nil {
		log.Printf("⚠️ Erro ao buscar eventos de mint: %v", err)
	}

	var mintTransfer *TransferEvent
	if mintEvents != nil {
		defer mintEvents.Close()
		
		for mintEvents.Next() {
			if mintEvents.Event.TokenId.Cmp(tokenId) == 0 {
				// Buscar timestamp do bloco
				block, _ := qs.client.Client.BlockByNumber(ctx, big.NewInt(int64(mintEvents.Event.Raw.BlockNumber)))
				timestamp := uint64(0)
				if block != nil {
					timestamp = block.Time()
				}

				mintTransfer = &TransferEvent{
					TokenID:     tokenId.Uint64(),
					CardID:      cardToken.CardId,
					From:        "0x0000000000000000000000000000000000000000", 
					To:          mintEvents.Event.Owner.Hex(),
					BlockNumber: mintEvents.Event.Raw.BlockNumber,
					TxHash:      mintEvents.Event.Raw.TxHash.Hex(),
					Timestamp:   timestamp,
				}
				break
			}
		}
	}

	// Buscar eventos de transferência
	transferEvents, err := qs.Contracts.Card.FilterCardTransferred(
		&bind.FilterOpts{
			Start:   0,
			End:     nil,
			Context: ctx,
		},
		nil, // tokenIds (nil = todos)
		nil, // from
	
	)
	if err != nil {
		log.Printf("⚠️ Erro ao buscar eventos de transferência: %v", err)
	}

	transfers := []TransferEvent{}
	
	// Adiciona mint primeiro
	if mintTransfer != nil {
		transfers = append(transfers, *mintTransfer)
	}

	// Adiciona transferências
	if transferEvents != nil {
		defer transferEvents.Close()
		
		for transferEvents.Next() {
			if transferEvents.Event.TokenId.Cmp(tokenId) == 0 {
				// Buscar timestamp do bloco
				block, _ := qs.client.Client.BlockByNumber(ctx, big.NewInt(int64(transferEvents.Event.Raw.BlockNumber)))
				timestamp := uint64(0)
				if block != nil {
					timestamp = block.Time()
				}

				transfers = append(transfers, TransferEvent{
					TokenID:     tokenId.Uint64(),
					CardID:      transferEvents.Event.CardId.String(),
					From:        transferEvents.Event.From.Hex(),
					To:          transferEvents.Event.To.Hex(),
					BlockNumber: transferEvents.Event.Raw.BlockNumber,
					TxHash:      transferEvents.Event.Raw.TxHash.Hex(),
					Timestamp:   timestamp,
				})
			}
		}
	}

	return &CardHistoryReport{
		CardID:       cardToken.CardId,
		TokenID:      tokenId.Uint64(),
		TemplateID:   cardToken.TemplateId,
		PackageID:    cardToken.PackageId,
		MintedAt:     cardToken.MintedAt.Uint64(),
		CurrentOwner: cardToken.CurrentOwner.Hex(),
		Transfers:    transfers,
	}, nil
}


// Helper para encurtar endereços
func shortAddress(addr string) string {
	if len(addr) < 10 {
		return addr
	}
	return addr[:6] + "..." + addr[len(addr)-4:]
}