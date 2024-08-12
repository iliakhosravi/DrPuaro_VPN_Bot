package models

import "gorm.io/gorm"

type Receipt struct {
	BaseModel
	OrderID   uint `gorm:"uniqueIndex" json:"order_id"`
	Order     Order
	MessageID int `gorm:"uniqueIndex" json:"message_id"`
}

func (receipt *Receipt) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Receipt{})
}
