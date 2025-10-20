package matchstate

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"time"
)
// representa fila globak 
// Estado replicado em todos os servidores via Raft
type MatchmakingState struct {
	 // Fila global de jogadores aguardando
	Queue   []*entities.QueueEntry     
	Matches map[string]*entities.Match  
}



func New() *MatchmakingState {
	return &MatchmakingState{
		Queue:   make([]*entities.QueueEntry , 0),
		Matches: make(map[string]*entities.Match),
	}
}

// AddToQueue adiciona jogador à fila
func (ms *MatchmakingState) AddToQueue(entry *entities.QueueEntry) {
	// Remove se já estiver na fila (idempotência)
	ms.RemoveFromQueue(entry.PlayerID)
	ms.Queue = append(ms.Queue, entry)
}


func (ms *MatchmakingState) RemoveFromQueue(playerID string) {
	newQueue := make([]*entities.QueueEntry, 0)
	for _, entry := range ms.Queue {
		if entry.PlayerID != playerID {
			newQueue = append(newQueue, entry)
		}
	}
	ms.Queue = newQueue
}


func (ms *MatchmakingState) GetQueueSize() int {
	return len(ms.Queue)
}

//retorna os 2 primeiros da fila (sem remover)
func (ms *MatchmakingState) GetNextPair() (*entities.QueueEntry, *entities.QueueEntry) {
	if len(ms.Queue) < 2 {
		return nil, nil
	}
	return ms.Queue[0], ms.Queue[1]
}

//  adiciona partida ao estado
func (ms *MatchmakingState) AddMatch(match *entities.Match) {
	ms.Matches[match.ID] = match
	
	ms.RemoveFromQueue(match.Player1ID)
	ms.RemoveFromQueue(match.Player2ID)
}


func (ms *MatchmakingState) UpdateMatch(matchID string, status string, newHost string) bool {
	match, exists := ms.Matches[matchID]
	if !exists {
		return false
	}
	match.Status = status
	if newHost != "" {
		match.HostServer = newHost
	}
	return true
}

func (ms *MatchmakingState) EndMatch(matchID, winnerID string) bool {
	match, exists := ms.Matches[matchID]
	if !exists {
		return false
	}
	match.Status = "finished"
	match.WinnerID = winnerID
	now := time.Now()
	match.EndedAt = &now
	return true
}


func (ms *MatchmakingState) GetMatch(matchID string) (*entities.Match, bool) {
	match, exists := ms.Matches[matchID]
	return match, exists
}

func (ms *MatchmakingState) GetActiveMatches() []*entities.Match {
	active := make([]*entities.Match, 0)
	for _, match := range ms.Matches {
		if match.Status == "in_progress" || match.Status == "waiting" {
			active = append(active, match)
		}
	}
	return active
}