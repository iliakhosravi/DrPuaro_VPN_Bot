package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Config struct {
	BaseModel
	Link      string    `json:"link"`
	StartDate time.Time `json:"start_date"`
	OrderID   uint      `json:"order_id"`
	Order     Order     `json:"order"`
}

func (config *Config) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Config{})
}

func DateValidator(value string) (bool, string) {
	_, err := time.Parse("2006-01-02", value)

	if err != nil {
		fmt.Println("warning: date format wrong ", err)
		return false, "فرمت ورودی نادرست است. لطفا ورودی را دوباره با فرمت درست وارد کنید."
	}
	return err == nil, ""
}
