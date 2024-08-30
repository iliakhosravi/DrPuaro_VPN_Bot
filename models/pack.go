package models

import (
	"fmt"
	"os"
	"strconv"

	"gorm.io/gorm"
	"techybat.org/go-vpn/widgets/form"
)

type PackStatus string
type PackType string

const (
	UndefinedPack PackStatus = "undefined"
	ActivePack    PackStatus = "active"
	UnactivePack  PackStatus = "unactive"
)

const (
	CustomPack PackType = "custom"
	SanaeiPack PackType = "sanaei"
)

type Pack struct {
	BaseModel
	Traffic    int        `json:"traffic"` //Gigabytes
	Period     int        `json:"period"`  //Days
	Price      int        `json:"price"`   //Toman
	CategoryID uint       `json:"category_id"`
	Category   Category   `json:"category"`
	Status     PackStatus `json:"status" gorm:"default:undefined"`
	InboundID  int        `json:"inbound_id" gorm:"default:-1"`
	Type       PackType   `json:"type"`
}

func (pack *Pack) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Pack{})
}

func GetActivePacksByCatID(db *gorm.DB, packs *[]Pack, catID uint) {
	db.Find(&packs, Pack{CategoryID: catID, Status: ActivePack})
}

func (pack Pack) Name() string {
	return fmt.Sprintf("%dGB %dD", pack.Traffic, pack.Period)
}

func (pack Pack) String() string {
	return fmt.Sprintf("حجم %d گیگابایت | %d روزه | %d تومان", pack.Traffic, pack.Period, pack.Price)
}

func (pack Pack) ConfigDesc() string {
	return fmt.Sprintf("%s | %s", os.Getenv("TELEGRAM_CHANNEL"), pack.Name())
}

func (pack Pack) FullStr() string {
	return fmt.Sprintf("دسته بندی:%s\nترافیک: %dGB\nدوره زمانی: %d روز\nقیمت: %d تومان\nوضعیت: %s", pack.Category.Name, pack.Traffic, pack.Period, pack.Price, pack.Status)
}

func (pack *Pack) Active(db *gorm.DB) error {
	pack.Status = ActivePack
	return pack.Store(db)
}

func (pack *Pack) Deactive(db *gorm.DB) error {
	pack.Status = UnactivePack
	return pack.Store(db)
}

func (pack *Pack) Store(db *gorm.DB) error {
	if result := db.Save(pack); result.RowsAffected == 0 {
		return fmt.Errorf("error: unable to store pack. Details: %s", result.Error)
	}
	return nil
}

func PackValidator(fieldName string) form.Validator {
	return func(value string) (bool, string) {
		var err error
		switch fieldName {
		case "category_id":
			_, err = strconv.ParseUint(value, 10, 0)
		case "type":
			if value != string(CustomPack) && value != string(SanaeiPack) {
				err = fmt.Errorf("no such pack type")
			}
		default:
			_, err = strconv.Atoi(value)
		}
		return err == nil, "فرمت ورودی نادرست است. لطفا ورودی را دوباره با فرمت درست وارد کنید."
	}
}

func (ps PackStatus) String() string {
	switch ps {
	case ActivePack:
		return "فعال"
	case UnactivePack:
		return "غیرفعال"
	default:
		return "نامشخص"
	}
}

func GetPackPeriods(packs []Pack) map[int][]Pack {
	periodsMap := make(map[int][]Pack)
	for _, pack := range packs {
		_, exist := periodsMap[pack.Period]
		if !exist {
			periodsMap[pack.Period] = make([]Pack, 0)
		}
		periodsMap[pack.Period] = append(periodsMap[pack.Period], pack)
	}

	return periodsMap
}
