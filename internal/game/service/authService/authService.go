package authService

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"errors"

	
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
)


type AuthService struct{
	repo *repository.PlayerRepository
	apiClient *client.Client
	knownServers map[string]*entities.ServerInfo
}

func New(repo *repository.PlayerRepository, apiClient *client.Client, KnownServers map[string]*entities.ServerInfo) *AuthService{
	return &AuthService{
		repo: repo,
		apiClient: apiClient,
		knownServers: KnownServers,
	}
}


// função que vai ser chamada no pub/sub para cirar conta
func (as *AuthService) CreateAccount(username string) error {
    if len(username) == 0 {
        return errors.New("username não pode ser vazio")
    }
    
    // 2. Verifica localmente
    if as.UserExists(username){
        return errors.New("este username ja existe")
    }
    
    // 3. Verifica globalmente ()
    if as.checkUsernameGlobal(username) {
        return errors.New("este username ja existe em outro servidor")
    }
    
    // 5. Cria no BD
    _, err := as.repo.Create(username)
    return err
}

// faz broadcast na api rest para verificar se o username esta disponivel 
func (as *AuthService) checkUsernameGlobal(username string) bool {
    for _, server := range as.knownServers {
        isAvailable, err := as.apiClient.AuthInterface.CheckUsernameGlobal(
            server.Address, 
            server.Port, 
            username,
        )
        
        if err != nil {
            continue // Ignora servers offline (depois inserir tratamento de falha dos outros servidores)
        }
        
        if !isAvailable {
            return true // Username já existe
        }
    }
    return false // Username disponível
}

func (ah *AuthService) UserExists(username string) bool{
	return ah.repo.UsernameExists(username)
}