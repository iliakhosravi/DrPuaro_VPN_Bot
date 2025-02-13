package models

import "gorm.io/gorm"

type Currency struct {
	BaseModel
	Name       string `json:"name"`
	Unit       string `json:"unit"`
	UnitFactor uint   `json:"unit_factor"`
}

func (cur *Currency) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Currency{})
}
