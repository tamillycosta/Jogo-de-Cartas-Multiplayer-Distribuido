
package remota


import (
	"fmt"
	"log"
	
	"time"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/aplication/usecases"

	
)


// ----------------------------- SINCRONIZAÇÃO  ------------------------------------

// Recebe sincronização do servidor host 
// e processa de acordo com o tipo de notificação de update 
// se for fim de partida
// inicio de partida
// ou mudança de estado 
func (s *RemoteGameSession) ReceiveSyncUpdate(update GameStateUpdate) error {
	if s.IsHost {
		return fmt.Errorf("host não deve receber sync updates")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldStatus := s.Status

	
	s.CurrentTurnPlayerID = update.CurrentTurnPlayerID
	s.TurnNumber = update.TurnNumber
	s.Status = update.Status

	s.RemotePlayer.Life = update.LocalPlayerLife
	s.RemotePlayer.CurrentCard = update.LocalPlayerCurrentCard

	s.LocalPlayer.Life = update.RemotePlayerLife
	s.LocalPlayer.CurrentCard = update.RemotePlayerCurrentCard

	s.LastHeartbeat = time.Now()

	
	var eventType string

	if !s.startNotified && s.Status == "in_progress" {
		// partida iniciou
		s.startNotified = true
		eventType = "match_started"
		log.Printf("[RemoteGame NÃO-HOST] Primeira sync recebida, notificando match_started")
	} else if s.Status == "finished" && oldStatus != "finished" {
		// Partida acabou de terminar
		eventType = "match_ended"
		log.Printf("[RemoteGame NÃO-HOST] Partida finalizada! LocalLife=%d RemoteLife=%d", 
			s.LocalPlayer.Life, s.RemotePlayer.Life)
		
		//  Agenda cleanup da partida
		if s.onMatchEnd != nil {
			go s.onMatchEnd(s.MatchID)
		}
	} else {
		// Sync normal
		eventType = "state_updated"
	}

	
	s.mu.Unlock()
	s.notifyLocalClient(eventType)
	s.mu.Lock()

	log.Printf("[RemoteGame] Estado sincronizado | Turn: %s | TurnNum: %d | LocalLife: %d | RemoteLife: %d | Status: %s",
		s.CurrentTurnPlayerID, s.TurnNumber, s.LocalPlayer.Life, s.RemotePlayer.Life, s.Status)
	
	return nil
}


// SINCRONIZA ESTADO DO SERVIDOR NÃO HOST 
func (s *RemoteGameSession) syncStateToRemote() {
	if !s.IsHost {
		return
	}

	servers := s.raft.GetServers()
	serverAddr := usecases.GetServer(s.remoteServerID, servers)

	if serverAddr == "" {
		log.Printf("[RemoteGame] Servidor remoto %s não encontrado", s.remoteServerID)
		return
	}

	s.mu.RLock()
	update := GameStateUpdate{
		MatchID:                 s.MatchID,
		CurrentTurnPlayerID:     s.CurrentTurnPlayerID,
		TurnNumber:              s.TurnNumber,
		Status:                  s.Status,
		LocalPlayerLife:         s.LocalPlayer.Life,
		LocalPlayerCurrentCard:  s.LocalPlayer.CurrentCard,
		RemotePlayerLife:        s.RemotePlayer.Life,
		RemotePlayerCurrentCard: s.RemotePlayer.CurrentCard,
		Timestamp:               time.Now().Unix(),
	}
	s.mu.RUnlock()

	log.Printf("[RemoteGame] Sincronizando com %s | Turn: %s | TurnNum: %d | Status: %s",
		serverAddr, update.CurrentTurnPlayerID, update.TurnNumber, update.Status)

	if err := s.apiClient.MatchInterface.SendMatchSync(serverAddr, update); err != nil {
		log.Printf("[RemoteGame] Erro ao sincronizar: %v", err)
	} else {
		log.Printf("[RemoteGame] Sincronização enviada com sucesso")
	}
}



func (s *RemoteGameSession) syncStateToRemoteBlocking() {
	if !s.IsHost {
		return
	}

	servers := s.raft.GetServers()
	serverAddr := usecases.GetServer(s.remoteServerID, servers)
	
	if serverAddr == "" {
		log.Printf("[RemoteGame] Servidor remoto %s não encontrado", s.remoteServerID)
		return
	}

	s.mu.RLock()
	update := GameStateUpdate{
		MatchID:                 s.MatchID,
		CurrentTurnPlayerID:     s.CurrentTurnPlayerID,
		TurnNumber:              s.TurnNumber,
		Status:                  s.Status,
		LocalPlayerLife:         s.LocalPlayer.Life,
		LocalPlayerCurrentCard:  s.LocalPlayer.CurrentCard,
		RemotePlayerLife:        s.RemotePlayer.Life,
		RemotePlayerCurrentCard: s.RemotePlayer.CurrentCard,
		Timestamp:               time.Now().Unix(),
	}
	s.mu.RUnlock()

	log.Printf("[RemoteGame] Sincronizando estado FINAL com %s | Status: %s",
		serverAddr, update.Status)

	// Tenta 3 vezes com timeout curto
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if err := s.apiClient.MatchInterface.SendMatchSync(serverAddr, update); err != nil {
			log.Printf("[RemoteGame] Tentativa %d/%d falhou: %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(500 * time.Millisecond)
			}
		} else {
			log.Printf("[RemoteGame] Estado final sincronizado com sucesso")
			return
		}
	}
	
	log.Printf("[RemoteGame] Falha ao sincronizar estado final após %d tentativas", maxRetries)
}
