package entities



// conta quantas cartas especiais foram distribuidas
var SpecialCardCount = make(map[string]int)
var CardVersions = make(map[string]int)

type CardRarity string

type CardTemplate struct {
	TemplateID string
	Nome       string
	Type       string
	Power      int
	Health     int
	Rarity     string
	IsSpecial  bool
	MaxCopies  int
}


const (
	COMMON    CardRarity = "COMMON"
	UNCOMMON  CardRarity = "UNCOMMON"
	RARE      CardRarity = "RARE"
	EPIC      CardRarity = "EPIC"
	LEGENDARY CardRarity = "LEGENDARY"
)

// Cartas Template (estoque global)

var BaseCards = map[string]CardTemplate{

	"starter_mage": {
		TemplateID: "starter_mage",
		Nome:       "Aprendiz Mago",
		Power:      100,
		Health:     100,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_goblin": {
		TemplateID: "starter_goblin",
		Nome:       "Goblin Com Bomba",
		Power:      100,
		Health:     140,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_witch": {
		TemplateID: "starter_witch",
		Nome:       "Bruxa",
		Power:      150,
		Health:     120,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_wolf": {
		TemplateID: "starter_wolf",
		Nome:       "Lobo",
		Power:      100,
		Health:     90,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_fire": {
		TemplateID: "starter_fire",
		Nome:       "Feiticeira de Fogo",
		Power:      100,
		Health:     150,
		Rarity:     string(COMMON),
	},
	"starter_knight": {
		TemplateID: "starter_knight",
		Nome:       "Escudeiro",
		Power:      70,
		Health:     100,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_raven": {
		TemplateID: "starter_raven",
		Nome:       "Corvo Místico",
		Power:      100,
		Health:     95,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_devil": {
		TemplateID: "starter_devil",
		Nome:       "Cavaleiro das Trevas",
		Power:      120,
		Health:     110,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_elf": {
		TemplateID: "starter_elf",
		Nome:       "Elfo Caçador",
		Power:      90,
		Health:     100,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_dragon": {
		TemplateID: "starter_dragon",
		Nome:       "Dragão Comum",
		Power:      50,
		Health:     100,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},

	// Cartas especiais raras
	"legend_dragon": {
		TemplateID: "legend_dragon",
		Nome:       "Dragão Ancião",
		Power:      350,
		Health:     300,
		Rarity:     string(LEGENDARY),
		IsSpecial:  true,
		MaxCopies:  100,
	},
	"legend_archmage": {
		TemplateID: "legend_archmage",
		Nome:       "Arquimago Supremo",
		Power:      230,
		Health:     280,
		Rarity:     string(LEGENDARY),
		IsSpecial:  true,
		MaxCopies:  100,
	},
	"epic_shadow_witch": {
		TemplateID: "epic_shadow_witch",
		Nome:       "Bruxa das Sombras",
		Power:      200,
		Health:     200,
		Rarity:     string(EPIC),

		IsSpecial: true,
		MaxCopies: 200,
	},

	"epic_phoenix": {
		TemplateID: "epic_phoenix",
		Nome:       "Fênix Dourada",
		Power:      170,
		Health:     200,
		Rarity:     string(EPIC),
		IsSpecial:  true,
		MaxCopies:  200,
	},

	"rare_best": {
		TemplateID: "rare_best",
		Nome:       "Besta Sombria",
		Power:      180,
		Health:     170,
		Rarity:     string(RARE),
		IsSpecial:  true,
		MaxCopies:  200,
	},

	"uncumon_bow": {
		TemplateID: "uncumon_bow",
		Nome:       "Arqueiro Fantasma",
		Power:      150,
		Health:     170,
		Rarity:     string(UNCOMMON),
		IsSpecial:  true,
		MaxCopies:  200,
	},
}
var StarterCardIDs = []string{
	"starter_mage", "starter_goblin", "starter_witch", "starter_wolf",
	"starter_fire", "starter_knight", "starter_raven", "starter_devil",
	"starter_elf", "starter_dragon",
}

