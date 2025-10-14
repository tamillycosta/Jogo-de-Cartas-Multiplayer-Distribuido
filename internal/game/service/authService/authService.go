package authService

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"

	raftService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	gameEntities "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"

	"github.com/google/uuid"
)

type AuthService struct {
	repo         *repository.PlayerRepository
	apiClient    *client.Client
	knownServers map[string]*entities.ServerInfo
	raft         *raftService.RaftService // ← RAFT ADICIONADO!
}

func New(repo *repository.PlayerRepository,apiClient *client.Client,knownServers map[string]*entities.ServerInfo,raft *raftService.RaftService) *AuthService {
	return &AuthService{
		repo:         repo,
		apiClient:    apiClient,
		knownServers: knownServers,
		raft:         raft,
	}
}


// ----------------  MÉTODO CHAMADO PELO PUB/SUB ----------------- 


// Chamado quando cliente WebSocket solicita criar conta
func (as *AuthService) CreateAccount(username string) error {
	log.Printf("[AuthService] Recebida solicitação de criação: %s", username)

	if len(username) == 0 {
		return errors.New("username não pode ser vazio")
	}


	// Verifica se já existe localmente 
	if as.UserExists(username) {
		log.Printf("[AuthService] Username '%s' já existe localmente", username)
		return errors.New("este username já existe")
	}

	// Verifica se este servidor é o lider , se não for busca o endereço do lider do cluster
	if !as.raft.IsLeader() {
		leaderAddr := as.raft.GetLeaderHTTPAddr()
		log.Printf("[AuthService] Não sou líder! Líder atual: %s", leaderAddr)
		
		if leaderAddr == "" {
			return errors.New("nenhum líder disponível no momento, tente novamente")
		}
		
		return fmt.Errorf("não sou o líder. Conecte-se ao líder: %s", leaderAddr)
	}

	
	log.Printf("[AuthService] Sou líder! Processando comando via Raft...")

	userID := uuid.New().String()

	cmdData := comands.CreateUserCommand{
		UserID:   userID,
		Username: username,
	}

	data, err := json.Marshal(cmdData)
	if err != nil {
		return fmt.Errorf("erro ao serializar comando: %v", err)
	}

	cmd := comands.Command{
		Type:      comands.CommandCreateUser,
		Data:      data,
		RequestID: uuid.New().String(),
	}

	
	
	// Aplica comando via Raft (será replicado para todos os servidores segidores)
	response, err := as.raft.ApplyCommand(cmd)
	if err != nil {
		log.Printf("[AuthService] Erro ao aplicar comando no Raft: %v", err)
		return fmt.Errorf("erro ao processar comando: %v", err)
	}

	if !response.Success {
		log.Printf("[AuthService] Comando rejeitado: %s", response.Error)
		return fmt.Errorf("falha ao criar usuário: %s", response.Error)
	}

	log.Printf("[AuthService] Usuário '%s' criado e replicado no cluster via Raft!", username)
	return nil
}

// 
func (as *AuthService) Login(username string) (*gameEntities.Player, error) {
    log.Printf("[AuthService] Tentativa de login para: %s", username)

    player, err := as.repo.FindByUsername(username)
    if err != nil {
        log.Printf("[AuthService] Erro ao buscar usuário '%s': %v", username, err)
        return nil, errors.New("erro interno ao tentar fazer login")
    }

    if player == nil {
        log.Printf("[AuthService] Usuário '%s' não encontrado", username)
        return nil, errors.New("usuário não encontrado")
    }

    log.Printf("[AuthService] Usuário '%s' autenticado com sucesso", username)
    return player, nil
}

// ----------------- AUXILIARES --------------------


// Verifica se username existe localmente (NÃO usa Raft, apenas leitura)
func (as *AuthService) UserExists(username string) bool {
	return as.repo.UsernameExists(username)
}

//  Retorna informações do líder atual
func (as *AuthService) GetLeaderInfo() map[string]interface{} {
	return map[string]interface{}{
		"is_leader":   as.raft.IsLeader(),
		"leader_id":   as.raft.GetLeaderID(),
		"leader_addr": as.raft.GetLeaderHTTPAddr(),
	}
}



// ----------------- MÉTODOS DA API REST (P2P) -------------------


// Chamado pela API quando outro servidor verifica username
func (as *AuthService) CheckUsernameLocal(username string) bool {
	return as.UserExists(username)
}