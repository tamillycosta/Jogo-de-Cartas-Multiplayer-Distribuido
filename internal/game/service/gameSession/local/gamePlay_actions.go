package local

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/aplication/usecases"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"log"
	"time"
	shared "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
)

// PROCESSA UMA JOGADA NA PARTIDA
// ESCOLHER CARTA
// SAIR DA PARTIDA
// ATACAR
func (s *LocalGameSession) ProcessAction(playerID string, action shared.GameAction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Verifica se partida já terminou
	if s.Status == "finished" {
		log.Printf("Partida já terminou!")
		return shared.NewGameError("partida já terminou")
	}
	player := s.getPlayer(playerID)
	opponent := s.getOpponent(playerID)
	

	if action.Type == ACTION_LEAVE_MATCH {
		log.Printf("%s desistiu", player.Username)
		s.LeaveMatch(playerID)
		return nil  
	}

	// Valida turno
	if playerID != s.CurrentTurnPlayerID {
		log.Printf("Não é turno de %s! CurrentTurn: %s", playerID, s.CurrentTurnPlayerID)
		return shared.ErrNotYourTurn
	}
	

	if player == nil || opponent == nil {
		log.Printf("Player/Opponent não encontrado!")
		return shared.NewGameError("erro ao recuperar players da partida")
	}
	
	log.Printf("Jogador: %s (Life: %d) | Oponente: %s (Life: %d)", 
		player.Username, player.Life, opponent.Username, opponent.Life)


	// Processa ação com recovery para evitar panic
	var actionErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC em ProcessAction: %v", r)
				actionErr = shared.NewGameError("erro interno ao processar ação")
			}
		}()
		
		switch action.Type {
		case ACTION_CHOOSE_CARD:
			actionErr = s.chooseCard(player, action.CardID)
			if actionErr == nil {
				log.Printf("%s escolheu carta", player.Username)
			}
			
		case ACTION_ATTACK:
			log.Printf("Processando ataque de %s...", player.Username)
			actionErr = s.attack(player, opponent, action)
			if actionErr == nil {
				log.Printf("Ataque de %s executado", player.Username)
			}
			
		 

		default:
			log.Printf("Ação inválida: %s", action.Type)
			actionErr = shared.NewGameError("ação inválida")
		}
	}()
	
	if actionErr != nil {
		log.Printf("Erro ao processar: %v", actionErr)
		return actionErr
	}
	
	
	// Verifica vitória 
	if winner := s.checkWinCondition(); winner != "" {
		winnerPlayer := s.getPlayer(winner)
		log.Printf("VITÓRIA! Vencedor: %s (Life: %d)", 
			winnerPlayer.Username, winnerPlayer.Life)
		
		s.endGame(winner)
		return nil 
	}
	
	// SE PARTIDA NÃO ACABAR TROCA TURNO 
	oldTurn := s.CurrentTurnPlayerID
	oldTurnPlayer := s.getPlayer(oldTurn)
	s.endTurn()
	newTurnPlayer := s.getPlayer(s.CurrentTurnPlayerID)
	
	log.Printf("Turno: %s -> %s | Turn #%d\n", 
		oldTurnPlayer.Username, newTurnPlayer.Username, s.TurnNumber)
		
	s.broadcastGameState("action_performed")
	return nil
}


// ESCOLHER CARTA
// OS JOGADORES PODEM ESCOLHER UMA CARTA DAS 3 
func (s *LocalGameSession) chooseCard(player *entities.GamePlayer, cardIndex string) error {
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
	
	if card.Health <= 0 {
		return shared.ErrNotEnoughLife
	}
	
	player.CurrentCard = card

	return nil
}


// ATACAR
// OS JOGADRES PRECISAM TER UMA CARTA NA MÃO COM VIDA
// O OPONENTE PRECISA TER CARTA NA MÃO COM VIDA 
func (s *LocalGameSession) attack(player, opponent *entities.GamePlayer, action shared.GameAction) error {
	
	log.Printf("   ⚔️ Atacante: %s | Defensor: %s", player.Username, opponent.Username)
	
	// Validações
	if player.CurrentCard == nil {
		return shared.NewGameError("você precisa escolher uma carta primeiro")
	}

	if opponent.CurrentCard == nil {
		return shared.ErrOppoentCard
	}

	if player.CurrentCard.Health <= 0 {
		return shared.ErrNotEnoughLife
	}

	attackerCard := player.CurrentCard
	defenderCard := opponent.CurrentCard
	damage := attackerCard.Power
	
	log.Printf("   %s (%s, Power: %d) vs %s (%s, HP: %d)",
		player.Username, attackerCard.Name, damage,
		opponent.Username, defenderCard.Name, defenderCard.Health)

	if defenderCard.Health > damage {
		// Carta sobrevive
		defenderCard.Health -= damage
	} else {
		// Carta destruída
		defenderCard.Health = 0
		opponent.Life -= 1
	
		usecases.RestoreCardsHp(player,opponent)
		
		player.CurrentCard = nil
		opponent.CurrentCard = nil
		
		log.Printf("Cartas limpas para próximo round")
	}
	

	return nil
}

// DESISTIR
func (s *LocalGameSession) LeaveMatch(playerID string) {
	player := s.getPlayer(playerID)
	opponent := s.getOpponent(playerID)
	
	if opponent != nil {
		s.endGame(opponent.ID)
		s.WinnerID = opponent.ID
	}
	
	if player != nil {
		log.Printf("[LocalGame] %s saiu da partida", player.Username)
	}
}


func (s *LocalGameSession) checkWinCondition() string {
	if s.Player1 != nil && s.Player1.Life <= 0 {
		return s.Player2.ID
	}
	if s.Player2 != nil && s.Player2.Life <= 0 {
		return s.Player1.ID
	}
	return ""
}



// Finaliza partida para os jogadores 
// notifica os jogadores e limpa as partidas 
func (s *LocalGameSession) endGame(winnerID string) {
	
	if s.Status == "finished" {
		log.Printf("[LocalGame] Partida %s já estava finalizada", s.MatchID)
		return
	}
	
	s.Status = "finished"
	
	winner := s.getPlayer(winnerID)
	loser := s.getOpponent(winnerID)

	s.WinnerID = winnerID

	if winner != nil && loser != nil {
		log.Printf("   [LocalGame] PARTIDA %s FINALIZADA!", s.MatchID)
		log.Printf("   Vencedor: %s (Life: %d)", winner.Username, winner.Life)
		log.Printf("   Perdedor: %s (Life: %d)", loser.Username, loser.Life)
	}
	
	
	s.broadcastGameState("match_ended")
	
	
	if s.onMatchEnd != nil {
		log.Printf("Agendando limpeza da partida %s...", s.MatchID)
		go func() {
			// Delay para garantir que broadcast chegue
			time.Sleep(500 * time.Millisecond)
			s.onMatchEnd(s.MatchID)
		}()
	}
}

func (s *LocalGameSession) endTurn() {
	if s.CurrentTurnPlayerID == s.Player1.ID {
		s.CurrentTurnPlayerID = s.Player2.ID
	} else {
		s.CurrentTurnPlayerID = s.Player1.ID
	}
	s.TurnNumber++
	currentPlayer := s.getPlayer(s.CurrentTurnPlayerID)
	if currentPlayer != nil {
		log.Printf("[LocalGame] Turno #%d: %s", s.TurnNumber, currentPlayer.Username)
	}
}


