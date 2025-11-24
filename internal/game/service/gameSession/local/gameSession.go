package local

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/aplication/usecases"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	"errors"

	"fmt"
	"log"
	"sync"
	"time"
	contracts "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/service"
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
	WinnerID string
	CurrentTurnPlayerID string
	TurnNumber          int
	Status              string
	broker              *pubsub.Broker	
	stopChan            chan struct{}
	
	//  Callback para limpar quando partida termina
	onMatchEnd func(matchID string) 

	// contratos da chain 
	chainService *contracts.ChainService  
}


func New(
	player1ID, player1Username, player1ClientID string,player2ID, player2Username, player2ClientID string, broker *pubsub.Broker,chainService *contracts.ChainService  ,onMatchEnd func(matchID string),) *LocalGameSession {
	
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
		chainService: chainService,
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
	
	time.Sleep(2 * time.Second)
	s.broadcastGameState("match_started")

	log.Printf("[LocalGame] Partida iniciada com sucesso")
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



func (s *LocalGameSession) LoadDecksFromBlockchain(playerRepo *repository.PlayerRepository ,cardRepo *repository.CardRepository) error {
	log.Printf("[LocalGame] Carregando decks da BLOCKCHAIN...")
	
	if s.Player1 == nil || s.Player2 == nil {
		return errors.New("players não inicializados")
	}

	// Busca players no banco para pegar endereços
	p1Entity, err := playerRepo.FindById(s.Player1.ID)
	if err != nil {
		return fmt.Errorf("erro ao buscar player1: %w", err)
	}

	p2Entity, err := playerRepo.FindById(s.Player2.ID)
	if err != nil {
		return fmt.Errorf("erro ao buscar player2: %w", err)
	}

	// Carrega cartas da blockchain
	p1Cards, err := usecases.LoadPlayerCardsFromChain(s.chainService,p1Entity.Address, s.Player1.ID, cardRepo)
	if err != nil {
		return fmt.Errorf("erro ao carregar deck P1 da blockchain: %w", err)
	}

	p2Cards, err :=  usecases.LoadPlayerCardsFromChain(s.chainService,p2Entity.Address, s.Player2.ID, cardRepo)
	if err != nil {
		return fmt.Errorf("erro ao carregar deck P2 da blockchain: %w", err)
	}

	// Pega apenas 3 cartas
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

	log.Printf("[LocalGame] Decks carregados da blockchain: P1=%d | P2=%d", 
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



