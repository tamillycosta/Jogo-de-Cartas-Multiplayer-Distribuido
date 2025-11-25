package tradeService

import (
	contracts "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/service"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	raftService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/session"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
)

// Serviço para gerenciar trocas
type TradeService struct {
	apiClient      *client.Client
	raft           *raftService.RaftService
	sessionManager *session.SessionManager

	chainService   *contracts.ChainService
    playerRepo     *repository.PlayerRepository
}

func New(
	apiClient *client.Client,
	raft *raftService.RaftService,
	sessionManager *session.SessionManager,
	chainService *contracts.ChainService,       
	playerRepo *repository.PlayerRepository,
) *TradeService {
	return &TradeService{
		apiClient:      apiClient,
		raft:           raft,
		sessionManager: sessionManager,
		chainService:   chainService,
		playerRepo:     playerRepo,
	}
}

// RequestTrade é chamado pelo Pub/Sub. Ele verifica a posse e encaminha ao líder.
func (ts *TradeService) RequestTrade(clientID string, cardID, targetPlayerID string) error {
	playerID, exists := ts.sessionManager.GetPlayerID(clientID)
	if !exists {
		return errors.New("sessao do jogador nao encontrada")
	}
	
	if playerID == targetPlayerID {
		return errors.New("voce nao pode enviar cartas para si mesmo")
	}

	log.Printf("[TradeService] %s enviando carta %s para %s", playerID, cardID, targetPlayerID)

	cmd := comands.TradeCardsCommand{
		FromPlayerID: playerID,
		ToPlayerID:   targetPlayerID,
		CardID:       cardID,
		RequestID:    uuid.New().String(),
	}

	if !ts.raft.IsLeader() {
		return ts.forwardTradeRequestToLeader(cmd)
	}

	return ts.ExecuteTradeAsLeader(cmd)
}

// ExecuteTradeAsLeader aplica o comando Raft (chamado diretamente ou via HTTP)
func (ts *TradeService) ExecuteTradeAsLeader(cmd comands.TradeCardsCommand) error {
	log.Printf("[TradeService] Sou líder! Iniciando processo de transferência %s...", cmd.RequestID)

	// 1. Validar e Buscar dados necessários para Blockchain
	if ts.chainService != nil {
		log.Println("[TradeService] Executando na Blockchain...")
		
		// Buscar jogador remetente para obter Chave Privada
		fromPlayer, err := ts.playerRepo.FindById(cmd.FromPlayerID)
		if err != nil {
			return fmt.Errorf("erro ao buscar remetente: %v", err)
		}

		// Buscar jogador destinatário para obter Endereço Público
		toPlayer, err := ts.playerRepo.FindById(cmd.ToPlayerID)
		if err != nil {
			return fmt.Errorf("erro ao buscar destinatário: %v", err)
		}

		// Buscar o TokenID da carta na blockchain usando o UUID (CardID)
		ctx := context.Background()
		tokenID, err := ts.chainService.CardChainService.GetTokenIDByCardID(ctx, cmd.CardID)
		if err != nil {
			return fmt.Errorf("erro ao resolver CardID para TokenID: %v", err)
		}

		// Executar transferência na Blockchain (TransferCard do card_service.go)
		// Isso vai usar a chave privada do fromPlayer para assinar a transação
		err = ts.chainService.CardChainService.TransferCard(
			ctx,
			tokenID,
			toPlayer.Address,
			fromPlayer.PrivateKey,
		)
		if err != nil {
			log.Printf("❌ [TradeService] Falha na Blockchain: %v", err)
			return fmt.Errorf("falha na transação blockchain: %v", err)
		}
		log.Println("✅ [TradeService] Sucesso na Blockchain via TransferCard!")
	} else {
		log.Println("⚠️ [TradeService] Blockchain service não configurado, pulando etapa on-chain.")
	}

	// 2. Se a blockchain confirmou (ou foi pulada), aplica no banco local via Raft
	log.Printf("[TradeService] Replicando estado no banco de dados (Raft)...")

	cmdData, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("erro ao serializar comando: %v", err)
	}

	raftCmd := comands.Command{
		Type:      comands.CommandTradeCards,
		Data:      cmdData,
		RequestID: cmd.RequestID,
	}

	response, err := ts.raft.ApplyCommand(raftCmd)
	if err != nil {
		log.Printf("[TradeService] Erro ao aplicar comando no Raft: %v", err)
		return fmt.Errorf("erro ao processar comando: %v", err)
	}

	if !response.Success {
		log.Printf("[TradeService] Comando rejeitado pelo Raft: %s", response.Error)
		return fmt.Errorf("falha ao atualizar banco de dados: %s", response.Error)
	}

	log.Printf("[TradeService] Transferência %s concluída com sucesso (Chain + DB)!", cmd.RequestID)
	return nil
}

// forwardTradeRequestToLeader encaminha a solicitação para o líder
func (ts *TradeService) forwardTradeRequestToLeader(cmd comands.TradeCardsCommand) error {
	leaderAddr := ts.raft.GetLeaderHTTPAddr()
	if leaderAddr == "" {
		return errors.New("nenhum lider disponivel no momento, tente novamente")
	}

	log.Printf("[TradeService] Encaminhando solicitacao de troca para o lider: %s", leaderAddr)

	if err := ts.apiClient.TradeInterface.AskForTrade(leaderAddr, cmd); err != nil {
		return fmt.Errorf("erro ao contatar lider: %v", err)
	}

	return nil
}