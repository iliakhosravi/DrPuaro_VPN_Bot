package models

import (
	"fmt"

	"gorm.io/gorm"
)

type CategoryStatus string

const (
	UndefinedCat CategoryStatus = "undefined"
	ActiveCat    CategoryStatus = "active"
	UnactiveCat  CategoryStatus = "unactive"
)

type Category struct {
	BaseModel
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Status      CategoryStatus `json:"category_status" gorm:"default:undefined"`
}

func (cat *Category) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Category{})
}

func (cat *Category) Store(db *gorm.DB) error {
	if result := db.Save(cat); result.RowsAffected == 0 {
		return fmt.Errorf("error: unable to store category. Details: %v\n", result.Error)
	}
	return nil
}

func GetActiveCategories(db *gorm.DB, cats *[]Category) {
	db.Where(&Category{Status: ActiveCat}).Find(cats)
}

func (cat *Category) Active(db *gorm.DB) error {
	cat.Status = ActiveCat
	if res := db.Save(cat); res.Error != nil {
		fmt.Println("Error unable to Active catefory. err: ", res.Error)
		return res.Error
	}
	return nil
}

func (cat *Category) Deactive(db *gorm.DB) error {
	cat.Status = UnactiveCat
	if res := db.Save(cat); res.Error != nil {
		fmt.Println("Error unable to deactive catefory. err: ", res.Error)
		return res.Error
	}
	return nil
}

func (cat *Category) String() string {
	return fmt.Sprintf("مشخصات دسته بندی به شرح زیر است:\nنام: %s\nجزئیات: %s\nوضعیت: %s", cat.Name, cat.Description, cat.Status)
}

func (cs CategoryStatus) String() string {
	switch cs {
	case ActiveCat:
		return "فعال"
	case UnactiveCat:
		return "غیرفعال"
	default:
		return "نامشخص"
	}
}
