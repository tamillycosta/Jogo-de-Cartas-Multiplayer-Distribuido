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
	WinnerID                string         `json:"winner_id"`
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
		"game_state": s.buildGameState(), // ✅ USA buildGameState normal (perspectiva local)
		"timestamp":  time.Now().Unix(),
	}

	// ✅ DEBUG: Log para ver o que está sendo enviado
	log.Printf("[RemoteGame] Notificando cliente local | Event: %s | Turn: %d | CurrentTurn: %s", 
		eventType, s.TurnNumber, s.CurrentTurnPlayerID)

	go s.broker.Publish(topic, message)
}

func (s *RemoteGameSession) buildGameState() map[string]interface{} {
	// Serializa deck do LOCAL player (do ponto de vista deste servidor)
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

	// Carta atual do LOCAL player
	var localCurrentCard map[string]interface{}
	if s.LocalPlayer != nil && s.LocalPlayer.CurrentCard != nil {
		localCurrentCard = map[string]interface{}{
			"id":     s.LocalPlayer.CurrentCard.ID,
			"name":   s.LocalPlayer.CurrentCard.Name,
			"power":  s.LocalPlayer.CurrentCard.Power,
			"health": s.LocalPlayer.CurrentCard.Health,
		}
	}

	// Remote player info
	remotePlayerInfo := map[string]interface{}{
		"id":       s.RemotePlayer.ID,
		"username": s.RemotePlayer.Username,
		"health":   s.RemotePlayer.Life,
	}

	if s.RemotePlayer.CurrentCard != nil {
		remotePlayerInfo["current_card"] = map[string]interface{}{
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
		"remote_player": remotePlayerInfo,
		"current_turn":  s.CurrentTurnPlayerID,
		"turn_number":   s.TurnNumber,
		"status":        s.Status,
		"is_host":       s.IsHost,
	}

	if s.Status == "finished" {
		if s.LocalPlayer.ID == s.WinnerID {
			gameState["winner_id"] = s.LocalPlayer.ID
			gameState["winner_username"] = s.LocalPlayer.Username
			log.Printf("[RemoteGame] Vencedor: %s (LOCAL)", s.LocalPlayer.Username)
		} else if s.RemotePlayer.ID == s.WinnerID {
			gameState["winner_id"] = s.RemotePlayer.ID
			gameState["winner_username"] = s.RemotePlayer.Username
			log.Printf("[RemoteGame] Vencedor: %s (REMOTO)", s.RemotePlayer.Username)
		}
	}

	return gameState
}


// buildGameStateForSync: para sincronizar com servidor REMOTO
func (s *RemoteGameSession) buildGameStateForSync() map[string]interface{} {
	return s.buildGameStateFor(true) // true = perspectiva invertida
}

// buildGameStateFor: função unificada que inverte conforme necessário
func (s *RemoteGameSession) buildGameStateFor(invertPerspective bool) map[string]interface{} {
	var localPlayer, remotePlayer *entities.GamePlayer
	
	if invertPerspective {
		// Para sync: inverte! O "remoto" deste servidor é o "local" do outro
		localPlayer = s.RemotePlayer
		remotePlayer = s.LocalPlayer
	} else {
		// Para notificação local: mantém perspectiva normal
		localPlayer = s.LocalPlayer
		remotePlayer = s.RemotePlayer
	}

	// Serializa deck do local
	localDeck := make([]map[string]interface{}, 0)
	if localPlayer != nil && len(localPlayer.Deck) > 0 {
		for i, card := range localPlayer.Deck {
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

	// Carta atual do local
	var localCurrentCard map[string]interface{}
	if localPlayer != nil && localPlayer.CurrentCard != nil {
		localCurrentCard = map[string]interface{}{
			"id":     localPlayer.CurrentCard.ID,
			"name":   localPlayer.CurrentCard.Name,
			"power":  localPlayer.CurrentCard.Power,
			"health": localPlayer.CurrentCard.Health,
		}
	}

	// Remote player info
	remotePlayerInfo := map[string]interface{}{
		"id":       remotePlayer.ID,
		"username": remotePlayer.Username,
		"health":   remotePlayer.Life,
	}

	if remotePlayer.CurrentCard != nil {
		remotePlayerInfo["current_card"] = map[string]interface{}{
			"id":     remotePlayer.CurrentCard.ID,
			"name":   remotePlayer.CurrentCard.Name,
			"power":  remotePlayer.CurrentCard.Power,
			"health": remotePlayer.CurrentCard.Health,
		}
	}

	gameState := map[string]interface{}{
		"local_player": map[string]interface{}{
			"id":           localPlayer.ID,
			"username":     localPlayer.Username,
			"health":       localPlayer.Life,
			"deck":         localDeck,
			"current_card": localCurrentCard,
		},
		"remote_player": remotePlayerInfo,
		"current_turn":  s.CurrentTurnPlayerID,
		"turn_number":   s.TurnNumber,
		"status":        s.Status,
		"is_host":       s.IsHost,
	}

	if s.Status == "finished" {
		if localPlayer.ID == s.WinnerID {
			gameState["winner_id"] = localPlayer.ID
			gameState["winner_username"] = localPlayer.Username
		} else if remotePlayer.ID == s.WinnerID {
			gameState["winner_id"] = remotePlayer.ID
			gameState["winner_username"] = remotePlayer.Username
		}
	}

	return gameState
}
