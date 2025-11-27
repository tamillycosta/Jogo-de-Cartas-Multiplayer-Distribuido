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
func (ts *TradeService) RequestTrade(clientID string, cardID, targetPlayerID, wantedCardID string) error {
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
		WantedCardID: wantedCardID,
		RequestID:    uuid.New().String(),
	}

	if !ts.raft.IsLeader() {
		return ts.forwardTradeRequestToLeader(cmd)
	}

	return ts.ExecuteTradeAsLeader(cmd)
}

// ExecuteTradeAsLeader aplica o comando Raft (chamado diretamente ou via HTTP)
func (ts *TradeService) ExecuteTradeAsLeader(cmd comands.TradeCardsCommand) error {
	log.Printf("[TradeService] Iniciando TROCA (Swap) %s...", cmd.RequestID)

    // 1. Buscar dados dos jogadores (incluindo chaves privadas)
    playerA, err := ts.playerRepo.FindById(cmd.FromPlayerID) // Quem pediu
    if err != nil { return err }
    
    playerB, err := ts.playerRepo.FindById(cmd.ToPlayerID)   // O alvo
    if err != nil { return err }

    if ts.chainService != nil {
        ctx := context.Background()
        
        // Converter UUIDs para TokenIDs
        tokenID_A, err := ts.chainService.CardChainService.GetTokenIDByCardID(ctx, cmd.CardID)
        if err != nil { return err }
        
        tokenID_B, err := ts.chainService.CardChainService.GetTokenIDByCardID(ctx, cmd.WantedCardID)
        if err != nil { return err }

        log.Println("[Blockchain] Executando SwapCards...")

        // PASSO 1: Jogador B precisa aprovar o Jogador A (ou o servidor) para mover a carta dele
        // Usamos a chave privada do Player B para assinar a aprovação
        err = ts.chainService.CardChainService.ApproveForSwap(
            ctx, 
            tokenID_B,       // Carta do Jogador B
            playerA.Address, // Jogador A pode mexer nela
            playerB.PrivateKey, // Assinado por B
        )
        if err != nil {
            return fmt.Errorf("falha ao aprovar carta na blockchain: %v", err)
        }

        // PASSO 2: Jogador A executa a troca (Swap)
        // O contrato vai mover a Carta A para B, e puxar a Carta B para A
        err = ts.chainService.CardChainService.SwapCards(
            ctx,
            tokenID_A, // Carta que A está dando
            tokenID_B, // Carta que A está recebendo
            playerA.PrivateKey, // Assinado por A
        )
        if err != nil {
            return fmt.Errorf("falha no SwapCards blockchain: %v", err)
        }
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