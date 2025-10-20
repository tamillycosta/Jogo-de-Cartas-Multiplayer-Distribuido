package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/gameSession/local"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/session"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin:     func(r *http.Request) bool { return true },
}

//  Handler genérico que delega os Topicos da aplicação 
type PubSubHandler struct {
    broker   *pubsub.Broker
    handlers map[string]pubsub.HandleTopics
    sessionManager *session.SessionManager
    gameSessionManager *local.GameSessionManager
}

func New(broker *pubsub.Broker, sm *session.SessionManager,   gameSessionManager *local.GameSessionManager) *PubSubHandler {
    return &PubSubHandler{
        broker:   broker,
        handlers: make(map[string]pubsub.HandleTopics),
        sessionManager: sm,
        gameSessionManager: gameSessionManager,
    }
}

// Registra um handler para um prefixo de tópico
func (h *PubSubHandler) RegisterHandler(prefix string, handler pubsub.HandleTopics) {
    h.handlers[prefix] = handler
}


// Endpoint WebSocket
func (h *PubSubHandler) SetWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    
    clientID := uuid.New().String()
    
    log.Printf("Client conectado: %s (IP: %s)", clientID, r.RemoteAddr)
    
    // Envia ID
    conn.WriteJSON(map[string]interface{}{
        "type":      "connected",
        "client_id": clientID,
    })
    
    go h.handleClientMessages(clientID, conn)
}


func (h *PubSubHandler) handleClientMessages(clientID string, conn *websocket.Conn) {
    defer func() {
		log.Printf("🔌 [WebSocket] Cliente desconectando: %s", clientID)
		
		// Remove de partidas/fila PRIMEIRO
		if h.gameSessionManager != nil {
			h.gameSessionManager.HandleClientDisconnect(clientID)
		}
		
		//  Remove sessão de autenticação
		if h.sessionManager != nil {
			removed := h.sessionManager.RemoveSession(clientID)
			if removed {
				log.Printf("Sessão de auth removida: %s", clientID)
			}
		}
		
		//  Remove do broker (unsubscribe de todos os tópicos)
		if h.broker != nil {
			h.broker.RemoveClient(clientID)
			log.Printf("Removido de todos os tópicos: %s", clientID)
		}
		
		
		conn.Close()
		
		log.Printf("[WebSocket] Limpeza completa: %s", clientID)
	}()

    for {
        _, messageBytes, err := conn.ReadMessage()
        if err != nil {
            break
        }
        
        var msg map[string]interface{}
        if err := json.Unmarshal(messageBytes, &msg); err != nil {
            continue
        }
        
        msgType, _ := msg["type"].(string)
        
        switch msgType {
        case "subscribe":
            h.handleSubscribe(clientID, conn, msg)
            
        case "unsubscribe":
            h.handleUnsubscribe(clientID, msg)
            
        case "publish":
            h.handlePublish(clientID, msg)
        }
    }
}

// Lida com uma inscrição 
func (h *PubSubHandler) handleSubscribe(clientID string, conn *websocket.Conn, msg map[string]interface{}) {
    topic, _ := msg["topic"].(string)
    h.broker.Subscribe(clientID, topic, conn) // chama o broker
    
    conn.WriteJSON(map[string]interface{}{
        "type":  "subscribed",
        "topic": topic,
    })
}

// Lida com uma desiscrição 
func (h *PubSubHandler) handleUnsubscribe(clientID string, msg map[string]interface{}) {
    topic, _ := msg["topic"].(string)
    h.broker.Unsubscribe(clientID, topic) // chama o broker
}

func (h *PubSubHandler) handlePublish(clientID string, msg map[string]interface{}) {
    topic, _ := msg["topic"].(string)
    data := msg["data"]
    
    prefix := strings.Split(topic, ".")[0]
    
    // Busca handler para esse prefix
    handler, exists := h.handlers[prefix]
    if !exists {
        log.Printf(" Este Topico não existe: %s (topico: %s)", prefix, topic)
        return
    }
    
    // Delega para o handler específico
    if err := handler.HandleTopic(clientID, topic, data); err != nil {
        log.Printf("%v", err)
    }
}