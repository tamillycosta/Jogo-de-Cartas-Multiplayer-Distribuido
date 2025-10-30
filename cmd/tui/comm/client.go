package comm

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings" // Precisa do "strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

// ... (Client struct, NewClient, Connect, Listen não mudam) ...
// Client gerencia a conexão WebSocket
type Client struct {
	conn     *websocket.Conn
	ClientID string
	URL      string
}

func NewClient(url string) *Client {
	return &Client{URL: url}
}

// Connect (Função sem alteração)
func (c *Client) Connect() tea.Cmd {
	return func() tea.Msg {
		conn, _, err := websocket.DefaultDialer.Dial(c.URL, nil)
		if err != nil {
			log.Printf("[WS] Erro ao conectar: %v", err)
			return ErrorMsg{fmt.Errorf("falha ao conectar com %s: %v", c.URL, err)}
		}
		c.conn = conn
		log.Println("[WS] Conectado! Aguardando ClientID...")
		
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

// Listen (Função sem alteração)
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

		if msg := c.parseServerMessage(serverMsg); msg != nil {
			return msg
		}

		return NoOpMsg{}
	}
}


// --- CORREÇÃO DO PARSE (Ignora "subscribed") ---
func (c *Client) parseServerMessage(msg ServerMsg) tea.Msg {
	
	// 1. Ignora mensagens de 'subscribed' para evitar o erro de JSON
	if msg.Type == "subscribed" {
		log.Printf("[WS] Confirmação de inscrição recebida para o tópico: %s. Ignorando.", msg.Topic)
		return nil // Retorna nil, que vira NoOpMsg no Listen()
	}

	// 2. Verificação de Auth (Baseado no seu log: Topic == "response")
	if msg.Topic == "response" {
		var authData AuthResponseData
		if err := json.Unmarshal(msg.Data, &authData); err != nil {
			log.Printf("[WS] Erro ao decodificar AuthResponseData: %v. Dados brutos: %s", err, string(msg.Data))
			return ErrorMsg{fmt.Errorf("erro ao decodificar AuthResponseData: %v", err)}
		}
		
		log.Printf("[WS] Recebido AuthResponse: %s (Success: %t)", authData.Type, authData.Success)
		return AuthResponseMsg{
			Success:  authData.Success,
			Message:  authData.Message,
			Error:    authData.Error,
			PlayerID: authData.Player.ID,
		}
	}
	
	// 3. Verificação de Pacote (Baseado no backend: "package.response.CLIENT_ID")
	// ✅ SOLUÇÃO 2: Decodificação em duas etapas
	if strings.HasPrefix(msg.Topic, "package.response.") {
		// Primeiro, decodifica o wrapper {"topic": "...", "data": {...}}
		var wrapper struct {
			Topic string          `json:"topic"`
			Data  json.RawMessage `json:"data"` // Mantém o JSON interno sem decodificar ainda
		}
		
		if err := json.Unmarshal(msg.Data, &wrapper); err != nil {
			log.Printf("[WS] Erro ao decodificar wrapper de PackageResponse: %v. Dados brutos: %s", err, string(msg.Data))
			return ErrorMsg{fmt.Errorf("erro ao decodificar wrapper: %v", err)}
		}
		
		log.Printf("[WS] Wrapper decodificado. Topic interno: %s", wrapper.Topic)
		
		// Depois, decodifica a resposta real que está dentro do 'data'
		var pkgData PackageResponseData
		if err := json.Unmarshal(wrapper.Data, &pkgData); err != nil {
			log.Printf("[WS] Erro ao decodificar PackageResponseData: %v. Dados brutos: %s", err, string(wrapper.Data))
			return ErrorMsg{fmt.Errorf("erro ao decodificar PackageResponseData: %v", err)}
		}

		log.Printf("[WS] Recebido PackageResponse: %s (Success: %t)", pkgData.Type, pkgData.Success)
		return PackageResponseMsg{
			Success: pkgData.Success,
			Message: pkgData.Message,
			Error:   pkgData.Error,
		}
	}

	// Retorna nil se não for nenhum dos dois
	return nil
}


// Subscribe (Função sem alteração)
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

// Publish (Função sem alteração)
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

// Close (Função sem alteração)
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}