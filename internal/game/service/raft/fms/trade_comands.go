package fms

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"encoding/json"
	"log"
)

func (f *GameFSM) applyTradeCards(data json.RawMessage) *comands.ApplyResponse {
    var cmd comands.TradeCardsCommand
    if err := json.Unmarshal(data, &cmd); err != nil {
        return &comands.ApplyResponse{Success: false, Error: "erro decode"}
    }

    // 1. Transferir Carta A -> Jogador B
    if err := f.cardRepository.UpdateCardOwner(cmd.CardID, cmd.ToPlayerID); err != nil {
        return &comands.ApplyResponse{Success: false, Error: "Erro ao transferir carta A"}
    }

    // 2. Transferir Carta B -> Jogador A (A troca real!)
    if err := f.cardRepository.UpdateCardOwner(cmd.WantedCardID, cmd.FromPlayerID); err != nil {
        // Rollback manual seria ideal aqui, mas para simplicidade:
        log.Printf("ERRO CRÍTICO: Falha ao transferir segunda carta na troca")
        return &comands.ApplyResponse{Success: false, Error: "Erro ao transferir carta B"}
    }

    log.Printf("[FSM] Troca concluída: %s <-> %s", cmd.CardID, cmd.WantedCardID)
    return &comands.ApplyResponse{Success: true, Data: "swap_complete"}
}