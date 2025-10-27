package gamesession

// ============== GERENCIADOR UNIFICADO DE PARTIDAS ==============

import (
	"errors"
	"log"
	"sync"
	"time"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/gameSession/local"
	remote "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/gameSession/remota"
	matchglobal "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/matchMacking/match_global"
	matchlocal "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/matchMacking/match_local"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/session"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
)

// Gerencia todas as partidas do servidor (locais ou remotas)
type GameSessionManager struct {
	mu sync.RWMutex
	
	localSessions  map[string]*local.LocalGameSession    // Partidas locais
	remoteSessions map[string]*remote.RemoteGameSession  // Partidas remotas
	playerMatches  map[string]string                     // playerID -> matchID
	
	playerRepository     *repository.PlayerRepository
	cardRepository       *repository.CardRepository
	sessionManager       *session.SessionManager
	localMatchmaking     *matchlocal.LocalMatchmaking
	globalMatchmaking    *matchglobal.GlobalMatchmakingService

	// comunicação 
	broker               *pubsub.Broker
	serverID             string
	apiClient 			*client.Client
	raft 				*raft.RaftService
}

func New(
	playerRepo *repository.PlayerRepository,
	cardRepo *repository.CardRepository,
	sessionMgr *session.SessionManager,
	localMatchmaking *matchlocal.LocalMatchmaking,
	globalMatchmaking *matchglobal.GlobalMatchmakingService,
	broker *pubsub.Broker,
	serverID string,
	client	*client.Client,
	raft 				*raft.RaftService,
) *GameSessionManager {
	gsm := &GameSessionManager{
		localSessions:        make(map[string]*local.LocalGameSession),
		remoteSessions:       make(map[string]*remote.RemoteGameSession),
		playerMatches:        make(map[string]string),
		playerRepository:     playerRepo,
		cardRepository:       cardRepo,
		sessionManager:       sessionMgr,
		localMatchmaking:     localMatchmaking,
		globalMatchmaking:    globalMatchmaking,
		broker:               broker,
		serverID:             serverID,
		apiClient: client,
		raft: raft,
	}
	
	// Registra callbacks
	matchlocal.OnLocalMatchFound = gsm.createLocalMatch
	localMatchmaking.SetTimeoutCallback(gsm.handleLocalQueueTimeout)
	
	return gsm
}

// ----------------------- PARTIDAS LOCAIS --------------------------

func (gsm *GameSessionManager) createLocalMatch(p1, p2 *matchlocal.QueueEntry) {
	log.Printf("[GameSessionManager] Criando partida local: %s vs %s", p1.Username, p2.Username)

	if !gsm.sessionManager.IsPlayerLoggedIn(p1.PlayerID) || !gsm.sessionManager.IsPlayerLoggedIn(p2.PlayerID) {
		log.Printf("[GameSessionManager] Jogadores não logados")
		return
	}
	
	session := local.New(
		p1.PlayerID, p1.Username, p1.ClientID,
		p2.PlayerID, p2.Username, p2.ClientID,
		gsm.broker,
		gsm.cleanupMatch, 
	)

	gsm.mu.Lock()
	gsm.localSessions[session.MatchID] = session
	gsm.playerMatches[p1.PlayerID] = session.MatchID
	gsm.playerMatches[p2.PlayerID] = session.MatchID
	gsm.mu.Unlock()

	if err := session.LoadDecks(gsm.cardRepository); err != nil {
		log.Printf("[GameSessionManager] Erro ao carregar decks: %v", err)
		gsm.cleanupMatch(session.MatchID)
		return
	}

	gsm.notifyLocalMatchCreated(session, p1.ClientID, p2.ClientID)
	go session.Start()

	log.Printf("[GameSessionManager] Partida local criada: %s", session.MatchID)
}


// ----------------- TIMEOUT DA FILA LOCAL → FILA GLOBAL -----------------

// Callback quando jogador passa 20s na fila local
func (gsm *GameSessionManager) handleLocalQueueTimeout(entry *matchlocal.QueueEntry) {
	
	
	err := gsm.globalMatchmaking.JoinGlobalQueue(
		entry.ClientID,
		entry.PlayerID,
		entry.Username,
		gsm.serverID,
	)
	
	if err != nil {
		log.Printf("[GameSessionManager] Erro ao adicionar à fila global: %v", err)
		
		// Notifica cliente sobre erro
		gsm.broker.Publish("response."+entry.ClientID, map[string]interface{}{
			"type":  "error",
			"error": "Erro ao entrar na fila : " + err.Error(),
		})
		return
	}
	

}


// ------------------------- PARTIDAS REMOTAS ----------------------------

// Cria partida remota (chamado pela API REST após notificação do líder)
func (gsm *GameSessionManager) CreateRemoteMatch(
	matchID string,
	localPlayerID, localClientID string,
	remotePlayerID, remotePlayerUsername string,
	remoteServerID string,
	isHost bool,
) error {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()
	
	log.Printf("[GameSessionManager] Criando partida REMOTA: %s | IsHost=%v | Remote=%s (srv=%s)",
		matchID, isHost, remotePlayerUsername, remoteServerID)
	
	// Busca jogador local
	localPlayer, err := gsm.playerRepository.FindById(localPlayerID)
	if err != nil {
		return err
	}
	
	// Cria sessão remota
	session := remote.New(
		matchID,
		isHost,
		localPlayer.ID, localPlayer.Username, localClientID,
		remotePlayerID, remotePlayerUsername, "",
		remoteServerID,
		gsm.broker,
		gsm.apiClient,
		gsm.raft,
		gsm.cleanupMatch,
	)
	
	gsm.remoteSessions[matchID] = session
	gsm.playerMatches[localPlayerID] = matchID
	
	// Carrega deck do jogador LOCAL
	if err := session.LoadDecks(gsm.cardRepository); err != nil {
		delete(gsm.remoteSessions, matchID)
		delete(gsm.playerMatches, localPlayerID)
		return err
	}
	
	//  Se for HOST, carrega também o deck do jogador REMOTO
	if isHost {
		
		remotePlayer, err := gsm.playerRepository.FindById(remotePlayerID)
		if err != nil {
			log.Printf("[GameSessionManager] Jogador remoto não encontrado no banco: %v", err)
		
		} else {
			remoteCards, err := gsm.cardRepository.FindByPlayerID(remotePlayer.ID)
			if err == nil {
				if len(remoteCards) > 3 {
					session.RemotePlayer.Deck = remoteCards[:3]
				} else {
					session.RemotePlayer.Deck = remoteCards
				}
				log.Printf("[GameSessionManager] Deck remoto carregado: %d cartas", len(session.RemotePlayer.Deck))
			}
		}
	}
	
	// Notifica cliente local e inica partida 
	gsm.notifyClientRemoteMatchCreated(session, localClientID)
	
	go session.Start()
	
	log.Printf("[GameSessionManager] Partida REMOTA criada: %s", matchID)
	return nil
}






// ----------------- PROCESSAMENTO DE AÇÕES -----------------

// VERIFICA SE A PARTIDA É LOCAL OU REMOTA 
// PROCESSA DE ACORDO COM O TIPO DE PARTIDA 
func (gsm *GameSessionManager) ProcessAction(matchID, playerID string, action entities.GameAction) error {
	// Tenta encontrar partida local
	gsm.mu.RLock()
	localSession, localExists := gsm.localSessions[matchID]
	remoteSession, remoteExists := gsm.remoteSessions[matchID]
	gsm.mu.RUnlock()
	
	if localExists {
		return localSession.ProcessAction(playerID, action)
	}
	
	if remoteExists {
		return remoteSession.ProcessAction(playerID, action)
	}
	
	return errors.New("partida não encontrada")
}



// Recebe sincronização de partida remota (do host via REST)
func (gsm *GameSessionManager) ReceiveRemoteSync(matchID string, update remote.GameStateUpdate) error {
	gsm.mu.RLock()
	session, exists := gsm.remoteSessions[matchID]
	gsm.mu.RUnlock()
	
	if !exists {
		return errors.New("partida remota não encontrada")
	}
	
	return session.ReceiveSyncUpdate(update)
}



func (gsm *GameSessionManager) Shutdown() {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()
	
	log.Println("[UnifiedSessionMgr] Encerrando todas as partidas...")
	
	for matchID, session := range gsm.localSessions {
		session.Close()
		log.Printf("Partida local %s encerrada", matchID)
	}
	
	for matchID, session := range gsm.remoteSessions {
		session.Close()
		log.Printf(" Partida remota %s encerrada", matchID)
	}
	
	gsm.localSessions = make(map[string]*local.LocalGameSession)
	gsm.remoteSessions = make(map[string]*remote.RemoteGameSession)
	gsm.playerMatches = make(map[string]string)
	
	log.Println("[UnifiedSessionMgr] Todas as partidas encerradas")
}





// ----------------- Heartbeat -----------------



func (gsm *GameSessionManager) UpdateHeartbeat(matchId string,heartbeat time.Time){
	session, exists := gsm.GetRemoteSession(matchId)
	if(!exists){
		return
	}
	
	session.LastHeartbeat = heartbeat
}


// ----------------- CONSULTAS -----------------

func (gsm *GameSessionManager) IsPlayerInMatch(playerID string) (bool, string) {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	
	matchID, exists := gsm.playerMatches[playerID]
	return exists, matchID
}


func (gsm *GameSessionManager) GetActiveMatchesCount() (int, int) {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	return len(gsm.localSessions), len(gsm.remoteSessions)
}



func (gsm *GameSessionManager) GetLocalSession(matchID string) (*local.LocalGameSession, bool) {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	session, exists := gsm.localSessions[matchID]
	return session, exists
}



func (gsm *GameSessionManager) GetRemoteSession(matchID string) (*remote.RemoteGameSession, bool) {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	session, exists := gsm.remoteSessions[matchID]
	return session, exists
}



func (gsm *GameSessionManager) GetGlobalMacthService() *matchglobal.GlobalMatchmakingService {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	return gsm.globalMatchmaking
}

