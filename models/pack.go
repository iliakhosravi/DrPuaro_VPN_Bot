package models

import (
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	SUIPack    PackType = "sui"
	MarzPack   PackType = "marz"
)

type Pack struct {
	BaseModel
	Title      string     `json:"title"`
	Traffic    int        `json:"traffic"` //Gigabytes
	Period     int        `json:"period"`  //Days
	Price      int        `json:"price"`   //Toman
	CurrencyID uint       `json:"currency_id"`
	Currency   Currency   `json:"-"`
	CategoryID uint       `json:"category_id"`
	Category   Category   `json:"category"`
	Status     PackStatus `json:"status" gorm:"default:undefined"`
	InboundID  int        `json:"inbound_id" gorm:"default:-1"`
	Type       PackType   `json:"type" gorm:"default:custom"`
	LimitIP    uint       `json:"limit_ip" gorm:"default:0"`
	ClientName string     `json:"client_name" gorm:"default:null"` //SUI Only
	ServiceID  int        `json:"service_id" gorm:"default:-1"`    //Marzneshin Only
	IsTest     bool       `json:"is_test" gorm:"default:false"`    //Free trial pack — see GetTestPack
}

func (pack *Pack) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Pack{})
}

func GetActivePacksByCatID(db *gorm.DB, packs *[]Pack, catID uint) {
	db.Preload(clause.Associations).Preload("Currency").
		Where("is_test = ?", false).
		Find(&packs, Pack{CategoryID: catID, Status: ActivePack})
}

func GetTestPack(db *gorm.DB) (*Pack, error) {
	var pack Pack
	res := db.Preload(clause.Associations).Preload("Currency").
		Where("is_test = ? AND status = ?", true, ActivePack).
		Order("created_at desc").
		First(&pack)

	if res.Error != nil || res.RowsAffected == 0 {
		return nil, fmt.Errorf("no active test pack is configured")
	}
	return &pack, nil
}

func (pack Pack) Name() string {
	if pack.Title != "" {
		return pack.Title
	}
	limit := fmt.Sprintf("%s Users", pack.UserLimitStr())
	return fmt.Sprintf("%s %dD %s", pack.TrafficName(), pack.Period, limit)
}

func (pack Pack) UserLimitStr() string {
	if pack.LimitIP == 0 {
		return "♾"
	}
	return fmt.Sprintf("%d", pack.LimitIP)
}

func (pack Pack) String() string {
	if pack.Title != "" {
		return pack.Title
	}
	return fmt.Sprintf("%s | %s | %s کاربره | %s %s", pack.TrafficString(), StringPeriod(pack.Period), pack.UserLimitStr(), pack.GetPrice(), pack.Currency.Unit)
}

func (pack Pack) GetPrice() decimal.Decimal {
	price := decimal.NewFromInt(int64(pack.Price))
	unitFactor := decimal.NewFromInt(int64(pack.Currency.UnitFactor))
	price = price.Div(unitFactor)
	return price
}

func (pack Pack) TrafficString() string {
	if pack.Traffic == 0 {
		return "♾"
	}
	gb, mb := pack.TrafficGbMb()
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
	if pack.Traffic == 0 {
		return "♾"
	}
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
	testStr := "خیر"
	if pack.IsTest {
		testStr = "بله"
	}
	return fmt.Sprintf("عنوان: %s\nدسته بندی:%s\nترافیک: %s\nدوره زمانی: %d روز\nمحدودیت کاربر: %s\nقیمت: %s %s\nوضعیت: %s\nبسته تست: %s", pack.Name(), pack.Category.Name, pack.TrafficName(), pack.Period, pack.UserLimitStr(), pack.GetPrice(), pack.Currency.Unit, pack.Status, testStr)
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

	if pack.IsTest {
		if res := db.Model(&Pack{}).
			Where("id <> ?", pack.ID).
			Where("is_test = ?", true).
			Update("is_test", false); res.Error != nil {
			return fmt.Errorf("error: unable to clear previous test pack. Details: %s", res.Error)
		}
	}

	return nil
}

func PackValidator(fieldName string) form.Validator {
	return func(value string) (bool, string) {
		var err error
		switch fieldName {
		case "category_id":
			_, err = strconv.ParseUint(value, 10, 0)
		case "currency_id":
			_, err = strconv.ParseUint(value, 10, 0)
		case "type":
			if value != string(CustomPack) && value != string(SanaeiPack) && value != string(SUIPack) && value != string(MarzPack) {
				err = fmt.Errorf("no such pack type")
			}
		case "title":
			if value != "" && len(value) <= 3 && len(value) >= 128 {
				err = fmt.Errorf("Title must be between 3 to 128 characters")
			}
		case "limitIP":
			_, err = strconv.ParseUint(value, 10, 0)
		case "is_test":
			if value != "true" && value != "false" {
				err = fmt.Errorf("is_test must be true or false")
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
	if period == 0 {
		return "نامحدود"
	}
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
