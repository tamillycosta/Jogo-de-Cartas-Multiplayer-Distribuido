package main

import (
	"fmt"
	"log"
	"time"
	"strings"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn          *websocket.Conn
	clientID      string
	matchID       string
	name          string
	port          string
	currentTurn   string
	myPlayerID    string
	
	turnNumber    int
	matchStarted  bool
	matchEnded    bool  // ✅ NOVO: Detecta quando partida terminou
	winnerID      string
	winnerName    string
}

func main() {
	fmt.Println("=== TESTE INTELIGENTE DE PARTIDA REMOTA ===")

	client1 := connectClient("MILLY", "8080")
	client2 := connectClient("MILLY2", "8081")

	time.Sleep(1 * time.Second)

	login(client1, "MILLY")
	login(client2, "MILLY2")

	time.Sleep(500 * time.Millisecond)

	subscribe(client1, client1.clientID)
	subscribe(client2, client2.clientID)

	fmt.Println("\n=== ENTRANDO NA FILA ===")
	joinQueue(client1)
	joinQueue(client2)

	// Aguarda match ser criado E iniciado
	fmt.Println("⏳ Aguardando match...")
	maxWait := 40
	for i := 0; i < maxWait; i++ {
		time.Sleep(1 * time.Second)
		
		if client1.matchStarted && client2.matchStarted {
			fmt.Println("✅ Match iniciado para ambos!")
			break
		}
		
		if i%5 == 0 {
			fmt.Printf("   [%ds] C1 started: %v | C2 started: %v\n", 
				i, client1.matchStarted, client2.matchStarted)
		}
	}

	if !client1.matchStarted || !client2.matchStarted {
		fmt.Println("❌ Match não iniciou a tempo!")
		fmt.Printf("   C1: started=%v, matchID=%s\n", client1.matchStarted, client1.matchID)
		fmt.Printf("   C2: started=%v, matchID=%s\n", client2.matchStarted, client2.matchID)
		client1.conn.Close()
		client2.conn.Close()
		return
	}

	fmt.Println("\n=== PARTIDA ENCONTRADA! ===")
	fmt.Printf("Match ID: %s\n", client1.matchID)
	fmt.Printf("P1 PlayerID: %s\n", client1.myPlayerID)
	fmt.Printf("P2 PlayerID: %s\n", client2.myPlayerID)
	fmt.Printf("Turno inicial: %s\n", client1.currentTurn)

	time.Sleep(2 * time.Second)

	fmt.Println("\n=== INICIANDO JOGO ===")
	
	// Identifica quem joga primeiro
	var firstPlayer, secondPlayer *Client
	if client1.currentTurn == client1.myPlayerID {
		firstPlayer = client1
		secondPlayer = client2
	} else if client2.currentTurn == client2.myPlayerID {
		firstPlayer = client2
		secondPlayer = client1
	} else {
		fmt.Printf("⚠️ Turno inicial inválido! C1.turn=%s, C2.turn=%s\n", 
			client1.currentTurn, client2.currentTurn)
		client1.conn.Close()
		client2.conn.Close()
		return
	}

	fmt.Printf("[%s] É o primeiro a jogar!\n", firstPlayer.name)
	
	// ===== TURNO 1: Primeiro jogador escolhe carta =====
	fmt.Printf("\n[TURNO 1] %s escolhendo carta...\n", firstPlayer.name)
	chooseCard(firstPlayer, "0")
	time.Sleep(3 * time.Second)

	// ===== TURNO 2: Segundo jogador escolhe carta =====
	fmt.Printf("\n[TURNO 2] %s escolhendo carta...\n", secondPlayer.name)
	chooseCard(secondPlayer, "0")
	time.Sleep(3 * time.Second)

	// ✅ VERIFICA SE PARTIDA JÁ TERMINOU (não deveria neste ponto)
	if firstPlayer.matchEnded || secondPlayer.matchEnded {
		fmt.Println("\n⚠️ Partida terminou inesperadamente após escolha de cartas")
		printFinalResults(client1, client2)
		client1.conn.Close()
		client2.conn.Close()
		return
	}

	// ===== TURNO 3: Primeiro jogador ataca =====
	fmt.Printf("\n[TURNO 3] %s atacando...\n", firstPlayer.name)
	attack(firstPlayer)
	
	// ✅ Aguarda até que AMBOS recebam notificação de fim de partida OU timeout
	fmt.Println("\n⏳ Aguardando finalização da partida...")
	waitForMatchEnd := 20 // 10 segundos para sincronizar
	for i := 0; i < waitForMatchEnd; i++ {
		time.Sleep(1 * time.Second)
		
		if firstPlayer.matchEnded && secondPlayer.matchEnded {
			fmt.Println("✅ Ambos jogadores receberam notificação de fim!")
			break
		}
		
		if i == waitForMatchEnd-1 {
			fmt.Printf("⚠️ Timeout! C1 ended: %v | C2 ended: %v\n", 
				firstPlayer.matchEnded, secondPlayer.matchEnded)
		}
	}

	// ===== RESULTADO FINAL =====
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("           RESULTADO FINAL")
	fmt.Println(strings.Repeat("=", 50))
	printFinalResults(client1, client2)
	
	// ✅ Aguarda antes de desconectar
	time.Sleep(3 * time.Second)

	fmt.Println("\n=== TESTE CONCLUÍDO ===")
	client1.conn.Close()
	client2.conn.Close()
}

func printFinalResults(c1, c2 *Client) {
	fmt.Printf("\n📊 Cliente 1 (%s):\n", c1.name)
	fmt.Printf("   Match Started: %v\n", c1.matchStarted)
	fmt.Printf("   Match Ended: %v\n", c1.matchEnded)
	if c1.winnerID != "" {
		fmt.Printf("   Vencedor: %s (%s)\n", c1.winnerName, c1.winnerID)
	}
	
	fmt.Printf("\n📊 Cliente 2 (%s):\n", c2.name)
	fmt.Printf("   Match Started: %v\n", c2.matchStarted)
	fmt.Printf("   Match Ended: %v\n", c2.matchEnded)
	if c2.winnerID != "" {
		fmt.Printf("   Vencedor: %s (%s)\n", c2.winnerName, c2.winnerID)
	}
	
	// Verifica consistência
	if c1.matchEnded && c2.matchEnded {
		if c1.winnerID == c2.winnerID {
			fmt.Printf("\n✅ RESULTADO CONSISTENTE! Vencedor: %s\n", c1.winnerName)
		} else {
			fmt.Printf("\n❌ RESULTADO INCONSISTENTE!\n")
			fmt.Printf("   C1 vê vencedor: %s\n", c1.winnerName)
			fmt.Printf("   C2 vê vencedor: %s\n", c2.winnerName)
		}
	} else {
		fmt.Println("\n⚠️ Nem todos os clientes receberam fim de partida")
	}
}

func connectClient(name, port string) *Client {
	url := fmt.Sprintf("ws://localhost:%s/ws", port)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Erro ao conectar:", err)
	}

	client := &Client{
		conn:         conn,
		name:         name,
		port:         port,
		matchStarted: false,
		matchEnded:   false,
	}

	var connMsg map[string]interface{}
	conn.ReadJSON(&connMsg)
	client.clientID = connMsg["client_id"].(string)

	fmt.Printf("✅ %s conectado no servidor %s | ClientID: %s\n", name, port, client.clientID)

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
			fmt.Printf("[%s] Entrou na fila\n", c.name)

		case "match_found":
			c.matchID, _ = msg["match_id"].(string)
			opponent, _ := msg["opponent"].(string)
			topic, _ := msg["topic"].(string)
			fmt.Printf("\n🎯 Partida encontrada!\nMatchID: %s\nOponente: %s\nTópico: %s\n", c.matchID, opponent, topic)
		
			// Auto-subscribe se "auto_subscribe" == true
			if autoSub, ok := msg["auto_subscribe"].(bool); ok && autoSub {
				c.subscribeToTopic(topic)
				fmt.Printf("📌 Auto-inscrito em: %s\n", topic)
			}
			
			fmt.Printf("[%s] Match encontrado! ID: %s\n", c.name, c.matchID)

			if deck, ok := msg["your_deck"].([]interface{}); ok && len(deck) > 0 {
				fmt.Printf("[%s] Deck: %d cartas\n", c.name, len(deck))
			}

			

		case "game_update":
			eventType, _ := msg["event_type"].(string)

			if gameState, ok := msg["game_state"].(map[string]interface{}); ok {
				// Extrai playerID se ainda não tiver
				if c.myPlayerID == "" {
					if localPlayer, ok := gameState["local_player"].(map[string]interface{}); ok {
						if playerID, ok := localPlayer["id"].(string); ok && playerID != "" {
							c.myPlayerID = playerID
							fmt.Printf("✅ [%s] PlayerID obtido (backup): %s\n", c.name, playerID)
						}
					}
				}

				// Atualiza turno atual
				if currentTurn, ok := gameState["current_turn"].(string); ok {
					c.currentTurn = currentTurn
				}

				// Atualiza número do turno
				if turnNum, ok := gameState["turn_number"].(float64); ok {
					c.turnNumber = int(turnNum)
				}

				// ✅ CRÍTICO: Marca partida como iniciada
				if eventType == "match_started" {
					c.matchStarted = true
					fmt.Printf("🎮 [%s] Partida INICIADA! Turn: %s\n", c.name, c.currentTurn)
				}

				// ✅ NOVO: Detecta fim de partida
				if eventType == "match_ended" {
					c.matchEnded = true
					
					// Extrai informações do vencedor
					if winnerID, ok := gameState["winner_id"].(string); ok {
						c.winnerID = winnerID
					}
					if winnerName, ok := gameState["winner_username"].(string); ok {
						c.winnerName = winnerName
					}
					
					fmt.Printf("🏁 [%s] PARTIDA FINALIZADA! Vencedor: %s\n", c.name, c.winnerName)
					return
				}

				isMyTurn := c.currentTurn == c.myPlayerID
				turnIcon := "⏸️"
				if isMyTurn {
					turnIcon = "▶️"
				}

				fmt.Printf("%s [%s] %s | Turn: %d | MyTurn: %v\n",
					turnIcon, c.name, eventType, c.turnNumber, isMyTurn)
			}

		case "error":
			errorMsg, _ := msg["error"].(string)
			fmt.Printf("❌ [%s] Erro: %s\n", c.name, errorMsg)
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
	fmt.Printf("📤 [%s] Carta %s escolhida\n", c.name, cardIndex)
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
	fmt.Printf("📤 [%s] Ataque enviado\n", c.name)
}

func leaveMatch(c *Client) {
	msg := map[string]interface{}{
		"type":  "publish",
		"topic": "match.leave_match",
		"data": map[string]interface{}{
			"match_id": c.matchID,
		},
	}
	c.conn.WriteJSON(msg)
	fmt.Printf("📤 [%s] Leave enviado\n", c.name)
}