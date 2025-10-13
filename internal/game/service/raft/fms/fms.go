package fms

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/hashicorp/raft"
)

// implementa a interface raft.FSM
// É a máquina de estados que aplica comandos no banco de dados
type GameFSM struct {
	mu         sync.RWMutex
	repository *repository.PlayerRepository
	
	// Cache de IDs de requisições processadas (para idempotência)
	processedRequests map[string]bool
}

func New(repo *repository.PlayerRepository) *GameFSM {
	return &GameFSM{
		repository:        repo,
		processedRequests: make(map[string]bool),
	}
}


// é chamado quando um comando é commitado pelo Raft
func (f *GameFSM) Apply(logs *raft.Log) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	var cmd comands.Command
	if err := json.Unmarshal(logs.Data, &cmd); err != nil {
		return &comands.ApplyResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to unmarshal command: %v", err),
		}
	}

	// Verifica idempotência
	if cmd.RequestID != "" && f.processedRequests[cmd.RequestID] {
		log.Printf("Comando já processado: %s", cmd.RequestID)
		return &comands.ApplyResponse{
			Success: true,
			Data:    "already processed",
		}
	}

	
	var response *comands.ApplyResponse
	switch cmd.Type {
	case comands.CommandCreateUser:
		response = f.applyCreateUser(cmd.Data)
	case comands.CommandDeleteUser:
		response = f.applyDeleteUser(cmd.Data)
	case comands.CommandUpdateUser:
		response = f.applyUpdateUser(cmd.Data)
	default:
		response = &comands.ApplyResponse{
			Success: false,
			Error:   fmt.Sprintf("unknown command type: %s", cmd.Type),
		}
	}

	// Marca como processado
	if cmd.RequestID != "" && response.Success {
		f.processedRequests[cmd.RequestID] = true
	}

	return response
}


// cria um snapshot do estado atual
func (f *GameFSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Obtém todos os players do banco
	players, err := f.repository.GetAll()
	if err != nil {
		return nil, err
	}

	return &GameSnapshot{
		players:           players,
		processedRequests: f.cloneProcessedRequests(),
	}, nil
}

// restaura o estado a partir de um snapshot
func (f *GameFSM) Restore(snapshot io.ReadCloser) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	defer snapshot.Close()

	var data struct {
		Players           []*entities.Player `json:"players"`
		ProcessedRequests map[string]bool    `json:"processed_requests"`
	}

	if err := json.NewDecoder(snapshot).Decode(&data); err != nil {
		return err
	}

	// Limpa banco atual
	if err := f.repository.DeleteAll(); err != nil {
		return fmt.Errorf("failed to clear database: %v", err)
	}

	// Restaura players
	for _, player := range data.Players {
		if _, err := f.repository.CreateWithID(&entities.Player{Username: player.Username, ID: player.ID}); err != nil {
			log.Printf("⚠️ Erro ao restaurar player %s: %v", player.Username, err)
		}
	}

	f.processedRequests = data.ProcessedRequests

	log.Printf("[FSM] Snapshot restaurado: %d players", len(data.Players))
	return nil
}


func (f *GameFSM) cloneProcessedRequests() map[string]bool {
	clone := make(map[string]bool, len(f.processedRequests))
	for k, v := range f.processedRequests {
		clone[k] = v
	}
	return clone
}

