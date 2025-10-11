package authService

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
    en "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"

	"errors"
	"fmt"
	"log"
)

type AuthService struct {
	repo         *repository.PlayerRepository
	apiClient    *client.Client
	knownServers map[string]*entities.ServerInfo
}

func New(repo *repository.PlayerRepository, apiClient *client.Client, knownServers map[string]*entities.ServerInfo) *AuthService {
	return &AuthService{
		repo:         repo,
		apiClient:    apiClient,
		knownServers: knownServers,
	}
}

// Função chamada no pub/sub para criar conta
func (as *AuthService) CreateAccount(username string) error {
	if len(username) == 0 {
		return errors.New("username não pode ser vazio")
	}

	//  Verifica localmente primeiro
	if as.UserExists(username) {
		return errors.New("este username já existe localmente")
	}

	//  Verifica globalmente em outros servidores
	if len(as.knownServers) > 0 {
		existsGlobally, serverID := as.checkUsernameGlobal(username)
		if existsGlobally {
			// Username existe em outro servidor, sincroniza localmente
			log.Printf("Username '%s' já existe no servidor %s, sincronizando localmente", username, serverID)
			_, err := as.repo.Create(username)
			if err != nil {
				return fmt.Errorf("erro ao sincronizar usuário: %w", err)
			}
			return errors.New("este username já existe em outro servidor")
		}
	}

	//  Username disponível, cria localmente
	player, err := as.repo.Create(username)
	if err != nil {
		return fmt.Errorf("erro ao criar usuário: %w", err)
	}

	log.Printf("Usuário '%s' criado localmente com ID: %s", username, player.ID)

	// Propaga para outros servidores
	as.propagateUserToServers(player)

	return nil
}


// Verifica se username existe globalmente
// Retorna (existe, serverID)
func (as *AuthService) checkUsernameGlobal(username string) (bool, string) {
	for serverID, server := range as.knownServers {
		exists, err := as.apiClient.AuthInterface.CheckUsernameExists(
			server.Address,
			server.Port,
			username,
		)

		if err != nil {
			log.Printf("⚠️ Erro ao verificar username no servidor %s: %v", serverID, err)
			continue // Ignora servers offline
		}

		if exists {
			return true, serverID
		}
	}
	return false, ""
}


// Propaga usuário criado para outros servidores
func (as *AuthService) propagateUserToServers(player *en.Player ) {
	successCount := 0
	failCount := 0

	for serverID, server := range as.knownServers {
		err := as.apiClient.AuthInterface.PropagateUser(
			server.Address,
			server.Port,
			player.ID,
			player.Username,
		)

		if err != nil {
			log.Printf("Falha ao propagar usuário '%s' para servidor %s: %v", player.Username, serverID, err)
			failCount++
		} else {
			log.Printf("Usuário '%s' propagado para servidor %s", player.Username, serverID)
			successCount++
		}
	}

	log.Printf("Propagação concluída: %d sucessos, %d falhas", successCount, failCount)
}


// Verifica se username existe localmente
func (as *AuthService) UserExists(username string) bool {
	return as.repo.UsernameExists(username)
}


// Recebe propagação de outro servidor (chamado pela API)
func (as *AuthService) ReceiveUserPropagation(userID, username string) error {
	// Verifica se já existe localmente
	if as.UserExists(username) {
		log.Printf("Usuário '%s' já existe localmente, ignorando propagação", username)
		return nil
	}

	// Cria usuário localmente (sincronização)
	_, err := as.repo.CreateWithID(userID, username)
	if err != nil {
		return fmt.Errorf("erro ao criar usuário propagado: %w", err)
	}

	log.Printf("Usuário '%s' (ID: %s) sincronizado via propagação", username, userID)
	return nil
}