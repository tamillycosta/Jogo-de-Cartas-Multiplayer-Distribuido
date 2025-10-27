package comm

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

// Client gerencia a conexão WebSocket
type Client struct {
	conn     *websocket.Conn
	ClientID string
	URL      string
}

func NewClient(url string) *Client {
	return &Client{URL: url}
}

// Connect é um tea.Cmd que tenta se conectar ao WebSocket
func (c *Client) Connect() tea.Cmd {
	return func() tea.Msg {
		conn, _, err := websocket.DefaultDialer.Dial(c.URL, nil)
		if err != nil {
			log.Printf("[WS] Erro ao conectar: %v", err)
			return ErrorMsg{fmt.Errorf("falha ao conectar com %s: %v", c.URL, err)}
		}
		c.conn = conn
		log.Println("[WS] Conectado! Aguardando ClientID...")
		
		// O servidor envia o 'connected' msg primeiro
		var msg ServerMsg
		if err := c.conn.ReadJSON(&msg); err != nil {
			return ErrorMsg{fmt.Errorf("erro ao ler msg de conexão: %v", err)}
		}

		if msg.Type == "connected" && msg.ClientID != "" {
			c.ClientID = msg.ClientID
			log.Printf("[WS] ClientID recebido: %s", c.ClientID)
			return ConnectedMsg{ClientID: c.ClientID}
		}
		
		return ErrorMsg{fmt.Errorf("protocolo inesperado: %s", msg.Type)}
	}
}

// Listen é um tea.Cmd que inicia a escuta de mensagens
func (c *Client) Listen() tea.Cmd {
	return func() tea.Msg {
		if c.conn == nil {
			return ErrorMsg{Err: errors.New("tentativa de escuta sem conexão")}
		}

		var serverMsg ServerMsg
		if err := c.conn.ReadJSON(&serverMsg); err != nil {
			return ErrorMsg{Err: fmt.Errorf("conexão perdida: %v", err)}
		}

		log.Printf("[WS] Mensagem bruta recebida: %+v", serverMsg)

		// Traduz a mensagem do servidor para uma tea.Msg
		if msg := c.parseServerMessage(serverMsg); msg != nil {
			return msg
		}

		return NoOpMsg{}
	}
}

// parseServerMessage traduz uma ServerMsg genérica em uma tea.Msg específica
func (c *Client) parseServerMessage(msg ServerMsg) tea.Msg {
	// Respostas de autenticação
	if msg.Topic == "auth.response" {
		var authData AuthResponseData
		if err := json.Unmarshal(msg.Data, &authData); err != nil {
			return ErrorMsg{fmt.Errorf("erro ao decodificar AuthResponseData: %v", err)}
		}
		
		log.Printf("[WS] Recebido AuthResponse: %s", authData.Type)
		return AuthResponseMsg{
			Success: authData.Success,
			Message: authData.Message,
			Error:   authData.Error,
		}
	}
	
	// Respostas da fila / matchmaking (tópico: response.{clientID})
	if msg.Topic == "response."+c.ClientID {
		var genericData map[string]interface{}
		if err := json.Unmarshal(msg.Data, &genericData); err != nil {
			return ErrorMsg{fmt.Errorf("erro ao decodificar resposta: %v", err)}
		}
		
		msgType, _ := genericData["type"].(string)
		
		switch msgType {
		case "queue_joined":
			queueSize := 0
			if qs, ok := genericData["queue_size"].(float64); ok {
				queueSize = int(qs)
			}
			log.Printf("[WS] Entrou na fila. Tamanho: %d", queueSize)
			return QueueJoinedMsg{QueueSize: queueSize}
			
		case "match_found":
			matchID, _ := genericData["match_id"].(string)
			playerID, _ := genericData["player_id"].(string)
			deck, _ := genericData["your_deck"].([]interface{})
			
			log.Printf("[WS] Match encontrado! ID: %s, PlayerID: %s", matchID, playerID)
			return MatchFoundMsg{
				MatchID:  matchID,
				PlayerID: playerID,
				Deck:     deck,
			}
			
		case "error":
			errorMsg, _ := genericData["error"].(string)
			return ErrorMsg{Err: errors.New(errorMsg)}
		}
	}
	
	// Atualizações do jogo (tópico: match.{matchID})
	if len(msg.Topic) > 6 && msg.Topic[:6] == "match." {
		var genericData map[string]interface{}
		if err := json.Unmarshal(msg.Data, &genericData); err != nil {
			return ErrorMsg{fmt.Errorf("erro ao decodificar game update: %v", err)}
		}
		
		msgType, _ := genericData["type"].(string)
		
		if msgType == "game_update" {
			// Extrai game_state
			gameStateData, ok := genericData["game_state"].(map[string]interface{})
			if !ok {
				return NoOpMsg{}
			}
			
			eventType, _ := gameStateData["event_type"].(string)
			currentTurn, _ := gameStateData["current_turn"].(string)
			turnNumber := 0
			if tn, ok := gameStateData["turn_number"].(float64); ok {
				turnNumber = int(tn)
			}
			
			var localPlayer, remotePlayer *PlayerData
			
			if lpData, ok := gameStateData["local_player"].(map[string]interface{}); ok {
				localPlayer = parsePlayerData(lpData)
			}
			
			if rpData, ok := gameStateData["remote_player"].(map[string]interface{}); ok {
				remotePlayer = parsePlayerData(rpData)
			}
			
			winnerUsername, _ := gameStateData["winner_username"].(string)
			
			log.Printf("[WS] Game Update: %s (Turn: %d)", eventType, turnNumber)
			return GameUpdateMsg{
				EventType:      eventType,
				CurrentTurn:    currentTurn,
				TurnNumber:     turnNumber,
				LocalPlayer:    localPlayer,
				RemotePlayer:   remotePlayer,
				WinnerUsername: winnerUsername,
			}
		}
	}
	
	return nil
}

// parsePlayerData converte map genérico em PlayerData
func parsePlayerData(data map[string]interface{}) *PlayerData {
	player := &PlayerData{}
	
	if id, ok := data["id"].(string); ok {
		player.ID = id
	}
	if username, ok := data["username"].(string); ok {
		player.Username = username
	}
	if hp, ok := data["hp"].(float64); ok {
		player.HP = int(hp)
	}
	
	if cardData, ok := data["current_card"].(map[string]interface{}); ok {
		player.CurrentCard = parseCardData(cardData)
	}
	
	return player
}

// parseCardData converte map genérico em CardData
func parseCardData(data map[string]interface{}) *CardData {
	card := &CardData{}
	
	if id, ok := data["id"].(string); ok {
		card.ID = id
	}
	if name, ok := data["name"].(string); ok {
		card.Name = name
	}
	if attack, ok := data["attack"].(float64); ok {
		card.Attack = int(attack)
	}
	if hp, ok := data["hp"].(float64); ok {
		card.HP = int(hp)
	}
	
	return card
}

// Subscribe é um tea.Cmd que envia uma mensagem de 'subscribe'
func (c *Client) Subscribe(topic string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("[WS] Inscrevendo no tópico: %s", topic)

		if c.conn == nil {
			log.Println("[WS] Erro: Tentativa de se inscrever sem conexão.")
			return ErrorMsg{Err: errors.New("não conectado ao servidor")}
		}

		msg := ClientMsg{
			Type:  "subscribe",
			Topic: topic,
		}
		if err := c.conn.WriteJSON(msg); err != nil {
			return ErrorMsg{Err: fmt.Errorf("falha ao se inscrever: %v", err)}
		}
		return nil
	}
}

// Publish é um tea.Cmd que envia uma mensagem de 'publish'
func (c *Client) Publish(topic string, data interface{}) tea.Cmd {
	return func() tea.Msg {
		log.Printf("[WS] Publicando no tópico: %s", topic)

		if c.conn == nil {
			log.Println("[WS] Erro: Tentativa de publicar sem conexão.")
			return ErrorMsg{Err: errors.New("não conectado ao servidor")}
		}

		msg := ClientMsg{
			Type:  "publish",
			Topic: topic,
			Data:  data,
		}
		if err := c.conn.WriteJSON(msg); err != nil {
			return ErrorMsg{Err: fmt.Errorf("falha ao publicar: %v", err)}
		}
		return nil
	}
}

// Close fecha a conexão
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}