package customerController

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/middlewares/auth"
	m "techybat.org/go-vpn/models"
	bp "techybat.org/go-vpn/widgets/buttonpage"
)

func ActiveOrdersHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	user := ctx.Value(auth.UserKey).(m.User)

	orders := user.RetrieveOrders(db, []m.OrderType{m.ActiveOrder}, "Pack")

	buttons := createOrderButtons(orders)

	buttonPage := bp.CreateButtonPage(bot.EscapeMarkdown("لیست بسته های فعال"), buttons, 5, true)
	buttonPage.Show(ctx, b, chatID)
}

func DepletedOrdersHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	user := ctx.Value(auth.UserKey).(m.User)

	orders := user.RetrieveOrders(db, []m.OrderType{m.DepletedOrder}, "Pack")

	buttons := createOrderButtons(orders)

	buttonPage := bp.CreateButtonPage(bot.EscapeMarkdown("لیست بسته های تمام شده"), buttons, 5, true)
	buttonPage.Show(ctx, b, chatID)
}

func DismissedOrdersHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	user := ctx.Value(auth.UserKey).(m.User)

	orders := user.RetrieveOrders(db, []m.OrderType{m.DismissedOrder}, "Pack")

	buttons := createOrderButtons(orders)

	buttonPage := bp.CreateButtonPage(bot.EscapeMarkdown("لیست بسته های رد شده"), buttons, 5, true)
	buttonPage.Show(ctx, b, chatID)
}

func PendingOrdersHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	user := ctx.Value(auth.UserKey).(m.User)

	orders := user.RetrieveOrders(db, []m.OrderType{m.PendingOrder, m.PendLinkOrder}, "Pack")

	buttons := createOrderButtons(orders)

	buttonPage := bp.CreateButtonPage(bot.EscapeMarkdown("لیست بسته های در انتظار تایید"), buttons, 5, true)
	buttonPage.Show(ctx, b, chatID)
}

func showOrderHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	db := database.GetDB()
	orderID := update.CallbackQuery.Data
	var order m.Order
	db.Preload("Pack").Preload("Pack.Category").Find(&order, orderID)

	txtMsg := order.UserStr(db)
	if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        txtMsg,
		ReplyMarkup: update.CallbackQuery.Message.Message.ReplyMarkup,
	}); err != nil {
		fmt.Println("Error for show order handler: ", err)
	}

	if order.Type == m.ActiveOrder {
		shortLink, qrPath := order.Config(db).ShortLink(db)
		fileContent, _ := os.ReadFile(qrPath)
		b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:    chatID,
			Caption:   fmt.Sprintf("لینک کانفیگ:\n`%s`", shortLink),
			Photo:     &tmodels.InputFileUpload{Filename: "qrcode.jpg", Data: bytes.NewReader(fileContent)},
			ParseMode: tmodels.ParseModeMarkdown,
		})
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
