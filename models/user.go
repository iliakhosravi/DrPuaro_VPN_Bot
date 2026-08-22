package models

import (
	"fmt"
	"strconv"

	tmodels "github.com/go-telegram/bot/models"
	"gorm.io/gorm"
	"techybat.org/go-vpn/vars"
	"techybat.org/go-vpn/widgets/form"
)

type UserType string

const (
	NoramlUser  UserType = "normal"
	TrustedUser UserType = "trusted"
	AdminUser   UserType = "admin"
)

type User struct {
	BaseModel
	TelID     int64
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Username  string   `json:"username"`
	Type      UserType `json:"user_type" gorm:"default:normal"`
	Charge    uint64   `json:"charge"`
}

func (user *User) Migrate(db *gorm.DB) {
	db.AutoMigrate(&User{})
}

func (user *User) CreateOrFindUserByTelegram(db *gorm.DB, tuser *tmodels.User) error {
	res := db.First(user, &User{TelID: tuser.ID})
	user.FirstName = tuser.FirstName
	user.LastName = tuser.LastName
	user.Username = tuser.Username
	user.TelID = tuser.ID
	if res.RowsAffected == 0 {
		user.Charge, _ = strconv.ParseUint(vars.Get("WALLET_INIT_BALANCE"), 0, 0)
	}

	if result := db.Save(user); result.RowsAffected == 0 {
		return fmt.Errorf("unable to create user: %v", result.Error)
	}
	return nil
}

func (user *User) RevivePackByCard(db *gorm.DB, pack *Pack, msgID int, configID any) (*Order, error) {
	var order Order
	err := db.Transaction(func(tx *gorm.DB) error {
		order = Order{
			UserID: user.ID,
			PackID: pack.ID,
			Type:   PendingOrder,
		}

		if err := order.CreateOrder(tx); err != nil {
			return err
		}

		var config Config
		if res := tx.Find(&config, configID); res.Error != nil {
			return res.Error
		}

		config.OrderID = order.ID
		if res := tx.Save(&config); res.Error != nil {
			return res.Error
		}

		if _, err := order.AddReceipt(tx, msgID); err != nil {
			return err
		}

		return nil
	})

	return &order, err
}

func (user *User) BuyPackByCard(db *gorm.DB, pack *Pack, msgID int) (*Order, error) {
	var order Order
	err := db.Transaction(func(tx *gorm.DB) error {
		order = Order{
			UserID: user.ID,
			PackID: pack.ID,
			Type:   PendingOrder,
		}

		if err := order.CreateOrder(tx); err != nil {
			return err
		}

		if _, err := order.AddReceipt(tx, msgID); err != nil {
			return err
		}

		return nil
	})

	return &order, err
}

func (user *User) RevivePackByCrypto(db *gorm.DB, pack *Pack, configID any) (*Order, error) {
	var order Order
	err := db.Transaction(func(tx *gorm.DB) error {
		order = Order{
			UserID:  user.ID,
			PackID:  pack.ID,
			Type:    PendingOrder,
			PayType: CryptoPay,
		}

		if err := order.CreateOrder(tx); err != nil {
			return err
		}

		var config Config
		if res := tx.Find(&config, configID); res.Error != nil {
			return res.Error
		}

		config.OrderID = order.ID
		if res := tx.Save(&config); res.Error != nil {
			return res.Error
		}

		return nil
	})

	return &order, err
}

func (user *User) BuyPackByCrypto(db *gorm.DB, pack *Pack) (*Order, error) {
	var order Order
	err := db.Transaction(func(tx *gorm.DB) error {
		order = Order{
			UserID:  user.ID,
			PackID:  pack.ID,
			Type:    PendingOrder,
			PayType: CryptoPay,
		}

		if err := order.CreateOrder(tx); err != nil {
			return err
		}

		return nil
	})

	return &order, err
}

func (user *User) RevivePackByCharge(db *gorm.DB, pack *Pack, configID any) (*Order, error) {
	if user.Charge < uint64(pack.Price) {
		return nil, fmt.Errorf("insufficient balance")
	}
	var order Order
	err := db.Transaction(func(tx *gorm.DB) error {
		order = Order{
			UserID: user.ID,
			PackID: pack.ID,
			Type:   ActiveOrder,
		}

		if err := order.CreateOrder(tx); err != nil {
			return err
		}

		var config Config
		if res := tx.Find(&config, configID); res.Error != nil {
			return res.Error
		}

		config.OrderID = order.ID
		if res := tx.Save(&config); res.Error != nil {
			return res.Error
		}

		if err := order.Verify(tx, "خرید سیستمی", "ثبت نشده"); err != nil {
			return err
		}

		if pack.Type == CustomPack {
			order.Type = PendLinkOrder
			if res := tx.Save(&order); res.Error != nil {
				return res.Error
			}
		}

		user.Charge -= uint64(pack.Price)

		if result := tx.Save(user); result.Error != nil {
			return result.Error
		}

		return nil
	})

	return &order, err
}

func (user *User) BuyPackByCharge(db *gorm.DB, pack *Pack) (*Order, error) {
	if user.Charge < uint64(pack.Price) {
		return nil, fmt.Errorf("insufficient balance")
	}
	var order Order
	err := db.Transaction(func(tx *gorm.DB) error {
		order = Order{
			UserID: user.ID,
			PackID: pack.ID,
			Type:   ActiveOrder,
		}

		if err := order.CreateOrder(tx); err != nil {
			return err
		}

		if err := order.Verify(tx, "خرید سیستمی", "ثبت نشده"); err != nil {
			return err
		}

		if pack.Type == CustomPack {
			order.Type = PendLinkOrder
			if res := tx.Save(&order); res.Error != nil {
				return res.Error
			}
		}

		user.Charge -= uint64(pack.Price)

		if result := tx.Save(user); result.Error != nil {
			return result.Error
		}

		return nil
	})

	return &order, err
}

func (user *User) MakeAdmin(db *gorm.DB) error {
	user.Type = AdminUser
	if result := db.Save(user); result.RowsAffected == 0 {
		return fmt.Errorf("unable to create user: %v", result.Error)
	}
	return nil
}

func (user *User) FirstSentOrder(db *gorm.DB, order *Order) error {
	if result := db.Where(&Order{Type: SentOrder, UserID: user.ID}).First(order); result.RowsAffected == 0 {
		return fmt.Errorf("unable to find sent order: %v", result.Error)
	}

	return nil
}

func (user *User) Fullname() string {
	return user.FirstName + " " + user.LastName
}

func (user *User) RetrieveOrders(db *gorm.DB, orderTypes []OrderType, preloads ...string) []Order {
	var orders []Order
	query := db.Where("user_id = ?", user.ID).Where("type in (?)", orderTypes)
	for _, preload := range preloads {
		query.Preload(preload)
	}
	if result := query.Order("created_at DESC").Find(&orders); result.RowsAffected == 0 {
		if result.Error != nil {
			fmt.Println("Error retrieve orders: ", result.Error)
		}
		return []Order{}
	}
	return orders
}

func (ut UserType) Fa() string {
	switch ut {
	case AdminUser:
		return "ادمین"
	default:
		return "عادی"
	}
}

func (user *User) String() string {
	return fmt.Sprintf("آیدی تلگرام: %d\nنام: %s\nنام خانوادگی: %s\nنام کاربری: %s\nنوع: %s\n", user.TelID, user.FirstName, user.LastName, user.Username, user.Type.Fa())
}

func UserIDValidator(db *gorm.DB) form.Validator {
	return func(userID string) (bool, string) {
		var user User
		res := db.Where("tel_id = ?", userID).First(&user)
		if res.RowsAffected == 0 {
			return false, "چنین کاربری یافت نشد. دوباره وارد کنید."
		}
		return true, ""
	}
}
