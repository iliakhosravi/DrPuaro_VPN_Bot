package models

import (
	"fmt"

	"gorm.io/gorm"
)

type CardStatus string

const (
	UndefinedCard CardStatus = "undefined"
	ActiveCard    CardStatus = "active"
	ReserveCard   CardStatus = "reserve"
)

type Card struct {
	BaseModel
	Fullname string     `json:"fullname"`
	Number   string     `json:"number"`
	Status   CardStatus `json:"status" gorm:"default:undefined"`
}

func (card *Card) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Card{})
}

func ReserveActiveCard(db *gorm.DB) error {
	var card Card
	result := db.Where("status = ?", ActiveCard).First(card)
	if result.RowsAffected == 0 {
		return nil
	}
	return card.ChangeStatus(db, ReserveCard)
}

func (card *Card) ChangeStatus(db *gorm.DB, cardSt CardStatus) error {
	card.Status = cardSt
	if result := db.Save(card); result.Error != nil {
		fmt.Println("Error unable to change card status in db. err: ", result.Error)
		return result.Error
	}
	return nil
}

func AddNewCard(db *gorm.DB, card *Card) error {
	if err := ReserveActiveCard(db); err != nil {
		return err
	}
	card.Status = ActiveCard
	if result := db.Save(&card); result.Error != nil {
		fmt.Println("Error unable to save card to db. err: ", result.Error)
		return result.Error
	}
	return nil
}

func GetActiveCard(db *gorm.DB) Card {
	var card Card
	db.Where("status = ?", ActiveCard).Order("created_at desc").First(&card)
	return card
}
