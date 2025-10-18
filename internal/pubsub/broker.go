
package pubsub

import (
    "encoding/json"
    "log"
    "sync"
    "github.com/gorilla/websocket"
)

type Broker struct {
    topics      map[string]map[string]*websocket.Conn // topico -> clientID -> conexão WebSockt
    mu          sync.RWMutex
    connections map[string]*websocket.Conn             // clientID -> conn
}

func New() *Broker {
    return &Broker{
        topics:      make(map[string]map[string]*websocket.Conn),
        connections: make(map[string]*websocket.Conn),
    }
}

//  Cliente se inscreve em um tópico
func (b *Broker) Subscribe(clientID, topic string, conn *websocket.Conn) {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    if b.topics[topic] == nil {
        b.topics[topic] = make(map[string]*websocket.Conn)
    }
    
    b.topics[topic][clientID] = conn
    b.connections[clientID] = conn
    
    log.Printf("Client %s se inscreveu no topic: %s", clientID, topic)
}


// Cliente cancela inscrição
func (b *Broker) Unsubscribe(clientID, topic string) {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    if b.topics[topic] != nil {
        delete(b.topics[topic], clientID)
    }
    
    log.Printf("Client %s se desiscrevreu dos topicos: %s", clientID, topic)
}


// Publica mensagem em um tópico
func (b *Broker) Publish(topic string, message interface{}) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    subscribers := b.topics[topic]
    if subscribers == nil {
        return
    }
    
    data, err := json.Marshal(message)
    if err != nil {
        return
    }
    
    // Envia para todos os inscritos no tópico
    for clientID, conn := range subscribers {
        if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
            log.Printf("erro ao enviar aos clientes %s: %v", clientID, err)
            delete(subscribers, clientID)
        }
    }
    
   
}


//  Remove cliente de todos os tópicos
func (b *Broker) RemoveClient(clientID string) {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    delete(b.connections, clientID)
    
    for _, subscribers := range b.topics {
        delete(subscribers, clientID)
    }
    
    log.Printf("Client %s removido dos topicos", clientID)
}