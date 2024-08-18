package models

import (
	"fmt"
	"time"

	tmodels "github.com/go-telegram/bot/models"
	"gorm.io/gorm"
)

type OrderType int

const (
	UndefinedOrder OrderType = iota
	SentOrder
	PendingOrder
	ActiveOrder
	DepletedOrder
	CancelledOrder
	DismissedOrder
)

type Order struct {
	BaseModel
	UserID    uint      `json:"user_id"`
	User      User      `json:"user"`
	PackID    uint      `json:"pack_id"`
	Pack      Pack      `json:"pack"`
	Type      OrderType `json:"type"`
	HandlerID string    `json:"-"`
}

func (order *Order) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Order{})
}

func (order *Order) CreateOrder(db *gorm.DB) error {
	result := db.Create(order)
	if result.RowsAffected == 0 {
		return fmt.Errorf("unable to create Order: %v", result.Error)
	}
	return nil
}

func (order *Order) AddReceiptByMsg(db *gorm.DB, msg *tmodels.Message) (*Receipt, error) {
	receipt := Receipt{
		OrderID:   order.ID,
		MessageID: msg.ID,
	}
	order.Type = PendingOrder
	if result := db.Save(order); result.RowsAffected == 0 {
		return nil, fmt.Errorf("unable to pend the order: %v", result.Error)
	}

	if result := db.Create(&receipt); result.RowsAffected == 0 {
		return nil, fmt.Errorf("unable to create receipt: %v", result.Error)
	}

	return &receipt, nil
}

func (order *Order) CancelOrder(db *gorm.DB) error {
	order.Type = CancelledOrder
	if result := db.Save(order); result.RowsAffected == 0 {
		return fmt.Errorf("unable to cancel order: %v", result.Error)
	}
	return nil
}

func (order *Order) GetReceipt(db *gorm.DB) Receipt {
	var receipt Receipt
	db.Where(&Receipt{OrderID: order.ID}).First(&receipt)
	return receipt
}

func (order *Order) Verify(db *gorm.DB, link string) error {
	order.Type = ActiveOrder
	var config Config = Config{}
	db.Where(&Config{OrderID: order.ID}).First(&config)
	config.OrderID = order.ID
	config.Link = link
	config.StartDate = time.Now()
	if result := db.Save(&config); result.RowsAffected == 0 {
		return fmt.Errorf("unable to update to verify order: %v", result.Error)
	}

	if result := db.Save(order); result.RowsAffected == 0 {
		return fmt.Errorf("unable to update order to verify order: %v", result.Error)
	}
	return nil
}

func (order *Order) Dismiss(db *gorm.DB) error {
	order.Type = DismissedOrder
	if result := db.Save(order); result.RowsAffected == 0 {
		return fmt.Errorf("unable to cancel order: %v", result.Error)
	}
	return nil
}
