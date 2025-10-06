package entities


type Package struct{
	Id  string   `gorm:"type:char(36);primaryKey" json:"id"`
	Cards       []*Card   `gorm:"foreignKey:PackgeID" json:"cards,omitempty"`
	Status    string `gorm:"size:20"` // available, locked, opened
	LockedBy  string `gorm:"size:50"` 
}