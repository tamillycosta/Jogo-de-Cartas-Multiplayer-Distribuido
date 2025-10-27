package matchlocal


import (
	"log"
	"sync"
	"time"
	raftService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
)

// Gere fila de jogadores para encontrar match local 
type LocalMatchmaking struct {
	mu          sync.RWMutex
	localQueue  []*QueueEntry     
	serverID    string
	raft         *raftService.RaftService 
	onTimeout   func(entry *QueueEntry)
}

type QueueEntry struct {
	ClientID  string  
	PlayerID  string
	Username  string
	JoinedAt  time.Time
}

func New(serverID string, raft *raftService.RaftService ) *LocalMatchmaking {
	lm := &LocalMatchmaking{
		localQueue: make([]*QueueEntry, 0),
		serverID:   serverID,
		raft: raft,
	}
	
	// loop para temtar 
	go lm.matchmakingLoop()

	go lm.timeoutCheckLoop()
	return lm
}



func (lm *LocalMatchmaking) SetTimeoutCallback(callback func(entry *QueueEntry)) {
	lm.onTimeout = callback
}



func (lm *LocalMatchmaking) AddToQueue(clientID, playerID, username string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	lm.removeFromQueueUnsafe(playerID)
	
	entry := &QueueEntry{
		ClientID: clientID,  
		PlayerID: playerID,
		Username: username,
		JoinedAt: time.Now(),
	}
	
	lm.localQueue = append(lm.localQueue, entry)
	
	log.Printf("[LocalMatchmaking] Player %s adicionado à fila local (total: %d)", 
		username, len(lm.localQueue))
}

func (lm *LocalMatchmaking) RemoveFromQueue(playerID string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.removeFromQueueUnsafe(playerID)
}



func (lm *LocalMatchmaking) removeFromQueueUnsafe(playerID string) {
	newQueue := make([]*QueueEntry, 0)
	for _, entry := range lm.localQueue {
		if entry.PlayerID != playerID {
			newQueue = append(newQueue, entry)
		}
	}
	lm.localQueue = newQueue
}

func (lm *LocalMatchmaking) GetQueueSize() int {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return len(lm.localQueue)
}


func (lm *LocalMatchmaking) matchmakingLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		lm.mu.Lock()
		
		if len(lm.localQueue) < 2 {
			lm.mu.Unlock()
			continue
		}
		
		p1 := lm.localQueue[0]
		p2 := lm.localQueue[1]
		lm.localQueue = lm.localQueue[2:]
		
		lm.mu.Unlock()
		
		log.Printf("[LocalMatchmaking] Match local encontrado: %s vs %s", 
			p1.Username, p2.Username)
		
		go lm.onMatchFound(p1, p2)
	}
}



// checa se jogador ja passou muito tempo na fila local sem achar partida 
func (lm *LocalMatchmaking) timeoutCheckLoop() {
	ticker := time.NewTicker(5 * time.Second) // Verifica a cada 5 segundos
	defer ticker.Stop()
	
	for range ticker.C {
		lm.mu.Lock()
		
		now := time.Now()
		remainingQueue := make([]*QueueEntry, 0)
		
		for _, entry := range lm.localQueue {
			// Se passou mais de 20 segundos
			if now.Sub(entry.JoinedAt) > 20*time.Second {
				log.Printf("[LocalMatchmaking] Timeout para %s (20s na fila) - Movendo para fila global", 
					entry.Username)
				
				// Chama callback (move para fila global)
				leaderAddr := lm.raft.GetLeaderHTTPAddr() ;
				// so move se tiver um leder no cluster
				if (lm.onTimeout != nil &&  leaderAddr != "") {
					go lm.onTimeout(entry)
				}
			} else {
				// Mantém na fila local
				remainingQueue = append(remainingQueue, entry)
			}
		}
		
		lm.localQueue = remainingQueue
		lm.mu.Unlock()
	}
}


var OnLocalMatchFound func(p1, p2 *QueueEntry)

func (lm *LocalMatchmaking) onMatchFound(p1, p2 *QueueEntry) {
	if OnLocalMatchFound != nil {
		OnLocalMatchFound(p1, p2)
	}
}

