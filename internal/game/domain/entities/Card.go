package entities

import (
    "time"
)


type Card struct {
    ID        string  `gorm:"type:char(36);primaryKey" json:"id"`
    Name      string  `gorm:"size:100;not null" json:"name"`
    Type      string  `gorm:"size:50" json:"type"`
    Power     int     `gorm:"default:0" json:"power"`
    Rarity    string  `gorm:"size:20" json:"rarity"`
    Health   int
    // Relações (somente uma pode ser != nil)
    PlayerID  *string `gorm:"type:char(36);index" json:"player_id,omitempty"`
    PackageID *string `gorm:"type:char(36);index" json:"package_id,omitempty"`
    TemplateID string 
    CreatedAt time.Time `json:"created_at"`
    InDeck   bool   `gorm:"default:false" json:"inDeck"` 
    MaxCopies int 
    IsSpecial bool
}