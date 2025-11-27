package main

import "fmt"
// topicos do pub sub 
func (c *Client) createAccount(username string) {
    c.conn.WriteJSON(map[string]interface{}{
		"type":  "publish",
        "topic": "auth.create_account",
        "data": map[string]interface{}{
            "username": username,
        },
    })
}


func (c *Client) login(username string) {
	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "auth.login",
		"data": map[string]interface{}{
			"username": username,
		},
	}
	c.conn.WriteJSON(msg)
}

func (c *Client) subscribe(topic string) {
	msg := map[string]interface{}{
		"type":  "subscribe",
		"topic": topic,
	}
	c.conn.WriteJSON(msg)
}

func (c *Client) joinQueue() {
	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "match.join_queue",
		"data":  map[string]interface{}{},
	}
	c.conn.WriteJSON(msg)
	fmt.Println("\n🔍 Procurando partida...")
}

func (c *Client) playCard(indexStr string) {

	if c.matchID == "" {
		fmt.Println("❌ Erro: matchID está vazio!")
		return
	}

	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "match.play_card",
		"data": map[string]interface{}{
			"match_id":   c.matchID,
			"card_index": indexStr, 
		},
	}
	
	
	c.conn.WriteJSON(msg)
	fmt.Printf("🃏 Jogando carta [%s]...\n", indexStr)
}

func (c *Client) attack() {
	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "match.attack",
		"data": map[string]interface{}{
			"match_id":         c.matchID,
			"attacker_card_id": "current",
		},
	}
	c.conn.WriteJSON(msg)
	fmt.Println("⚔️ Atacando...")
}

func (c *Client) openPack() {
	if c.playerID == "" {
		fmt.Println("❌ Erro: Você precisa fazer login primeiro!")
		fmt.Println("⚠️ PlayerID não foi definido. Tente fazer login novamente.")
		return
	}
	
	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "package.open_pack",
		"data": map[string]interface{}{
			"player_id": c.playerID,
		},
	}
	
	fmt.Printf("[DEBUG] Enviando openPack com player_id: %s", c.playerID)
	c.conn.WriteJSON(msg)
	fmt.Println("\n📦 Abrindo pacote de cartas...")
}

func (c *Client) leaveMatch() {
	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "match.leave_match",
		"data": map[string]interface{}{
			"match_id": c.matchID,
		},
	}
	c.conn.WriteJSON(msg)
	fmt.Println("desistindo da partida")
}

func (c *Client) listCards(targetUser string) {
	if c.playerID == "" {
		fmt.Println("❌ Você precisa fazer login primeiro!")
		return
	}

	data := map[string]interface{}{
		"player_id": c.playerID,
	}

	// Se o usuário passou um nome, adiciona ao payload
	if targetUser != "" {
		data["target_username"] = targetUser
		fmt.Printf("\n🔍 Espiando coleção de %s...\n", targetUser)
	} else {
		fmt.Println("\n📚 Buscando sua coleção de cartas...")
	}

	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "inventory.list",
		"data":  data,
	}
	c.conn.WriteJSON(msg)
}

func (c *Client) giveCard(cardUUID, targetUsername string) {
    // Verifica se o usuário está logado
    if c.playerID == "" {
        fmt.Println("❌ Você precisa fazer login primeiro!")
        return
    }

    // Monta a mensagem para o tópico trade.request_trade
    msg := map[string]interface{}{
        "type":  "publish",
        "topic": "trade.request_trade",
        "data": map[string]interface{}{
            "card_id":         cardUUID,       // O UUID que aparece no comando 'list'
            "target_username": targetUsername, // O nome do jogador destino
        },
    }

    c.conn.WriteJSON(msg)
    fmt.Printf("\n🎁 Enviando solicitação de transferência...\n")
    fmt.Printf("   Carta: %s\n", cardUUID)
    fmt.Printf("   Para:  %s\n", targetUsername)
}