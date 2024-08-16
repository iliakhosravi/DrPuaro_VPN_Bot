package models

import (
	"fmt"

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

func (cat *Category) Store(db *gorm.DB) error {
	if result := db.Save(cat); result.RowsAffected == 0 {
		return fmt.Errorf("Error: unable to store category. Details: %v\n", result.Error)
	}
	return nil
}
