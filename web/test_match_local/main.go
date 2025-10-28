package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn     *websocket.Conn
	clientID string
	matchID  string
	username string
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("🌐 Conectando ao servidor WebSocket...")

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("Erro ao conectar:", err)
	}

	client := &Client{conn: conn}

	// Recebe client_id
	var msg map[string]interface{}
	conn.ReadJSON(&msg)
	client.clientID = msg["client_id"].(string)
	fmt.Printf("✅ Conectado! ClientID: %s\n", client.clientID)

	// Goroutine para receber mensagens
	go client.listen()

	// Login
	fmt.Print("\nDigite seu nome de usuário: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	client.username = username
	client.login(username)

	client.subscribe("response." + client.clientID)

	fmt.Println("\nComandos disponíveis:")
	fmt.Println("  queue        → entrar na fila")
	fmt.Println("  play [index] → jogar carta")
	fmt.Println("  attack       → atacar")
	fmt.Println("  exit         → sair")

	for {
		fmt.Print("> ")
		cmdLine, _ := reader.ReadString('\n')
		cmdLine = strings.TrimSpace(cmdLine)
		parts := strings.Split(cmdLine, " ")

		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "queue":
			client.joinQueue()
		case "play":
			if len(parts) < 2 {
				fmt.Println("Uso: play [index]")
				continue
			}
			client.playCard(parts[1])
		case "attack":
			client.attack()
		case "exit":
			fmt.Println("👋 Saindo...")
			client.conn.Close()
			return
		default:
			fmt.Println("❌ Comando inválido.")
		}
	}
}

func (c *Client) listen() {
	for {
		var msg map[string]interface{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("🔌 Conexão encerrada:", err)
			return
		}

		msgType, _ := msg["type"].(string)

		switch msgType {
		case "queue_joined":
			fmt.Printf("📥 Entrou na fila | Tamanho: %.0f\n", msg["queue_size"])
			

		case "match_found":
			c.matchID, _ = msg["match_id"].(string)
			opponent, _ := msg["opponent"].(string)
			topic, _ := msg["topic"].(string)
			fmt.Printf("\n🎯 Partida encontrada!\nMatchID: %s\nOponente: %s\nTópico: %s\n", c.matchID, opponent, topic)
			c.subscribe(topic)
			
		case "match_started":
			eventType, _ := msg["event_type"].(string)
			fmt.Printf("\n🎮 Game Update recebido: %s\n", eventType)

			if eventType == "match_started" {
				if gameState, ok := msg["game_state"].(map[string]interface{}); ok {
					currentTurn, _ := gameState["current_turn"].(string)
					fmt.Printf("🟢 Partida iniciada! Current Turn: %s\n", currentTurn)
				} else {
					fmt.Println("⚠️ match_started chegou SEM game_state!")
				}
			}

		case "game_update":
			eventType, _ := msg["event_type"].(string)
			fmt.Printf("\n🎮 Game Update recebido: %s\n", eventType)

			if eventType == "match_started" {
				if gameState, ok := msg["game_state"].(map[string]interface{}); ok {
					currentTurn, _ := gameState["current_turn"].(string)
					fmt.Printf("🟢 Partida iniciada! Current Turn: %s\n", currentTurn)
				} else {
					fmt.Println("⚠️ match_started chegou SEM game_state!")
				}
			}

			if eventType == "action_performed" {
				if gameState, ok := msg["game_state"].(map[string]interface{}); ok {
					currentTurn, _ := gameState["current_turn"].(string)
					turnNum, _ := gameState["turn_number"].(float64)
					fmt.Printf("🔁 Turno %.0f | Jogador atual: %s\n", turnNum, currentTurn)
				}
			}

			if eventType == "match_ended" {
				fmt.Println("🏆 Partida FINALIZADA!")
			}

		case "subscribed":
			topic, _ := msg["topic"].(string)
			fmt.Printf("📌 Inscrito em tópico: %s\n", topic)

		case "error":
			fmt.Println("❌ Erro:", msg["error"])

		default:
			b, _ := json.MarshalIndent(msg, "", "  ")
			fmt.Printf("📩 Mensagem desconhecida:\n%s\n", string(b))
		}
	}
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
	fmt.Printf("🔐 Login enviado para '%s'\n", username)
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
	fmt.Println("🎮 Entrando na fila...")

	
}

func (c *Client) playCard(index string) {
	if c.matchID == "" {
		fmt.Println("⚠️ Nenhuma partida ativa!")
		return
	}
	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "match.play_card",
		"data": map[string]interface{}{
			"match_id":   c.matchID,
			"card_index": index,
		},
	}
	c.conn.WriteJSON(msg)
	fmt.Printf("🃏 Jogando carta %s\n", index)
}

func (c *Client) attack() {
	if c.matchID == "" {
		fmt.Println("⚠️ Nenhuma partida ativa!")
		return
	}
	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "match.attack",
		"data": map[string]interface{}{
			"match_id":         c.matchID,
			"attacker_card_id": "current",
		},
	}
	c.conn.WriteJSON(msg)
	fmt.Println("⚔️ Atacando!")

}