package usecases

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"log"
	shared "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
)

// RESTAURA HP DAS CARTAS APOS UMA BATALHA
func RestoreCardsHp(player1, player2 *entities.GamePlayer) {
	log.Printf(" Restaurando HP das cartas...")
	
	if player1 != nil && len(player1.Deck) > 0 {
		for i := 0; i < len(player1.Deck) && i < 3; i++ {
			card := player1.Deck[i]
			if card == nil {
				continue
			}
			templateCard, exists := entities.BaseCards[card.TemplateID]
			
			// DEPOIS VERIFICAR PQ N TA RESTAURANDO 
			if !exists {
				log.Printf("Template não encontrado: %s", card.TemplateID)
				continue
			}
			
			player1.Deck[i].Health = templateCard.Health
		}
	}

	if player2 != nil && len(player2.Deck) > 0 {
		for i := 0; i < len( player2.Deck) && i < 3; i++ {
			card :=  player2.Deck[i]
			if card == nil {
				continue
			}
			templateCard, exists := entities.BaseCards[card.TemplateID]
			
			if !exists {
				log.Printf("   ⚠️ Template não encontrado: %s", card.TemplateID)
				continue
			}
			
			player2.Deck[i].Health = templateCard.Health
			
		}
	}
}


func GetServer(serverId string, servers []shared.ServerInfo) string{
	for _, server := range servers{
		if(server.ID == serverId){
			return  server.Address
		}
	}
	return  ""
}