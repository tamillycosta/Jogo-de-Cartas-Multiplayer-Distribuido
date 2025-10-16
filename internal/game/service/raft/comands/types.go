package comands

import (
	"encoding/json"
	"time"
)

// define os tipos de comandos que podem ser replicados entre os servidores 
type CommandType string

const (
	CommandCreateUser CommandType = "CREATE_USER"
	CommandDeleteUser CommandType = "DELETE_USER"
	CommandUpdateUser CommandType = "UPDATE_USER"
	CommanOpenPackage CommandType = "OPEN_PACKAGE"
)

// representa um comando a ser replicado via Raft
type Command struct {
	Type      CommandType     `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
	RequestID string          `json:"request_id"` // Para idempotência
}

// E a resposta após aplicar um comando
type ApplyResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}