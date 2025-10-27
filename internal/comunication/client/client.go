package client

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client/auth"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client/match"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client/packages"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client/raft"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// responsável por realizar requisições HTTP entre os servidores do jogo.
// faz comunicação **server-to-server**
type Client struct {
	HttpClient *http.Client
	AuthInterface  *auth.AuthClientInterface
	MatchInterface *match.MatchClientInterface
	RaftInterface  *raft.RaftClientInterface
	PackageInterface *packages.PackageClientInterface
}

func New() *Client {
	// ✅ CORRIGIDO: Timeout aumentado para operações de jogo
	httpClient := http.Client{
		Timeout: 30 * time.Second, // Era muito curto antes!
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
			// ✅ Timeouts adicionais para evitar travamentos
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	client := Client{
		HttpClient: &httpClient,
	}
	client.AuthInterface = auth.New(*client.HttpClient)
	client.RaftInterface = raft.New(*client.HttpClient)
	client.PackageInterface = packages.New(*client.HttpClient)
	client.MatchInterface = match.New(*client.HttpClient)
	return  &client
}



//envia uma requisição GET para o endpoint `/api/v1/info`
// de outro servidor e retorna suas informaçõese
func (c *Client) AskServerInfo(serverAddress string, port int) (*entities.ServerInfo, error){
	url := fmt.Sprintf("http://%s:%d/api/v1/info", serverAddress, port)

	resp , err := c.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get health: %w", err)
	}
	defer resp.Body.Close()

	var info entities.ServerInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
        return nil, fmt.Errorf("failed to decode health: %w", err)
    }
	return &info, nil
}



