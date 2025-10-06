package entities

import (
    "time"
)


type Player struct {
    ID          string    `gorm:"type:char(36);primaryKey" json:"id"`
    Username    string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
    Score       int       `gorm:"default:0" json:"score"`
    ServerID    string    `gorm:"size:50" json:"server_id"` // Servidor onde está logado
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    
    // Relacionamentos (carregados quando necessário)
    Cards       []*Card   `gorm:"foreignKey:PlayerID" json:"cards,omitempty"`
    
    // Runtime apenas (não persistido)
    MatchID     *string   `gorm:"-" json:"match_id,omitempty"`
    BattleDeck  []*Card   `gorm:"-" json:"battle_deck,omitempty"`
    CurrentCard *Card     `gorm:"-" json:"current_card,omitempty"`
}