package authService

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"crypto/ecdsa"

	gameEntities "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	raftService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/session"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"context"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/google/uuid"
)

type AuthService struct {
	repo         *repository.PlayerRepository
	apiClient    *client.Client
	knownServers map[string]*entities.ServerInfo
	raft         *raftService.RaftService 
	sessionManager *session.SessionManager
	blockchainClient *blockchain.BlockchainClient
}

func New(repo *repository.PlayerRepository,apiClient *client.Client,knownServers map[string]*entities.ServerInfo,raft *raftService.RaftService, sm *session.SessionManager,blockchainClient *blockchain.BlockchainClient ) *AuthService {
	return &AuthService{
		repo:         repo,
		apiClient:    apiClient,
		knownServers: knownServers,
		raft:         raft,
		sessionManager: sm,
		blockchainClient: blockchainClient,
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
		return as.forwardToLeader(username)
		
	}

	return as.createAccountAsLeader(username)
}


func (as *AuthService) createAccountAsLeader(username string) error {
	log.Printf("[AuthService] Sou líder! Processando comando via Raft...")

	userID := uuid.New().String()
	hexKey, address, err := as.GeneratePrivateKey()
	if err != nil {
		return fmt.Errorf("erro ao gerar chave privada: %v", err)
	}

	// Criar usuário via Raft
	cmdData := comands.CreateUserCommand{
		UserID:        userID,
		Username:      username,
		PrivateKey:    hexKey,
		AddressAcount: address,
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

	response, err := as.raft.ApplyCommand(cmd)
	if err != nil {
		log.Printf("[AuthService] Erro ao aplicar comando no Raft: %v", err)
		return fmt.Errorf("erro ao processar comando: %v", err)
	}

	if !response.Success {
		log.Printf("[AuthService] Comando rejeitado: %s", response.Error)
		return fmt.Errorf("falha ao criar usuário: %s", response.Error)
	}

	log.Printf("[AuthService] Usuário '%s' criado no Raft!", username)

	// financia a conta do jogador 
	err = as.finaceAccount(address)
	if err != nil {
		return  err
	}	

	log.Printf("[AuthService] Usuário '%s' criado e replicado no cluster via Raft!", username)
	return nil
}

//Envia 1 ETH para a conta do  jogador (para que ele possa assinar suas transações )
func (as *AuthService) finaceAccount( address string ) error{
	if as.blockchainClient != nil {
		
			ctx := context.Background()
			
			
			amount := blockchain.EthToWei(1.0)
			
			log.Printf(" [AuthService] Financiando conta %s com 1 ETH...", address)
			
			err := as.blockchainClient.FundAccount(ctx, address, amount)
			if err != nil {
				log.Printf("[AuthService] Erro ao financiar conta %s: %v", address, err)
				return err
			}
			
		
			balance, err := as.blockchainClient.GetBalance(ctx, address)
			if err != nil {
				log.Printf("[AuthService] Erro ao verificar saldo: %v", err)
				return err
			}
			
			log.Printf(" [AuthService] Conta %s financiada! Saldo: %f ETH", 
				address, blockchain.WeiToEth(balance))
		
	}
 return  nil
}


// redireciona criação de conta para o lider chamando rota da api rest 
func (as *AuthService) forwardToLeader(username string)error{
	leaderAddr := as.raft.GetLeaderHTTPAddr()
	
	if leaderAddr == "" {
		return errors.New("nenhum líder disponível no momento, tente novamente")
	}

	if err :=  as.apiClient.AuthInterface.AskForCreatePlayerAccount(leaderAddr, username); err != nil{
		return fmt.Errorf("erro ao contatar líder: %v", err)
	}

	log.Printf("Conta criada via líder: %s", username)

	return  nil
}



// Login autentica um usuário, verifica se ele já está logado em outro servidor
// no cluster, e cria uma sessão em memória.
func (as *AuthService) Login(username string, clientID string) (*gameEntities.Player, error) {
	log.Printf("[AuthService] Tentativa de login para: %s (ClientID: %s)", username, clientID)

	// Verifica se este CLIENTE já tem uma sessão ativa.
	if as.sessionManager.IsClientLoggedIn(clientID) {
		log.Printf("[AuthService] Login negado: ClientID %s já tem uma sessão ativa.", clientID)
		return nil, errors.New("este cliente já está logado (faça logout primeiro)")
	}

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
			log.Printf("Erro ao verificar sessão no servidor %s: %v", serverID, err)
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

	// A função RemoveSession agora diz-nos se o logout foi efetivo.
	wasRemoved := as.sessionManager.RemoveSession(clientID)

	if !wasRemoved {
		log.Printf("[AuthService] Tentativa de logout para o clientID %s, mas não havia sessão ativa.", clientID)
		return errors.New("nenhuma sessão ativa para fazer logout")
	}

	log.Printf("[AuthService] Sessão removida com sucesso para o clientID: %s", clientID)
	return nil
}


// ----------------- AUXILIARES --------------------
func (as *AuthService) GeneratePrivateKey() (string, string, error) {
    // Gera chave privada
    privateKey, err := crypto.GenerateKey()
    if err != nil {
        return "", "", fmt.Errorf("erro ao gerar chave: %v", err)
    }

    // Converte para hex (formato 64 chars)
    privateKeyBytes := crypto.FromECDSA(privateKey)            // 32 bytes
    privateKeyHex := hex.EncodeToString(privateKeyBytes)       // 64 chars hex

    // Gera endereço público
    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        return "", "", fmt.Errorf("erro ao converter public key")
    }

    address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

    return privateKeyHex, address, nil
}




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