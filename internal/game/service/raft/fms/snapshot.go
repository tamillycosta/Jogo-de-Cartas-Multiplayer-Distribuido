package fms

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	matchstate "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/matchMacking/matchState"

	"encoding/json"

	"github.com/hashicorp/raft"
)

// implementa raft.FSMSnapshot

type GameSnapshot struct {
	players           []*entities.Player
	packages          []*entities.Package
	cards             []*entities.Card
	matchmakingState *matchstate.MatchmakingState
	processedRequests map[string]bool

}

func (s *GameSnapshot) Persist(sink raft.SnapshotSink) error {
	data := struct {
		Players           []*entities.Player  `json:"players"`
		Packages          []*entities.Package `json:"packages"`
		Cards             []*entities.Card    `json:"cards"`
		MatchMakingState   *matchstate.MatchmakingState  `json:"matchmaking_state"`
		ProcessedRequests map[string]bool     `json:"processed_requests"`
	}{
		Players:           s.players,
		Packages:          s.packages,
		Cards:             s.cards,
		ProcessedRequests: s.processedRequests,
		MatchMakingState: s.matchmakingState,
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