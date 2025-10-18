package authhandler

import (
	"fmt"
	"log"

	auth "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/authService"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	authprotocol "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/protocol/authProtocol"
)

// Handler para tópicos de autenticação via Pub/Sub
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


//Processa mensagens recebidas via Pub/Sub
func (h *AuthTopicHandler) HandleTopic(clientID string, topic string, data interface{}) error {
	log.Printf("[AuthHandler] Topic: %s, Cliente: %s", topic, clientID)

	switch topic {
	case "auth.create_account":
		return h.handleCreateAccount(clientID, data)

	
	case "auth.login":
		return h.handleLogin(clientID, data)

	case "auth.logout":
		return h.handleLogout(clientID)

	default:
		return fmt.Errorf("topico n encontrado: %s", topic)
	}
}


// Handler para criação de conta
func (h *AuthTopicHandler) handleCreateAccount(clientID string, data interface{}) error {
	
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		h.publishErrorResponse(clientID, "formato de dados inválido")
		return fmt.Errorf("invalid data format")
	}

	username, ok := dataMap["username"].(string)
	if !ok || username == "" {
		h.publishErrorResponse(clientID, "username não fornecido")
		return fmt.Errorf("username not provided")
	}

	log.Printf("[AuthHandler] Cliente %s quer criar conta: %s", clientID, username)

	// Chama AuthService (METODO DO COM RAFT)
	err := h.authService.CreateAccount(username)

	
	response := authprotocol.AuthResponse{
		Type:    "account_created",
		Success: err == nil,
	}

	if err != nil {
		response.Error = err.Error()
		log.Printf("[AuthHandler] Erro ao criar conta '%s': %v", username, err)
	} else {
		response.Message = fmt.Sprintf("Conta '%s' criada com sucesso!", username)
		log.Printf("[AuthHandler] Conta '%s' criada com sucesso!", username)
	}

	
	h.publishResponse(clientID, response)

	return err
}


// Handler para login de conta
func (h *AuthTopicHandler) handleLogin(clientID string, data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		h.publishErrorResponse(clientID, "formato de dados inválido")
		return fmt.Errorf("invalid data format")
	}

	username, ok := dataMap["username"].(string)
	if !ok || username == "" {
		h.publishErrorResponse(clientID, "username não fornecido")
		return fmt.Errorf("username not provided")
	}

	log.Printf("[AuthHandler] Cliente %s quer fazer login: %s", clientID, username)

	// Chama AuthService
	player, err := h.authService.Login(username, clientID)

	response := authprotocol.AuthResponse{
		Type:    "login_response",
		Success: err == nil,
	}

	if err != nil {
		response.Error = err.Error()
		log.Printf("[AuthHandler] Erro no login de '%s': %v", username, err)
	} else {
		response.Message = fmt.Sprintf("Login de '%s' bem-sucedido!", username)
		response.Player = player
		log.Printf("[AuthHandler] Login de '%s' bem-sucedido!", username)
	}

	h.publishResponse(clientID, response)

	return err
}

func (h *AuthTopicHandler) handleLogout(clientID string) error {
	err := h.authService.Logout(clientID)

	response := authprotocol.AuthResponse{
		Type:    "logout_response",
		Success: err == nil,
	}

	if err != nil {
		response.Error = err.Error()
		log.Printf("[AuthHandler] Erro ao fazer logout para o cliente %s: %v", clientID, err)
	} else {
		response.Message = "Logout realizado com sucesso!"
		log.Printf("[AuthHandler] Logout bem-sucedido para o cliente %s", clientID)
	}

	h.publishResponse(clientID, response)
	return err
}



// ---------------------- AUXILIARES -----------------------------


// Envia resposta de sucesso para o cliente
func (h *AuthTopicHandler) publishResponse(clientID string, response interface{}) {
	responseTopic := fmt.Sprintf("auth.response.%s", clientID)
	
	h.broker.Publish(responseTopic, map[string]interface{}{
		"topic": "auth.response",
		"data":  response,
	})
	
	log.Printf("[AuthHandler] Resposta enviada para cliente %s", clientID)
}


// Envia resposta de erro para o cliente
func (h *AuthTopicHandler) publishErrorResponse(clientID string, errorMsg string) {
	response := authprotocol.AuthResponse{
		Type:    "account_created",
		Success: false,
		Error:   errorMsg,
	}
	
	h.publishResponse(clientID, response)
}


// retorna tópicos que este handler gerencia
func (h *AuthTopicHandler) GetTopics() []string {
	return []string{
		"auth.create_account",
		"auth.login",
		"auth.logout",
	}
}