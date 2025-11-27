package comands

// Comando para troca (replicado via Raft)
const (
	CommandTradeCards CommandType = "TRADE_CARDS"
)

// Representa dados para uma troca atômica
type TradeCardsCommand struct {
    FromPlayerID string `json:"from_player_id"`
    ToPlayerID   string `json:"to_player_id"`
    CardID       string `json:"card_id"`        // Carta que eu dou
    WantedCardID string `json:"wanted_card_id"` // Carta que eu recebo (NOVO CAMPO)
    RequestID    string `json:"request_id"`
}