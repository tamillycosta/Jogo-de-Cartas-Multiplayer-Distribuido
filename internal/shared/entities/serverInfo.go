package entities


import (
	"time"
)


type ServerInfo struct {
	ID       string    `json:"server_id"`
	Region   string    `json:"region"`
	Address  string    `json:"address"`
	Port     int       `json:"port"`
	Load     float64   `json:"current_load"`
	Players  int       `json:"active_players"`
	GossipPort int    `json:"gossip_port"` 
	IsLeader bool	`json:"leader"`
	Status   string    `json:"status"`
	
}

type NotificationMessage struct {
	From    string      `json:"from_server"`
	Type    string      `json:"message_type"`
	Data    map[string]string `json:"data"`
	SentAt  time.Time   `json:"sent_at"`
}


