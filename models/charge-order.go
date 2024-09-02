package models

import (
	"fmt"
	"strconv"

	ptime "github.com/yaa110/go-persian-calendar"
	"gorm.io/gorm"
)

type ChargeType string

const (
	UndefinedCharge ChargeType = "undefined"
	PendingCharge   ChargeType = "pending"
	AcceptedCharge  ChargeType = "accepted"
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

func (order *ChargeOrder) AcceptCharge(db *gorm.DB) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		var user User
		if result := tx.First(&user, order.UserID); result.Error != nil {
			return result.Error
		}
		user.Charge += uint64(order.Amount)
		if result := tx.Save(&user); result.Error != nil {
			return result.Error
		}
		order.Type = AcceptedCharge
		if result := tx.Save(order); result.Error != nil {
			return result.Error
		}
		return nil
	})
	fmt.Println("Error: Unable to accept charge. err: ", err)
	return err
}

func (order *ChargeOrder) DismissCharge(db *gorm.DB) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		order.Type = DismissedCharge
		if result := tx.Save(order); result.Error != nil {
			return result.Error
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error: Unable to dismiss charge. err: ", err)
	}
	return err
}

func (order *ChargeOrder) Receipt(db *gorm.DB) *ChargeReceipt {
	var receipt *ChargeReceipt = &ChargeReceipt{}
	db.Find(receipt, order.ID)
	return receipt
}

func (order *ChargeOrder) FullStr() string {
	dateFormat := "d MMM y"
	orderDate := ptime.New(order.CreatedAt).Format(dateFormat)
	updateOrderDate := ptime.New(order.UpdatedAt).Format(dateFormat)
	return fmt.Sprintf("میزان شارژ: %d تومان\nتاریخ درخواست: %s\nیادداشت ادمین: %s\nوضعیت: %s\nتاریخ آخرین تغییرات: %s", order.Amount, orderDate, order.AdminNote, order.Type, updateOrderDate)
}

func (chargeType ChargeType) String() string {
	switch chargeType {
	case PendingCharge:
		return "در انتظار تایید"
	case AcceptedCharge:
		return "تایید شده"
	case DismissedCharge:
		return "رد شده"
	default:
		return "نامشخص"
	}
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
