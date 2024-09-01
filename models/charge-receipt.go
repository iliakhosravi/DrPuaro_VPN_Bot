package models

import "gorm.io/gorm"

type ChargeReceipt struct {
	BaseModel
	ChargeOrderID uint `gorm:"uniqueIndex" json:"charge_order_id"`
	ChargeOrder   ChargeOrder
	MessageID     int `gorm:"uniqueIndex" json:"message_id"`
}

func (receipt *ChargeReceipt) Migrate(db *gorm.DB) {
	db.AutoMigrate(&ChargeReceipt{})
}
