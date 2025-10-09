package auth

import ("net/http"
	"fmt"
	"encoding/json"
	"net/url"
)


type AuthInterface struct{
	httpClient	http.Client 
}

func New(client http.Client) *AuthInterface{
	return &AuthInterface{
		httpClient: client,
	}
}

// HTTP GET api/v1/unsername-avaliblre
// verifica se o username ja existe em outro servidor
func (a *AuthInterface) CheckUsernameGlobal(serverAddress string, port int, username string) (bool, error) {
	escapedUsername := url.QueryEscape(username)
	url := fmt.Sprintf("http://%s:%d/api/v1/username-available?username=%s", serverAddress, port, escapedUsername)

	resp, err := a.httpClient.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to request username check: %w", err)
	}
	defer resp.Body.Close()

	// Se o servidor responder != 200, trata erro
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var isAvailable bool
	if err := json.NewDecoder(resp.Body).Decode(&isAvailable); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return isAvailable, nil
}