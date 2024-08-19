package models

import (
	"fmt"

	tmodels "github.com/go-telegram/bot/models"
	ptime "github.com/yaa110/go-persian-calendar"
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
	AdminNote string    `json:"admin_note"`
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

func (order *Order) Verify(db *gorm.DB, adminNote string) error {
	order.Type = ActiveOrder
	order.AdminNote = adminNote
	var config Config = Config{}
	db.Where(&Config{OrderID: order.ID}).First(&config)
	config.OrderID = order.ID
	config.Link = adminNote
	config.StartDate = ptime.Now().Time()
	if result := db.Save(&config); result.RowsAffected == 0 {
		return fmt.Errorf("unable to update to verify order: %v", result.Error)
	}

	if result := db.Save(order); result.RowsAffected == 0 {
		return fmt.Errorf("unable to update order to verify order: %v", result.Error)
	}
	return nil
}

func (order *Order) Dismiss(db *gorm.DB, adminNote string) error {
	order.Type = DismissedOrder
	order.AdminNote = adminNote
	if result := db.Save(order); result.RowsAffected == 0 {
		return fmt.Errorf("unable to cancel order: %v", result.Error)
	}
	return nil
}

func (orderType OrderType) String() string {
	switch orderType {
	case SentOrder:
		return "در انتظار ارسال رسید توسط کاربر"
	case PendingOrder:
		return "در انتظار تایید"
	case ActiveOrder:
		return "فعال"
	case DepletedOrder:
		return "اتمام حجم یا دوره"
	case CancelledOrder:
		return "انصراف کاربر از تکمیل سفارش"
	case DismissedOrder:
		return "رد شده"
	}
	return "نامشخص"
}

func (order *Order) Config(db *gorm.DB) *Config {
	var config Config
	if result := db.Where(&Config{OrderID: order.ID}).Find(&config); result.RowsAffected == 0 {
		return &Config{}
	}

	return &config
}
