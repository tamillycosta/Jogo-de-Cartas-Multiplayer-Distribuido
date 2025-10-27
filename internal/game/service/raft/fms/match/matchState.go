package match

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"time"
)

// RESPONSSAVEL POR FAZER GERENCIAMENTO DE CLIENTES NA FILA DE PARTIDAS GLOBAIS
type GlobalMatchmakingState struct {
	GlobalQueue   []*entities.GlobalQueueEntry       // Fila global replicada
	RemoteMatches map[string]*entities.RemoteMatch   // Partidas remotas ativas
}



func New() *GlobalMatchmakingState {
	return &GlobalMatchmakingState{
		GlobalQueue:   make([]*entities.GlobalQueueEntry, 0),
		RemoteMatches: make(map[string]*entities.RemoteMatch),
	}
}


func (gms *GlobalMatchmakingState) AddToGlobalQueue(entry *entities.GlobalQueueEntry) {
	gms.RemoveFromGlobalQueue(entry.PlayerID)
	gms.GlobalQueue = append(gms.GlobalQueue, entry)
}


func (gms *GlobalMatchmakingState) RemoveFromGlobalQueue(playerID string) {
	newQueue := make([]*entities.GlobalQueueEntry, 0)
	for _, entry := range gms.GlobalQueue {
		if entry.PlayerID != playerID {
			newQueue = append(newQueue, entry)
		}
	}
	gms.GlobalQueue = newQueue
}

func (gms *GlobalMatchmakingState) GetGlobalQueueSize() int {
	return len(gms.GlobalQueue)
}

func (gms *GlobalMatchmakingState) GetNextGlobalPair() (*entities.GlobalQueueEntry, *entities.GlobalQueueEntry) {
	if len(gms.GlobalQueue) < 2 {
		return nil, nil
	}
	return gms.GlobalQueue[0], gms.GlobalQueue[1]
}


func (gms *GlobalMatchmakingState) AddRemoteMatch(match *entities.RemoteMatch) {
	gms.RemoteMatches[match.ID] = match
	gms.RemoveFromGlobalQueue(match.Player1ID)
	gms.RemoveFromGlobalQueue(match.Player2ID)
}


func (gms *GlobalMatchmakingState) UpdateRemoteMatch(matchID, status, newHost, winnerID string) bool {
	match, exists := gms.RemoteMatches[matchID]
	if !exists {
		return false
	}
	match.Status = status
	if newHost != "" {
		match.HostServer = newHost
	}
	if winnerID != "" {
		match.WinnerID = winnerID
	}
	return true
}


func (gms *GlobalMatchmakingState) EndRemoteMatch(matchID, winnerID string) bool {
	match, exists := gms.RemoteMatches[matchID]
	if !exists {
		return false
	}
	match.Status = "finished"
	match.WinnerID = winnerID
	now := time.Now()
	match.EndedAt = &now
	return true
}


func (gms *GlobalMatchmakingState) GetRemoteMatch(matchID string) (*entities.RemoteMatch, bool) {
	match, exists := gms.RemoteMatches[matchID]
	return match, exists
}