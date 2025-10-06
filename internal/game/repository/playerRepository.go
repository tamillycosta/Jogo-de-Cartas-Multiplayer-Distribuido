package repository

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlayerRepository struct {
    db *gorm.DB
}


func New(db *gorm.DB)*PlayerRepository{
	return &PlayerRepository{db: db}
}


// Cria objeto player no banco 
func (r *PlayerRepository) Create(username string, ) (*entities.Player, error){
	player := &entities.Player{
		ID: uuid.NewString(),
		Username: username,
		Score: 0,
	}

	if err := r.db.Create(player).Error; err != nil {
        return nil, err
    }
	return  player, nil
}


// Busca usuário por username
func (r *PlayerRepository) FindByUsername(username string)(*entities.Player, error){
	var player entities.Player
	if err := r.db.Where("username = ?", username).First(&player).Error; err != nil{
		if errors.Is(err, gorm.ErrRecordNotFound){ // caso o objeto n exista no banco
			return  nil, nil
		}
	return nil, err // caso ocorra erro 
	}
	return  &player, nil 
}


//Verifica existência de username no banco
func (r *PlayerRepository) UsernameExists(username string) bool{
	var count int64
	r.db.Model(&entities.Player{}).Where("username = ?", username).Count(&count)
    return count > 0
}


// Modifica id do servidor do jogador
func (r *PlayerRepository) UpdateServerID(playerID, serverID string) error {
    return r.db.Model(&entities.Player{}).
        Where("id = ?", playerID).
        Update("server_id", serverID).Error
}


// Lista cartas do jogador 
func (r *PlayerRepository) GetPlayerCards(playerID string) ([]*entities.Card, error) {
    var cards []*entities.Card
    err := r.db.Where("player_id = ?", playerID).Find(&cards).Error
    return cards, err
}