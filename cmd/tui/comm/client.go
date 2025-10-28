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
	
	// *** CORREÇÃO AQUI ***
	// O log mostra que o tópico da mensagem de resposta é "response"
	if msg.Topic == "response" {
		var authData AuthResponseData
		if err := json.Unmarshal(msg.Data, &authData); err != nil {
			log.Printf("[WS] Erro ao decodificar AuthResponseData: %v. Dados brutos: %s", err, string(msg.Data))
			return ErrorMsg{fmt.Errorf("erro ao decodificar AuthResponseData: %v", err)}
		}
		
		log.Printf("[WS] Recebido AuthResponse: %s (Success: %t)", authData.Type, authData.Success)
		return AuthResponseMsg{
			Success: authData.Success,
			Message: authData.Message,
			Error:   authData.Error,
		}
	}
	
	// Retorna nil se não for uma mensagem de auth (será tratado como NoOpMsg)
	return nil
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