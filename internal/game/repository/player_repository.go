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


func NewPlayerRepository(db *gorm.DB)*PlayerRepository{
	return &PlayerRepository{db: db}
}


// Cria objeto player no banco 
func (r *PlayerRepository) Create( player *entities.Player) (*entities.Player, error){
	player.ID = uuid.NewString()
	if err := r.db.Create(player).Error; err != nil {
        return nil, err
    }
	return  player, nil
}


// CreateWithID cria um jogador com ID específico (para sincronização Raft)
func (r *PlayerRepository) CreateWithID( player *entities.Player) (*entities.Player, error) {
	result := r.db.Create(player)
	if result.Error != nil {
		return nil, result.Error
	}

	return player, nil
}	


// retorna todos os players (para snapshot)
func (r *PlayerRepository) GetAll() ([]*entities.Player, error) {
	var players []*entities.Player
	result := r.db.Find(&players)
	if result.Error != nil {
		return nil, result.Error
	}
	return players, nil
}


// deleta todos os players (para restore snapshot)
func (r *PlayerRepository)DeleteAll() error {
	result := r.db.Exec("DELETE FROM players")
	return  result.Error
}



func (r *PlayerRepository) Delete(id string){
	r.db.Delete(id)
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


// Modifica id do servidor do jogador (USAR NO LOGIN)
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