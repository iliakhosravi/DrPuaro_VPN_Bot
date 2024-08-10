package models

import (
	"gorm.io/gorm"
)

type Category struct {
	BaseModel
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (cat *Category) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Category{})
}
