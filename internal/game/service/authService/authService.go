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
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/session"

	"github.com/google/uuid"
)

type AuthService struct {
	repo         *repository.PlayerRepository
	apiClient    *client.Client
	knownServers map[string]*entities.ServerInfo
	raft         *raftService.RaftService // ← RAFT ADICIONADO!
	sessionManager *session.SessionManager
}

func New(repo *repository.PlayerRepository,apiClient *client.Client,knownServers map[string]*entities.ServerInfo,raft *raftService.RaftService, sm *session.SessionManager) *AuthService {
	return &AuthService{
		repo:         repo,
		apiClient:    apiClient,
		knownServers: knownServers,
		raft:         raft,
		sessionManager: sm,
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

// Login autentica um usuário, verifica se ele já está logado em outro servidor
// no cluster, e cria uma sessão em memória.
func (as *AuthService) Login(username string, clientID string) (*gameEntities.Player, error) {
	log.Printf("[AuthService] Tentativa de login para: %s (ClientID: %s)", username, clientID)

	// PASSO 1: Verificar se o jogador já está logado em outro servidor do cluster
	for serverID, serverInfo := range as.knownServers {
		// Não precisa de verificar no próprio servidor, pois isso será feito localmente a seguir
		if serverID == as.raft.GetMyID() {
			continue
		}

		log.Printf("[AuthService] Verificando sessão de '%s' no servidor %s...", username, serverID)
		isLoggedIn, err := as.apiClient.AuthInterface.CheckPlayerLoggedIn(serverInfo.Address, serverInfo.Port, username)
		if err != nil {
			// Se houver um erro, pode ser que o servidor esteja temporariamente offline.
			// É mais seguro continuar a verificação nos outros.
			log.Printf("⚠️  Erro ao verificar sessão no servidor %s: %v", serverID, err)
			continue
		}

		if isLoggedIn {
			log.Printf("[AuthService] Login negado: '%s' já tem uma sessão ativa no servidor %s.", username, serverID)
			return nil, fmt.Errorf("usuário já está logado no servidor %s", serverID)
		}
	}

	// PASSO 2: Proceder com a lógica de login local (como estava antes)
	player, err := as.repo.FindByUsername(username)
	if err != nil {
		log.Printf("[AuthService] Erro ao buscar usuário '%s': %v", username, err)
		return nil, errors.New("erro interno ao tentar fazer login")
	}
	if player == nil {
		log.Printf("[AuthService] Usuário '%s' não encontrado", username)
		return nil, errors.New("usuário não encontrado")
	}
	if as.sessionManager.IsPlayerLoggedIn(player.ID) {
		log.Printf("[AuthService] Usuário '%s' (ID: %s) já está logado.", username, player.ID)
		return nil, errors.New("usuário já está logado")
	}

	as.sessionManager.CreateSession(clientID, player)
	log.Printf("[AuthService] Sessão criada para '%s'. ClientID: %s -> PlayerID: %s", username, clientID, player.ID)
	return player, nil
}

// Logout remove a sessão de um jogador com base no clientID da sua conexão
func (as *AuthService) Logout(clientID string) error {
	log.Printf("[AuthService] Recebido pedido de logout para o clientID: %s", clientID)
	
	// A lógica de remoção já está no SessionManager, apenas a chamamos.
	as.sessionManager.RemoveSession(clientID)
	
	log.Printf("[AuthService] Sessão removida com sucesso para o clientID: %s", clientID)
	return nil
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