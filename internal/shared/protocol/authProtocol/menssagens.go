package authprotocol

// Representa protocolo de comunicação para autenticação na aplicação (pub/sub) _> cliente -> servidor

// Request de criar conta (publicação do cliente)
type CreateAccountRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Nome     string `json:"nome"`
}

// Request de login
type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

//  Resposta de autenticação (reply)
type AuthResponse struct {
    Type    string      `json:"type"`
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Error   string      `json:"error,omitempty"`
    Player  interface{} `json:"player,omitempty"`
}