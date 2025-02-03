package config_crons

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"gorm.io/gorm/clause"
	"techybat.org/go-vpn/database"
	m "techybat.org/go-vpn/models"
)

var (
	notifs     *Notifs
	notifsOnce sync.Once
)

type Notifs struct {
	Depletions []uint         `json:"depletion"`
	EndDays    map[int][]uint `json:"end-days"`
	Traffic    map[int][]uint `json:"traffic"`
}

func getNotifs() *Notifs {
	notifsOnce.Do(func() {
		importNotifs()
		if len(notifs.Depletions) > 0 {
			return
		}
		notifs = &Notifs{
			Depletions: []uint{
				0,
			},

			EndDays: map[int][]uint{
				1: {},
				2: {},
				3: {},
				4: {},
				5: {},
				6: {},
				7: {},
			},

			Traffic: map[int][]uint{
				200:  {},
				500:  {},
				1000: {},
			},
		}
	})
	return notifs
}

func NotifyAll(ctx context.Context, b *bot.Bot) {
	db := database.GetDB()
	var configs []m.Config
	getNotifs()
	packQuery := db.Model(&m.Pack{}).Select("id").Where("type = ?", m.SanaeiPack)
	orderQuery := db.Model(&m.Order{}).Select("id").Where("pack_id in (?)", packQuery).Where("type in (?)", []string{string(m.ActiveOrder), string(m.DepletedOrder)}).Where("id not in (?)", notifs.Depletions)
	db.Where("order_id in (?)", orderQuery).Preload(clause.Associations).Preload("Order.User").Preload("Order.Pack").Find(&configs)

	fmt.Printf("NotifyAll configs\n")

	for _, config := range configs {
		client, err := config.GetClient()
		if err != nil {
			fmt.Println("Error: unable to retrieve client for notify all. err: ", err)
			continue
		}
		endTime := time.UnixMilli(client.ExpiryTime)
		duration := time.Until(endTime)
		daysDuration := int(duration.Hours()) / 24
		notified := false
		switch {
		case duration.Hours() <= 0:
			notifyDepletion(ctx, b, config, true)
			notified = true

		case daysDuration <= 7:
			notifyEndDays(ctx, b, config, daysDuration)
		}

		if !notified {
			remainedTraffic := client.RemainedTraffic()
			switch {
			case remainedTraffic <= 0:
				if config.Order.Pack.Traffic == 0 {
					continue
				}
				notifyDepletion(ctx, b, config, false)
			case remainedTraffic <= 200:
				notifyRemainedTraffic(ctx, b, config, 200)
			case remainedTraffic <= 500:
				notifyRemainedTraffic(ctx, b, config, 500)
			case remainedTraffic <= 1000:
				notifyRemainedTraffic(ctx, b, config, 1000)
			}
		}

		config.SyncByClient(db, client)
	}
	saveNotifs()
}

func notifyRemainedTraffic(ctx context.Context, b *bot.Bot, config m.Config, remainedMB int) {
	if contains(notifs.Traffic[remainedMB], config.OrderID) {
		return
	}
	db := database.GetDB()
	txtMsg := fmt.Sprintf("کمتر از %d مگابایت از حجم بسته شما باقی مانده است. شما می توانید پیش از اتمام بسته نسبت به خرید بسته جدید از منو اصلی اقدام فرمایید.\n\nمشخصات بسته:%s", remainedMB, config.Order.UserStr(db))
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: config.Order.User.TelID,
		Text:   txtMsg,
	})
	notifs.Traffic[remainedMB] = append(notifs.Traffic[remainedMB], config.OrderID)
}

func notifyEndDays(ctx context.Context, b *bot.Bot, config m.Config, daysLeft int) {
	if contains(notifs.EndDays[daysLeft], config.OrderID) {
		return
	}
	db := database.GetDB()
	txtMsg := fmt.Sprintf("کمتر از %d روز از مدت زمان بسته شما باقی مانده است. شما می توانید پیش از اتمام بسته نسبت به خرید بسته جدید از منو اصلی اقدام فرمایید.\n\nمشخصات بسته:%s", daysLeft, config.Order.UserStr(db))
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: config.Order.User.TelID,
		Text:   txtMsg,
	})
	notifs.EndDays[daysLeft] = append(notifs.EndDays[daysLeft], config.OrderID)
}

func notifyDepletion(ctx context.Context, b *bot.Bot, config m.Config, isTime bool) {
	db := database.GetDB()
	var txtMsg string
	if isTime {
		txtMsg = fmt.Sprintf("شما به انتهای دوره بسته خود با مشخصات زیر رسیده اید. برای خرید مجدد از منو اقدام فرمایید.\n\n%s", config.Order.UserStr(db))
	} else {
		txtMsg = fmt.Sprintf("حجم بسته شما با مشخصات زیر به پایان رسیده است لطفا برای خرید مجدد از منو اقدام فرمایید.\n\n%s", config.Order.UserStr(db))
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: config.Order.User.TelID,
		Text:   txtMsg,
	})
	notifs.Depletions = append(notifs.Depletions, config.OrderID)
}

func contains(slice []uint, item uint) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func saveNotifs() {
	file, _ := os.OpenFile("notifs.json", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	encoder := json.NewEncoder(file)
	encoder.Encode(notifs)
}

func importNotifs() {
	file, _ := os.ReadFile("notifs.json")
	if notifs == nil {
		notifs = &Notifs{}
	}
	json.Unmarshal(file, notifs)
}
