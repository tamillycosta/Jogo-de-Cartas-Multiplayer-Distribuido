package main

import "fmt"
// topicos do pub sub 

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