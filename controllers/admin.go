package controllers

import (
	"context"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"techybat.org/go-vpn/middlewares/auth"
)

func AdminController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	ShowAdminDialog(ctx, b, update)
}

func ShowAdminDialog(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Welcome to admin panel!",
	})
	p := dialog.New(NewAdminDialog(ctx), dialog.Inline())
	p.Show(ctx, b, update.Message.Chat.ID, "admin-panel")
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
					{Text: "افزودن بسته"},
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

// func AddCatMiddleware(next bot.HandlerFunc, adminCtx context.Context) bot.HandlerFunc {
// 	return func(ctx context.Context, bot *bot.Bot, update *tmodels.Update) {
// 		if update.CallbackQuery.From.ID != ctx.Value(auth.UserKey).(tmodels.User).ID {
// 			return
// 		}

// 		return next(ctx, )
// 	}
// }
