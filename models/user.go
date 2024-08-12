package models

import (
	"fmt"

	tmodels "github.com/go-telegram/bot/models"
	"gorm.io/gorm"
)

type User struct {
	BaseModel
	TelID     int64
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
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
