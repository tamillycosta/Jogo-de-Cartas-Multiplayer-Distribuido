
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

	
	
	// Aplica turno e status
	s.CurrentTurnPlayerID = update.CurrentTurnPlayerID
	s.TurnNumber = update.TurnNumber
	s.Status = update.Status
	s.WinnerID = update.WinnerID

	// O "Local" do host é o "Remote" deste servidor!
	// O host envia: LocalPlayer = jogador dele, RemotePlayer = jogador daqui
	// Então ao receber: LocalPlayer aqui = RemotePlayer do update!
	
	s.LocalPlayer.Life = update.RemotePlayerLife
	s.LocalPlayer.CurrentCard = update.RemotePlayerCurrentCard
	
	s.RemotePlayer.Life = update.LocalPlayerLife
	s.RemotePlayer.CurrentCard = update.LocalPlayerCurrentCard

	s.LastHeartbeat = time.Now()
	
	var eventType string

	if !s.startNotified && s.Status == "in_progress" {
		s.startNotified = true
		eventType = "match_started"
		log.Printf("[RemoteGame NÃO-HOST] Primeira sync recebida, notificando match_started")
	} else if s.Status == "finished" && oldStatus != "finished" {
		eventType = "match_ended"
		log.Printf("[RemoteGame NÃO-HOST] Partida finalizada!")
		
		if s.onMatchEnd != nil {
			go s.onMatchEnd(s.MatchID)
		}
	} else {
		eventType = "state_updated"
	}

	//  Notifica ANTES de dar unlock (para garantir que o estado está consistente)
	s.notifyLocalClient(eventType)

	log.Printf("[RemoteGame]  Estado sincronizado | Turn: %s (#%d) | LocalLife: %d | RemoteLife: %d",
		s.CurrentTurnPlayerID, s.TurnNumber, s.LocalPlayer.Life, s.RemotePlayer.Life)
	
	return nil
}


// SINCRONIZA ESTADO DO SERVIDOR NÃO HOST 
func (s *RemoteGameSession) syncStateToRemote() {
	if !s.IsHost {
		return
	}

	servers := s.raft.GetServers()
	remoteAddr := usecases.GetServer(s.remoteServerID, servers)
	if remoteAddr == "" {
		log.Printf("[RemoteGame] Servidor remoto %s não encontrado", s.remoteServerID)
		return
	}

	update := GameStateUpdate{
		MatchID:                 s.MatchID,
		CurrentTurnPlayerID:     s.CurrentTurnPlayerID,
		TurnNumber:              s.TurnNumber,
		Status:                  s.Status,
		LocalPlayerLife:         s.LocalPlayer.Life,
		WinnerID:                s.WinnerID, 
		LocalPlayerCurrentCard:  s.LocalPlayer.CurrentCard,
		RemotePlayerLife:        s.RemotePlayer.Life,
		RemotePlayerCurrentCard: s.RemotePlayer.CurrentCard,
		Timestamp:               time.Now().Unix(),
	}

	//  LOG PARA DEBUG
	log.Printf("[RemoteGame SYNC] Enviando para %s:", s.remoteServerID)
	log.Printf("  Turn: %d | CurrentTurn: %s", update.TurnNumber, update.CurrentTurnPlayerID)
	log.Printf("  LocalPlayer (%s): Life=%d, Card=%v", 
		s.LocalPlayer.Username, s.LocalPlayer.Life, s.LocalPlayer.CurrentCard != nil)
	log.Printf("  RemotePlayer (%s): Life=%d, Card=%v", 
		s.RemotePlayer.Username, s.RemotePlayer.Life, s.RemotePlayer.CurrentCard != nil)

	go func() {
		if err := s.apiClient.MatchInterface.SendMatchSync(remoteAddr, update); err != nil {
			log.Printf("[RemoteGame] Erro ao sincronizar: %v", err)
		} else {
			log.Printf("[RemoteGame]  Estado sincronizado com %s", s.remoteServerID)
		}
	}()
}

func (s *RemoteGameSession) syncStateToRemoteBlocking() {
	if !s.IsHost {
		return
	}

	servers := s.raft.GetServers()
	remoteAddr := usecases.GetServer(s.remoteServerID, servers)
	if remoteAddr == "" {
		log.Printf("[RemoteGame] Servidor remoto %s não encontrado", s.remoteServerID)
		return
	}

	update := GameStateUpdate{
		MatchID:                 s.MatchID,
		CurrentTurnPlayerID:     s.CurrentTurnPlayerID,
		TurnNumber:              s.TurnNumber,
		Status:                  s.Status,
		LocalPlayerLife:         s.LocalPlayer.Life,
		WinnerID:                s.WinnerID,
		LocalPlayerCurrentCard:  s.LocalPlayer.CurrentCard,
		RemotePlayerLife:        s.RemotePlayer.Life,
		RemotePlayerCurrentCard: s.RemotePlayer.CurrentCard,
		Timestamp:               time.Now().Unix(),
	}

	
	log.Printf("[RemoteGame SYNC BLOCKING] Enviando para %s", s.remoteServerID)

	if err := s.apiClient.MatchInterface.SendMatchSync(remoteAddr, update); err != nil {
		log.Printf("[RemoteGame] Erro ao sincronizar (blocking): %v", err)
	} else {
		log.Printf("[RemoteGame]  Estado sincronizado (blocking) com %s", s.remoteServerID)
	}
}



