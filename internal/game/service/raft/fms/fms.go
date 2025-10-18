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
	playerRepository *repository.PlayerRepository
	packageRepository *repository.PackageRepository
	cardRepository *repository.CardRepository
	
	// Cache de IDs de requisições processadas (para idempotência)
	processedRequests map[string]bool
}

func New(
	playerRepo *repository.PlayerRepository, packageRepo *repository.PackageRepository,cardRepo *repository.CardRepository,) *GameFSM {
	return &GameFSM{
		playerRepository:        playerRepo,
		packageRepository:       packageRepo,
		cardRepository:          cardRepo,
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
		log.Printf("[FSM] Comando já processado: %s", cmd.RequestID)
		return &comands.ApplyResponse{
			Success: true,
			Data:    "already processed",
		}
	}

	// Processa comando
	var response *comands.ApplyResponse
	switch cmd.Type {
	// Players
	case comands.CommandCreateUser:
		response = f.applyCreateUser(cmd.Data)
	case comands.CommandDeleteUser:
		response = f.applyDeleteUser(cmd.Data)
	case comands.CommandUpdateUser:
		response = f.applyUpdateUser(cmd.Data)
	
	// Packages
	case comands.CommandCreatePackage:
		response = f.applyCreatePackage(cmd.Data)

	case comands.CommandLockPackage:
		response = f.applyLockPackage(cmd.Data)
	case comands.CommandOpenPackage:
		response = f.applyOpenPackage(cmd.Data)
	
	// Cards
	case comands.CommandCreateCard:
		response = f.applyCreateCard(cmd.Data)
	case comands.CommandTransferCard:
		response = f.applyTransferCard(cmd.Data)
	
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

	// Obtém todos os dados
	players, err := f.playerRepository.GetAll()
	if err != nil {
		return nil, err
	}

	packages, err := f.packageRepository.GetAll()
	if err != nil {
		return nil, err
	}

	cards, err := f.cardRepository.GetAll()
	if err != nil {
		return nil, err
	}

	return &GameSnapshot{
		players:           players,
		packages:          packages,
		cards:             cards,
		processedRequests: f.cloneProcessedRequests(),
	}, nil
}

// restaura o estado a partir de um snapshot
func (f *GameFSM) Restore(snapshot io.ReadCloser) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	defer snapshot.Close()

	var data struct {
		Players           []*entities.Player  `json:"players"`
		Packages          []*entities.Package `json:"packages"`
		Cards             []*entities.Card    `json:"cards"`
		ProcessedRequests map[string]bool     `json:"processed_requests"`
	}

	if err := json.NewDecoder(snapshot).Decode(&data); err != nil {
		return err
	}

	// Limpa bancos
	if err := f.playerRepository.DeleteAll(); err != nil {
		return err
	}
	if err := f.packageRepository.DeleteAll(); err != nil {
		return err
	}
	if err := f.cardRepository.DeleteAll(); err != nil {
		return err
	}

	// Restaura players
	for _, player := range data.Players {
		if _, err := f.playerRepository.CreateWithID(player); err != nil {
			log.Printf("Erro ao restaurar player: %v", err)
		}
	}

	// Restaura packages
	for _, pkg := range data.Packages {
		if _, err := f.packageRepository.CreateWithID(pkg); err != nil {
			log.Printf("Erro ao restaurar package: %v", err)
		}
	}

	// Restaura cards
	for _, card := range data.Cards {
		if _, err := f.cardRepository.CreateWithID(card); err != nil {
			log.Printf("Erro ao restaurar card: %v", err)
		}
	}

	f.processedRequests = data.ProcessedRequests

	log.Printf("[FSM] Snapshot restaurado: %d players, %d packages, %d cards",
		len(data.Players), len(data.Packages), len(data.Cards))
	return nil
}

func (f *GameFSM) cloneProcessedRequests() map[string]bool {
	clone := make(map[string]bool, len(f.processedRequests))
	for k, v := range f.processedRequests {
		clone[k] = v
	}
	return clone
}