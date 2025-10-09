package handlers

import (
    "fmt"
    "log"
    
    "Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
    auth "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/authService"
    "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/protocol"
)

// AuthTopicHandler - Handler para tópicos de autenticação
type AuthTopicHandler struct {
    authService *auth.AuthService
    broker      *pubsub.Broker
}

func New(authService *auth.AuthService, broker *pubsub.Broker) *AuthTopicHandler {
    return &AuthTopicHandler{
        authService: authService,
        broker:      broker,
    }
}

// GetTopics - Retorna tópicos que este handler gerencia
func (h *AuthTopicHandler) GetTopics() []string {
    return []string{
        "auth.create_account",
        "auth.login",
        "auth.logout",
    }
}

// HandleTopic - Processa mensagens de auth
func (h *AuthTopicHandler) HandleTopic(clientID string, topic string, data interface{}) error {
    
    switch topic {
    case "auth.create_account":
        return h.handleCreateAccount(clientID, data)
        
    // Falta iplementar metodos handlers 
    //case "auth.login":
        //return h.handleLogin(clientID, data)
        
    //case "auth.logout":
        //return h.handleLogout(clientID, data)
        
    default:
        return fmt.Errorf("unknown auth topic: %s", topic)
    }
}

// Handler para topico de ciração de conta 
func (h *AuthTopicHandler) handleCreateAccount(clientID string, data interface{}) error {
    
    dataMap, ok := data.(map[string]interface{})
    if !ok {
        return fmt.Errorf("invalid data format")
    }
    
    username, _ := dataMap["username"].(string)
   
    
    log.Printf("📝 Creating account: %s", username)
    
    // Chama service do servidor
    err := h.authService.CreateAccount(username)
    
    // Monta resposta
    response := protocol.AuthResponse{
        Type:    "account_created",
        Success: err == nil,
    }
    
    if err != nil {
        response.Error = err.Error()
    } else {
        response.Message = "Account created successfully"
    }
    
    // Publica resposta
    h.publishResponse(clientID, response)
    
    return err
}



func (h *AuthTopicHandler) publishResponse(clientID string, response interface{}) {
    h.broker.Publish("auth.response."+clientID, map[string]interface{}{
        "topic": "auth.response",
        "data":  response,
    })
}