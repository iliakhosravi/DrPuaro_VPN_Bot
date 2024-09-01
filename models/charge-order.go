package models

import (
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

type ChargeType string

const (
	SentCharge      ChargeType = "sent"
	PendingCharge   ChargeType = "pending"
	AcceptedCharge  ChargeType = "accepted"
	CancelledCharge ChargeType = "cancelled"
	DismissedCharge ChargeType = "dismissed"
)

type ChargeOrder struct {
	BaseModel
	UserID    uint       `json:"user_id"`
	User      User       `json:"user"`
	Amount    uint       `json:"amount"`
	Type      ChargeType `json:"type"`
	AdminNote string     `json:"admin_note"`
}

func (order *ChargeOrder) Migrate(db *gorm.DB) {
	db.AutoMigrate(&ChargeOrder{})
}

func (order *ChargeOrder) AddReceipt(db *gorm.DB, msgID int) (*ChargeReceipt, error) {
	receipt := ChargeReceipt{
		ChargeOrderID: order.ID,
		MessageID:     msgID,
	}
	order.Type = PendingCharge
	if result := db.Save(order); result.RowsAffected == 0 {
		return nil, fmt.Errorf("unable to pend the charge-order: %v", result.Error)
	}

	if result := db.Create(&receipt); result.RowsAffected == 0 {
		return nil, fmt.Errorf("unable to create charge-receipt: %v", result.Error)
	}

	return &receipt, nil
}

func MoneyValidator(value string) (bool, string) {
	intValue, err := strconv.ParseUint(value, 0, 0)
	if err != nil {
		return false, "لطفا تنها ورودی عددی بزرگتر از 10000 وارد نمایید."
	}
	if intValue < 10000 {
		return false, "مقدار شارژ نمی تواند کمتر از 10,000 تومان باشد."
	}
	return true, ""
}
