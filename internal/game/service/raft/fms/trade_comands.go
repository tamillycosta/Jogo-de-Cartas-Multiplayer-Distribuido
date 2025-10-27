package fms

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"encoding/json"
	"fmt"
	"log"
)

// applyTradeCards executa a troca atômica de duas cartas entre dois jogadores
func (f *GameFSM) applyTradeCards(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.TradeCardsCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: "formato de comando invalido"}
	}

	log.Printf("[FSM] Processando troca: Card %s (Player A: %s) por Card %s (Player B: %s)",
		cmd.CardAID, cmd.PlayerAID, cmd.CardBID, cmd.PlayerBID)

	// --- 1. Verificar posse e validade ---
	cardA, err := f.cardRepository.FindById(cmd.CardAID)
	if err != nil {
		return &comands.ApplyResponse{Success: false, Error: fmt.Sprintf("Erro ao buscar carta A: %v", err)}
	}
	if cardA == nil || cardA.PlayerID == nil || *cardA.PlayerID != cmd.PlayerAID {
		return &comands.ApplyResponse{Success: false, Error: "Jogador A nao possui a carta A"}
	}
	// (Opcional) Verificar se a carta está em um deck, em partida, etc.
	if cardA.InDeck {
		return &comands.ApplyResponse{Success: false, Error: "Carta A esta em um deck"}
	}

	cardB, err := f.cardRepository.FindById(cmd.CardBID)
	if err != nil {
		return &comands.ApplyResponse{Success: false, Error: fmt.Sprintf("Erro ao buscar carta B: %v", err)}
	}
	if cardB == nil || cardB.PlayerID == nil || *cardB.PlayerID != cmd.PlayerBID {
		return &comands.ApplyResponse{Success: false, Error: "Jogador B nao possui a carta B"}
	}
	if cardB.InDeck {
		return &comands.ApplyResponse{Success: false, Error: "Carta B esta em um deck"}
	}

	// --- 2. Executar a troca (atômica) ---
	
	// Passo A: Dar Carta A para Jogador B
	if err := f.cardRepository.UpdateCardOwner(cmd.CardAID, cmd.PlayerBID); err != nil {
		log.Printf("[FSM] ERRO TROCA (Passo A): %v", err)
		return &comands.ApplyResponse{Success: false, Error: "Falha ao transferir carta A para B"}
	}

	// Passo B: Dar Carta B para Jogador A
	if err := f.cardRepository.UpdateCardOwner(cmd.CardBID, cmd.PlayerAID); err != nil {
		log.Printf("[FSM] ERRO TROCA (Passo B): %v. REVERTENDO Passo A...", err)
		// --- ROLLBACK ---
		// Se o passo B falhar, DESFAZEMOS o passo A para garantir a atomicidade.
		if errRollback := f.cardRepository.UpdateCardOwner(cmd.CardAID, cmd.PlayerAID); errRollback != nil {
			log.Printf("[FSM] ERRO CRITICO DE ROLLBACK: %v", errRollback)
			// O estado agora está inconsistente. Isso não deveria acontecer.
			return &comands.ApplyResponse{Success: false, Error: "ERRO CRITICO DE ROLLBACK"}
		}
		
		return &comands.ApplyResponse{Success: false, Error: "Falha ao transferir carta B para A (troca revertida)"}
	}

	log.Printf("[FSM] Troca concluida com sucesso.")
	return &comands.ApplyResponse{Success: true, Data: "trade_complete"}
}