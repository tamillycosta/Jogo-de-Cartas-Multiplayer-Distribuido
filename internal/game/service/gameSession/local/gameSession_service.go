package local

import (
	"errors"
	"log"
	"sync"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	matchmaking "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/matchMacking/match_local"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/session"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
)

// Grencia todoas as partidas locais que estiverem ativas no serivdor
type GameSessionManager struct {
	mu sync.RWMutex

	activeSessions map[string]*LocalGameSession
	playerMatches  map[string]string // playerID -> matchID (para limpeza rápida)

	playerRepository *repository.PlayerRepository
	cardRepository   *repository.CardRepository
	sessionManager   *session.SessionManager
	matchmaking      *matchmaking.LocalMatchmaking
	broker           *pubsub.Broker
}

func NewGameSessionManager(
	playerRepo *repository.PlayerRepository, cardRepo *repository.CardRepository, sessionMgr *session.SessionManager, matchmakingMgr *matchmaking.LocalMatchmaking, broker *pubsub.Broker) *GameSessionManager {
	gsm := &GameSessionManager{
		activeSessions:   make(map[string]*LocalGameSession),
		playerMatches:    make(map[string]string),
		playerRepository: playerRepo,
		cardRepository:   cardRepo,
		sessionManager:   sessionMgr,
		matchmaking:      matchmakingMgr,
		broker:           broker,
	}

	matchmaking.OnLocalMatchFound = gsm.createLocalMatch
	return gsm
}



// CRIA UMA PARTIDA LOCAL 
func (gsm *GameSessionManager) createLocalMatch(p1, p2 *matchmaking.QueueEntry) {
	log.Printf("[SessionManager] Criando partida local: %s vs %s", p1.Username, p2.Username)

	if !gsm.sessionManager.IsPlayerLoggedIn(p1.PlayerID) || !gsm.sessionManager.IsPlayerLoggedIn(p2.PlayerID) {
		log.Printf("[SessionManager] Jogadores não logados")
		return
	}
	
	session := New(
		p1.PlayerID, p1.Username, p1.ClientID,
		p2.PlayerID, p2.Username, p2.ClientID,
		gsm.broker,
		gsm.cleanupMatch, 
	)

	gsm.mu.Lock()
	gsm.activeSessions[session.MatchID] = session
	gsm.playerMatches[p1.PlayerID] = session.MatchID
	gsm.playerMatches[p2.PlayerID] = session.MatchID
	gsm.mu.Unlock()

	if err := session.LoadDecks(gsm.cardRepository); err != nil {
		log.Printf("[SessionManager] Erro ao carregar decks: %v", err)
		gsm.cleanupMatch(session.MatchID)
		return
	}

	gsm.notifyLocalMatchCreated(session, p1.ClientID, p2.ClientID)
	go session.Start()

	log.Printf("[SessionManager] Partida local criada: %s", session.MatchID)
}




func (gsm *GameSessionManager) ProcessAction(matchID, playerID string, action entities.GameAction) error {
	session, exists := gsm.GetSession(matchID)

	if !exists {
		return errors.New("partida não encontrada")
	}

	return session.ProcessAction(playerID, action)
}



func (gsm *GameSessionManager) LeaveMatch(matchID, playerID string) error {
	session, exists := gsm.GetSession(matchID)

	if !exists {
		log.Printf("⚠️ [SessionManager] Partida %s não encontrada", matchID)
		return errors.New("partida não encontrada")
	}
	session.LeaveMatch(playerID)
	return nil
}






// ------------------------ lida com desconexão dos jogadores  -------------------------

// Chamar quando player desconectar
func (gsm *GameSessionManager) HandlePlayerDisconnect(playerID string) {
	gsm.mu.RLock()
	matchID, exists := gsm.playerMatches[playerID]
	gsm.mu.RUnlock()

	if !exists {
		// Player pode estar na fila
		log.Printf("[SessionManager] Player %s não tinha partida ativa", playerID)
		gsm.matchmaking.RemoveFromQueue(playerID)
		return
	}

	// Remove da partida ativa
	gsm.LeaveMatch(matchID, playerID)
}

func (gsm *GameSessionManager) HandleClientDisconnect(clientID string) {
	playerID, exists := gsm.sessionManager.GetPlayerID(clientID)
	if !exists {
		log.Printf("⚠️ [SessionManager] ClientID %s não mapeado", clientID)
		return
	}

	log.Printf("[SessionManager] Cliente desconectado: %s (playerID: %s)", clientID, playerID)
	gsm.HandlePlayerDisconnect(playerID)
}




// ------------------------- - LIMPA REFERENCIA DAS PARTIDAS 
func (gsm *GameSessionManager) cleanupMatch(matchID string) {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()

	session, exists := gsm.activeSessions[matchID]
	if !exists {
		log.Printf("⚠️ [SessionManager] Partida %s já foi limpa", matchID)
		return
	}


	if session.Player1 != nil {
		delete(gsm.playerMatches, session.Player1.ID)
	}
	if session.Player2 != nil {
		delete(gsm.playerMatches, session.Player2.ID)
	}

	
	delete(gsm.activeSessions, matchID)

	log.Printf("[SessionManager] Partida %s limpa (players: %d, sessions: %d)",
		matchID, len(gsm.playerMatches), len(gsm.activeSessions))
}
















// ---------------------- AUXILIARES -----------------------


func (gsm *GameSessionManager) GetSession(matchID string) (*LocalGameSession, bool) {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	session, exists := gsm.activeSessions[matchID]
	return session, exists
}

func (gsm *GameSessionManager) IsPlayerInMatch(playerID string) (bool, string) {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()

	matchID, exists := gsm.playerMatches[playerID]
	return exists, matchID
}

func (gsm *GameSessionManager) GetActiveMatchesCount() int {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	return len(gsm.activeSessions)
}

func (gsm *GameSessionManager) Shutdown() {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()

	log.Println("[SessionManager] Encerrando todas as partidas...")

	for matchID, session := range gsm.activeSessions {
		session.Close()
		log.Printf("Partida %s encerrada", matchID)
	}

	gsm.activeSessions = make(map[string]*LocalGameSession)
	gsm.playerMatches = make(map[string]string)

	log.Println("[SessionManager] Todas as partidas encerradas")
}


