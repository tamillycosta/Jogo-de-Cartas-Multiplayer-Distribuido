package session

import (
	gameEntities "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"sync"
)

// SessionManager gere as sessões ativas dos jogadores.
// Utiliza dois mapas para uma gestão eficiente.
type SessionManager struct {
	clientToPlayer map[string]string                 // Mapeia clientID -> playerID para remoção rápida
	playerSessions map[string]*gameEntities.Player // Mapeia playerID -> *Player para acesso rápido aos dados
	mu             sync.RWMutex
}

// New cria uma nova instância do SessionManager.
func New() *SessionManager {
	return &SessionManager{
		clientToPlayer: make(map[string]string),
		playerSessions: make(map[string]*gameEntities.Player),
	}
}

// CreateSession cria uma nova sessão para um jogador.
func (sm *SessionManager) CreateSession(clientID string, player *gameEntities.Player) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.clientToPlayer[clientID] = player.ID
	sm.playerSessions[player.ID] = player
}

// IsPlayerLoggedIn verifica se um jogador já possui uma sessão ativa.
func (sm *SessionManager) IsPlayerLoggedIn(playerID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	_, exists := sm.playerSessions[playerID]
	return exists
}

// IsClientLoggedIn verifica se um clientID já possui uma sessão ativa.
func (sm *SessionManager) IsClientLoggedIn(clientID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	_, exists := sm.clientToPlayer[clientID]
	return exists
}

// RemoveSession remove uma sessão com base no clientID, que é a única informação
// que temos quando um cliente se desconecta.
func (sm *SessionManager) RemoveSession(clientID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	playerID, exists := sm.clientToPlayer[clientID]
	if exists {
		// Remove dos dois mapas para manter a consistência
		delete(sm.playerSessions, playerID)
		delete(sm.clientToPlayer, clientID)
		return true // Sessão encontrada e removida
	}

	return false // Nenhuma sessão foi encontrada para este clientID
}

// IsPlayerLoggedInByUsername verifica se um jogador com um determinado username
// possui uma sessão ativa. É menos eficiente do que procurar por ID,
// mas necessário para as verificações entre servidores.
func (sm *SessionManager) IsPlayerLoggedInByUsername(username string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Itera sobre as sessões ativas para encontrar o username
	for _, player := range sm.playerSessions {
		if player.Username == username {
			return true // Encontrou uma sessão para este username
		}
	}
	return false
}