package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	paym "github.com/sinasadeghi83/go-crypto-paywall/models"
	ptime "github.com/yaa110/go-persian-calendar"
	"gorm.io/gorm"
	"techybat.org/go-vpn/tools/qr"
	"techybat.org/go-vpn/vars"
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
	InvoiceID uint       `json:"invoice_id"`
	CoinID    uint       `json:"coin_id"`
	Amount    float32    `json:"amount"`
	Type      ChargeType `json:"type"`
	AdminNote string     `json:"admin_note"`
}

func (order *ChargeOrder) Migrate(db *gorm.DB) {
	db.AutoMigrate(&ChargeOrder{})
}

func (order *ChargeOrder) CreateInvoice(db *gorm.DB) (*paym.Invoice, error) {
	var coin paym.Coin
	if res := db.First(&coin, order.CoinID); res.Error != nil {
		return nil, res.Error
	}
	invoice := paym.Invoice{
		Price:        uint64(order.Amount * float32(coin.UnitFactor)),
		CoinID:       coin.ID,
		AcceptOthers: false,
		ExpiresAt:    time.Now().Add(15 * time.Minute),
	}

	if err := invoice.Create(db); err != nil {
		return nil, err
	}

	order.InvoiceID = invoice.ID

	return &invoice, db.Save(order).Error
}

func RetrieveInDollars(coin paym.Coin, amount float32) (float64, error) {
	url := fmt.Sprintf("https://api.coinpaprika.com/v1/coins/%s/markets?quotes=USD", coin.Name)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	data := struct {
		RAW struct {
			PRICE float64
		}
	}{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		return 0, err
	}

	return data.RAW.PRICE, nil
}

func (o *ChargeOrder) CryptoAddrMemo(db *gorm.DB) (string, string) {
	return paym.GetAddrMemoByInvoiceID(db, o.InvoiceID)
}

func (o *ChargeOrder) GetAmount(db *gorm.DB) float32 {
	return o.Amount
}

func (o *ChargeOrder) CryptoLink(db *gorm.DB) (string, string) {
	transferLink := paym.GetURLByInvoiceID(db, o.InvoiceID)
	qrPath := fmt.Sprintf("./%s/%d.jpg", vars.Get("QR_PATH"), o.InvoiceID)
	err := qr.GenerateQRLogo(transferLink, vars.Get("LOGO_PATH"), qrPath)
	if err != nil {
		fmt.Println("Unable to create QR Logo. err: ", err)
	}
	return transferLink, qrPath
}

func (order *ChargeOrder) IsPending(db *gorm.DB) bool {
	var invoice paym.Invoice
	db.Find(&invoice, order.InvoiceID)
	return order.Type == PendingCharge && invoice.ExpiresAt.After(time.Now())
}

func (order *ChargeOrder) PaymentType() PayType {
	return CryptoPay
}

func (order *ChargeOrder) CoinName(db *gorm.DB) string {
	var coin paym.Coin
	db.First(&coin, order.CoinID)
	return coin.Name
}

func (order *ChargeOrder) CoinUnit(db *gorm.DB) string {
	var coin paym.Coin
	db.First(&coin, order.CoinID)
	return coin.Unit
}

func (order *ChargeOrder) ProductName() string {
	return fmt.Sprintf("%g$", order.Amount)
}

func (order *ChargeOrder) AcceptCharge(db *gorm.DB) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		var coin paym.Coin
		if result := tx.First(&coin, order.CoinID); result.Error != nil {
			return result.Error
		}
		var user User
		if result := tx.First(&user, order.UserID); result.Error != nil {
			return result.Error
		}
		user.Charge += uint64(order.Amount * float32(coin.UnitFactor))
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
	return fmt.Sprintf("میزان شارژ: %g دلار\nتاریخ درخواست: %s\nیادداشت ادمین: %s\nوضعیت: %s\nتاریخ آخرین تغییرات: %s", order.Amount, orderDate, order.AdminNote, order.Type, updateOrderDate)
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

func DollarValidator(value string) (bool, string) {
	_, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return false, "لطفا تنها ورودی عددی وارد نمایید."
	}
	return true, ""
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
