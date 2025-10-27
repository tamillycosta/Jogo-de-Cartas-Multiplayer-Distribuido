package packages


import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	
)

type PackageClientInterface struct {
	httpClient http.Client
}

func New(client http.Client) *PackageClientInterface{
	return &PackageClientInterface{
		httpClient: client,
	}
}


// Pede ao servidor lider para abrir pacote de player .
// // POST /api/v1/auth/open-package
func (p *PackageClientInterface) AskForOpenPackge(leaderAddr, player_id string) error {
	url := fmt.Sprintf("%s/api/v1/package/open-package", leaderAddr)
	

	
	payload := map[string]string{
		"player_id": player_id,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := p.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
	
		return fmt.Errorf("failed to call leader: %w", err)
	}
	defer resp.Body.Close()


	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return fmt.Errorf("server returned status %d: %v", resp.StatusCode, errorResp)
	}


	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	
	if success, ok := response["success"].(bool); ok && !success {
		if errorMsg, ok := response["error"].(string); ok {
			return fmt.Errorf("leader error: %s", errorMsg)
		}
		return fmt.Errorf("leader returned success=false")
	}


	return nil
}