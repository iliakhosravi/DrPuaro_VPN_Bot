package models

import "gorm.io/gorm"

type InlineKeyboard struct {
	BaseModel
	Name string
	Text string
}

func (kb *InlineKeyboard) Migrate(db *gorm.DB) {
	db.AutoMigrate(&InlineKeyboard{})
}

func (kb *InlineKeyboard) Guides(db *gorm.DB) []Guide {
	var guides []Guide
	db.Where("inline_keyboard_id = ?", kb.ID).Find(&guides)
	return guides
}
