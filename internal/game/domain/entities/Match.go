package entities

import "time"

// Metadados da partida 

type Match struct {
	ID            string    `json:"id"`
	Player1ID     string    `json:"player1_id"`
	Player1Server string    `json:"player1_server"`
	Player2ID     string    `json:"player2_id"`
	Player2Server string    `json:"player2_server"`
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

// ----------------- Match Global --------------------

type GlobalQueueEntry struct {
	PlayerID  string    `json:"player_id"`
	Username  string    `json:"username"`
	ServerID  string    `json:"server_id"`
	ClientID  string    `json:"client_id"`
	JoinedAt  time.Time `json:"joined_at"`
}



type RemoteMatch struct {
	ID            string     `json:"id"`
	Player1ID     string     `json:"player1_id"`
	Player1Server string     `json:"player1_server"`
	Player1ClientID string   `json:"player1_client_id"`
	Player2ID     string     `json:"player2_id"`
	Player2Server string     `json:"player2_server"`
	Player2ClientID string   `json:"player2_client_id"`
	HostServer    string     `json:"host_server"`
	Status        string     `json:"status"` // "waiting", "in_progress", "finished"
	CreatedAt     time.Time  `json:"created_at"`
	EndedAt       *time.Time `json:"ended_at,omitempty"`
	WinnerID      string     `json:"winner_id,omitempty"`
}