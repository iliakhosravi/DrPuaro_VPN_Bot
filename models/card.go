package models

import "gorm.io/gorm"

type Card struct {
	BaseModel
	Fullname string `json:"fullname"`
	Number   string `json:"number"`
}

func (card *Card) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Card{})
}
