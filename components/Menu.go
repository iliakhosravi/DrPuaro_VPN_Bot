package components

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"github.com/sinasadeghi83/go-telegram-bot-ui/keyboard/reply"
	"gorm.io/gorm"
	"techybat.org/go-vpn/controllers/buy_controller"
	customerController "techybat.org/go-vpn/controllers/customer"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/models"
)

var (
	mainKB     *reply.ReplyKeyboard
	mainKBOnce sync.Once
)

func BuildMainKeyboard(b *bot.Bot) {
	mainKB = reply.New(b, reply.ResizableKeyboard()).
		Button("🛍خرید بسته", b, bot.MatchTypeExact, handleBuy).
		Button("👨‍💻بسته های خریداری شده", b, bot.MatchTypeExact, handleBuyList).
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
		FromChatID: os.Getenv("STORAGE_CHANNEL_ID"),
		ChatID:     update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:  msgID,
	})
	// b.ForwardMessage(ctx, &bot.ForwardMessageParams{
	// 	FromChatID: os.Getenv("STORAGE_CHANNEL_ID"),
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
			Text:     fmt.Sprintf("%s\n\n%s", os.Getenv("BRAND_NAME"), os.Getenv("TG_CHANNEL")),
			Keyboard: nil,
		},

		{
			ID:   "how-connect",
			Text: os.Getenv("HOW_CONNECT_TEXT"),
			Keyboard: [][]dialog.Button{
				{
					{
						Text: os.Getenv("HOW_IPHONE_TEXT"),
						URL:  os.Getenv("HOW_IPHONE_LINK"),
					},
					{
						Text: os.Getenv("HOW_ANDROID_TEXT"),
						URL:  os.Getenv("HOW_ANDROID_LINK"),
					},
				},
				{
					{
						Text: os.Getenv("HOW_UNIX_TEXT"),
						URL:  os.Getenv("HOW_UNIX_LINK"),
					},
					{
						Text: os.Getenv("HOW_WIN_TEXT"),
						URL:  os.Getenv("HOW_WIN_LINK"),
					},
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
	dialogNodes = append(dialogNodes, CreateCatPackNodes(db, buy_controller.BuyController)...)

	return dialogNodes
}

func CreateCatPackNodes(db *gorm.DB, packHandler bot.HandlerFunc) []dialog.Node {
	packNodes := make([]dialog.Node, 0)

	var categories []models.Category

	models.GetActiveCategories(db, &categories)
	catNode := dialog.Node{
		ID:       "categories",
		Text:     "یکی از دسته بندی های زیر را انتخاب کنید",
		Keyboard: make([][]dialog.Button, (len(categories) + 1)),
	}
	catNode.Keyboard[len(categories)] = make([]dialog.Button, 1)
	catNode.Keyboard[len(categories)][0] = dialog.Button{
		Text:   "بازگشت",
		NodeID: "start",
	}

	for i, category := range categories {
		catNode.Keyboard[i] = make([]dialog.Button, 1)

		strCatID := "cat_" + strconv.FormatUint(uint64(category.ID), 10)
		catNode.Keyboard[i][0] = dialog.Button{
			Text:   category.Name,
			NodeID: strCatID,
		}

		var packs []models.Pack
		models.GetActivePacksByCatID(db, &packs, category.ID)

		packPeriods := models.GetPackPeriods(packs)
		periodsNode := dialog.Node{
			ID:       strCatID,
			Text:     "مدت مورد نظر بسته خود را انتخاب کنید",
			Keyboard: make([][]dialog.Button, 0),
		}

		packsNodes := make([]dialog.Node, 0)

		for period := range packPeriods {
			packNodeID := fmt.Sprintf("cat_%d_%d", category.ID, period)

			row := []dialog.Button{
				{
					Text:   models.StringPeriod(period),
					NodeID: packNodeID,
				},
			}
			periodsNode.Keyboard = append(periodsNode.Keyboard, row)

			packNode := dialog.Node{
				ID:       packNodeID,
				Text:     bot.EscapeMarkdown(fmt.Sprintf("لطفا بسته مورد نظر خود را انتخاب کنید.\nدسته بندی: %s\nتوضیحات: %s\nمدت: %d روزه", category.Name, category.Description, period)),
				Keyboard: make([][]dialog.Button, 0),
			}

			for _, pack := range packPeriods[period] {
				strPackID := strconv.FormatUint(uint64(pack.ID), 10)
				btnRow := []dialog.Button{{
					ID:              "pack_" + strPackID,
					Text:            pack.String(),
					CallbackHandler: packHandler,
					CallbackData:    strPackID,
				}}
				packNode.Keyboard = append(packNode.Keyboard, btnRow)
			}
			backBtnRow := []dialog.Button{{
				Text:   "بازگشت",
				NodeID: strCatID,
			}}
			packNode.Keyboard = append(packNode.Keyboard, backBtnRow)

			packsNodes = append(packsNodes, packNode)
		}

		backBtnRow := []dialog.Button{{
			Text:   "بازگشت",
			NodeID: "categories",
		}}
		periodsNode.Keyboard = append(periodsNode.Keyboard, backBtnRow)
		packNodes = append(packNodes, periodsNode)
		packNodes = append(packNodes, packsNodes...)
	}

	return append(packNodes, catNode)
}
