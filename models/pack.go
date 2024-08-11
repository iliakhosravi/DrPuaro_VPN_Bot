package models

import (
	"fmt"

	"gorm.io/gorm"
)

type Pack struct {
	BaseModel
	Traffic    int      `json:"traffic"` //Gigabytes
	Period     int      `json:"period"`  //Days
	Price      int      `json:"price"`   //Toman
	CategoryID uint     `json:"category_id"`
	Category   Category `json:"category"`
}

func (pack *Pack) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Pack{})
}

func (pack *Pack) String() string {
	return fmt.Sprintf("حجم %d گیگابایت | %d روزه | %d تومان", pack.Traffic, pack.Period, pack.Price)
}
