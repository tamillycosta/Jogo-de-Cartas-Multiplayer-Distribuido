package fms

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"encoding/json"
	"fmt"
	"log"
)

func (f *GameFSM) applyTradeCards(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.TradeCardsCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: "formato de comando invalido"}
	}

	log.Printf("[FSM] Processando transferencia: Card %s de %s para %s",
		cmd.CardID, cmd.FromPlayerID, cmd.ToPlayerID)

	// 1. Verificar se a carta existe e pertence ao remetente
	card, err := f.cardRepository.FindById(cmd.CardID)
	if err != nil {
		return &comands.ApplyResponse{Success: false, Error: fmt.Sprintf("Erro ao buscar carta: %v", err)}
	}
	if card == nil {
		return &comands.ApplyResponse{Success: false, Error: "Carta nao encontrada"}
	}
	
	if card.PlayerID == nil || *card.PlayerID != cmd.FromPlayerID {
		return &comands.ApplyResponse{Success: false, Error: "Voce nao possui esta carta"}
	}

	if card.InDeck {
		return &comands.ApplyResponse{Success: false, Error: "Nao pode trocar carta que esta no deck"}
	}

	// 2. Executar a transferência (Update Owner)
	// Usamos a função já existente no repositório
	if err := f.cardRepository.UpdateCardOwner(cmd.CardID, cmd.ToPlayerID); err != nil {
		log.Printf("[FSM] ERRO TRANSFERENCIA: %v", err)
		return &comands.ApplyResponse{Success: false, Error: "Falha ao transferir carta no banco de dados"}
	}

	log.Printf("[FSM] Transferencia concluida com sucesso.")
	return &comands.ApplyResponse{Success: true, Data: "transfer_complete"}
}