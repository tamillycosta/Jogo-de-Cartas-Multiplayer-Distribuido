package raft

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"bytes"
)
// impelemnta a interface de cliente para reuqisções e gereciamento interno do raft (metodos n rpc)
type RaftClientInterface struct{

	httpClient http.Client
}

func New(client http.Client) *RaftClientInterface{
	return &RaftClientInterface{
		httpClient: client,
	}
}

// Metodo chamado pelo node para desobrir o lider do cluster
// GET api/v1/raft/status
// retorna o response body completo
func (c *RaftClientInterface) GetLeader(serverAddress string, port int) (map[string]interface{}, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/raft/status", serverAddress, port)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("falha ao enviar requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro HTTP %d: %s", resp.StatusCode, string(body))
	}

	var bodyMap map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&bodyMap); err != nil {
		return nil, fmt.Errorf("falha ao decodificar resposta: %w", err)
	}

	return bodyMap, nil
}


// POST /api/v1/raft/join
func (c *RaftClientInterface) RequestJoin(server *entities.ServerInfo, body map[string]string) (*http.Response, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/raft/join", server.Address, server.Port)

	jsonBody, _ := json.Marshal(body)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	return resp, nil
}