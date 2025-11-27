package inventoryhandler

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub"
	"fmt"
	"log"
)

type InventoryHandler struct {
	cardRepo *repository.CardRepository
	playerRepo *repository.PlayerRepository
	broker   *pubsub.Broker
}

func New(cardRepo *repository.CardRepository, playerRepo *repository.PlayerRepository, broker *pubsub.Broker) *InventoryHandler {
	return &InventoryHandler{
		cardRepo:   cardRepo,
		playerRepo: playerRepo,
		broker:     broker,
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
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("formato de dados inválido")
	}

	// Tenta ler "target_username" do payload
	targetUsername, _ := dataMap["target_username"].(string)
	playerID, _ := dataMap["player_id"].(string)

	// Se um username alvo foi fornecido, busque o ID dele
	if targetUsername != "" {
		targetPlayer, err := h.playerRepo.FindByUsername(targetUsername)
		if err != nil || targetPlayer == nil {
			return h.sendError(clientID, fmt.Sprintf("Jogador '%s' não encontrado", targetUsername))
		}
		playerID = targetPlayer.ID
		log.Printf("[Inventory] Listando inventário de %s (solicitado por %s)", targetUsername, clientID)
	} else if playerID == "" {
		return h.sendError(clientID, "player_id ou target_username é obrigatório")
	}

	// Buscar cartas (o resto da função permanece igual)
	cards, err := h.cardRepo.FindByPlayerID(playerID)
	if err != nil {
		return h.sendError(clientID, "erro ao buscar cartas: "+err.Error())
	}

	// Montar resposta simplificada
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
		"type":            "inventory_list",
		"cards":           cardsResponse,
		"count":           len(cards),
		"target_username": targetUsername, // Útil para o cliente saber de quem é a lista
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