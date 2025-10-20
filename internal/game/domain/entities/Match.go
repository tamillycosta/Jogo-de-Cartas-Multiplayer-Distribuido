package entities

import "time"

// Metadados da partida (replicados via Raft)

type Match struct {
	ID            string    `json:"id"`
	Player1ID     string    `json:"player1_id"`
	Player1Server string    `json:"player1_server"`
	Player2ID     string    `json:"player2_id"`
	Player2Server string    `json:"player2_server"`
	HostServer    string    `json:"host_server"`
	IsLocal       bool      `json:"is_local"`
	Status        string    `json:"status"` // "waiting", "in_progress", "finished"
	CreatedAt     time.Time `json:"created_at"`
	EndedAt       *time.Time `json:"ended_at,omitempty"`
	WinnerID      string    `json:"winner_id,omitempty"`
}

// Representa Jogador na fila de espera
type QueueEntry struct {
	ClientID  string	`json:"client_id"`
	PlayerID  string    `json:"player_id"`
	ServerID  string    `json:"server_id"`
	JoinedAt  time.Time `json:"joined_at"`
}


type GamePlayer struct {
	ID       string
	Username string
	ClientID string // ID da conexão WebSocket
	Life   int
	Deck     []*Card
	CurrentCard  *Card
	
}
