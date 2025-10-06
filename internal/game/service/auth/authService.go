package auth

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"errors"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
)


type Auth struct{
	repo *repository.PlayerRepository
	apiClient *client.Client
	knownServers map[string]*entities.ServerInfo
}

func New(repo *repository.PlayerRepository, apiClient *client.Client, KnownServers map[string]*entities.ServerInfo) *Auth{
	return &Auth{
		repo: repo,
		apiClient: apiClient,
		knownServers: KnownServers,
	}
}


// finzalizar implementação
func (a *Auth) CreateAccount(username string, ) error{
	if(len(username) == 0){
		return errors.New("username não pode ser vazio")
	}
	// verfica localmente
	if(a.repo.UsernameExists(username)){
		return errors.New("este username ja existe")
	}//verifica globalmente
	if(a.apiClient.CheckUsernameGlobal()){
		return  errors.New("este username ja existe")
	}
	return  nil
}