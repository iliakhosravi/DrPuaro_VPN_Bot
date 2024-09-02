package models

import (
	"fmt"

	"github.com/go-telegram/bot"
	ptime "github.com/yaa110/go-persian-calendar"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"techybat.org/go-vpn/panel"
)

type OrderType string

const (
	UndefinedOrder OrderType = "undefined"
	SentOrder      OrderType = "sent"
	PendingOrder   OrderType = "pending"
	ActiveOrder    OrderType = "active"
	DepletedOrder  OrderType = "depleted"
	CancelledOrder OrderType = "cancelled"
	DismissedOrder OrderType = "dismissed"
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

func (order *Order) AddReceipt(db *gorm.DB, msgID int) (*Receipt, error) {
	receipt := Receipt{
		OrderID:   order.ID,
		MessageID: msgID,
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
	var o Order
	db.Preload(clause.Associations).Find(&o, order.ID)
	order.Type = ActiveOrder
	order.AdminNote = adminNote

	var config Config = Config{}
	db.Where(&Config{OrderID: order.ID}).First(&config)
	config.OrderID = order.ID
	config.StartDate = ptime.Now().Time()

	if o.Pack.Type == SanaeiPack {
		p := panel.GetPanel()
		clientForm := panel.ClientForm{
			ID:         o.User.UUID,
			Email:      fmt.Sprintf("U%d O%d", order.UserID, order.ID),
			TotalGB:    int64(o.Pack.Traffic) * panel.ONE_GB,
			ExpiryTime: config.StartDate.AddDate(0, 0, o.Pack.Period).UnixMilli(),
			Enable:     true,
			TgID:       fmt.Sprint(o.User.TelID),
			SubID:      bot.RandomString(8),
		}

		if _, err := p.StoreClient(o.Pack.InboundID, clientForm); err != nil {
			return err
		}
		if _, err := p.ResetClientStats(o.Pack.InboundID, clientForm.Email); err != nil {
			return err
		}

		config.SubID = clientForm.SubID
		config.Email = clientForm.Email

	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if result := tx.Save(&config); result.RowsAffected == 0 {
			return fmt.Errorf("unable to update to verify order: %v", result.Error)
		}

		if result := tx.Save(order); result.RowsAffected == 0 {
			return fmt.Errorf("unable to update order to verify order: %v", result.Error)
		}
		return nil
	})

	return err
}

func (order *Order) Dismiss(db *gorm.DB, adminNote string) error {
	order.Type = DismissedOrder
	order.AdminNote = adminNote
	if result := db.Save(order); result.RowsAffected == 0 {
		return fmt.Errorf("unable to dismiss order: %v", result.Error)
	}
	return nil
}

func (order *Order) Deplete(db *gorm.DB) error {
	order.Type = DepletedOrder
	if result := db.Save(order); result.RowsAffected == 0 {
		return fmt.Errorf("unable to deplete order: %v", result.Error)
	}
	return nil
}

func (order *Order) ChangeType(db *gorm.DB, orderType OrderType) error {
	switch orderType {
	case order.Type:
		return nil
	case ActiveOrder:
		return order.Verify(db, order.AdminNote)
	case DepletedOrder:
		return order.Deplete(db)
	case DismissedOrder:
		return order.Dismiss(db, order.AdminNote)
	}
	return fmt.Errorf("error: order.ChangeType, order type is not supported")
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

func (order *Order) HasConfig(db *gorm.DB) bool {
	return order.Config(db).ID != 0
}

func (order *Order) Config(db *gorm.DB) *Config {
	var config Config
	if result := db.Where(&Config{OrderID: order.ID}).Preload(clause.Associations).Find(&config); result.RowsAffected == 0 {
		return &Config{}
	}
	config.Order = *order

	return &config
}

func (o *Order) UserStr(db *gorm.DB) string {
	return o.FullStr(db)
}

func (o *Order) FullStr(db *gorm.DB) string {
	var txtMsg string
	var order Order
	db.Preload("Pack").Preload("Pack.Category").Find(&order, o.ID)
	if order.Type == ActiveOrder {
		config := order.Config(db)
		dateFormat := "d MMM y"
		pt := ptime.New(config.StartDate)
		startDate := pt.Format(dateFormat)
		pt = ptime.New(order.CreatedAt)
		orderDate := pt.Format(dateFormat)
		endDate, _ := config.EndDate(db)
		endDateStr := endDate.Format(dateFormat)
		remainedTraffic, err := config.RemainedTraffic(db)
		remainedTrafficStr := fmt.Sprintf("%.2f", remainedTraffic)
		if err != nil {
			remainedTrafficStr = "N/A"
		}
		txtMsg = fmt.Sprintf("شماره سفارش: %d\nوضعیت سفارش: %s\nگروه بسته: %s\nنوع بسته: %s\nتاریخ درخواست: %s\nتوضیحات ادمین: %s\nتاریخ تایید بسته: %s\nحجم باقی مانده: %s GB\nتاریخ اتمام دوره:%s\nلینک بسته: %s\nلینک جیسون بسته: %s", order.ID, order.Type, order.Pack.Category.Name, order.Pack, orderDate, order.AdminNote, startDate, remainedTrafficStr, endDateStr, config.Link(db), config.JSONLink(db))
	} else {
		pt := ptime.New(order.CreatedAt)
		showDate := pt.Format("d MMM y")
		txtMsg = fmt.Sprintf("شماره سفارش: %d\nوضعیت سفارش: %s\nگروه بسته: %s\nنوع بسته: %s\nتاریخ درخواست: %s\nتوضیحات ادمین: %s", order.ID, order.Type, order.Pack.Category.Name, order.Pack, showDate, order.AdminNote)
	}

	return txtMsg
}
