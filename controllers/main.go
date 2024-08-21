package controllers

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"gorm.io/gorm"
	customerController "techybat.org/go-vpn/controllers/customer"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/models"
)

func MainController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	if update.CallbackQuery != nil {
		return
	}
	ShowMainDialog(ctx, b, update)
}

func ShowMainDialog(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
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

func NewMainDialog() []dialog.Node {
	db := database.GetDB()
	dialogNodes := []dialog.Node{
		{
			ID:   "start",
			Text: "☄️ Ultra Fast VPN☄️\n",
			Keyboard: [][]dialog.Button{
				{
					{Text: "خرید بسته", NodeID: "categories"},
					{Text: "بسته های خریداری شده", NodeID: "orders"},
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
	dialogNodes = append(dialogNodes, CreateCatPackNodes(db, BuyController)...)

	return dialogNodes
}

func CreateGuideKeyboard(db *gorm.DB) [][]dialog.Button {
	var guides []models.Guide
	db.Find(&guides)

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

func ForwardGuideHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	msgID, _ := strconv.Atoi(update.CallbackQuery.Data)
	b.ForwardMessage(ctx, &bot.ForwardMessageParams{
		FromChatID: os.Getenv("STORAGE_CHANNEL_ID"),
		ChatID:     update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:  msgID,
	})
}

func CreateCatPackNodes(db *gorm.DB, packHandler bot.HandlerFunc) []dialog.Node {
	packNodes := make([]dialog.Node, 0)

	var categories []models.Category
	db.Find(&categories)
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
		db.Find(&packs, models.Pack{CategoryID: category.ID})

		packsNode := dialog.Node{
			ID:       strCatID,
			Text:     bot.EscapeMarkdown(fmt.Sprintf("دسته بندی:%s\nتوضیحات:%s\nبسته مورد نظر خود را انتخاب کنید", category.Name, category.Description)),
			Keyboard: make([][]dialog.Button, len(packs)+1),
		}

		packsNode.Keyboard[len(packs)] = make([]dialog.Button, 1)
		packsNode.Keyboard[len(packs)][0] = dialog.Button{
			Text:   "بازگشت",
			NodeID: "categories",
		}

		for j, pack := range packs {
			packsNode.Keyboard[j] = make([]dialog.Button, 1)

			strPackID := strconv.FormatUint(uint64(pack.ID), 10)
			packsNode.Keyboard[j][0] = dialog.Button{
				ID:              "pack_" + strPackID,
				Text:            pack.String(),
				CallbackHandler: packHandler,
				CallbackData:    strPackID,
			}
		}

		packNodes = append(packNodes, packsNode)
	}

	return append(packNodes, catNode)
}
