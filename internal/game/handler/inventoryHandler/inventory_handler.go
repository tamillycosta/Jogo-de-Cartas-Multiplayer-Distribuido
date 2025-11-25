package inventoryhandler

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	"fmt"
	"log"
)

type InventoryHandler struct {
	cardRepo *repository.CardRepository
	broker   *pubsub.Broker
}

func New(cardRepo *repository.CardRepository, broker *pubsub.Broker) *InventoryHandler {
	return &InventoryHandler{
		cardRepo: cardRepo,
		broker:   broker,
	}
}

func (h *InventoryHandler) HandleTopic(clientID string, topic string, data interface{}) error {
	switch topic {
	case "inventory.list":
		return h.handleListInventory(clientID, data)
	default:
		return fmt.Errorf("tópico desconhecido: %s", topic)
	}
}

func (h *InventoryHandler) handleListInventory(clientID string, data interface{}) error {
	// Extrair playerID do payload
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("formato de dados inválido")
	}
	
	playerID, _ := dataMap["player_id"].(string)
	if playerID == "" {
		return h.sendError(clientID, "player_id é obrigatório")
	}

	log.Printf("[Inventory] Listando cartas para %s", playerID)

	// Buscar cartas no repositório existente
	cards, err := h.cardRepo.FindByPlayerID(playerID)
	if err != nil {
		return h.sendError(clientID, "erro ao buscar cartas: "+err.Error())
	}

	// Montar resposta simplificada para o cliente
	var cardsResponse []map[string]interface{}
	for i, card := range cards {
		cardsResponse = append(cardsResponse, map[string]interface{}{
			"index":  i,
			"id":     card.ID,
			"name":   card.Name,
			"rarity": card.Rarity,
			"power":  card.Power,
			"health": card.Health,
		})
	}

	response := map[string]interface{}{
		"type":  "inventory_list",
		"cards": cardsResponse,
		"count": len(cards),
	}

	h.broker.Publish("response."+clientID, response)
	return nil
}

func (h *InventoryHandler) sendError(clientID, msg string) error {
	h.broker.Publish("response."+clientID, map[string]interface{}{
		"type":  "error",
		"error": msg,
	})
	return fmt.Errorf(msg)
}

func (h *InventoryHandler) GetTopics() []string {
	return []string{"inventory.list"}
}