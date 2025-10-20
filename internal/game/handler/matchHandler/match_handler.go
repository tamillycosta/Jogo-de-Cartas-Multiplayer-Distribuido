package matchhandler

import (
	"errors"
	"fmt"
	"log"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/gameSession/local"
	matchlocal "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/matchMacking/match_local"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/session"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
)

// MatchHandler - Gerencia tópicos relacionados a partidas
type MatchTopicHandler struct {
	localMatchmaking *matchlocal.LocalMatchmaking
	sessionManager   *local.GameSessionManager
	authSession      *session.SessionManager
	broker           *pubsub.Broker
}

func New(localMatchmaking *matchlocal.LocalMatchmaking, sessionManager *local.GameSessionManager, authSession *session.SessionManager, broker *pubsub.Broker) *MatchTopicHandler {
	return &MatchTopicHandler{
		localMatchmaking: localMatchmaking,
		sessionManager:   sessionManager,
		authSession:      authSession,
		broker:           broker,
	}
}






//  Implementa interface pubsub.HandleTopics
func (h *MatchTopicHandler) HandleTopic(clientID, topic string, data interface{}) error {
	switch topic {
	case "match.join_queue":
		return h.handleJoinQueue(clientID, data)

	case "match.play_card":
		return h.handlePlayCard(clientID, data)

	case "match.attack":
		return h.handleAttack(clientID, data)

	case "match.surrender":
		return h.handleLeaveMatch(clientID, data)

	default:
		return fmt.Errorf("tópico desconhecido: %s", topic)
	}
}






// ----------------- MATCHMAKING -----------------

func (h *MatchTopicHandler) handleJoinQueue(clientID string, data interface{}) error {
	log.Printf("[MatchHandler] Cliente %s quer entrar na fila", clientID)

	playerID, exists := h.authSession.GetPlayerID(clientID)
	if !exists {
		return h.sendError(clientID, "você não está autenticado")
	}

	if inMatch, matchID := h.sessionManager.IsPlayerInMatch(playerID); inMatch {
		return h.sendError(clientID, "você já está na partida: "+matchID)
	}

	player, err := h.authSession.GetSession(playerID)
	if err != nil {
		return h.sendError(clientID, "sessão não encontrada")
	}


	h.localMatchmaking.AddToQueue(clientID, playerID, player.Username)

	response := map[string]interface{}{
		"type":       "queue_joined",
		"status":     "queued",
		"queue_size": h.localMatchmaking.GetQueueSize(),
	}

	h.broker.Publish("response."+clientID, response)

	log.Printf("[MatchHandler] Player %s (%s) adicionado à fila local", player.Username, clientID)
	return nil
}


// -------------------------- AÇÕES DE JOGO -----------------------------

func (h *MatchTopicHandler) handlePlayCard(clientID string, data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return h.sendError(clientID, "dados inválidos")
	}

	matchID, _ := dataMap["match_id"].(string)
	cardID, _ := dataMap["card_index"].(string)

	if matchID == "" || cardID == "" {
		return h.sendError(clientID, "match_id e card_index são obrigatórios")
	}

	playerID, exists := h.authSession.GetPlayerID(clientID)
	if !exists {
		return h.sendError(clientID, "não autenticado")
	}
	// criar get card id by index
	action := entities.GameAction{
		Type:   "play_card",
		CardID: cardID,
	}

	if err := h.sessionManager.ProcessAction(matchID, playerID, action); err != nil {
		return h.sendError(clientID, err.Error())
	}

	log.Printf("[MatchHandler] Player %s jogou carta %s", playerID, cardID)
	return nil
}





func (h *MatchTopicHandler) handleAttack(clientID string, data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return h.sendError(clientID, "dados inválidos")
	}

	matchID, _ := dataMap["match_id"].(string)
	attackerCardID, _ := dataMap["attacker_card_id"].(string)

	if matchID == "" || attackerCardID == "" {
		return h.sendError(clientID, "match_id e attacker_card_id são obrigatórios")
	}

	playerID, exists := h.authSession.GetPlayerID(clientID)
	if !exists {
		return h.sendError(clientID, "não autenticado")
	}

	action := entities.GameAction{
		Type:           "attack",
		AttackerCardID: attackerCardID,
	}

	if err := h.sessionManager.ProcessAction(matchID, playerID, action); err != nil {
		return h.sendError(clientID, err.Error())
	}

	log.Printf("[MatchHandler] Player %s atacou", playerID)
	return nil
}




func (h *MatchTopicHandler) handleLeaveMatch(clientID string, data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return h.sendError(clientID, "dados inválidos")
	}

	matchID, _ := dataMap["match_id"].(string)

	if matchID == "" {
		return h.sendError(clientID, "match_id é obrigatório")
	}

	playerID, exists := h.authSession.GetPlayerID(clientID)
	if !exists {
		return h.sendError(clientID, "não autenticado")
	}


	action := entities.GameAction{
		Type:           "leave_match",
		AttackerCardID: "",
	}

	if err := h.sessionManager.ProcessAction(matchID, playerID, action); err != nil {
		return h.sendError(clientID, err.Error())
	}

	log.Printf("[MatchHandler] Player %s desistiu", playerID)
	return nil
}

// ----------------- HELPERS -----------------

func (h *MatchTopicHandler) sendError(clientID, errorMsg string) error {
	response := map[string]interface{}{
		"type":  "error",
		"error": errorMsg,
	}

	h.broker.Publish("response."+clientID, response)

	return errors.New(errorMsg)
}

// retorna tópicos que este handler gerencia
func (h *MatchTopicHandler) GetTopics() []string {
	return []string{
		"match.join_queue",
		"match.choose_card",
		"match.attack",
		"match.leave_match",
	}
}
