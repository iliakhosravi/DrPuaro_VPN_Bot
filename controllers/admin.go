package controllers

import (
	"context"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"techybat.org/go-vpn/middlewares/auth"
)

func AdminController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	var chatID int64
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Message.Chat.ID
	} else {
		chatID = update.Message.Chat.ID
	}
	ShowAdminDialog(ctx, b, update, chatID)
}

func ShowAdminDialog(ctx context.Context, b *bot.Bot, update *tmodels.Update, chatID int64) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Welcome to admin panel!",
		ReplyMarkup: tmodels.ReplyKeyboardRemove{
			RemoveKeyboard: true,
		},
	})
	p := dialog.New(NewAdminDialog(ctx), dialog.Inline())
	p.Show(ctx, b, chatID, "admin-panel")
}

func NewAdminDialog(ctx context.Context) []dialog.Node {
	// db := database.GetDB()
	dialogNodes := []dialog.Node{
		{
			ID:   "admin-panel",
			Text: "Ultra VPN Admin Panel",
			Keyboard: [][]dialog.Button{
				{
					{Text: "مدیریت کانفیگ ها", NodeID: "config-manage"},
					{Text: "مدیریت دکمه های راهنما", NodeID: "help-btns"},
				},
				{
					{Text: "بررسی درخواست های خرید", NodeID: "buy-requests"},
					{Text: "کانفیگ های خریداری شده", NodeID: "bought-configs"},
				},
				{
					{Text: "ارسال پیام همگانی", NodeID: "bulk-message"},
				},
			},
		},
		{
			ID:   "config-manage",
			Text: "چه بخشی را میخواهید تغییر دهید؟",
			Keyboard: [][]dialog.Button{
				{
					{ID: "add-cat", Text: "افزودن دسته بندی", CallbackHandler: auth.AdminMiddleware(AddCatController)},
					{Text: "ویرایش دسته بندی"},
				},
				{
					{ID: "add-pack", Text: "افزودن بسته", CallbackHandler: auth.AdminMiddleware(AddPackController)},
					{Text: "ویرایش بسته"},
				},
				{
					{Text: "بازگشت", NodeID: "admin-panel"},
				},
			},
		},
	}
	return dialogNodes
}
