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

// POST /api/v1/propagate-user
// Propaga criação de usuário para outro servidor
func (a *AuthClientInterface) PropagateUser(serverAddress string, port int, userID, username string) error {
	url := fmt.Sprintf("http://%s:%d/api/v1/propagate-user", serverAddress, port)

	payload := map[string]string{
		"user_id":  userID,
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