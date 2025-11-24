package usecases

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	shared "Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"log"
	"context"
	contracts "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/service"
	"fmt"
	"strings"
)

// Valida se as cartas de um jogador na blockchain correspondem as cartas no banco de dados
func  LoadPlayerCardsFromChain( chainService *contracts.ChainService,playerAddress string,playerID string, cardRepo *repository.CardRepository,) ([]*entities.Card, error) {
	if chainService == nil || chainService.CardChainService == nil {
		log.Printf("⚠️ Blockchain indisponível, carregando do banco...")
		return cardRepo.FindByPlayerID(playerID)
	}

	ctx := context.Background()

	tokenIDs, err := chainService.CardChainService.GetPlayerCards(ctx, playerAddress)
	if err != nil {
		log.Printf("Erro ao buscar cartas na blockchain: %v", err)
		return cardRepo.FindByPlayerID(playerID)
	}

	if len(tokenIDs) == 0 {
		return nil, fmt.Errorf("jogador não possui cartas na blockchain")
	}

	log.Printf("Jogador possui %d cartas na blockchain", len(tokenIDs))

	cards := make([]*entities.Card, 0, len(tokenIDs))

	for _, tokenID := range tokenIDs {
		cardMeta, err := chainService.CardChainService.GetCardMetadata(ctx, tokenID)
		if err != nil {
			log.Printf("Erro ao buscar metadados da carta token %d: %v", tokenID, err)
			continue
		}

		dbCard, err := cardRepo.FindById(cardMeta.CardID)
		if err != nil {
			log.Printf("Carta %s não encontrada no banco: %v", cardMeta.CardID, err)
			continue
		}

		if !strings.EqualFold(cardMeta.CurrentOwner, playerAddress) {
			log.Printf("Carta %s não pertence ao jogador!", cardMeta.CardID)
			continue
		}

		cards = append(cards, dbCard)
		log.Printf("Carta validada: %s (%s)", dbCard.Name, dbCard.TemplateID)
	}

	if len(cards) == 0 {
		return nil, fmt.Errorf("nenhuma carta válida encontrada")
	}

	return cards, nil
}



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