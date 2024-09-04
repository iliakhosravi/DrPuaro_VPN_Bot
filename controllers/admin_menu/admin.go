package admin_menu

import (
	"context"

	"techybat.org/go-vpn/controllers/buy_controller"
	"techybat.org/go-vpn/controllers/category_controller"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	adminController "techybat.org/go-vpn/controllers/admin"
	"techybat.org/go-vpn/middlewares/auth"
	"techybat.org/go-vpn/models"
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
					{ID: "verify-buy-requests", Text: "بررسی درخواست های خرید", CallbackHandler: auth.AdminMiddleware(buy_controller.VerifyBuyController)},
					{Text: "لیست سفارشات", NodeID: "orders-list"},
				},
				{
					{ID: "bulk-message", Text: "ارسال پیام همگانی", CallbackHandler: auth.AdminMiddleware(adminController.SendToAllMsgHandler)},
					{ID: "pv-message", Text: "ارسال پیام به کاربر خاص", CallbackHandler: auth.AdminMiddleware(adminController.SendMsgHandler)},
				},
				{
					{
						Text: "بررسی درخواست های شارژ", NodeID: "charges-list",
					},
				},
				{
					{Text: "مدیریت کارت ها", NodeID: "card-manage"},
				},
			},
		},
		{
			ID:   "charges-list",
			Text: "نوع درخواست شارژ را انتخاب کنید",
			Keyboard: [][]dialog.Button{
				{
					{
						ID:              "Pending",
						Text:            "درانتظار تایید",
						CallbackHandler: auth.AdminMiddleware(adminController.ChargeOrdersHandler),
						CallbackData:    string(models.PendingCharge),
					},
				},
				{
					{
						ID:              "Accepted",
						Text:            "تایید شده",
						CallbackHandler: auth.AdminMiddleware(adminController.ChargeOrdersHandler),
						CallbackData:    string(models.AcceptedCharge),
					},
					{
						ID:              "Dismissed",
						Text:            "رد شده",
						CallbackHandler: auth.AdminMiddleware(adminController.ChargeOrdersHandler),
						CallbackData:    string(models.DismissedCharge),
					},
				},
			},
		},
		{
			ID:   "config-manage",
			Text: "چه بخشی را میخواهید تغییر دهید؟",
			Keyboard: [][]dialog.Button{
				{
					{ID: "add-cat", Text: "افزودن دسته بندی", CallbackHandler: auth.AdminMiddleware(category_controller.AddCatController)},
					{ID: "edit-cat", Text: "ویرایش دسته بندی", CallbackHandler: auth.AdminMiddleware(category_controller.EditCatController)},
				},
				{
					{ID: "add-pack", Text: "افزودن بسته", CallbackHandler: auth.AdminMiddleware(buy_controller.AddPackController)},
					{ID: "edit-pack", Text: "ویرایش بسته", CallbackHandler: auth.AdminMiddleware(buy_controller.EditPackController)},
				},
				{
					{Text: "بازگشت", NodeID: "admin-panel"},
				},
			},
		},
		{
			ID:   "orders-list",
			Text: "انتخاب کنید",
			Keyboard: [][]dialog.Button{
				{
					{ID: "active-orders", Text: "بسته های فعال", CallbackHandler: adminController.OrdersHandler, CallbackData: string(models.ActiveOrder)},
					{ID: "depleted-orders", Text: "بسته های تمام شده", CallbackHandler: adminController.OrdersHandler, CallbackData: string(models.DepletedOrder)},
				},
				{
					{ID: "pending-orders", Text: "بسته های درانتظار تایید", CallbackHandler: adminController.OrdersHandler, CallbackData: string(models.PendingOrder)},
					{ID: "dismissed-orders", Text: "بسته های رد شده", CallbackHandler: adminController.OrdersHandler, CallbackData: string(models.DismissedOrder)},
				},
				{
					{Text: "بازگشت", NodeID: "admin-panel"},
				},
			},
		},
		{
			ID:   "help-btns",
			Text: "انتخاب کنید",
			Keyboard: [][]dialog.Button{
				{
					{
						ID:              "add-inline-kb",
						Text:            "افزودن کیبورد شیشه ای",
						CallbackHandler: auth.AdminMiddleware(adminController.AddInlineKBHandler),
					},
					{
						ID:              "remove-inline-kb",
						Text:            "حذف کیبورد شیشه ای",
						CallbackHandler: auth.AdminMiddleware(adminController.RemoveInlineKBHandler),
					},
				},
				{
					{
						ID:              "add-guide",
						Text:            "افزودن دکمه راهنما",
						CallbackHandler: auth.AdminMiddleware(adminController.AddGuideHandler),
					},
					{
						ID:              "remove-guide",
						Text:            "حذف دکمه راهنما",
						CallbackHandler: auth.AdminMiddleware(adminController.RemoveGuideHandler),
					},
				},
				{
					{Text: "بازگشت", NodeID: "admin-panel"},
				},
			},
		},
		{
			ID:   "card-manage",
			Text: "انتخاب کنید",
			Keyboard: [][]dialog.Button{
				{
					{
						ID:              "add-card",
						Text:            "افزودن کارت",
						CallbackHandler: auth.AdminMiddleware(adminController.AddCardHandler),
					},
				},
				{
					{Text: "بازگشت", NodeID: "admin-panel"},
				},
			},
		},
	}
	return dialogNodes
}
