package tradehandler

import (
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
}

func New(tradeService *tradeService.TradeService, broker *pubsub.Broker, authSession *session.SessionManager) *TradeTopicHandler {
	return &TradeTopicHandler{
		tradeService: tradeService,
		broker:       broker,
		authSession:  authSession,
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

	// O cliente (Jogador A) envia a carta dele, a carta do Jogador B, e o ID do Jogador B
	cardAID, _ := dataMap["card_a_id"].(string)
	cardBID, _ := dataMap["card_b_id"].(string)
	playerBID, _ := dataMap["player_b_id"].(string)

	if cardAID == "" || cardBID == "" || playerBID == "" {
		return h.sendError(clientID, "card_a_id, card_b_id, e player_b_id sao obrigatorios")
	}

	// O serviço de troca cuida de verificar o líder e aplicar o comando Raft
	err := h.tradeService.RequestTrade(clientID, cardAID, cardBID, playerBID)

	if err != nil {
		log.Printf("[TradeHandler] Erro ao processar troca: %v", err)
		return h.sendError(clientID, err.Error())
	}

	// Se chegou aqui, o comando foi *aceito* (encaminhado ou aplicado)
	response := map[string]interface{}{
		"type":    "trade_requested",
		"success": true,
		"message": "Solicitacao de troca processada pelo cluster.",
	}

	h.publishResponse(clientID, response)

	// (Idealmente) Você também publicaria uma notificação para o Jogador B
	// h.notifyPlayerB(playerBID, ...);

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