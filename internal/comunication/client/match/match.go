package match

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type MatchClientInterface struct {
	httpClient http.Client
}

func New(client http.Client) *MatchClientInterface {

	if client.Timeout == 0 {
		client.Timeout = 30 * time.Second
	}
	
	return &MatchClientInterface{
		httpClient: client,
	}
}

//Envia requisição ao líder para adicionar jogador à fila global
func (m *MatchClientInterface) JoinGlobalQueue(leaderAddr, playerID, username, serverID, clientID string) error {
	url := fmt.Sprintf("%s/api/v1/match/global/join", leaderAddr)

	payload := map[string]string{
		"player_id": playerID,
		"username":  username,
		"server_id": serverID,
		"client_id": clientID,
	}

	body, _ := json.Marshal(payload)

	resp, err := m.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("erro ao chamar líder (%s): %w", leaderAddr, err)
	}
	defer resp.Body.Close()

	return handleResponse(resp, "JoinGlobalQueue")
}

//notifica outro servidor que uma partida remota foi criada
func (m *MatchClientInterface) NotifyRemoteMatchCreated(hostServerAddr string, notification map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/v1/match/created", hostServerAddr)

	body, _ := json.Marshal(notification)

	resp, err := m.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("erro ao notificar criação de partida remota: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("falha na notificação: status %d", resp.StatusCode)
	}

	return nil
}

// envia sincornização da parida para servidor não host 
func (m *MatchClientInterface) SendMatchSync(remoteServerAddr string, update interface{}) error {
	url := fmt.Sprintf("%s/api/v1/match/sync", remoteServerAddr)

	body, _ := json.Marshal(update)

	resp, err := m.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("erro ao enviar sync: %w", err)
	}
	defer resp.Body.Close()

	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("falha ao sincronizar estado: status %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

//envia ação de um jogador remoto para o host
func (m *MatchClientInterface) SendMatchAction(hostServerAddr, matchID, playerID string, action entities.GameAction) error {
	url := fmt.Sprintf("%s/api/v1/match/action", hostServerAddr)

	payload := map[string]interface{}{
		"match_id":  matchID,
		"player_id": playerID,
		"action": map[string]interface{}{
			"type":             action.Type,
			"card_id":          action.CardID,
			"attacker_card_id": action.AttackerCardID,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	log.Printf(" [MatchClient] Enviando ação para %s", url)
	

	resp, err := m.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("erro ao enviar ação: %w", err)
	}
	defer resp.Body.Close()


	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("[MatchClient] Erro %d: %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("falha ao processar ação: status %d - %s", resp.StatusCode, string(respBody))
	}

	log.Printf("[MatchClient] Ação enviada com sucesso (status %d)", resp.StatusCode)
	return nil
}

// envia heartbeat periódico para o servidor remoto (locutor)
func (m *MatchClientInterface) SendHeartbeat(remoteServerAddr, matchID string) error {
	url := fmt.Sprintf("%s/api/v1/match/heartbeat", remoteServerAddr)

	payload := map[string]string{"match_id": matchID}
	body, _ := json.Marshal(payload)

	resp, err := m.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("erro ao enviar heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("falha no heartbeat: status %d", resp.StatusCode)
	}

	return nil
}

func handleResponse(resp *http.Response, operation string) error {
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("[%s] requisição falhou (%d): %s", operation, resp.StatusCode, string(data))
	}
	return nil
}