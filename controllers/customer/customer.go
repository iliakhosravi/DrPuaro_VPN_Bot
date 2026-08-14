package customerController

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"github.com/sinasadeghi83/go-telegram-bot-ui/keyboard/inline"
	"techybat.org/go-vpn/controllers/buy_controller"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/middlewares/auth"
	m "techybat.org/go-vpn/models"
	dialog_tools "techybat.org/go-vpn/tools/dialog"
	msgTool "techybat.org/go-vpn/tools/message"
	bp "techybat.org/go-vpn/widgets/buttonpage"
)

func ActiveOrdersHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	user := ctx.Value(auth.UserKey).(m.User)

	orders := user.RetrieveOrders(db, []m.OrderType{m.ActiveOrder}, "Pack", "Pack.Currency")

	buttons := createOrderButtons(orders, showOrderHandler)

	buttonPage := bp.CreateButtonPage(bot.EscapeMarkdown("لیست بسته های فعال"), buttons, 5, true)
	buttonPage.Show(ctx, b, chatID)
}

func DepletedOrdersHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	user := ctx.Value(auth.UserKey).(m.User)

	orders := user.RetrieveOrders(db, []m.OrderType{m.DepletedOrder}, "Pack", "Pack.Currency")

	buttons := createOrderButtons(orders, showDepletedOrderHandler)

	buttonPage := bp.CreateButtonPage(bot.EscapeMarkdown("لیست بسته های تمام شده"), buttons, 5, true)
	buttonPage.Show(ctx, b, chatID)
}

func DismissedOrdersHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	user := ctx.Value(auth.UserKey).(m.User)

	orders := user.RetrieveOrders(db, []m.OrderType{m.DismissedOrder}, "Pack", "Pack.Currency")

	buttons := createOrderButtons(orders, showOrderHandler)

	buttonPage := bp.CreateButtonPage(bot.EscapeMarkdown("لیست بسته های رد شده"), buttons, 5, true)
	buttonPage.Show(ctx, b, chatID)
}

func PendingOrdersHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	user := ctx.Value(auth.UserKey).(m.User)

	orders := user.RetrieveOrders(db, []m.OrderType{m.PendingOrder, m.PendLinkOrder}, "Pack", "Pack.Currency")

	buttons := createOrderButtons(orders, showOrderHandler)

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
	db.Preload("Pack").Preload("Pack.Category").Preload("Pack.Currency").Find(&order, orderID)

	txtMsg := order.UserStr(db)
	if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        txtMsg,
		ReplyMarkup: update.CallbackQuery.Message.Message.ReplyMarkup,
	}); err != nil {
		fmt.Println("Error for show order handler: ", err)
	}

	msgTool.SendPanelSubLink(ctx, b, chatID, order)
}

func showDepletedOrderHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	db := database.GetDB()
	orderID := update.CallbackQuery.Data
	var order m.Order
	db.Preload("Pack").Preload("Pack.Category").Preload("Pack.Currency").Find(&order, orderID)

	var kb tmodels.ReplyMarkup
	if (order.Config(db).ID > 0) && order.Type != m.ActiveOrder {
		kb = inline.New(b).
			Row().Button("تمدید", []byte(fmt.Sprintf("%d", order.Config(db).ID)), handleRevive)
	}

	txtMsg := order.UserStr(db)
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        txtMsg,
		ReplyMarkup: kb,
	}); err != nil {
		fmt.Println("Error for show order handler: ", err)
	}
}

func handleRevive(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	chatID := mes.Message.Chat.ID
	db := database.GetDB()

	nodes := dialog_tools.CreateCatPackNodes(db, passConfigID(buy_controller.BuyController, string(data)))
	dialog := dialog.New(nodes, dialog.Inline())
	dialog.Show(ctx, b, chatID, "categories")
	// var packs []m.Pack
	// db.Find(&packs, "status = ?", m.ActivePack)

	// btns := createPackButtons(packs, string(data))
	// buttonPage := bp.CreateButtonPage("لطفا بسته ای که میخواهید کانفیگ شما به آن اختصاص یابد را انتخاب کنید", btns, 5, true)
	// buttonPage.Show(ctx, b, chatID)
}

func passConfigID(next bot.HandlerFunc, configID string) bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *tmodels.Update) {
		update.CallbackQuery.Data = fmt.Sprintf("%s_%s", update.CallbackQuery.Data, configID)
		next(ctx, bot, update)
	}
}

func createPackButtons(packs []m.Pack, configID string) []dialog.Button {
	buttons := []dialog.Button{}
	for _, pack := range packs {
		buttons = append(buttons, dialog.Button{
			ID:              fmt.Sprintf("pack%d", pack.ID),
			Text:            pack.Name(),
			CallbackHandler: buy_controller.BuyController,
			CallbackData:    fmt.Sprintf("%d_%s", pack.ID, configID),
		})
	}
	return buttons
}

func createOrderButtons(orders []m.Order, handler bot.HandlerFunc) []dialog.Button {
	buttons := []dialog.Button{}
	for _, order := range orders {
		buttons = append(buttons, dialog.Button{
			ID:              fmt.Sprintf("order%d", order.ID),
			Text:            order.Name(database.GetDB()),
			CallbackHandler: handler,
			CallbackData:    fmt.Sprint(order.ID),
		})
	}
	return buttons
}
