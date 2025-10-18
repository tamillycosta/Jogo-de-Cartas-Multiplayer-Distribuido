package usecases

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	
	"time"
	
	"math/rand"
		
	"errors"
	
)
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// ----------------- Lógica de Geração de Cartas

// gera IDs de templates aleatórios
func GenerateRandomCards(count int) []string {
	templates := make([]string, count)
	
	for i := 0; i < count; i++ {
		
		rarity := rollRarity()
		
		
		template := getRandomTemplateByRarity(rarity)
		templates[i] = template
	}
	
	return templates
}

// sorteia a raridade baseado em probabilidades
func rollRarity() string {
	roll := rand.Intn(100)
	
	switch {
	case roll < 50: // 50% comum
		return string(entities.COMMON)
	case roll < 75: // 25% incomum
		return string(entities.UNCOMMON)
	case roll < 90: // 15% rara
		return string(entities.RARE)
	case roll < 97: // 7% épica
		return string(entities.EPIC)
	default: // 3% lendária
		return string(entities.LEGENDARY)
	}
}

//busca um template aleatório da raridade
func getRandomTemplateByRarity(rarity string) string {
	
	var matching []string
	for id, card := range entities.BaseCards {
		if card.Rarity == rarity {
			matching = append(matching, id)
		}
	}
	
	if len(matching) == 0 {
		
		return entities.StarterCardIDs[rand.Intn(len(entities.StarterCardIDs))]
	}
	
	return matching[rand.Intn(len(matching))]
}


// ---------------------------- Métodos de Consulta 


// seleciona um pacote disponível aleatoriamente
func SelectAvailablePackage(packages []*entities.Package) (*entities.Package, error) {
	if len(packages) == 0 {
		return nil, errors.New("nenhum pacote fornecido")
	}

	// Filtra apenas pacotes disponíveis
	var available []*entities.Package
	for _, pkg := range packages {
		if pkg.Status == "avalible"    {
			available = append(available, pkg)
		}
	}

	if len(available) == 0 {
		return nil, errors.New("nenhum pacote disponível")
	}

	randomIndex := rand.Intn(len(available))
	return available[randomIndex], nil
}