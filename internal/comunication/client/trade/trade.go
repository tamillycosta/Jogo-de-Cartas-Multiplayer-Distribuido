package trade

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type TradeClientInterface struct {
	httpClient http.Client
}

func New(client http.Client) *TradeClientInterface {
	return &TradeClientInterface{
		httpClient: client,
	}
}

// Pede ao servidor lider para executar a troca.
// POST /api/v1/trade/execute
func (t *TradeClientInterface) AskForTrade(leaderAddr string, cmd comands.TradeCardsCommand) error {
	url := fmt.Sprintf("%s/api/v1/trade/execute", leaderAddr)

	jsonData, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := t.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to propagate trade: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return fmt.Errorf("server returned status %d: %v", resp.StatusCode, errorResp)
	}

	return nil
}