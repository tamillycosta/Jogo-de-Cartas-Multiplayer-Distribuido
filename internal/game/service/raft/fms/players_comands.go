package fms

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"

	"encoding/json"

	"fmt"
	"log"
)

// comandos da GameFSM para cadastro de um usuario
func (f *GameFSM) applyCreateUser(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.CreateUserCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	player := &entities.Player{
		ID:         cmd.UserID,
		Username:   cmd.Username,
		PrivateKey: cmd.PrivateKey, 
		Address: cmd.AddressAcount,
	}

	if _, err := f.playerRepository.CreateWithID(player); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	log.Printf("[FSM] Player criado: %s (ID: %s)", cmd.Username, cmd.UserID)
	return &comands.ApplyResponse{Success: true, Data: player}
}

// --------------------- FALTA Implementar lógica de delete e update
func (f *GameFSM) applyDeleteUser(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.DeleteUserCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid delete user command: %v", err),
		}
	}

	log.Printf("[FSM] Deletando usuário: %s", cmd.UserID)

	return &comands.ApplyResponse{
		Success: true,
		Data:    "user deleted",
	}
}

func (f *GameFSM) applyUpdateUser(data json.RawMessage) *comands.ApplyResponse {

	return &comands.ApplyResponse{
		Success: true,
		Data:    "user updated",
	}
}
