package comands

const (
	// ... seus comandos existentes ...
	
	//  - Comandos de matchmaking global
	CommandJoinGlobalQueue    CommandType = "JOIN_GLOBAL_QUEUE"
	CommandLeaveGlobalQueue   CommandType = "LEAVE_GLOBAL_QUEUE"
	CommandCreateRemoteMatch  CommandType = "CREATE_REMOTE_MATCH"
	CommandUpdateRemoteMatch  CommandType = "UPDATE_REMOTE_MATCH"
	CommandEndRemoteMatch     CommandType = "END_REMOTE_MATCH"
)

// Jogador entra na fila global
type JoinGlobalQueueCommand struct {
	PlayerID  string `json:"player_id"`
	Username  string `json:"username"`
	ServerID  string `json:"server_id"`  // Qual servidor o jogador está
	ClientID  string `json:"client_id"`
	JoinedAt  int64  `json:"joined_at"`
}

//  Jogador sai da fila global
type LeaveGlobalQueueCommand struct {
	PlayerID string `json:"player_id"`
}

//  Líder cria partida remota
type CreateRemoteMatchCommand struct {
	MatchID       string `json:"match_id"`
	Player1ID     string `json:"player1_id"`
	Player1Server string `json:"player1_server"`
	Player1ClientID string `json:"player1_client_id"`
	Player2ID     string `json:"player2_id"`
	Player2Server string `json:"player2_server"`
	Player2ClientID string `json:"player2_client_id"`
	HostServer    string `json:"host_server"` // Servidor escolhido como host
}

//  Atualiza status da partida
type UpdateRemoteMatchCommand struct {
	MatchID    string `json:"match_id"`
	Status     string `json:"status"` // "in_progress", "finished"
	WinnerID   string `json:"winner_id,omitempty"`
	NewHost    string `json:"new_host,omitempty"` // Em caso de failover
}

//  Finaliza partida remota
type EndRemoteMatchCommand struct {
	MatchID  string `json:"match_id"`
	WinnerID string `json:"winner_id"`
	Reason   string `json:"reason"` // "victory", "disconnect", "surrender"
	EndedAt  int64  `json:"ended_at"`
}