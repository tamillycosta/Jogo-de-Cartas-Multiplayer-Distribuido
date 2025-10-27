package gamesession

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"errors"
	"log"
)

// LIMPA AS REFERENCIAS DE UMA PARTIDA QUE ACABAOU 
// REMOTA OU LOCAL 
func (gsm *GameSessionManager) cleanupMatch(matchID string) {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()

	// LOCAL 
	if session, exists := gsm.localSessions[matchID]; exists {
		log.Printf(" [SessionManager] Limpando partida LOCAL %s", matchID)
		
		// Remove dos mapas
		if session.Player1 != nil {
			delete(gsm.playerMatches, session.Player1.ID)
			log.Printf("   Removido player: %s", session.Player1.Username)
		}
		if session.Player2 != nil {
			delete(gsm.playerMatches, session.Player2.ID)
			log.Printf("   Removido player: %s", session.Player2.Username)
		}
		
		delete(gsm.localSessions, matchID)
		
		
		session.Close()
		
		log.Printf("[SessionManager] Partida LOCAL %s limpa", matchID)
		return
	}

	// REMOTA
	if session, exists := gsm.remoteSessions[matchID]; exists {
		log.Printf("[SessionManager] Limpando partida REMOTA %s", matchID)
		
		if session.LocalPlayer != nil {
			delete(gsm.playerMatches, session.LocalPlayer.ID)
			log.Printf("   Removido local player: %s", session.LocalPlayer.Username)
		}
		
		delete(gsm.remoteSessions, matchID)
		
		session.Close()
		
		log.Printf("[SessionManager] Partida REMOTA %s limpa", matchID)
		return
	}

	log.Printf("⚠️ [SessionManager] Partida %s não encontrada para limpeza", matchID)
}

// 
func (gsm *GameSessionManager) HandleClientDisconnect(clientID string) {
	playerID, exists := gsm.sessionManager.GetPlayerID(clientID)
	if !exists {
		log.Printf("⚠️ [SessionManager] ClientID %s não mapeado", clientID)
		return
	}

	log.Printf("🔌 [SessionManager] Cliente desconectado: %s (playerID: %s)", clientID, playerID)
	gsm.HandlePlayerDisconnect(playerID)
}


func (gsm *GameSessionManager) HandlePlayerDisconnect(playerID string) {
	gsm.mu.RLock()
	matchID, exists := gsm.playerMatches[playerID]
	gsm.mu.RUnlock()

	if !exists {
		log.Printf("⚠️ [SessionManager] Player %s não tinha partida ativa", playerID)
		gsm.localMatchmaking.RemoveFromQueue(playerID)
		gsm.globalMatchmaking.LeaveGlobalQueue(playerID)
		return
	}

//	Verifica se a partida ainda existe 
	gsm.mu.RLock()
	_, localExists := gsm.localSessions[matchID]
	_, remoteExists := gsm.remoteSessions[matchID]
	gsm.mu.RUnlock()

	if !localExists && !remoteExists {
		log.Printf("[SessionManager] Partida %s já foi encerrada, apenas limpando referência", matchID)
		
		gsm.mu.Lock()
		delete(gsm.playerMatches, playerID)
		gsm.mu.Unlock()
		
		return
	}

	log.Printf("[SessionManager] Processando leave de %s da partida %s", playerID, matchID)
	gsm.ProcessAction(matchID, playerID, entities.GameAction{
		Type: "leave_match",
	})
}

func (gsm *GameSessionManager) LeaveMatch(matchID, playerID string) error {
	gsm.mu.RLock()
	localSession, localExists := gsm.localSessions[matchID]
	remoteSession, remoteExists := gsm.remoteSessions[matchID]
	gsm.mu.RUnlock()
	
	if localExists {
		localSession.LeaveMatch(playerID)
		return nil
	}
	
	if remoteExists {
		remoteSession.HandleLeave(playerID)
		return nil
	}
	
	return errors.New("partida não encontrada")
}


