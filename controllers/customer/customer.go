package customerController

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	ptime "github.com/yaa110/go-persian-calendar"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/middlewares/auth"
	m "techybat.org/go-vpn/models"
	bp "techybat.org/go-vpn/widgets/buttonpage"
)

func ActiveOrdersHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	user := ctx.Value(auth.UserKey).(m.User)

	orders := user.RetrieveOrders(db, m.ActiveOrder, "Pack")

	buttons := createOrderButtons(orders)

	buttonPage := bp.CreateButtonPage("لیست بسته های فعال", buttons, 5, true)
	buttonPage.Show(ctx, b, chatID)
}

func showOrderHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	db := database.GetDB()
	orderID := update.CallbackQuery.Data
	var order m.Order
	db.Preload("Pack").Preload("Pack.Category").Find(&order, orderID)

	config := order.Config(db)
	pt := ptime.New(config.StartDate)
	showDate := pt.Format("d MMM y")
	txtMsg := fmt.Sprintf("شماره سفارش:%d\nوضعیت سفارش:%s\nگروه بسته:%s\nنوع بسته:%s\nتاریخ شروع:%v\nتوضیحات:%s", order.ID, order.Type, order.Pack.Category.Name, order.Pack, showDate, config.Link)

	if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        txtMsg + "\n\n" + update.CallbackQuery.Message.Message.Text,
		ReplyMarkup: update.CallbackQuery.Message.Message.ReplyMarkup,
	}); err != nil {
		fmt.Println("Error for show order handler: ", err)
	}
}

func createOrderButtons(orders []m.Order) []dialog.Button {
	buttons := []dialog.Button{}
	for _, order := range orders {
		buttons = append(buttons, dialog.Button{
			ID:              fmt.Sprintf("order%d", order.ID),
			Text:            order.Pack.String(),
			CallbackHandler: showOrderHandler,
			CallbackData:    fmt.Sprint(order.ID),
		})
	}
	return buttons
}
