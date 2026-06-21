package models

import (
	"fmt"

	ptime "github.com/yaa110/go-persian-calendar"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	PendLinkOrder  OrderType = "pend-link"
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

func (order *Order) Verify(db *gorm.DB, adminNote, customLink string) error {
	var o Order
	db.Preload(clause.Associations).Find(&o, order.ID)
	order.Type = ActiveOrder
	order.AdminNote = adminNote

	config := o.Config(db)
	config.OrderID = order.ID
	config.StartDate = ptime.Now().Time()
	config.CustomLink = customLink

	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		if o.Pack.Type == SanaeiPack {
			err = config.SetupSanaei(tx, o)
		}

		if o.Pack.Type == SUIPack {
			err = config.SetupSUI(tx, o)
		}

		if o.Pack.Type == MarzPack {
			err = config.SetupMarz(tx, o)
		}

		if err != nil {
			return fmt.Errorf("unable to setup with panel: %w", err)
		}

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
	var o Order
	db.Preload(clause.Associations).Find(&o, order.ID)
	order.Type = DepletedOrder

	err := db.Transaction(func(tx *gorm.DB) error {
		if o.Pack.Type == SanaeiPack {
			o.Config(tx).DepleteSanaei(tx, o)
		}

		if o.Pack.Type == SUIPack {
			o.Config(tx).DepleteSUI(tx, o)
		}

		if o.Pack.Type == MarzPack {
			o.Config(tx).DepleteMarz(tx, o)
		}
		if result := db.Save(order); result.RowsAffected == 0 {
			return fmt.Errorf("unable to deplete order: %v", result.Error)
		}
		return nil
	})
	return err
}

func (order *Order) ChangeType(db *gorm.DB, orderType OrderType) error {
	switch orderType {
	case order.Type:
		return nil
	case ActiveOrder:
		return order.Verify(db, order.AdminNote, order.Config(db).CustomLink)
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
	case PendLinkOrder:
		return "در انتظار ثبت لینک کانفیگ"
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

func (config Config) TrafficString(db *gorm.DB) (string, error) {
	gb, mb, err := config.TrafficGbMb(db)
	if err != nil {
		return "", err
	}
	result := ""
	if gb != 0 {
		result += fmt.Sprintf("%d گیگابایت", gb)
	}

	if mb != 0 {
		if gb != 0 {
			result += " و "
		}
		result += fmt.Sprintf("%d مگابایت", mb)
	}

	return result, nil
}

func (config Config) TrafficName(db *gorm.DB) (string, error) {
	gb, mb, err := config.TrafficGbMb(db)
	if err != nil {
		return "", err
	}
	result := ""
	if gb != 0 {
		result += fmt.Sprintf("%d GB", gb)
	}

	if mb != 0 {
		if gb != 0 {
			result += " "
		}
		result += fmt.Sprintf("%d MB", mb)
	}

	return result, nil
}

func (config Config) TrafficGbMb(db *gorm.DB) (int, int, error) {
	traffic, err := config.RemainedTraffic(db)
	sign := 1
	if traffic < 0 {
		sign = -1
		traffic *= -1
	}
	gb := traffic / 1024
	mb := traffic % 1024
	return sign * gb, sign * mb, err
}

func (o *Order) UserStr(db *gorm.DB) string {
	return o.NormalStr(db)
}

func (order Order) Name(db *gorm.DB) string {
	var o Order
	db.Preload(clause.Associations).Preload("Pack.Category").Find(&o, order.ID)
	return fmt.Sprintf("%d | %s", o.ID, o.Pack.String())
}

func (o *Order) NormalStr(db *gorm.DB) string {
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
		if endDate.Unix() == 0 {
			endDateStr = "نامحدود"
		}
		remainedTrafficStr, err := config.TrafficString(db)
		if err != nil {
			remainedTrafficStr = "N/A"
		}
		txtMsg = fmt.Sprintf("شماره سفارش: %d\nشماره کانفیگ:%d\nوضعیت سفارش: %s\nگروه بسته: %s\nنوع بسته: %s\nتاریخ درخواست: %s\nتوضیحات ادمین: %s\nتاریخ تایید بسته: %s\nحجم باقی مانده: %s\nتاریخ اتمام دوره:%s ", order.ID, config.ID, order.Type, order.Pack.Category.Name, order.Pack, orderDate, order.AdminNote, startDate, remainedTrafficStr, endDateStr)
	} else {
		pt := ptime.New(order.CreatedAt)
		showDate := pt.Format("d MMM y")
		txtMsg = fmt.Sprintf("شماره سفارش: %d\nوضعیت سفارش: %s\nگروه بسته: %s\nنوع بسته: %s\nتاریخ درخواست: %s\nتوضیحات ادمین: %s", order.ID, order.Type, order.Pack.Category.Name, order.Pack, showDate, order.AdminNote)
	}

	return txtMsg
}

func (o *Order) FullStr(db *gorm.DB) string {
	txtMsg := o.NormalStr(db)
	if o.Type == ActiveOrder {
		txtMsg += fmt.Sprintf("\nلینک ساب بسته: %s\nلینک ساب جیسون بسته: %s", o.Config(db).Link(db), o.Config(db).JSONLink(db))
	}
	return txtMsg
}
