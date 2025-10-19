package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type AuthClientInterface struct {
	httpClient http.Client
}

func New(client http.Client) *AuthClientInterface{
	return & AuthClientInterface{
		httpClient: client,
	}
}

// GET /api/v1/user-exists?username=xyz
// Verifica se o username existe em outro servidor
// Retorna true se existe, false se disponível
func (a *AuthClientInterface) CheckUsernameExists(serverAddress string, port int, username string) (bool, error) {
	escapedUsername := url.QueryEscape(username)
	url := fmt.Sprintf("http://%s:%d/api/v1/user-exists?username=%s", serverAddress, port, escapedUsername)

	resp, err := a.httpClient.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to request username check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var result struct {
		Exists   bool   `json:"exists"`
		Username string `json:"username"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Exists, nil
}


// CheckPlayerLoggedIn verifica se um jogador está logado em outro servidor.
// GET /api/v1/auth/is-player-logged-in?username=xyz
func (a *AuthClientInterface) CheckPlayerLoggedIn(serverAddress string, port int, username string) (bool, error) {
	escapedUsername := url.QueryEscape(username)
	url := fmt.Sprintf("http://%s:%d/api/v1/auth/is-player-logged-in?username=%s", serverAddress, port, escapedUsername)

	resp, err := a.httpClient.Get(url)
	if err != nil {
		return false, fmt.Errorf("falha ao requisitar a verificação de login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("servidor retornou o status: %d", resp.StatusCode)
	}

	var result struct {
		IsLoggedIn bool `json:"is_logged_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("falha ao descodificar a resposta: %w", err)
	}

	return result.IsLoggedIn, nil
}

// Pede ao servidor lider para criar a conta de um player .
// // POST /api/v1/auth/create-account
func (a *AuthClientInterface) AskForCreatePlayerAccount(leaderAddr, username string) error{
	// Faz requisição HTTP para o líder
	url := fmt.Sprintf("%s/api/v1/auth/create-account", leaderAddr)
	
	payload := map[string]string{
		"username": username,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := a.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to propagate user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return fmt.Errorf("server returned status %d: %v", resp.StatusCode, errorResp)
	}

	return nil
}
