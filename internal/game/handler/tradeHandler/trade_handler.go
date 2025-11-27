package tradehandler

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/session"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/tradeService"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	"errors"
	"fmt"
	"log"
)

// Handler para tópicos de troca via Pub/Sub
type TradeTopicHandler struct {
	tradeService *tradeService.TradeService
	broker       *pubsub.Broker
	authSession  *session.SessionManager
	playerRepo   *repository.PlayerRepository
}

func New(tradeService *tradeService.TradeService, broker *pubsub.Broker, authSession *session.SessionManager, playerRepo *repository.PlayerRepository) *TradeTopicHandler {
	return &TradeTopicHandler{
		tradeService: tradeService,
		broker:       broker,
		authSession:  authSession,
		playerRepo:   playerRepo,
	}
}

func (h *TradeTopicHandler) HandleTopic(clientID string, topic string, data interface{}) error {
	log.Printf("[TradeHandler] Topic: %s, Cliente: %s", topic, clientID)

	switch topic {
	case "trade.request_trade":
		return h.handleRequestTrade(clientID, data)
	default:
		return fmt.Errorf("topico n encontrado: %s", topic)
	}
}

// HANDLER DO PUB SUB PARA SOLICITACAO DE TROCA
func (h *TradeTopicHandler) handleRequestTrade(clientID string, data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return h.sendError(clientID, "formato de dados invalido")
	}

	cardID, _ := dataMap["card_id"].(string)
	targetUsername, _ := dataMap["target_username"].(string)
	wantedCardID, _ := dataMap["wanted_card_id"].(string)

	if cardID == "" || targetUsername == "" || wantedCardID == "" {
		return h.sendError(clientID, "Use: trade <SuaCartaID> <UsuarioDestino> <CartaDeleID>")
	}

	// 1. Buscar o ID do jogador destinatário pelo NOME usando o Repositório
	targetPlayer, err := h.playerRepo.FindByUsername(targetUsername)
	if err != nil {
		log.Printf("Erro ao buscar usuario: %v", err)
		return h.sendError(clientID, "Erro interno ao buscar usuario destino")
	}
	if targetPlayer == nil {
		return h.sendError(clientID, fmt.Sprintf("Usuario '%s' nao encontrado", targetUsername))
	}

	log.Printf("[TradeHandler] Username '%s' resolvido para ID: %s", targetUsername, targetPlayer.ID)

	// 2. Chamar o serviço com os IDs
	err = h.tradeService.RequestTrade(clientID, cardID, targetPlayer.ID, wantedCardID)

	if err != nil {
		log.Printf("[TradeHandler] Erro ao processar troca: %v", err)
		return h.sendError(clientID, err.Error())
	}

	response := map[string]interface{}{
		"type":    "trade_requested",
		"success": true,
		"message": fmt.Sprintf("Carta enviada para %s com sucesso!", targetUsername),
	}

	h.publishResponse(clientID, response)
	return nil
}

// --- AUXILIARES ---

func (h *TradeTopicHandler) publishResponse(clientID string, response interface{}) {
	responseTopic := fmt.Sprintf("trade.response.%s", clientID)
	h.broker.Publish(responseTopic, map[string]interface{}{
		"topic": "trade.response",
		"data":  response,
	})
}

func (h *TradeTopicHandler) sendError(clientID string, errorMsg string) error {
	response := map[string]interface{}{
		"type":    "trade_error",
		"success": false,
		"error":   errorMsg,
	}
	h.publishResponse(clientID, response)
	return errors.New(errorMsg)
}

func (h *TradeTopicHandler) GetTopics() []string {
	return []string{
		"trade.request_trade",
	}
}