package local

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"log"
	"time"
)


//	faz publicação no topico da partida  para
// compartilhar estado do jogo para ambos jogadores via WebSocket
func (s *LocalGameSession) broadcastGameState(eventType string) {
	message := map[string]interface{}{
		"type":       "game_update",
		"event_type": eventType,
		"match_id":   s.MatchID,
		"game_state": s.buildGameState(),
		"timestamp":  time.Now().Unix(),
	}

	// Publica no tópico da partida
	topic := "match." + s.MatchID
	s.broker.Publish(topic, message)

	log.Printf("[LocalGame] Estado enviado: %s | Event: %s", s.MatchID, eventType)
}



// Notifica os jogadores a ciração da  partida
// Serializa o deck
func (gsm *GameSessionManager) notifyMatchCreated(session *LocalGameSession, client1ID, client2ID string) {
	matchTopic := "match." + session.MatchID

	p1DeckInfo := gsm.serializeDeck(session.Player1.Deck)
	p2DeckInfo := gsm.serializeDeck(session.Player2.Deck)

	// Notificação para Player1
	notification1 := map[string]interface{}{
		"type":           "match_found",
		"match_id":       session.MatchID,
		"topic":          matchTopic,
		"auto_subscribe": true,
		"opponent":       session.Player2.Username,
		"your_deck":      p1DeckInfo,
		"message":        "Partida encontrada! Aguarde início...",
	}

	gsm.broker.Publish("response."+client1ID, notification1)

	// Notificação para Player2
	notification2 := map[string]interface{}{
		"type":           "match_found",
		"match_id":       session.MatchID,
		"topic":          matchTopic,
		"auto_subscribe": true,
		"opponent":       session.Player1.Username,
		"your_deck":      p2DeckInfo,
		"message":        "Partida encontrada! Aguarde início...",
	}

	gsm.broker.Publish("response."+client2ID, notification2)

	log.Printf("[SessionManager] Notificações enviadas | Tópico: %s", matchTopic)
}

// ------------------------ Auxiliares -----------------------

//  constroi estado do jogo para os jogadores 
func (s *LocalGameSession) buildGameState() map[string]interface{} {
	// Serializa deck do Player1
	p1Deck := make([]map[string]interface{}, 0)
	for i, card := range s.Player1.Deck {
		p1Deck = append(p1Deck, map[string]interface{}{
			"index":  i,
			"id":     card.ID,
			"name":   card.Name,
			"power":  card.Power,
			"health": card.Health,
			"rarity": card.Rarity,
		})
	}

	// Serializa deck do Player2
	p2Deck := make([]map[string]interface{}, 0)
	for i, card := range s.Player2.Deck {
		p2Deck = append(p2Deck, map[string]interface{}{
			"index":  i,
			"id":     card.ID,
			"name":   card.Name,
			"power":  card.Power,
			"health": card.Health,
			"rarity": card.Rarity,
		})
	}

	// Serializa carta atual do Player1
	var p1CurrentCard map[string]interface{}
	if s.Player1.CurrentCard != nil {
		p1CurrentCard = map[string]interface{}{
			"id":     s.Player1.CurrentCard.ID,
			"name":   s.Player1.CurrentCard.Name,
			"power":  s.Player1.CurrentCard.Power,
			"health": s.Player1.CurrentCard.Health,
		}
	}

	// Serializa carta atual do Player2
	var p2CurrentCard map[string]interface{}
	if s.Player2.CurrentCard != nil {
		p2CurrentCard = map[string]interface{}{
			"id":     s.Player2.CurrentCard.ID,
			"name":   s.Player2.CurrentCard.Name,
			"power":  s.Player2.CurrentCard.Power,
			"health": s.Player2.CurrentCard.Health,
		}
	}

	return map[string]interface{}{
		"player1": map[string]interface{}{
			"id":           s.Player1.ID,
			"username":     s.Player1.Username,
			"health":       s.Player1.Life,
			"deck":         p1Deck,
			"current_card": p1CurrentCard,
		},
		"player2": map[string]interface{}{
			"id":           s.Player2.ID,
			"username":     s.Player2.Username,
			"health":       s.Player2.Life,
			"deck":         p2Deck,
			"current_card": p2CurrentCard,
		},
		"current_turn": s.CurrentTurnPlayerID, 
		"turn_number":  s.TurnNumber,
		"status":       s.Status,
	}
}

func (gsm *GameSessionManager) serializeDeck(deck []*entities.Card) []map[string]interface{} {
	deckInfo := make([]map[string]interface{}, 0)

	for i, card := range deck {
		deckInfo = append(deckInfo, map[string]interface{}{
			"index":  i,
			"id":     card.ID,
			"name":   card.Name,
			"power":  card.Power,
			"health": card.Health,
			"rarity": card.Rarity,
		})
	}

	return deckInfo
}
