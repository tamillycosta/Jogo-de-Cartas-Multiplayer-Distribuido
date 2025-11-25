package tradeService

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	raftService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/session"
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
}

func New(
	apiClient *client.Client,
	raft *raftService.RaftService,
	sessionManager *session.SessionManager) *TradeService {
	return &TradeService{
		apiClient:      apiClient,
		raft:           raft,
		sessionManager: sessionManager,
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
	log.Printf("[TradeService] Sou líder! Processando troca %s via Raft...", cmd.RequestID)

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
		log.Printf("[TradeService] Comando rejeitado: %s", response.Error)
		return fmt.Errorf("falha ao realizar troca: %s", response.Error)
	}

	log.Printf("[TradeService] Troca %s concluida e replicada!", cmd.RequestID)
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