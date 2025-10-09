package entities

import(
	"time"
)



type Package struct {
    ID        string   `gorm:"type:char(36);primaryKey" json:"id"`
    Cards     []*Card  `gorm:"foreignKey:PackageID" json:"cards,omitempty"`
    Status    string   `gorm:"size:20"`  // "available", "locked", "opened"
    LockedBy  *string  `gorm:"type:char(36);index" json:"locked_by,omitempty"` // PlayerID que bloqueou
    CreatedAt time.Time
    UpdatedAt time.Time
}