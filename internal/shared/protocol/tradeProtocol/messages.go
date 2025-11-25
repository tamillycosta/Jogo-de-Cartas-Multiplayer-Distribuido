package tradeprotocol

// Representa protocolo de comunicação para troca (pub/sub)

// Request de troca (publicação do cliente)
type TradeRequest struct {
	CardID         string `json:"card_id"`          // Carta que eu (cliente) quero dar
	TargetUsername string `json:"target_username"`  // Nome do jogador que vai receber
}

// Resposta da troca
type TradeResponse struct {
	Type    string      `json:"type"`
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}