package fms

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	
	matchglobal "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/fms/match"

	"encoding/json"

	"github.com/hashicorp/raft"
)

// implementa raft.FSMSnapshot

type GameSnapshot struct {
	players           []*entities.Player
	packages          []*entities.Package
	cards             []*entities.Card
	globalMatchmakingState *matchglobal.GlobalMatchmakingState
	processedRequests map[string]bool

}

func (s *GameSnapshot) Persist(sink raft.SnapshotSink) error {
	data := struct {
		Players           []*entities.Player  `json:"players"`
		Packages          []*entities.Package `json:"packages"`
		Cards             []*entities.Card    `json:"cards"`
		GlobalMatchmakingState  *matchglobal.GlobalMatchmakingState  `json:"matchmaking_state"`
		ProcessedRequests map[string]bool     `json:"processed_requests"`
	}{
		Players:           s.players,
		Packages:          s.packages,
		Cards:             s.cards,
		ProcessedRequests: s.processedRequests,
		GlobalMatchmakingState: s.globalMatchmakingState,
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