package remota

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"log"
	"time"
)

// ----------------- NOTIFICAÇÕES DE UMA PARTIDA REMOTA  -----------------

type GameStateUpdate struct {
	MatchID                 string         `json:"match_id"`
	CurrentTurnPlayerID     string         `json:"current_turn_player_id"`
	TurnNumber              int            `json:"turn_number"`
	Status                  string         `json:"status"`
	LocalPlayerLife         int            `json:"local_player_life"`
	LocalPlayerCurrentCard  *entities.Card `json:"local_player_current_card"`
	RemotePlayerLife        int            `json:"remote_player_life"`
	RemotePlayerCurrentCard *entities.Card `json:"remote_player_current_card"`
	Timestamp               int64          `json:"timestamp"`
}





func (s *RemoteGameSession) notifyLocalClient(eventType string) {
	topic := "match." + s.MatchID
	status := s.Status
	
	
	if status == "finished" && eventType != "match_ended" {
		log.Printf("[RemoteGame] Status=finished detectado, mudando evento para match_ended")
		eventType = "match_ended"
	}
	
	message := map[string]interface{}{
		"type":       "game_update",
		"event_type": eventType,
		"match_id":   s.MatchID,
		"game_state": s.buildGameState(),
		"timestamp":  time.Now().Unix(),
	}

	log.Printf("Publicando %s (status: %s)", eventType, status)
	
	
	go s.broker.Publish(topic, message)
}

func (s *RemoteGameSession) buildGameState() map[string]interface{} {
	localDeck := make([]map[string]interface{}, 0)
	if s.LocalPlayer != nil && len(s.LocalPlayer.Deck) > 0 {
		for i, card := range s.LocalPlayer.Deck {
			if card != nil {
				localDeck = append(localDeck, map[string]interface{}{
					"index":  i,
					"id":     card.ID,
					"name":   card.Name,
					"power":  card.Power,
					"health": card.Health,
				})
			}
		}
	}

	var localCurrentCard map[string]interface{}
	if s.LocalPlayer != nil && s.LocalPlayer.CurrentCard != nil {
		localCurrentCard = map[string]interface{}{
			"id":     s.LocalPlayer.CurrentCard.ID,
			"name":   s.LocalPlayer.CurrentCard.Name,
			"power":  s.LocalPlayer.CurrentCard.Power,
			"health": s.LocalPlayer.CurrentCard.Health,
		}
	}

	remotePlayer := map[string]interface{}{
		"id":       s.RemotePlayer.ID,
		"username": s.RemotePlayer.Username,
		"health":   s.RemotePlayer.Life,
	}

	if s.RemotePlayer.CurrentCard != nil {
		remotePlayer["current_card"] = map[string]interface{}{
			"id":     s.RemotePlayer.CurrentCard.ID,
			"name":   s.RemotePlayer.CurrentCard.Name,
			"power":  s.RemotePlayer.CurrentCard.Power,
			"health": s.RemotePlayer.CurrentCard.Health,
		}
	}

	gameState := map[string]interface{}{
		"local_player": map[string]interface{}{
			"id":           s.LocalPlayer.ID,
			"username":     s.LocalPlayer.Username,
			"health":       s.LocalPlayer.Life,
			"deck":         localDeck,
			"current_card": localCurrentCard,
		},
		"remote_player": remotePlayer,
		"current_turn":  s.CurrentTurnPlayerID,
		"turn_number":   s.TurnNumber,
		"status":        s.Status,
		"is_host":       s.IsHost,
	}

	
	if s.Status == "finished" {
		if s.LocalPlayer != nil && s.LocalPlayer.Life > 0 {
			gameState["winner_id"] = s.LocalPlayer.ID
			gameState["winner_username"] = s.LocalPlayer.Username
			log.Printf("[RemoteGame HOST] Vencedor: %s (LOCAL)", s.LocalPlayer.Username)
		} else if s.RemotePlayer != nil && s.RemotePlayer.Life > 0 {
			gameState["winner_id"] = s.RemotePlayer.ID
			gameState["winner_username"] = s.RemotePlayer.Username
			log.Printf("[RemoteGame HOST] Vencedor: %s (REMOTO)", s.RemotePlayer.Username)
		}
	}

	return gameState
}

