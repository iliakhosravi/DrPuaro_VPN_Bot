package config_crons

import (
	"context"
	"fmt"
	"time"

	"github.com/go-telegram/bot"
	"gorm.io/gorm/clause"
	"techybat.org/go-vpn/database"
	m "techybat.org/go-vpn/models"
)

var notifiedDepletions = []uint{
	0,
}
var notifiedEndDays = map[int][]uint{
	1: {},
	2: {},
	3: {},
	4: {},
	5: {},
	6: {},
	7: {},
}

var notifiedTraffic = map[int][]uint{
	200:  {},
	500:  {},
	1000: {},
}

func NotifyAll(ctx context.Context, b *bot.Bot) {
	db := database.GetDB()
	var configs []m.Config
	orderQuery := db.Model(&m.Order{}).Select("id").Where("type in (?)", []string{string(m.ActiveOrder), string(m.DepletedOrder)}).Where("id not in (?)", notifiedDepletions)
	db.Where("order_id in (?)", orderQuery).Preload(clause.Associations).Preload("Order.User").Find(&configs)

	fmt.Printf("NotifyAll configs: %v\n", configs)

	for _, config := range configs {
		client, err := config.GetClient()
		if err != nil {
			fmt.Println("Error: unable to retrieve client for notify all. err: ", err)
			continue
		}
		endTime := time.UnixMilli(client.ExpiryTime)
		duration := time.Until(endTime)
		daysDuration := int(duration.Hours()) / 24
		fmt.Printf("Config: %d | Days Duration: %d\n", config.ID, daysDuration)
		switch {
		case duration.Hours() <= 0:
			notifyDepletion(ctx, b, config, true)
		case daysDuration <= 7:
			notifyEndDays(ctx, b, config, daysDuration)
		}

		remainedTraffic := client.RemainedTraffic()
		switch {
		case remainedTraffic <= 0:
			notifyDepletion(ctx, b, config, false)
		case remainedTraffic <= 0.2:
			notifyRemainedTraffic(ctx, b, config, 200)
		case remainedTraffic <= 0.5:
			notifyRemainedTraffic(ctx, b, config, 500)
		case remainedTraffic <= 1:
			notifyRemainedTraffic(ctx, b, config, 1000)
		}

		config.SyncByClient(db, client)
	}
}

func notifyRemainedTraffic(ctx context.Context, b *bot.Bot, config m.Config, remainedMB int) {
	if contains(notifiedTraffic[remainedMB], config.OrderID) {
		return
	}
	db := database.GetDB()
	txtMsg := fmt.Sprintf("کمتر از %d مگابایت از حجم بسته شما باقی مانده است. شما می توانید پیش از اتمام بسته نسبت به خرید بسته جدید از منو اصلی اقدام فرمایید.\n\nمشخصات بسته:%s", remainedMB, config.Order.UserStr(db))
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: config.Order.User.TelID,
		Text:   txtMsg,
	})
	notifiedTraffic[remainedMB] = append(notifiedEndDays[remainedMB], config.OrderID)
}

func notifyEndDays(ctx context.Context, b *bot.Bot, config m.Config, daysLeft int) {
	if contains(notifiedEndDays[daysLeft], config.OrderID) {
		return
	}
	db := database.GetDB()
	txtMsg := fmt.Sprintf("کمتر از %d روز از مدت زمان بسته شما باقی مانده است. شما می توانید پیش از اتمام بسته نسبت به خرید بسته جدید از منو اصلی اقدام فرمایید.\n\nمشخصات بسته:%s", daysLeft, config.Order.UserStr(db))
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: config.Order.User.TelID,
		Text:   txtMsg,
	})
	notifiedEndDays[daysLeft] = append(notifiedEndDays[daysLeft], config.OrderID)
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
	notifiedDepletions = append(notifiedDepletions, config.OrderID)
}

func contains(slice []uint, item uint) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
