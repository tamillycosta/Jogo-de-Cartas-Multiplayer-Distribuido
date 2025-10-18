package packageprotocol


// Representa protocolo de comunicação para abertura de pacote na aplicação (pub/sub) _> cliente -> servidor

// Request de criar conta (publicação do cliente)
type OpenPackageRequest struct {
    PlayerID string  `json:"player_id"`
	
}


type OpenPackageResponse struct {
    Type    string      `json:"type"`
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Error   string      `json:"error,omitempty"`
    Player  interface{} `json:"player,omitempty"`
}