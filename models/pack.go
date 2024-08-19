package models

import (
	"fmt"
	"strconv"

	"gorm.io/gorm"
	"techybat.org/go-vpn/widgets/form"
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

func (pack Pack) String() string {
	return fmt.Sprintf("حجم %d گیگابایت | %d روزه | %d تومان", pack.Traffic, pack.Period, pack.Price)
}

func (pack *Pack) Store(db *gorm.DB) error {
	if result := db.Save(pack); result.RowsAffected == 0 {
		return fmt.Errorf("Error: unable to store pack. Details: %v\n", result.Error)
	}
	return nil
}

func PackValidator(fieldName string) form.Validator {
	return func(value string) (bool, string) {
		var err error
		switch fieldName {
		case "category_id":
			_, err = strconv.ParseUint(value, 10, 0)
		default:
			_, err = strconv.Atoi(value)
		}
		return err == nil, "فرمت ورودی نادرست است. لطفا ورودی را دوباره با فرمت درست وارد کنید."
	}
}
