package models

import (
	"fmt"

	tmodels "github.com/go-telegram/bot/models"
	"gorm.io/gorm"
)

type UserType string

const (
	NoramlUser UserType = "normal"
	AdminUser           = "admin"
)

type User struct {
	BaseModel
	TelID     int64
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Username  string   `json:"username"`
	Type      UserType `json:"user_type" gorm:"default:normal"`
}

func (user *User) Migrate(db *gorm.DB) {
	db.AutoMigrate(&User{})
}

func (user *User) CreateOrFindUserByTelegram(db *gorm.DB, tuser *tmodels.User) error {
	db.First(user, &User{TelID: tuser.ID})
	user.FirstName = tuser.FirstName
	user.LastName = tuser.LastName
	user.Username = tuser.Username
	user.TelID = tuser.ID

	if result := db.Save(user); result.RowsAffected == 0 {
		return fmt.Errorf("unable to create user: %v", result.Error)
	}
	return nil
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

func (user *User) RetrieveOrders(db *gorm.DB, orderType OrderType, preloads ...string) []Order {
	var orders []Order
	query := db.Where(&Order{UserID: user.ID, Type: orderType})
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

func (ut UserType) String() string {
	switch ut {
	case AdminUser:
		return "ادمین"
	default:
		return "عادی"
	}
}

func (user *User) String() string {
	return fmt.Sprintf("آیدی تلگرام: %d\nنام: %s\nنام خانوادگی: %s\nنام کاربری: %s\nنوع: %s\n", user.TelID, user.FirstName, user.LastName, user.Username, user.Type)
}
