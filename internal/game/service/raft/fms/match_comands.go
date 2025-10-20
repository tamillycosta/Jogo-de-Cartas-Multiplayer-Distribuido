package fms

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"

	matchstate "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/matchMacking/matchState"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"encoding/json"
	"log"
	"time"
)



func (f *GameFSM) applyJoinQueue(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.JoinQueueCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	entry := &entities.QueueEntry{
		PlayerID: cmd.PlayerID,
		ServerID: cmd.ServerID,
		JoinedAt:time.Now(),
	}

	f.matchmakingState.AddToQueue(entry)

	log.Printf("[FSM] Player %s adicionado à fila (servidor: %s) - Fila: %d",
		cmd.PlayerID, cmd.ServerID, f.matchmakingState.GetQueueSize())

	return &comands.ApplyResponse{
		Success: true,
		Data: map[string]interface{}{
			"player_id":  cmd.PlayerID,
			"queue_size": f.matchmakingState.GetQueueSize(),
		},
	}
}

func (f *GameFSM) applyLeaveQueue(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.LeaveQueueCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	f.matchmakingState.RemoveFromQueue(cmd.PlayerID)

	log.Printf("[FSM] Player %s removido da fila", cmd.PlayerID)

	return &comands.ApplyResponse{Success: true}
}

func (f *GameFSM) applyCreateMatch(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.CreateMatchCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	match := &entities.Match{
		ID:            cmd.MatchID,
		Player1ID:     cmd.Player1ID,
		Player1Server: cmd.Player1Server,
		Player2ID:     cmd.Player2ID,
		Player2Server: cmd.Player2Server,
		HostServer:    cmd.HostServer,
		IsLocal:       cmd.IsLocal,
		Status:        "waiting",
		CreatedAt:     time.Now(),
	}

	f.matchmakingState.AddMatch(match)

	log.Printf("[FSM] Match criado: %s | P1=%s (srv=%s) vs P2=%s (srv=%s) | Host=%s | Local=%v",
		cmd.MatchID, cmd.Player1ID, cmd.Player1Server,
		cmd.Player2ID, cmd.Player2Server, cmd.HostServer, cmd.IsLocal)

	return &comands.ApplyResponse{Success: true, Data: match}
}

func (f *GameFSM) applyUpdateMatch(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.UpdateMatchCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	success := f.matchmakingState.UpdateMatch(cmd.MatchID, cmd.Status, cmd.NewHost)
	if !success {
		return &comands.ApplyResponse{Success: false, Error: "match not found"}
	}

	log.Printf("[FSM] Match atualizado: %s -> Status=%s", cmd.MatchID, cmd.Status)

	if cmd.NewHost != "" {
		log.Printf("[FSM] Match %s: Novo host = %s (failover)", cmd.MatchID, cmd.NewHost)
	}

	return &comands.ApplyResponse{Success: true}
}

func (f *GameFSM) applyEndMatch(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.EndMatchCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	success := f.matchmakingState.EndMatch(cmd.MatchID, cmd.WinnerID)
	if !success {
		return &comands.ApplyResponse{Success: false, Error: "match not found"}
	}

	log.Printf("[FSM] Match finalizado: %s | Vencedor=%s | Motivo=%s",
		cmd.MatchID, cmd.WinnerID, cmd.Reason)

	// TODO: Atualizar estatísticas dos jogadores aqui
	// f.updatePlayerStats(cmd.WinnerID, cmd.LoserID)

	return &comands.ApplyResponse{Success: true}
}

func (f *GameFSM) GetMatchmakingState() *matchstate.MatchmakingState {
    f.mu.RLock()
    defer f.mu.RUnlock()
    return f.matchmakingState
}