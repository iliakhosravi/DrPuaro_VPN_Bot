package models

import (
	"time"

	"gorm.io/gorm"
)

type Config struct {
	BaseModel
	Link      string    `json:"link"`
	StartDate time.Time `json:"start_date"`
	OrderID   uint      `json:"order_id"`
	Order     Order     `json:"order"`
}

func (config *Config) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Config{})
}
