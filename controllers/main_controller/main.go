package main_controller

import (
	"context"
	"fmt"
	"os"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"gorm.io/gorm"
	comp "techybat.org/go-vpn/components"
	"techybat.org/go-vpn/controllers/buy_controller"
	customerController "techybat.org/go-vpn/controllers/customer"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/models"
	dialog_tools "techybat.org/go-vpn/tools/dialog"
)

func MainController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	if update.CallbackQuery != nil {
		return
	}
	ShowMainDialog(ctx, b, update)
}

func ShowMainDialog(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	if os.Getenv("MAIN_KB_INLINE") != "true" {
		b.SendMessage(ctx, &bot.SendMessageParams{

			ChatID:      update.Message.Chat.ID,
			Text:        fmt.Sprintf("%s\n\n%s", os.Getenv("BRAND_NAME"), os.Getenv("TG_CHANNEL")),
			ReplyMarkup: comp.GetMenuKeyboard(b),
		})
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{

			ChatID: update.Message.Chat.ID,
			Text:   "خوش آمدید",
			ReplyMarkup: tmodels.ReplyKeyboardRemove{
				RemoveKeyboard: true,
				Selective:      false,
			},
		})
		p := dialog.New(NewMainDialog(), dialog.Inline())
		p.Show(ctx, b, update.Message.Chat.ID, "start")
	}
}

func handleHowConnect(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	p := dialog.New(comp.NewMainNodes(), dialog.Inline())
	p.Show(ctx, b, update.Message.Chat.ID, "how-connect")
}

func handleGuide(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	guideNodes := []dialog.Node{
		{
			ID:       "guide",
			Text:     "در چه موضوعی نیاز به راهنمایی دارید؟",
			Keyboard: CreateGuideKeyboard(database.GetDB()),
		},
	}
	p := dialog.New(guideNodes, dialog.Inline())
	p.Show(ctx, b, update.Message.Chat.ID, "guide")
}

func NewMainDialog() []dialog.Node {
	db := database.GetDB()
	dialogNodes := []dialog.Node{
		{
			ID:   "start",
			Text: fmt.Sprintf("%s\n\n%s", os.Getenv("BRAND_NAME"), os.Getenv("TG_CHANNEL")),
			Keyboard: [][]dialog.Button{
				{
					{Text: "خرید بسته", NodeID: "categories"},
					{Text: "بسته های خریداری شده", NodeID: "orders"},
				},
				{
					{ID: "balance", Text: "کیف پول", CallbackHandler: customerController.BalanceHandler},
					{ID: "charge-account", Text: "شارژ اکانت", CallbackHandler: buy_controller.ChargeHandler},
				},
			},
		},

		{
			ID:   "orders",
			Text: "انتخاب کنید",
			Keyboard: [][]dialog.Button{
				{
					{ID: "active-orders", Text: "بسته های فعال", CallbackHandler: customerController.ActiveOrdersHandler},
					{ID: "depleted-orders", Text: "بسته های تمام شده", CallbackHandler: customerController.DepletedOrdersHandler},
				},
				{
					{ID: "pending-orders", Text: "بسته های درانتظار تایید", CallbackHandler: customerController.PendingOrdersHandler},
					{ID: "dismissed-orders", Text: "بسته های رد شده", CallbackHandler: customerController.DismissedOrdersHandler},
				},
				{
					{Text: "بازگشت", NodeID: "start"},
				},
			},
		},
	}

	dialogNodes[0].Keyboard = append(dialogNodes[0].Keyboard, CreateGuideKeyboard(db)...)
	dialogNodes = append(dialogNodes, dialog_tools.CreateCatPackNodes(db, buy_controller.BuyController)...)

	return dialogNodes
}

func CreateGuideKeyboard(db *gorm.DB) [][]dialog.Button {
	var guides []models.Guide
	db.Find(&guides)
	return comp.BuildGuideInlineKeyboard(guides)
}
