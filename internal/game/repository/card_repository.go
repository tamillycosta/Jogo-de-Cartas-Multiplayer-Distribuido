package repository

import (
	
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"

	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CardRepository struct {
    db *gorm.DB
}


func NewCardRepository(db *gorm.DB)*CardRepository {
	return &CardRepository {db: db}
}


// Cria objeto package no banco 
func (r *CardRepository ) Create( cards *entities.Card) (*entities.Card, error){
	cards.ID = uuid.NewString()
	if err := r.db.Create(cards).Error; err != nil {
        return nil, err
    }
	return  cards, nil
}


// CreateWithID cria um pacote com ID específico (para sincronização Raft)
func (r *CardRepository ) CreateWithID( cards *entities.Card) (*entities.Card, error) {
	result := r.db.Create(cards)
	if result.Error != nil {
		return nil, result.Error
	}
	return cards, nil
}	


// retorna todos os cards (para snapshot)
func (r *CardRepository ) GetAll() ([]*entities.Card, error) {
	var cards []*entities.Card
	result := r.db.Find(&cards)
	if result.Error != nil {
		return nil, result.Error
	}
	return cards, nil
}


// deleta todos os cards (para restore snapshot)
func (r *CardRepository )DeleteAll() error {
	result := r.db.Exec("DELETE FROM cards")
	return  result.Error
}


func (r *CardRepository ) Delete(id string){
	r.db.Delete(id)
}


func (r *CardRepository ) FindById(id string)(*entities.Card, error){
	var cards entities.Card
	if err := r.db.Where("id = ?", id).First(&cards).Error; err != nil{
		if errors.Is(err, gorm.ErrRecordNotFound){
			return  nil, nil
		}
	return nil, err
	}
	return  &cards, nil 
}


// Modifica de quem é a carta, um pacote ou um player 
func (r *CardRepository) UpdateCardStatus(playerID, cardID string) error {
    return r.db.Model(&entities.Card{}).
        Where("id = ?", cardID).
        Updates(map[string]interface{}{
            "package_id": nil,
            "player_id":  playerID,
        }).Error
}


// FindByPlayerID retorna todas as cartas de um jogador
func (r *CardRepository) FindByPlayerID(playerID string) ([]*entities.Card, error) {
	var cards []*entities.Card
	
	err := r.db.Where("player_id = ?", playerID).Find(&cards).Error
	if err != nil {
		return nil, err
	}
	

	return cards, nil
}