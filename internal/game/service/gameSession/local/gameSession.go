package local

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	"errors"
	"log"
	"sync"
	"time"
	

	"github.com/google/uuid"
)

const (
	ACTION_CHOOSE_CARD = "play_card"
	ACTION_ATTACK      = "attack"
	ACTION_LEAVE_MATCH = "leave_match"

	INITIAL_PLAYER_LIFE = 1
)




// REPRESENTA UMA PARTIDA LOCAL 
// É GERENCIADA PELO GAME SESSION MANAGER 
type LocalGameSession struct {
	mu sync.RWMutex
	
	MatchID  string
	Player1  *entities.GamePlayer
	Player2  *entities.GamePlayer
	
	CurrentTurnPlayerID string
	TurnNumber          int
	Status              string
	broker              *pubsub.Broker	
	stopChan            chan struct{}
	
	//  Callback para limpar quando partida termina
	onMatchEnd func(matchID string) 
}


func New(
	player1ID, player1Username, player1ClientID string,player2ID, player2Username, player2ClientID string, broker *pubsub.Broker,onMatchEnd func(matchID string),) *LocalGameSession {
	
	matchID := uuid.New().String()
	
	session := &LocalGameSession{
		MatchID: matchID,
		Player1: &entities.GamePlayer{
			ID:       player1ID,
			Username: player1Username,
			ClientID: player1ClientID,
			Life:     INITIAL_PLAYER_LIFE,
			Deck:     make([]*entities.Card, 0),
		},
		Player2: &entities.GamePlayer{
			ID:       player2ID,
			Username: player2Username,
			ClientID: player2ClientID,
			Life:     INITIAL_PLAYER_LIFE,
			Deck:     make([]*entities.Card, 0),
		},
		CurrentTurnPlayerID: player1ID, 
		TurnNumber:          1,
		Status:              "waiting",
		broker:              broker,
		stopChan:            make(chan struct{}),
		onMatchEnd:          onMatchEnd,
	}
	
	log.Printf("[LocalGameSession] Criada: %s | P1: %s vs P2: %s",
		matchID, player1Username, player2Username)
	
	return session
}



// inicia partida 
func (s *LocalGameSession) Start() {
	s.mu.Lock()
	s.Status = "in_progress"
	s.mu.Unlock()
	
	log.Printf("[LocalGame] Iniciando partida %s: %s vs %s | Turno: %s", 
		s.MatchID, s.Player1.Username, s.Player2.Username, s.CurrentTurnPlayerID)
	
	s.broadcastGameState("match_started")
	log.Printf("[LocalGame] Partida iniciada com sucesso")
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






// ---------------------------- AUXILIARES ---------------------------------------



// CARREGA 3 CARTAS DO DECK PARA USAR NA PARTIDA 
func (s *LocalGameSession) LoadDecks(cardRepo *repository.CardRepository) error {
	log.Printf("[LocalGame] Carregando decks...")
	
	if s.Player1 == nil || s.Player2 == nil {
		return errors.New("players não inicializados")
	}
	
	p1Cards, err := cardRepo.FindByPlayerID(s.Player1.ID)
	if err != nil {
		log.Printf("Erro ao carregar deck P1: %v", err)
		return err
	}
	
	p2Cards, err := cardRepo.FindByPlayerID(s.Player2.ID)
	if err != nil {
		log.Printf("Erro ao carregar deck P2: %v", err)
		return err
	}
	
	if len(p1Cards) > 3 {
		s.Player1.Deck = p1Cards[:3]
	} else {
		s.Player1.Deck = p1Cards
	}
	
	if len(p2Cards) > 3 {
		s.Player2.Deck = p2Cards[:3]
	} else {
		s.Player2.Deck = p2Cards
	}
	
	log.Printf("Decks carregados: P1=%d cartas | P2=%d cartas", 
		len(s.Player1.Deck), len(s.Player2.Deck))
	
	return nil
}


func (s *LocalGameSession) getPlayer(playerID string) *entities.GamePlayer {
	if s.Player1 != nil && s.Player1.ID == playerID {
		return s.Player1
	}
	if s.Player2 != nil && s.Player2.ID == playerID {
		return s.Player2
	}
	return nil
}


func (s *LocalGameSession) getOpponent(playerID string) *entities.GamePlayer {
	if s.Player1 != nil && s.Player1.ID == playerID {
		return s.Player2
	}
	if s.Player2 != nil && s.Player2.ID == playerID {
		return s.Player1
	}
	log.Printf("⚠️ [getOpponent] Oponente de %s não encontrado", playerID)
	return nil
}




func (s *LocalGameSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	close(s.stopChan)
	s.Status = "finished"
	log.Printf("[LocalGame] Partida %s encerrada", s.MatchID)
}



func (s *LocalGameSession) endTurn() {
	if s.CurrentTurnPlayerID == s.Player1.ID {
		s.CurrentTurnPlayerID = s.Player2.ID
	} else {
		s.CurrentTurnPlayerID = s.Player1.ID
		s.TurnNumber++
	}
	
	currentPlayer := s.getPlayer(s.CurrentTurnPlayerID)
	if currentPlayer != nil {
		log.Printf("[LocalGame] Turno #%d: %s", s.TurnNumber, currentPlayer.Username)
	}
}

