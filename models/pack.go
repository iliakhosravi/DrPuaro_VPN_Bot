package models

import (
	"fmt"
	"strconv"

	"gorm.io/gorm"
	"techybat.org/go-vpn/vars"
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
	Title      string     `json:"title"`
	Traffic    int        `json:"traffic"` //Gigabytes
	Period     int        `json:"period"`  //Days
	Price      int        `json:"price"`   //Toman
	CategoryID uint       `json:"category_id"`
	Category   Category   `json:"category"`
	Status     PackStatus `json:"status" gorm:"default:undefined"`
	InboundID  int        `json:"inbound_id" gorm:"default:-1"`
	Type       PackType   `json:"type" gorm:"default:custom"`
}

func (pack *Pack) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Pack{})
}

func GetActivePacksByCatID(db *gorm.DB, packs *[]Pack, catID uint) {
	db.Find(&packs, Pack{CategoryID: catID, Status: ActivePack})
}

func (pack Pack) Name() string {
	if pack.Title != "" {
		return pack.Title
	}
	return fmt.Sprintf("%s %dD", pack.TrafficName(), pack.Period)
}

func (pack Pack) String() string {
	if pack.Title != "" {
		return pack.Title
	}
	return fmt.Sprintf("%s | %s | %d تومان", pack.TrafficString(), StringPeriod(pack.Period), pack.Price)
}

func (pack Pack) TrafficString() string {
	gb := pack.Traffic / 1024
	mb := pack.Traffic % 1024
	result := ""
	if gb > 0 {
		result += fmt.Sprintf("%d گیگابایت", gb)
	}

	if mb > 0 {
		if gb > 0 {
			result += " و "
		}
		result += fmt.Sprintf("%d مگابایت", mb)
	}

	return result
}

func (pack Pack) TrafficName() string {
	gb, mb := pack.TrafficGbMb()
	result := ""
	if gb > 0 {
		result += fmt.Sprintf("%d GB", gb)
	}

	if mb > 0 {
		if gb > 0 {
			result += " "
		}
		result += fmt.Sprintf("%d MB", mb)
	}

	return result
}

func (pack Pack) TrafficGbMb() (int, int) {
	gb := pack.Traffic / 1024
	mb := pack.Traffic % 1024
	return gb, mb
}

func (pack Pack) ConfigDesc() string {
	return fmt.Sprintf("%s | %s", vars.Get("TELEGRAM_CHANNEL"), pack.Name())
}

func (pack Pack) FullStr() string {
	return fmt.Sprintf("عنوان: %s\nدسته بندی:%s\nترافیک: %s\nدوره زمانی: %d روز\nقیمت: %d تومان\nوضعیت: %s", pack.Name(), pack.Category.Name, pack.TrafficName(), pack.Period, pack.Price, pack.Status)
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
		case "title":
			if value != "" && len(value) <= 3 && len(value) >= 128 {
				err = fmt.Errorf("Title must be between 3 to 128 characters")
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

func StringPeriod(period int) string {
	result := ""
	years := period / 365
	months := period / 30
	days := period % 30
	if years != 0 {
		result += fmt.Sprintf("%d سال", years)
		if months+days != 0 {
			result += " و "
		}
	}
	if months != 0 {
		result += fmt.Sprintf("%d ماه", months)
		if days != 0 {
			result += " و "
		}
	}
	if days != 0 {
		result += fmt.Sprintf("%d روز", days)
	}
	result += "ه"
	return result
}
