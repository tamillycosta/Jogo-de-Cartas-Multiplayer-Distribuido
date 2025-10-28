package main

import (
	"fmt"
	"log"
)

// ==================== MATCH / GAME ====================


func (c *Client) handleMatchFound(msg map[string]interface{}) {
	c.matchID, _ = msg["match_id"].(string)
	c.inMatch = true
	
	
	if playerID, ok := msg["player_id"].(string); ok {
		c.playerID = playerID
		log.Printf("[DEBUG] playerID: '%s'", c.playerID)
	}
	
	opponent, _ := msg["opponent"].(string)
	
	clearScreen()
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║       🎮  PARTIDA ENCONTRADA!  🎮     ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("\n🆔 Match ID: %s\n", c.matchID)
	fmt.Printf("⚔️  Oponente: %s\n", opponent)
	
	if deck, ok := msg["your_deck"].([]interface{}); ok {
		fmt.Printf("🎴 Seu deck: %d cartas\n", len(deck))
		fmt.Println("\n📋 Cartas disponíveis:")
		for i, card := range deck {
			if cardMap, ok := card.(map[string]interface{}); ok {
				name, _ := cardMap["name"].(string)
				power, _ := cardMap["power"].(float64)
				health, _ := cardMap["health"].(float64)
				fmt.Printf("   [%d] %s (⚔️ %.0f | ❤️ %.0f)\n", i, name, power, health)
			}
		}
	}
	
	fmt.Println("\n⏳ Aguardando início da partida...")
}

func (c *Client) handleGameUpdate(msg map[string]interface{}) {
	eventType, _ := msg["event_type"].(string)
	gameState, ok := msg["game_state"].(map[string]interface{})
	if !ok {
		return
	}

	// ✅ DEBUG: Veja qual evento está chegando
	log.Printf("[DEBUG] Evento recebido: %s", eventType)

	previousTurn := c.currentTurn
	
	if currentTurn, ok := gameState["current_turn"].(string); ok {
		c.currentTurn = currentTurn
	}
	
	if turnNum, ok := gameState["turn_number"].(float64); ok {
		c.turnNumber = int(turnNum)
	}

	switch eventType {
	case "match_started":
		clearScreen()
		fmt.Println("\n╔════════════════════════════════════════╗")
		fmt.Println("║       🎮  PARTIDA INICIADA!  🎮       ║")
		fmt.Println("╚════════════════════════════════════════╝")
		c.showGameState(gameState)
		c.showTurnInfo()

	case "card_played":
		fmt.Println("\n📇 ═══ CARTA JOGADA ═══")
		c.showGameState(gameState)
		c.showTurnInfo()

	case "attack_executed":
		fmt.Println("\n⚔️ ═══ ATAQUE EXECUTADO ═══")
		c.showGameState(gameState)
		c.showTurnInfo()

	case "match_ended":
		c.inMatch = false
		c.showMatchEnded(gameState)

	case "action_performed", "state_updated":  // ✅ Trata ambos eventos
		if previousTurn != c.currentTurn {
			fmt.Println("\n🔄 ═══ TURNO TROCADO ═══")
		}
		c.showGameState(gameState)
		c.showTurnInfo()

	default:
		// ✅ Mostra evento desconhecido
		fmt.Printf("\n🔔 [%s] Turno %d\n", eventType, c.turnNumber)
		c.showGameState(gameState)
		c.showTurnInfo()
	}
}

func (c *Client) showGameState(gameState map[string]interface{}) {
	fmt.Printf("\n📊 Estado do Jogo (Turno %d)\n", c.turnNumber)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// ========== SEU JOGADOR ==========
	var myPlayer map[string]interface{}
	if localPlayer, ok := gameState["local_player"].(map[string]interface{}); ok {
		myPlayer = localPlayer
	} else if player1, ok := gameState["player1"].(map[string]interface{}); ok {
		if id, _ := player1["id"].(string); id == c.playerID {
			myPlayer = player1
		} else {
			myPlayer, _ = gameState["player2"].(map[string]interface{})
		}
	}

	if myPlayer != nil {
		username, _ := myPlayer["username"].(string)
		health, _ := myPlayer["health"].(float64)
		
		cardName := "Nenhuma"
		cardPower := 0.0
		cardHealth := 0.0

		// ✅ FIX: Código limpo, sem duplicação
		if card, ok := myPlayer["current_card"].(map[string]interface{}); ok && card != nil {
			if name, ok := card["name"].(string); ok {
				cardName = name
			}
			if power, ok := card["power"].(float64); ok {
				cardPower = power
			}
			if hp, ok := card["health"].(float64); ok {
				cardHealth = hp
			}
		}
		
		fmt.Printf("\n👤 Você (%s)\n", username)
		fmt.Printf("   ❤️  HP: %.0f\n", health)
		if cardName != "Nenhuma" {
			fmt.Printf("   🎴 Carta: %s (⚔️ %.0f | ❤️ %.0f)\n", cardName, cardPower, cardHealth)
		} else {
			fmt.Printf("   🎴 Carta: Nenhuma\n")
		}
	}

	// ========== OPONENTE ==========
	var opponent map[string]interface{}
	if remotePlayer, ok := gameState["remote_player"].(map[string]interface{}); ok {
		opponent = remotePlayer
	} else if opponentData, ok := gameState["opponent"].(map[string]interface{}); ok {
		opponent = opponentData
	} else if player1, ok := gameState["player1"].(map[string]interface{}); ok {
		if id, _ := player1["id"].(string); id != c.playerID {
			opponent = player1
		} else {
			opponent, _ = gameState["player2"].(map[string]interface{})
		}
	}

	if opponent != nil {
		username, _ := opponent["username"].(string)
		health, _ := opponent["health"].(float64)
		
		cardName := "Nenhuma"
		cardPower := 0.0
		cardHealth := 0.0

		// ✅ FIX: Mesmo código limpo
		if card, ok := opponent["current_card"].(map[string]interface{}); ok && card != nil {
			if name, ok := card["name"].(string); ok {
				cardName = name
			}
			if power, ok := card["power"].(float64); ok {
				cardPower = power
			}
			if hp, ok := card["health"].(float64); ok {
				cardHealth = hp
			}
		}
		
		fmt.Printf("\n👥 Oponente (%s)\n", username)
		fmt.Printf("   ❤️  HP: %.0f\n", health)
		if cardName != "Nenhuma" {
			fmt.Printf("   🎴 Carta: %s (⚔️ %.0f | ❤️ %.0f)\n", cardName, cardPower, cardHealth)
		} else {
			fmt.Printf("   🎴 Carta: Nenhuma\n")
		}
	}
	
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}



func (c *Client) showTurnInfo() {
	isMyTurn := c.currentTurn == c.playerID
	
	if isMyTurn {
		fmt.Println("\n🎯 É SEU TURNO!")
		fmt.Println("💡 Comandos: card <index> | attack | help")
	} else {
		fmt.Println("\n⏳ Aguardando turno do oponente...")
	}
}

func (c *Client) showMatchEnded(gameState map[string]interface{}) {
	clearScreen()
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║       🏁  PARTIDA FINALIZADA!  🏁     ║")
	fmt.Println("╚════════════════════════════════════════╝")

	winnerID, _ := gameState["winner_id"].(string)
	winnerName, _ := gameState["winner_username"].(string)

	if winnerID == c.playerID {
		fmt.Println("\n🎉🎉🎉 VOCÊ VENCEU! 🎉🎉🎉")
	} else {
		fmt.Printf("\n😔 Você perdeu. Vencedor: %s\n", winnerName)
	}
	

	c.matchID = ""
	c.currentTurn = ""
	c.turnNumber = 0
	
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("💡 Digite 'queue(q)' para jogar novamente")
	fmt.Println("💡 Digite 'menu(m)' para voltar ao menu ")
}


