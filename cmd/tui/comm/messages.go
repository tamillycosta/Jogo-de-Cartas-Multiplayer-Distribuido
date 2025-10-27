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

// QueueResponseData para respostas da fila
type QueueResponseData struct {
	Type      string `json:"type"`
	Status    string `json:"status"`
	QueueSize int    `json:"queue_size"`
}

// MatchFoundData para quando uma partida é encontrada
type MatchFoundData struct {
	Type     string        `json:"type"`
	MatchID  string        `json:"match_id"`
	PlayerID string        `json:"player_id"`
	Deck     []interface{} `json:"your_deck"`
}

// GameStateData representa o estado do jogo
type GameStateData struct {
	EventType      string       `json:"event_type"`
	CurrentTurn    string       `json:"current_turn"`
	TurnNumber     int          `json:"turn_number"`
	LocalPlayer    *PlayerData  `json:"local_player"`
	RemotePlayer   *PlayerData  `json:"remote_player"`
	WinnerID       string       `json:"winner_id,omitempty"`
	WinnerUsername string       `json:"winner_username,omitempty"`
}

type PlayerData struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	HP          int       `json:"hp"`
	CurrentCard *CardData `json:"current_card"`
}

type CardData struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Attack int    `json:"attack"`
	HP     int    `json:"hp"`
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

// QueueJoinedMsg quando entramos na fila
type QueueJoinedMsg struct {
	QueueSize int
}

// MatchFoundMsg quando uma partida é encontrada
type MatchFoundMsg struct {
	MatchID  string
	PlayerID string
	Deck     []interface{}
}

// GameUpdateMsg para atualizações do estado do jogo
type GameUpdateMsg struct {
	EventType      string
	CurrentTurn    string
	TurnNumber     int
	LocalPlayer    *PlayerData
	RemotePlayer   *PlayerData
	WinnerUsername string
}

// SubscribeToMatchMsg para subscrever no tópico da partida
type SubscribeToMatchMsg struct {
	MatchID string
}

// ErrorMsg é enviada quando ocorre um erro
type ErrorMsg struct {
	Err error
}

// NoOpMsg é uma mensagem interna para forçar o loop de escuta
type NoOpMsg struct{}