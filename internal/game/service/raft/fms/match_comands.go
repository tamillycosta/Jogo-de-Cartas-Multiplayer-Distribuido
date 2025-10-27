package fms

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/fms/match"
	"encoding/json"
	"log"
	"time"
)


func (f *GameFSM) applyJoinGlobalQueue(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.JoinGlobalQueueCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	entry := &entities.GlobalQueueEntry{
		PlayerID: cmd.PlayerID,
		Username: cmd.Username,
		ServerID: cmd.ServerID,
		ClientID: cmd.ClientID,
		JoinedAt: time.Now(),
	}

	f.globalMatchmakingState.AddToGlobalQueue(entry)

	log.Printf("[FSM] Player %s adicionado à FILA GLOBAL (servidor: %s) - Fila: %d",
		cmd.Username, cmd.ServerID, f.globalMatchmakingState.GetGlobalQueueSize())

	return &comands.ApplyResponse{
		Success: true,
		Data: map[string]interface{}{
			"player_id":        cmd.PlayerID,
			"global_queue_size": f.globalMatchmakingState.GetGlobalQueueSize(),
		},
	}
}

func (f *GameFSM) applyLeaveGlobalQueue(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.LeaveGlobalQueueCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	f.globalMatchmakingState.RemoveFromGlobalQueue(cmd.PlayerID)

	log.Printf("[FSM] Player %s removido da fila global", cmd.PlayerID)

	return &comands.ApplyResponse{Success: true}
}

func (f *GameFSM) applyCreateRemoteMatch(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.CreateRemoteMatchCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	match := &entities.RemoteMatch{
		ID:              cmd.MatchID,
		Player1ID:       cmd.Player1ID,
		Player1Server:   cmd.Player1Server,
		Player1ClientID: cmd.Player1ClientID,
		Player2ID:       cmd.Player2ID,
		Player2Server:   cmd.Player2Server,
		Player2ClientID: cmd.Player2ClientID,
		HostServer:      cmd.HostServer,
		Status:          "waiting",
		CreatedAt:       time.Now(),
	}

	f.globalMatchmakingState.AddRemoteMatch(match)

	return &comands.ApplyResponse{Success: true, Data: match}
}

func (f *GameFSM) applyUpdateRemoteMatch(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.UpdateRemoteMatchCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	success := f.globalMatchmakingState.UpdateRemoteMatch(cmd.MatchID, cmd.Status, cmd.NewHost, cmd.WinnerID)
	if !success {
		return &comands.ApplyResponse{Success: false, Error: "match not found"}
	}

	log.Printf("[FSM] Match remoto atualizado: %s -> Status=%s", cmd.MatchID, cmd.Status)


	return &comands.ApplyResponse{Success: true}
}

func (f *GameFSM) applyEndRemoteMatch(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.EndRemoteMatchCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	success := f.globalMatchmakingState.EndRemoteMatch(cmd.MatchID, cmd.WinnerID)
	if !success {
		return &comands.ApplyResponse{Success: false, Error: "match not found"}
	}


	return &comands.ApplyResponse{Success: true}
}

func (f *GameFSM) GetGlobalMatchmakingState() *match.GlobalMatchmakingState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.globalMatchmakingState
}