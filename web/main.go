package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	// Conecta 2 clientes
	client1 := connectClient("MILLY")
	client2 := connectClient("MILLY2")

	time.Sleep(1 * time.Second)

	// Login
	login(client1, "MILLY")
	login(client2, "MILLY2")

	time.Sleep(500 * time.Millisecond)

	// Inscreve nos tópicos de resposta
	subscribe(client1, client1.clientID)
	subscribe(client2, client2.clientID)

	// Entra na fila
	fmt.Println("\n=== ENTRANDO NA FILA ===")
	joinQueue(client1)
	joinQueue(client2)

	// Aguarda match
	time.Sleep(4 * time.Second)

	if client1.matchID != "" {
		fmt.Println("\n=== PARTIDA ENCONTRADA ===")
		fmt.Printf("Match ID: %s\n", client1.matchID)

		// Inscreve no tópico da partida
		subscribe(client1, "match."+client1.matchID)
		subscribe(client2, "match."+client2.matchID)

		time.Sleep(1 * time.Second)

		
		// Aguardar game_update com match_started para ver current_turn
		time.Sleep(2 * time.Second)

		// Player que começa escolhe carta
		fmt.Println("\n=== TURNO 1: ESCOLHER CARTAS ===")
		chooseCard(client1, "0")
		time.Sleep(1 * time.Second)

		chooseCard(client2, "0")
		time.Sleep(1 * time.Second)

		// ta funcionando , so fiquei com preguiça de criar o handler da menssagem kkkkk
		//leave(client1)
	


		fmt.Println("\n=== TURNO 2: ATAQUES ===")
		attack(client1)
		time.Sleep(1 * time.Second)

		

		attack(client2)
		time.Sleep(1 * time.Second)

		
		fmt.Println("\n=== FIM DO TESTE ===")
	} else {
		fmt.Println("❌ Match não foi criado!")
	}

	client1.conn.Close()
	client2.conn.Close()
}

type Client struct {
	conn     *websocket.Conn
	clientID string
	matchID  string
	name     string
}

func connectClient(name string) *Client {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("Erro ao conectar:", err)
	}

	client := &Client{
		conn: conn,
		name: name,
	}

	// Recebe mensagem de conexão
	var connMsg map[string]interface{}
	conn.ReadJSON(&connMsg)
	client.clientID = connMsg["client_id"].(string)

	fmt.Printf(" %s conectado | ClientID: %s\n", name, client.clientID)

	// Goroutine para receber mensagens
	go client.receiveMessages()

	return client
}

func (c *Client) receiveMessages() {
	for {
		var msg map[string]interface{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			fmt.Printf("[%s] Desconectado\n", c.name)
			return
		}

		msgType, _ := msg["type"].(string)

		switch msgType {
		case "queue_joined":
			queueSize := msg["queue_size"]
			fmt.Printf("[%s] Entrou na fila | Tamanho: %.0f\n", c.name, queueSize)

		case "match_found":
			matchID, _ := msg["match_id"].(string)
			topic, _ := msg["topic"].(string)
			opponent, _ := msg["opponent"].(string)
			
			c.matchID = matchID
			
			fmt.Printf("[%s] Match encontrado!\n", c.name)
			fmt.Printf("   Match ID: %s\n", matchID)
			fmt.Printf("   Oponente: %s\n", opponent)
			fmt.Printf("   Tópico: %s\n", topic)
			
			// Auto-subscribe
			c.subscribeToTopic(topic)
			fmt.Printf(" [%s] Auto-inscrito em: %s\n", c.name, topic)
			
			// Mostra deck
			if deck, ok := msg["your_deck"].([]interface{}); ok {
				fmt.Printf(" [%s] Seu deck (%d cartas):\n", c.name, len(deck))
				for _, card := range deck {
					cardMap := card.(map[string]interface{})
					fmt.Printf("   [%v] %s (Power: %v, HP: %v)\n", 
						cardMap["index"], cardMap["name"], 
						cardMap["power"], cardMap["health"])
				}
			}

		case "game_update":
			eventType, _ := msg["event_type"].(string)
			fmt.Printf("[%s] Game Update: %s\n", c.name, eventType)

			if eventType == "match_started" {
				gameState := msg["game_state"].(map[string]interface{})
				currentTurn := gameState["current_turn"].(string)
				fmt.Printf("[%s] Partida INICIADA! Current Turn: %s\n", c.name, currentTurn)
			}

			if eventType == "action_performed" {
				gameState := msg["game_state"].(map[string]interface{})
				currentTurn, _ := gameState["current_turn"].(string)
				turnNum, _ := gameState["turn_number"].(float64)
				fmt.Printf("   Turn #%.0f | Current: %s\n", turnNum, currentTurn)
			}

			if eventType == "match_ended" {
				fmt.Printf("[%s] Partida FINALIZADA!\n", c.name)
				
			
			}

		case "error":
			errorMsg, _ := msg["error"].(string)
			fmt.Printf("[%s] Erro: %s\n", c.name, errorMsg)

		case "subscribed":
			topic, _ := msg["topic"].(string)
			fmt.Printf(" [%s] Inscrito em: %s\n", c.name, topic)
		}
	}
}

func (c *Client) subscribeToTopic(topic string) {
	msg := map[string]interface{}{
		"type":  "subscribe",
		"topic": topic,
	}
	c.conn.WriteJSON(msg)
}

func login(c *Client, username string) {
	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "auth.login",
		"data": map[string]interface{}{
			"username": username,
		},
	}
	c.conn.WriteJSON(msg)
	fmt.Printf("[%s] Login enviado\n", c.name)
}

func subscribe(c *Client, topic string) {
	msg := map[string]interface{}{
		"type":  "subscribe",
		"topic": "response." + topic,
	}
	c.conn.WriteJSON(msg)
}

func joinQueue(c *Client) {
	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "match.join_queue",
		"data":  map[string]interface{}{},
	}
	c.conn.WriteJSON(msg)
	fmt.Printf("🎮 [%s] Entrando na fila...\n", c.name)
}

func chooseCard(c *Client, cardIndex string) {
	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "match.play_card",
		"data": map[string]interface{}{
			"match_id":   c.matchID,
			"card_index": cardIndex,
		},
	}
	c.conn.WriteJSON(msg)
	fmt.Printf("[%s] Escolhendo carta %s\n", c.name, cardIndex)
}

func attack(c *Client) {
	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "match.attack",
		"data": map[string]interface{}{
			"match_id":         c.matchID,
			"attacker_card_id": "current",
		},
	}
	c.conn.WriteJSON(msg)
	fmt.Printf("[%s] Atacando!\n", c.name)
}


func leave(c *Client){
	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "match.surrender",
		"data": map[string]interface{}{
			"match_id":         c.matchID,
			"attacker_card_id": "current",
		},
	}
	c.conn.WriteJSON(msg)
	fmt.Printf("[%s] Desistindo!\n", c.name)
}