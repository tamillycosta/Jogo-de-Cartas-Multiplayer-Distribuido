package client

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	
	"time"
)

// responsável por realizar requisições HTTP entre os servidores do jogo.
// faz comunicação **server-to-server**
type Client struct {
	httpClient *http.Client
	timeout    time.Duration
	
}


func New(timeout time.Duration) *Client {
    return &Client{
        httpClient: &http.Client{Timeout: timeout},
        timeout:    timeout,
    }
}


//envia uma requisição GET para o endpoint `/api/v1/info`
// de outro servidor e retorna suas informaçõese
func (c *Client) AskServerInfo(serverAddress string, port int) (*entities.ServerInfo, error){
	url := fmt.Sprintf("http://%s:%d/api/v1/info", serverAddress, port)

	resp , err := c.httpClient.Get(url)
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


// envia uma notificação para outro servidor,
// através do endpoint HTTP POST `/api/v1/notify`.
func (c *Client) SendNotification(serverAddress string, port int, notification *entities.NotificationMessage) error{
	url	:= fmt.Sprintf("http://%s:%d/api/v1/notify",serverAddress, port)

	jsonData, err := json.Marshal(notification)
    if err != nil {
        return fmt.Errorf("failed to marshal notification: %w", err)
    }

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("failed to send notification: %w", err)
    }
    defer resp.Body.Close()

    return nil
}


// implementar 
// verificar se o username ja existe em outro servidor
// faz broadcast para servres conhecidos 
func (c *Client) CheckUsernameGlobal()bool{
	return  true
}