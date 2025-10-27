package tradeprotocol

// Representa protocolo de comunicação para troca (pub/sub)

// Request de troca (publicação do cliente)
type TradeRequest struct {
	CardAID   string `json:"card_a_id"`   // Carta que eu (cliente) quero dar
	PlayerBID string `json:"player_b_id"` // ID do jogador com quem quero trocar
	CardBID   string `json:"card_b_id"`   // Carta que eu quero receber
}

// Resposta da troca
type TradeResponse struct {
	Type    string      `json:"type"`
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}