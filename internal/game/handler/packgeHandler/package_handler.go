package packgehandler

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/packageService"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	packageprotocol "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/protocol/packageProtocol"
	"fmt"
	"log"
)

// Handler para tópicos de abertura de pacotes via Pub/Sub
type PackageTopicHandler struct {
	packageService *packageService.PackageService
	broker      *pubsub.Broker
}

func New(packageService *packageService.PackageService, broker *pubsub.Broker) *PackageTopicHandler {
	return &PackageTopicHandler{
		packageService: packageService,
		broker:      broker,
	}
}


func (p *PackageTopicHandler) HandleTopic(clientID string, topic string, data interface{}) error {
	log.Printf("[PackageHandler] Topic: %s, Cliente: %s", topic, clientID)

	switch topic {
	case "package.open_pack":
		return p.handleOpenPackage(clientID, data)

	// num sei se vão ter mais topicos kkkk

	default:
		return fmt.Errorf("topico n encontrado: %s", topic)
	}
}


// HANDLER DO PUB SUB PARA ABERTURA DOS PACOTES 
func (p *PackageTopicHandler) handleOpenPackage(clientID string, data interface{}) error {

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		p.publishErrorResponse(clientID, "formato de dados inválido")
		return fmt.Errorf("invalid data format")
	}

	
	playerID, _ := dataMap["player_id"].(string)
	if !ok || playerID == "" {
		p.publishErrorResponse(playerID, "username não fornecido")
		return fmt.Errorf("username not provided")
	}

	err := p.packageService.OpenPackage(playerID)
	log.Printf("[AuthHandler] Cliente %s quer criar conta: ", clientID, )

	response := packageprotocol.OpenPackageResponse{
		Type:    "package_opend",
		Success: err == nil,
	}

	if err != nil {
		response.Error = err.Error()
		
	} else {
		response.Message = "Pacote aberto com sucesso!"
		
	}

	p.publishResponse(clientID, response)

	return err

}



// ---------------------- AUXILIARES -----------------------------


// Envia resposta de sucesso para o cliente (REPLY)
func (p *PackageTopicHandler) publishResponse(clientID string, response interface{}) {
	responseTopic := fmt.Sprintf("package.response.%s", clientID)
	
	p.broker.Publish(responseTopic, map[string]interface{}{
		"topic": "package.response",
		"data":  response,
	})
	
	log.Printf("[PackageHandler] Resposta enviada para cliente %s", clientID)
}


// Envia resposta de erro para o cliente
func (p *PackageTopicHandler) publishErrorResponse(clientID string, errorMsg string) {
	response := packageprotocol.OpenPackageResponse{
		Type:    "package.open_pack",
		Success: false,
		Error:   errorMsg,
	}
	
	p.publishResponse(clientID, response)
}


func (p *PackageTopicHandler) GetTopics() []string {
	return []string{
		"package.open_pack",
	}
}
