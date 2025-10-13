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
		return &comands.ApplyResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid create user command: %v", err),
		}
	}

	// Cria usuário com ID específico 
	player, err := f.repository.CreateWithID(&entities.Player{Username: cmd.Username, ID: cmd.UserID})
	if err != nil {
	
		if f.repository.UsernameExists(cmd.Username) {
			log.Printf("ℹUsuário '%s' já existe, ignorando", cmd.Username)
			return &comands.ApplyResponse{
				Success: true,
				Data:    "user already exists",
			}
		}
		return &comands.ApplyResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to create user: %v", err),
		}
	}

	log.Printf("[FSM] Usuário '%s' criado com ID: %s", cmd.Username, cmd.UserID)
	return &comands.ApplyResponse{
		Success: true,
		Data:    player,
	}
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
