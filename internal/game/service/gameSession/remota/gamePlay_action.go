package remota

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/aplication/usecases"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	shared "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"log"
	
)

const (
	ACTION_CHOOSE_CARD  = "play_card"
	ACTION_ATTACK       = "attack"
	ACTION_LEAVE_MATCH  = "leave_match"
	INITIAL_PLAYER_LIFE = 1
)

// METODO PARA SERVIDOR HOST PROCESSAR UMA AÇÃO DO JOGO 
// um jogador pode :
	// jogar carta
	// atacar 
	// sair da partida
func (s *RemoteGameSession) ProcessAction(playerID string, action shared.GameAction) error {
	// Se não é host, encaminha
	if !s.IsHost {
		log.Printf("[RemoteGame] Encaminhando para host...")
		return s.forwardActionToHost(playerID, action)
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[RemoteGame HOST] Processando | Turn=%s | Status=%s", 
		s.CurrentTurnPlayerID, s.Status)

	
	if s.Status == "finished" {
		return shared.NewGameError("partida já terminou")
	}

	
	if action.Type == ACTION_LEAVE_MATCH {
		return s.HandleLeave(playerID)
	}

	// Valida turno
	if playerID != s.CurrentTurnPlayerID {
		log.Printf("NÃO É SEU TURNO! Esperado: %s | Recebido: %s", 
			s.CurrentTurnPlayerID, playerID)
		return shared.ErrNotYourTurn
	}

	player := s.getPlayer(playerID)
	opponent := s.getOpponent(playerID)

	if player == nil || opponent == nil {
		return shared.NewGameError("jogador não encontrado")
	}

	// Processa ação
	var err error
	switch action.Type {
	case ACTION_CHOOSE_CARD:
		err = s.chooseCard(player, action.CardID)
		if err == nil {
			log.Printf("%s escolheu carta %s", player.Username, action.CardID)
		}

	case ACTION_ATTACK:
		err = s.attack(player, opponent)
		if err == nil {
			log.Printf("%s atacou", player.Username)
		}

	default:
		return shared.NewGameError("ação inválida")
	}

	if err != nil {
		return err
	}

	// Verifica vitória
	if winner := s.checkWinCondition(); winner != "" {
		winnerPlayer := s.getPlayer(winner)
		log.Printf("VITÓRIA! Vencedor: %s", winnerPlayer.Username)
		
		s.Status = "finished"
		s.WinnerID = winnerPlayer.ID
		s.notifyLocalClient("match_ended")
		
		
		s.mu.Unlock()
		s.syncStateToRemoteBlocking()
		s.mu.Lock()
		
		if s.onMatchEnd != nil {
			go s.onMatchEnd(s.MatchID)
		}
		
		return nil
	}

	// Troca turno
	oldPlayer := s.getPlayer(s.CurrentTurnPlayerID)
	s.endTurn()
	newPlayer := s.getPlayer(s.CurrentTurnPlayerID)

	log.Printf("Turno: %s -> %s | Turn #%d", 
		oldPlayer.Username, newPlayer.Username, s.TurnNumber)

	
	s.notifyLocalClient("action_performed")

	s.mu.Unlock()
	s.syncStateToRemote()
	s.mu.Lock()

	return nil
}

// LIDA COM SAIDA DE UM JOGADOR DA PARTIDA 
func (s *RemoteGameSession) HandleLeave(playerID string) error {
	if s.Status == "finished" {
		log.Printf("⚠️ [RemoteGame] Leave ignorado: partida já terminou (playerID: %s)", playerID)
		return nil
	}
	
	log.Printf(" [RemoteGame] %s está saindo", playerID)
	
	
	opponent := s.getOpponent(playerID)

	//  FIX: Oponente SEMPRE vence (ou empate se não houver oponente)
	var winnerID string
	if opponent != nil {
		winnerID = opponent.ID
		log.Printf("Vencedor: %s (oponente de quem desistiu)", opponent.Username)
	} else {
		log.Printf("⚠️ Sem oponente - partida encerrada sem vencedor")
	}

	//  FIX: Salva o winner no estado da sessão
	s.WinnerID = winnerID
	s.Status = "finished"

	s.notifyLocalClient("match_ended")

	s.mu.Unlock()
	if s.IsHost {
		s.syncStateToRemoteBlocking()
	}
	s.mu.Lock()

	if s.onMatchEnd != nil {
		go s.onMatchEnd(s.MatchID)
	}

	return nil
}


// ESCOLHE UMA CARTA 
func (s *RemoteGameSession) chooseCard(player *entities.GamePlayer, cardIndex string) error {
	if len(player.Deck) == 0 {
		return shared.NewGameError("deck vazio")
	}

	index := 0
	switch cardIndex {
	case "0":
		index = 0
	case "1":
		index = 1
	case "2":
		index = 2
	default:
		return shared.ErrCardNotInHand
	}

	if index < 0 || index >= len(player.Deck) {
		return shared.ErrCardNotInHand
	}

	card := player.Deck[index]
	if card == nil || card.Health <= 0 {
		return shared.ErrNotEnoughLife
	}

	player.CurrentCard = card
	return nil
}



func (s *RemoteGameSession) attack(player, opponent *entities.GamePlayer) error {
	if player.CurrentCard == nil {
		return shared.NewGameError("escolha uma carta primeiro")
	}

	if opponent.CurrentCard == nil {
		return shared.ErrOppoentCard
	}

	attackerCard := player.CurrentCard
	defenderCard := opponent.CurrentCard
	damage := attackerCard.Power

	log.Printf(" %s (%d) vs %s (%d)", 
		attackerCard.Name, damage, defenderCard.Name, defenderCard.Health)

	if defenderCard.Health > damage {
		defenderCard.Health -= damage
		log.Printf("Defensor sobreviveu com %d HP", defenderCard.Health)
	} else {
		defenderCard.Health = 0
		opponent.Life -= 1

		usecases.RestoreCardsHp(player, opponent)

		player.CurrentCard = nil
		opponent.CurrentCard = nil

		log.Printf("Carta destruída! %s perdeu 1 vida (Total: %d)", 
			opponent.Username, opponent.Life)
	}

	return nil
}


// ---------------------------- AUXILIARES --------------------------------


func (s *RemoteGameSession) endTurn() {
	if s.CurrentTurnPlayerID == s.LocalPlayer.ID {
		s.CurrentTurnPlayerID = s.RemotePlayer.ID
	} else {
		s.CurrentTurnPlayerID = s.LocalPlayer.ID
	
	}
	s.TurnNumber++
}


func (s *RemoteGameSession) checkWinCondition() string {
	if s.LocalPlayer != nil && s.LocalPlayer.Life <= 0 {
		return s.RemotePlayer.ID
	}
	if s.RemotePlayer != nil && s.RemotePlayer.Life <= 0 {
		return s.LocalPlayer.ID
	}
	return ""
}

func (s *RemoteGameSession) getPlayer(playerID string) *entities.GamePlayer {
	if s.LocalPlayer != nil && s.LocalPlayer.ID == playerID {
		return s.LocalPlayer
	}
	if s.RemotePlayer != nil && s.RemotePlayer.ID == playerID {
		return s.RemotePlayer
	}
	return nil
}

func (s *RemoteGameSession) getOpponent(playerID string) *entities.GamePlayer {
	if s.LocalPlayer != nil && s.LocalPlayer.ID == playerID {
		return s.RemotePlayer
	}
	return s.LocalPlayer
}