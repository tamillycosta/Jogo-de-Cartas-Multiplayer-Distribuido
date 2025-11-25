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
)

type CardMetadata struct {
	CardID       string
	TemplateID   string
	PackageID    string
	MintedAt     uint64
	CurrentOwner string
}




type CardChainService struct {
	client    *c.BlockchainClient
	contracts *loader.Contracts
}

func NewCardChainService(client *c.BlockchainClient, contracts *loader.Contracts) *CardChainService {
	return &CardChainService{
		client:    client,
		contracts: contracts,
	}
}

// ===== MINT (não é chamado diretamente - PackageRegistry faz isso) =====

// Apenas para testes ou casos especiais
func (cs *CardChainService) MintCard(
	ctx context.Context,
	cardID string,
	templateID string,
	packageID string,
	playerAddress string,
) (uint64, error) {
	
	auth, err := cs.client.NewTransactor(ctx)
	if err != nil {
		return 0, fmt.Errorf("erro ao criar transactor: %w", err)
	}

	
	
	playerAddr := common.HexToAddress(playerAddress)

	tx, err := cs.contracts.Card.MintCard(
		auth,
		cardID,
		templateID,
		packageID,
		playerAddr,
	)
	if err != nil {
		return 0, fmt.Errorf("erro ao mintar NFT: %w", err)
	}

	log.Printf(" [Blockchain] Carta %s mintada como NFT. TX: %s", cardID, tx.Hash().Hex())

	// Esperar confirmação
	receipt, err := bind.WaitMined(ctx, cs.client.Client, tx)
	if err != nil {
		return 0, fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	// Pegar tokenId do evento
	for _, vLog := range receipt.Logs {
		event, err := cs.contracts.Card.ParseCardMinted(*vLog)
		if err == nil {
			log.Printf(" NFT tokenId: %d para carta %s", event.TokenId.Uint64(), cardID)
			return event.TokenId.Uint64(), nil
		}
	}

	return 0, fmt.Errorf("evento CardMinted não encontrado")
}


// ===== TRANSFERIR CARTA =====
func (cs *CardChainService) TransferCard(ctx context.Context,tokenID uint64,toAddress string,fromPrivateKey string) error {
	
	auth, err := cs.client.NewTransactorFromPrivateKey(ctx, fromPrivateKey)
	if err != nil {
		return fmt.Errorf("erro ao criar transactor: %w", err)
	}

	toAddr := common.HexToAddress(toAddress)

	tx, err := cs.contracts.Card.TransferCard(
		auth,
		big.NewInt(int64(tokenID)),
		toAddr,
	)
	if err != nil {
		return fmt.Errorf("erro ao transferir carta: %w", err)
	}

	log.Printf(" [Blockchain] Carta tokenId=%d transferida. TX: %s", tokenID, tx.Hash().Hex())
	
	// Aguardar confirmação (opcional)
	_, err = bind.WaitMined(ctx, cs.client.Client, tx)
	if err != nil {
		return fmt.Errorf("erro ao confirmar transação: %w", err)
	}
	
	log.Printf(" [Blockchain] Transferência confirmada")
	return nil
}

// =================== CONSULTAS =================================

// Verificar dono de uma carta específica
func (cs *CardChainService) GetCardOwner(ctx context.Context, cardID string) (string, error) {

	owner, err := cs.contracts.Card.GetCardOwner(cs.client.CallOpts(), cardID)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar dono da carta: %w", err)
	}

	return owner.Hex(), nil
}

// Pegar todas as cartas de um jogador
func (cs *CardChainService) GetPlayerCards(ctx context.Context, playerAddress string) ([]uint64, error) {
	playerAddr := common.HexToAddress(playerAddress)
	
	tokenIds, err := cs.contracts.Card.GetPlayerCards(cs.client.CallOpts(), playerAddr)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar cartas do jogador: %w", err)
	}

	result := make([]uint64, len(tokenIds))
	for i, tokenId := range tokenIds {
		result[i] = tokenId.Uint64()
	}

	log.Printf(" [Blockchain] Jogador %s possui %d cartas", playerAddress, len(result))
	return result, nil
}


func (cs *CardChainService) GetCardMetadata(ctx context.Context, tokenID uint64) (*CardMetadata, error) {
	cardData, err := cs.contracts.Card.GetCardMetadata(
		cs.client.CallOpts(),
		big.NewInt(int64(tokenID)),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar metadados: %w", err)
	}

	return &CardMetadata{
		CardID:      cardData.CardId,
		TemplateID:   cardData.TemplateId,
		PackageID:    cardData.PackageId,
		MintedAt:     cardData.MintedAt.Uint64(),
		CurrentOwner: cardData.CurrentOwner.Hex(),
	}, nil
}


func (cs *CardChainService) GetTotalCards(ctx context.Context) (uint64, error) {
	total, err := cs.contracts.Card.GetTotalCards(cs.client.CallOpts())
	if err != nil {
		return 0, err
	}
	return total.Uint64(), nil
}

// SEBASEAR NO TESTE NO ARQUIVO DO BLOCKCHAIN/TESTE/TRADE


// METODO QUE DEVE SER CHAMADO NO SERVIÇO DE TRADE (QUEM REQUSITA A TROCA)
func (cs *CardChainService) SwapCards(
	ctx context.Context,
	myTokenID uint64,        // Carta que você está oferecendo
	theirTokenID uint64,     // Carta que você quer receber
	myPrivateKey string,     // Sua chave privada
) error {
	
	auth, err := cs.client.NewTransactorFromPrivateKey(ctx, myPrivateKey)
	if err != nil {
		return fmt.Errorf("erro ao criar transactor: %w", err)
	}

	tx, err := cs.contracts.Card.SwapCards(
		auth,
		big.NewInt(int64(myTokenID)),
		big.NewInt(int64(theirTokenID)),
	)
	if err != nil {
		return fmt.Errorf("erro ao trocar cartas: %w", err)
	}

	log.Printf(" [Blockchain] Troca iniciada: Token %d <-> Token %d. TX: %s", 
		myTokenID, theirTokenID, tx.Hash().Hex())
	
	receipt, err := bind.WaitMined(ctx, cs.client.Client, tx)
	if err != nil {
		return fmt.Errorf("erro ao confirmar troca: %w", err)
	}

	log.Printf(" [Blockchain] Troca confirmada no bloco %d", receipt.BlockNumber.Uint64())
	return nil
}

//  permite que outro jogador execute a troca
// METODOD QUE DEVE SER CHAMADO QUANDO VC RECEBE UM REQUSISÇÃO DE TROCA DE CARTA (TBM NO TRADE SERVICE)
func (cs *CardChainService) ApproveForSwap(
	ctx context.Context,
	tokenID uint64,
	swapperAddress string,
	ownerPrivateKey string,
) error {
	
	auth, err := cs.client.NewTransactorFromPrivateKey(ctx, ownerPrivateKey)
	if err != nil {
		return fmt.Errorf("erro ao criar transactor: %w", err)
	}

	swapperAddr := common.HexToAddress(swapperAddress)

	tx, err := cs.contracts.Card.ApproveForSwap(
		auth,
		big.NewInt(int64(tokenID)),
		swapperAddr,
	)
	if err != nil {
		return fmt.Errorf("erro ao aprovar carta: %w", err)
	}

	log.Printf(" [Blockchain] Carta %d aprovada para troca com %s", tokenID, swapperAddress)
	
	_, err = bind.WaitMined(ctx, cs.client.Client, tx)
	return err
}

func (cs *CardChainService) GetTokenIDByCardID(ctx context.Context, cardID string) (uint64, error) {
    // Chama o mapeamento público cardIdToTokenId do contrato Card.sol
    tokenIdBig, err := cs.contracts.Card.CardIdToTokenId(cs.client.CallOpts(), cardID)
    if err != nil {
        return 0, fmt.Errorf("erro ao consultar contrato: %w", err)
    }

    if tokenIdBig.Cmp(big.NewInt(0)) == 0 {
        return 0, fmt.Errorf("carta %s não encontrada na blockchain", cardID)
    }

    return tokenIdBig.Uint64(), nil
}