package client

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client/auth"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client/packages"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client/raft"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client/trade"
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
	timeout    time.Duration
	AuthInterface *auth.AuthClientInterface
	RaftInterface *raft.RaftClientInterface
	PackageInterface *packages.PackageClientInterface
	TradeInterface   *trade.TradeClientInterface
}


func New(timeout time.Duration) *Client {
    client := Client{
        HttpClient: &http.Client{Timeout: timeout},
        timeout:    timeout,
		
    }
	client.AuthInterface = auth.New(*client.HttpClient)
	client.RaftInterface = raft.New(*client.HttpClient)
	client.PackageInterface = packages.New(*client.HttpClient)
	client.TradeInterface = trade.New(*client.HttpClient)
	return &client
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



