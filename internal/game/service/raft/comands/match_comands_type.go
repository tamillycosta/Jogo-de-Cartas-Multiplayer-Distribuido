package comands

// Comandos para matchmaking (replicados via Raft)
const (
	CommandJoinQueue     CommandType = "JOIN_QUEUE"
	CommandLeaveQueue    CommandType = "LEAVE_QUEUE"
	CommandCreateMatch   CommandType = "CREATE_MATCH"
	CommandUpdateMatch   CommandType = "UPDATE_MATCH"
	CommandEndMatch      CommandType = "END_MATCH"
)

// reuquest Jogador entra na fila
type JoinQueueCommand struct {
	PlayerID  string `json:"player_id"`
	ServerID  string `json:"server_id"`  
	Timestamp int64  `json:"timestamp"`
}

//  request Jogador sai da fila
type LeaveQueueCommand struct {
	PlayerID string `json:"player_id"`
}

// reuqest  Líder cria uma partida
type CreateMatchCommand struct {
	MatchID       string `json:"match_id"`
	Player1ID     string `json:"player1_id"`
	Player1Server string `json:"player1_server"`
	Player2ID     string `json:"player2_id"`
	Player2Server string `json:"player2_server"`
	HostServer    string `json:"host_server"`    // Servidor que vai hospedar
	IsLocal       bool   `json:"is_local"`       // Ambos jogadores no mesmo servidor?
}

//  Atualiza status da partida
type UpdateMatchCommand struct {
	MatchID    string `json:"match_id"`
	Status     string `json:"status"` // "waiting", "in_progress", "finished"
	NewHost    string `json:"new_host,omitempty"` // Em caso de failover
}

//  Finaliza partida e atualiza stats
type EndMatchCommand struct {
	MatchID   string `json:"match_id"`
	WinnerID  string `json:"winner_id"`
	LoserID   string `json:"loser_id"`
	Reason    string `json:"reason"` // "victory", "surrender", "disconnect"
	EndedAt   int64  `json:"ended_at"`
}