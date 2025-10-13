package fms

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"encoding/json"

	"github.com/hashicorp/raft"
)

// GameSnapshot implementa raft.FSMSnapshot

type GameSnapshot struct {
	players           []*entities.Player
	processedRequests map[string]bool
}

func (s *GameSnapshot) Persist(sink raft.SnapshotSink) error {
	data := struct {
		Players           []*entities.Player `json:"players"`
		ProcessedRequests map[string]bool    `json:"processed_requests"`
	}{
		Players:           s.players,
		ProcessedRequests: s.processedRequests,
	}

	if err := json.NewEncoder(sink).Encode(data); err != nil {
		sink.Cancel()
		return err
	}

	return sink.Close()
}

func (s *GameSnapshot) Release() {
	// Libera recursos se necessário
}