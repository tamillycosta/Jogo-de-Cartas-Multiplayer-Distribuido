package remota

import (
	"fmt"
	"log"
	"sync"
	"time"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/aplication/usecases"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	shared "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	contracts "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/service"
)
// REPRESENTA UMA PARTIDA CRIADA REMOTAMENTE 
type RemoteGameSession struct {
	mu sync.RWMutex

	MatchID string
	IsHost  bool

	LocalPlayer  *entities.GamePlayer
	RemotePlayer *entities.GamePlayer
	WinnerID 		string
	CurrentTurnPlayerID string
	TurnNumber          int
	Status              string

	broker         *pubsub.Broker
	remoteServerID string
	apiClient      *client.Client
	raft           *raft.RaftService

	LastHeartbeat time.Time
	heartbeatStop chan struct{}
	stopChan      chan struct{}
	onMatchEnd    func(matchID string)

	// contratos da chain 
	chainService *contracts.ChainService  

	// verificações de estados 
	closed        bool
	startNotified bool 
}

func New(matchID string,isHost bool,localPlayerID, localUsername, 
	localClientID string,remotePlayerID, remoteUsername, remoteClientID string,
	remoteServerID string,broker *pubsub.Broker,client *client.Client,
	raft *raft.RaftService, chainService *contracts.ChainService  ,onMatchEnd func(matchID string), ) *RemoteGameSession {

	var initialTurnPlayerID string
	if isHost {
		initialTurnPlayerID = remotePlayerID
		log.Printf("[RemoteGame] HOST decidiu: %s (%s) começa", remoteUsername, remotePlayerID)
	} else {
		initialTurnPlayerID = ""
		log.Printf("[RemoteGame] NÃO-HOST aguardando sincronização de turno inicial")
	}

	session := &RemoteGameSession{
		MatchID: matchID,
		IsHost:  isHost,
		LocalPlayer: &entities.GamePlayer{
			ID:       localPlayerID,
			Username: localUsername,
			ClientID: localClientID,
			Life:     INITIAL_PLAYER_LIFE,
			Deck:     make([]*entities.Card, 0),
		},
		RemotePlayer: &entities.GamePlayer{
			ID:       remotePlayerID,
			Username: remoteUsername,
			ClientID: remoteClientID,
			Life:     INITIAL_PLAYER_LIFE,
			Deck:     make([]*entities.Card, 0),
		},
		CurrentTurnPlayerID: initialTurnPlayerID,
		TurnNumber:          1,
		Status:              "waiting",
		remoteServerID:      remoteServerID,
		broker:              broker,
		apiClient:           client,
		raft:                raft,
		LastHeartbeat:       time.Now(),
		heartbeatStop:       make(chan struct{}),
		stopChan:            make(chan struct{}),
		onMatchEnd:          onMatchEnd,
		closed:              false,
		startNotified:       false,
		chainService: chainService,
	}

	go session.startHeartbeat()

	log.Printf("[RemoteGame] Criada: %s | IsHost=%v | Local=%s vs Remote=%s | InitialTurn=%s",
		matchID, isHost, localUsername, remoteUsername, initialTurnPlayerID)

	return session
}


// Sinaliza inicio de uma partida remota 
// sincorniza o servidor não host e jogador local 
func (s *RemoteGameSession) Start() {
	s.mu.Lock()
	s.Status = "in_progress"
	s.mu.Unlock()

	log.Printf("[RemoteGame] Iniciando partida %s | Host=%v", s.MatchID, s.IsHost)

	if s.IsHost {
		time.Sleep(100 * time.Millisecond)
		s.syncStateToRemote()
		log.Printf("[RemoteGame] HOST sincronizou turno inicial: %s", s.CurrentTurnPlayerID)
	}
	
	
	s.notifyLocalClient("match_started")

	log.Printf("[RemoteGame] Partida iniciada | CurrentTurn: %s", s.CurrentTurnPlayerID)
}




// ------------------------- HEARTBEAT ----------------------

func (s *RemoteGameSession) startHeartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Printf("[RemoteGame] Heartbeat iniciado | IsHost=%v | Match=%s", s.IsHost, s.MatchID)

	for {
		select {
		case <-ticker.C:
			s.mu.RLock()
			isClosed := s.closed
			isHost := s.IsHost
			status := s.Status
			s.mu.RUnlock()

			if isClosed || status == "finished" {
				log.Printf("[RemoteGame] Partida finalizada/fechada, parando heartbeat")
				return
			}

			if isHost {
				s.sendHeartbeat()
			} else {
				timeSince := time.Since(s.LastHeartbeat)

				if timeSince > 15*time.Second {
					log.Printf("[RemoteGame] Host não responde! Promovendo a host...")
					s.promoteToHost()
				}
			}

		case <-s.heartbeatStop:
			log.Printf("[RemoteGame] Heartbeat encerrado para match %s", s.MatchID)
			return
		}
	}
}



func (s *RemoteGameSession) sendHeartbeat() {
	servers := s.raft.GetServers()
	serverAddr := usecases.GetServer(s.remoteServerID, servers)
	if serverAddr == "" {
		return
	}

	if err := s.apiClient.MatchInterface.SendHeartbeat(serverAddr, s.MatchID); err != nil {
		log.Printf("[RemoteGame] Erro ao enviar heartbeat: %v", err)
	}
}










// ------------------------- AUXILIARES  ---------------------

// Carrega o deck de um jogador na blockachain
func (s *RemoteGameSession) LoadDecks(cardRepo *repository.CardRepository) error {
	log.Printf("[RemoteGame] Carregando deck do jogador LOCAL...")

	localCards, err := cardRepo.FindByPlayerID(s.LocalPlayer.ID)
	if err != nil {
		log.Printf("Erro ao carregar deck local: %v", err)
		return err
	}

	if len(localCards) > 3 {
		s.LocalPlayer.Deck = localCards[:3]
	} else {
		s.LocalPlayer.Deck = localCards
	}

	log.Printf("[RemoteGame] Deck local carregado: %d cartas", len(s.LocalPlayer.Deck))
	return nil
}


func (s *RemoteGameSession) LoadDecksFromBlockchain(playerRepo *repository.PlayerRepository, cardRepo *repository.CardRepository) error {
	log.Printf("[RemoteGame] Carregando deck LOCAL da BLOCKCHAIN...")

	// Busca jogador no banco para pegar endereço
	player, err := playerRepo.FindById(s.LocalPlayer.ID)
	if err != nil {
		return fmt.Errorf("erro ao buscar player local: %w", err)
	}

	// Carrega cartas da blockchain
	cards, err := usecases.LoadPlayerCardsFromChain(s.chainService ,player.Address, s.LocalPlayer.ID, cardRepo)
	if err != nil {
		return fmt.Errorf("erro ao carregar deck da blockchain: %w", err)
	}

	// Pega apenas 3 cartas
	if len(cards) > 3 {
		s.LocalPlayer.Deck = cards[:3]
	} else {
		s.LocalPlayer.Deck = cards
	}

	log.Printf("[RemoteGame] Deck local carregado: %d cartas", len(s.LocalPlayer.Deck))
	return nil
}


func (s *RemoteGameSession) forwardActionToHost(playerID string, action shared.GameAction) error {
	servers := s.raft.GetServers()
	hostAddr := usecases.GetServer(s.remoteServerID, servers)
	if hostAddr == "" {
		return fmt.Errorf("endereço do host %s não encontrado", s.remoteServerID)
	}

	if err := s.apiClient.MatchInterface.SendMatchAction(hostAddr, s.MatchID, playerID, action); err != nil {
		log.Printf("[RemoteGame] Erro ao encaminhar: %v", err)
		return err
	}

	log.Printf("[RemoteGame] Ação encaminhada")
	return nil
}






func (s *RemoteGameSession) promoteToHost() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IsHost {
		return
	}

	s.IsHost = true
	log.Printf("[RemoteGame] Servidor promovido a HOST da partida %s", s.MatchID)

	s.notifyLocalClient("host_changed")
}

func (s *RemoteGameSession) UpdateHeartbeat() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastHeartbeat = time.Now()
}

func (s *RemoteGameSession) Close() {
	s.mu.Lock()

	if s.closed {
		s.mu.Unlock()
		return
	}

	s.closed = true
	s.Status = "finished"
	s.mu.Unlock()

	log.Printf("[RemoteGame] Encerrando partida %s", s.MatchID)

	close(s.heartbeatStop)
	close(s.stopChan)

	log.Printf("[RemoteGame] Partida %s encerrada", s.MatchID)
}


