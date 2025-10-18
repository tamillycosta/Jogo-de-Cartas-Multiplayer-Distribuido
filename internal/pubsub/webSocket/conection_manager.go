package websocket

import (
    "sync"
    "github.com/gorilla/websocket"
)

//  Gerencia conexões WebSocket ativas
type ConnectionManager struct {
    connections map[string]*websocket.Conn
    mu          sync.RWMutex
}

func NewConnectionManager() *ConnectionManager {
    return &ConnectionManager{
        connections: make(map[string]*websocket.Conn),
    }
}

func (cm *ConnectionManager) Add(clientID string, conn *websocket.Conn) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    cm.connections[clientID] = conn
}

func (cm *ConnectionManager) Remove(clientID string) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    delete(cm.connections, clientID)
}

func (cm *ConnectionManager) Get(clientID string) (*websocket.Conn, bool) {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    conn, exists := cm.connections[clientID]
    return conn, exists
}

func (cm *ConnectionManager) Count() int {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    return len(cm.connections)
}