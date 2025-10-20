package matchlocal


import (
	"log"
	"sync"
	"time"
)

// Gere fila de jogadores para encontrar match local 
type LocalMatchmaking struct {
	mu          sync.RWMutex
	localQueue  []*QueueEntry     
	serverID    string
}

type QueueEntry struct {
	ClientID  string  
	PlayerID  string
	Username  string
	JoinedAt  time.Time
}

func New(serverID string) *LocalMatchmaking {
	lm := &LocalMatchmaking{
		localQueue: make([]*QueueEntry, 0),
		serverID:   serverID,
	}
	
	go lm.matchmakingLoop()
	return lm
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

var OnLocalMatchFound func(p1, p2 *QueueEntry)

func (lm *LocalMatchmaking) onMatchFound(p1, p2 *QueueEntry) {
	if OnLocalMatchFound != nil {
		OnLocalMatchFound(p1, p2)
	}
}

