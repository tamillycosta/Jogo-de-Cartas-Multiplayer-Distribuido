package gamesession

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/gameSession/local"
	remote "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/gameSession/remota"
	"log"
)


//------------------- Notificações enviadas durate uma partida remota para player local ----------------

// 
func (gsm *GameSessionManager) notifyClientRemoteMatchCreated(session *remote.RemoteGameSession, clientID string) {
	matchTopic := "match." + session.MatchID


	localDeckInfo := gsm.serializeDeck(session.LocalPlayer.Deck)

	notification := map[string]interface{}{
		"type":           "match_found",
		"match_id":       session.MatchID,
		"player_id":      session.LocalPlayer.ID,  
		"topic":          matchTopic,
		"auto_subscribe": true,
		"opponent":       session.RemotePlayer.Username,
		"your_deck":      localDeckInfo,
		"is_remote":      true,
		"is_host":        session.IsHost,
		"message":        "Partida remota encontrada! Aguarde início...",
	}

	gsm.broker.Publish("response."+clientID, notification)

	log.Printf(" [GameSessionManager] Notificação REMOTA enviada | PlayerID: %s | Tópico: %s", 
		session.LocalPlayer.ID, matchTopic)
}


func (gsm *GameSessionManager) notifyLocalMatchCreated(session *local.LocalGameSession, client1ID, client2ID string) {
	matchTopic := "match." + session.MatchID

	p1DeckInfo := gsm.serializeDeck(session.Player1.Deck)
	p2DeckInfo := gsm.serializeDeck(session.Player2.Deck)

	// Notificação para Player1
	notification1 := map[string]interface{}{
		"type":           "match_found",
		"match_id":       session.MatchID,
		"player_id":      session.Player1.ID,  
		"topic":          matchTopic,
		"auto_subscribe": true,
		"opponent":       session.Player2.Username,
		"your_deck":      p1DeckInfo,
		"is_remote":      false,
		"message":        "Partida encontrada! Aguarde início...",
	}

	gsm.broker.Publish("response."+client1ID, notification1)

	// Notificação para Player2
	notification2 := map[string]interface{}{
		"type":           "match_found",
		"match_id":       session.MatchID,
		"player_id":      session.Player2.ID,  
		"topic":          matchTopic,
		"auto_subscribe": true,
		"opponent":       session.Player1.Username,
		"your_deck":      p2DeckInfo,
		"is_remote":      false,
		"message":        "Partida encontrada! Aguarde início...",
	}

	gsm.broker.Publish("response."+client2ID, notification2)

	log.Printf("[SessionManager] Notificações LOCAL enviadas | P1: %s | P2: %s | Tópico: %s", 
		session.Player1.ID, session.Player2.ID, matchTopic)
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