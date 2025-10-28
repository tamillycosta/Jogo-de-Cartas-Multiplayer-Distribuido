package comm

import (
	"encoding/json"
)

// --- Mensagens do Protocolo JSON (Cliente <-> Servidor) ---

// ClientMsg é a estrutura que enviamos para o servidor
type ClientMsg struct {
	Type  string      `json:"type"`
	Topic string      `json:"topic"`
	Data  interface{} `json:"data,omitempty"`
}

// ServerMsg é a estrutura genérica que recebemos do servidor
type ServerMsg struct {
	Type     string          `json:"type"`
	ClientID string          `json:"client_id,omitempty"`
	Topic    string          `json:"topic,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

// AuthResponseData é a estrutura aninhada dentro de ServerMsg.Data
type AuthResponseData struct {
	Type    string      `json:"type"`
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Player  interface{} `json:"player,omitempty"`
}

// --- Mensagens Internas do Bubble Tea (tea.Msg) ---

// ConnectedMsg é enviada quando o WebSocket conecta com sucesso
type ConnectedMsg struct {
	ClientID string
}

// AuthResponseMsg é enviada quando recebemos uma resposta de auth
type AuthResponseMsg struct {
	Success bool
	Message string
	Error   string
}

// ErrorMsg é enviada quando ocorre um erro
type ErrorMsg struct {
	Err error
}

// NoOpMsg é uma mensagem interna para forçar o loop de escuta
type NoOpMsg struct{}