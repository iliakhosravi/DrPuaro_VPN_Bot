package components

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"github.com/sinasadeghi83/go-telegram-bot-ui/keyboard/reply"
	"techybat.org/go-vpn/controllers/buy_controller"
	customerController "techybat.org/go-vpn/controllers/customer"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/models"
	dialog_tools "techybat.org/go-vpn/tools/dialog"
	"techybat.org/go-vpn/vars"
)

var (
	mainKB     *reply.ReplyKeyboard
	mainKBOnce sync.Once
)

func BuildMainKeyboard(b *bot.Bot) {
	mainKB = reply.New(b, reply.ResizableKeyboard()).
		Button("🛍خرید بسته", b, bot.MatchTypeExact, handleBuy).
		Button("👨‍💻بسته ها", b, bot.MatchTypeExact, handleBuyList).
		Row().
		Button("💰کیف پول", b, bot.MatchTypeExact, customerController.BalanceHandler).
		Button("💳 شارژ اکانت", b, bot.MatchTypeExact, buy_controller.ChargeHandler)

	addGuideBtnsToReply(b)
}

func handleBuy(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	p := dialog.New(NewMainNodes(), dialog.Inline())
	p.Show(ctx, b, update.Message.Chat.ID, "categories")
}

func handleBuyList(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	p := dialog.New(NewMainNodes(), dialog.Inline())
	p.Show(ctx, b, update.Message.Chat.ID, "orders")
}

func GetMenuKeyboard(b *bot.Bot) *reply.ReplyKeyboard {
	mainKBOnce.Do(func() {
		BuildMainKeyboard(b)
	})
	return mainKB
}

func addGuideBtnsToReply(b *bot.Bot) {
	db := database.GetDB()
	var kbs []models.InlineKeyboard
	db.Find(&kbs)
	for i, kb := range kbs {
		if i%2 == 0 {
			mainKB.Row()
		}

		mainKB.Button(kb.Name, b, bot.MatchTypeExact, passKB(handleGuideKB, &kb))
	}
}

func passKB(next func(ctx context.Context, b *bot.Bot, update *tmodels.Update, kb *models.InlineKeyboard), kb *models.InlineKeyboard) bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *tmodels.Update) {
		next(ctx, bot, update, kb)
	}
}

func handleGuideKB(ctx context.Context, b *bot.Bot, update *tmodels.Update, kb *models.InlineKeyboard) {
	db := database.GetDB()
	guides := kb.Guides(db)

	guideNodes := []dialog.Node{
		{
			ID:       "guide",
			Text:     bot.EscapeMarkdown(kb.Text),
			Keyboard: BuildGuideInlineKeyboard(guides),
		},
	}
	p := dialog.New(guideNodes, dialog.Inline())
	p.Show(ctx, b, update.Message.Chat.ID, "guide")
}

func ForwardGuideHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	msgID, _ := strconv.Atoi(update.CallbackQuery.Data)
	b.CopyMessage(ctx, &bot.CopyMessageParams{
		FromChatID: vars.Get("STORAGE_CHANNEL_ID"),
		ChatID:     update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:  msgID,
	})
	// b.ForwardMessage(ctx, &bot.ForwardMessageParams{
	// 	FromChatID: vars.Get("STORAGE_CHANNEL_ID"),
	// 	ChatID:     update.CallbackQuery.Message.Message.Chat.ID,
	// 	MessageID:  msgID,
	// })
}

func BuildGuideInlineKeyboard(guides []models.Guide) [][]dialog.Button {
	keyboard := [][]dialog.Button{}
	for i, guide := range guides {
		if i%2 == 0 {
			keyboard = append(keyboard, []dialog.Button{})
		}
		idx := len(keyboard) - 1
		button := dialog.Button{
			ID:   fmt.Sprintf("guide%d", i),
			Text: guide.Title,
		}
		switch guide.Type {
		case models.LINK_GUIDE:
			button.URL = guide.Link
		case models.FWD_MSG_GUIDE:
			button.CallbackHandler = ForwardGuideHandler
			button.CallbackData = fmt.Sprintf("%d", guide.FwdMsgID)
		}
		keyboard[idx] = append(keyboard[idx], button)
	}
	return keyboard
}

func NewMainNodes() []dialog.Node {
	db := database.GetDB()
	dialogNodes := []dialog.Node{
		{
			ID:       "start",
			Text:     fmt.Sprintf("%s\n\n%s", vars.Get("BRAND_NAME"), vars.Get("TG_CHANNEL")),
			Keyboard: nil,
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
	dialogNodes = append(dialogNodes, dialog_tools.CreateCatPackNodes(db, buy_controller.BuyController)...)

	return dialogNodes
}
