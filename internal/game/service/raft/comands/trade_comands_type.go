package comands

// Comando para troca (replicado via Raft)
const (
	CommandTradeCards CommandType = "TRADE_CARDS"
)

// Representa dados para uma troca atômica
type TradeCardsCommand struct {
	// Jogador A
	PlayerAID string `json:"player_a_id"`
	CardAID   string `json:"card_a_id"` // Carta que A vai dar

	// Jogador B
	PlayerBID string `json:"player_b_id"`
	CardBID   string `json:"card_b_id"` // Carta que B vai dar

	// RequestID para idempotência
	RequestID string `json:"request_id"`
}