package admin_controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	ptime "github.com/yaa110/go-persian-calendar"
	"techybat.org/go-vpn/database"
	m "techybat.org/go-vpn/models"
	bp "techybat.org/go-vpn/widgets/buttonpage"
	"techybat.org/go-vpn/widgets/form"
)

type onEditOrder func(ctx context.Context, b *bot.Bot, update *tmodels.Update, order m.Order)

func OrdersHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	orderType := update.CallbackQuery.Data
	db := database.GetDB()
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	var orders []m.Order
	db.Where("type = ?", orderType).Preload("Pack").Preload("Pack.Category").Preload("User").Find(&orders)

	buttons := createOrderButtons(orders)

	buttonPage := bp.CreateButtonPage(bot.EscapeMarkdown("لیست بسته های خواسته شده:"), buttons, 5, true)
	if _, err := buttonPage.Show(ctx, b, chatID); err != nil {
		fmt.Println("Error showing orders handler:", err)
	}
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

	var txtMsg string
	if order.Type == m.ActiveOrder {
		config := order.Config(db)
		pt := ptime.New(config.StartDate)
		showDate := pt.Format("d MMM y")
		txtMsg = fmt.Sprintf("شماره سفارش:%d\nوضعیت سفارش:%s\nگروه بسته:%s\nنوع بسته:%s\nتاریخ شروع:%v\nتوضیحات:%s", order.ID, order.Type, order.Pack.Category.Name, order.Pack, showDate, config.Link)
	} else {
		pt := ptime.New(order.CreatedAt)
		showDate := pt.Format("d MMM y")
		txtMsg = fmt.Sprintf("شماره سفارش:%d\nوضعیت سفارش:%s\nگروه بسته:%s\nنوع بسته:%s\nتاریخ درخواست:%v\nتوضیحات ادمین:%s", order.ID, order.Type, order.Pack.Category.Name, order.Pack, showDate, order.AdminNote)
	}

	txtMsg = bot.EscapeMarkdown(txtMsg)

	nodes := []dialog.Node{
		{
			ID:   "manage-orders",
			Text: txtMsg,
			Keyboard: [][]dialog.Button{
				{
					{
						Text:   "اعلام اتمام حجم یا دوره",
						NodeID: "deplete-order",
					},
					{
						Text:   "رد کردن",
						NodeID: "dismiss-order",
					},
				},
				{
					{
						Text:   "فعال سازی",
						NodeID: "activate-order",
					},
					{
						ID:              "admin-note-order",
						Text:            "تغییر توضیحات",
						CallbackHandler: editNoteOrderHandler,
						CallbackData:    orderID,
					},
				},
			},
		},
		{
			ID:   "deplete-order",
			Text: "آیا از اعلام اتمام حجم یا دوره برای این کاربر مطمئن هستید؟",
			Keyboard: [][]dialog.Button{
				{
					{
						ID:              "deplete",
						Text:            "بله",
						CallbackHandler: changeOrderTypeHandler,
						CallbackData:    orderID + "_" + m.DepletedOrder,
					},
					{
						Text:   "برگشت",
						NodeID: "manage-orders",
					},
				},
			},
		},
		{
			ID:   "dismiss-order",
			Text: "آیا از رد کردن این سفارش مطمئن هستید؟",
			Keyboard: [][]dialog.Button{
				{
					{
						ID:              "dismiss",
						Text:            "بله",
						CallbackHandler: changeOrderTypeHandler,
						CallbackData:    orderID + "_" + m.DismissedOrder,
					},
					{
						Text:   "برگشت",
						NodeID: "manage-orders",
					},
				},
			},
		},
		{
			ID:   "activate-order",
			Text: "آیا از فعال سازی این سفارش مطمئن هستید؟",
			Keyboard: [][]dialog.Button{
				{
					{
						ID:              "activate",
						Text:            "بله",
						CallbackHandler: changeOrderTypeHandler,
						CallbackData:    orderID + "_" + m.ActiveOrder,
					},
					{
						Text:   "برگشت",
						NodeID: "manage-orders",
					},
				},
			},
		},
	}

	dialog := dialog.New(nodes, dialog.Inline())

	if _, err := dialog.Show(ctx, b, chatID, "manage-orders"); err != nil {
		fmt.Println("Error cannot show manage orders", err)
	}
}

func editNoteOrderHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	orderID := update.CallbackQuery.Data
	var order m.Order
	database.GetDB().Find(&order, orderID)
	fields := []form.Field{
		{
			Name:        "admin_note",
			MessageText: fmt.Sprintf("لطفا توضیحات خود را برای این سفارش قرار دهید.\nمقدار فعلی: %s", order.AdminNote),
			Value:       order.AdminNote,
		},
	}
	form := form.CreateForm("انصراف", fields, chatID, update.CallbackQuery.From.ID, passOrder(onEditNoteOrder, order), onCancelEditNote, nil)
	form.Show(ctx, b, update)
}

func onCancelEditNote(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
}

func onEditNoteOrder(ctx context.Context, b *bot.Bot, update *tmodels.Update, order m.Order) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)
	order.AdminNote = form.FindField("admin_note").Value
	db := database.GetDB()

	var txtMsg string
	if result := db.Save(&order); result.Error != nil {
		txtMsg = "خطایی پیش آمده"
		fmt.Println("Error onEditNoteOrder saving order: ", result.Error)
	}

	txtMsg = "توضیحات سفارش با موفقیت ویرایش شد."

	chatID := update.Message.Chat.ID
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   txtMsg,
	})
}

func passOrder(next onEditOrder, order m.Order) bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *tmodels.Update) {
		next(ctx, bot, update, order)
	}
}

func changeOrderTypeHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	data := strings.Split(update.CallbackQuery.Data, "_")
	orderID := data[0]
	var orderType m.OrderType = m.OrderType(data[1])
	db := database.GetDB()
	var order m.Order
	db.Preload("User").Preload("Pack").Preload("Pack.Category").Find(&order, orderID)
	if err := order.ChangeType(db, orderType); err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "خطایی پیش آمده",
		})
	}

	pt := ptime.New(order.CreatedAt)
	showDate := pt.Format("d MMM y")
	txtMsg := fmt.Sprintf("سفارش شما با مشخصات زیر تغییر وضعیت یافته است:\n\nشماره سفارش:%d\nوضعیت سفارش:%s\nگروه بسته:%s\nنوع بسته:%s\nتاریخ درخواست:%v\nتوضیحات ادمین:%s", order.ID, order.Type, order.Pack.Category.Name, order.Pack, showDate, order.AdminNote)

	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: order.User.TelID,
		Text:   txtMsg,
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "تغییر وضعیت بسته به کاربر اعلام و در سیستم ثبت شد",
	})
}

func createOrderButtons(orders []m.Order) []dialog.Button {
	buttons := []dialog.Button{}
	for _, order := range orders {
		buttons = append(buttons, dialog.Button{
			ID:              fmt.Sprintf("order%d", order.ID),
			Text:            fmt.Sprintf("#%d | %s | o(%s)o | %s", order.ID, order.User.Fullname(), order.Pack.Category.Name, order.Pack.Name()),
			CallbackHandler: showOrderHandler,
			CallbackData:    fmt.Sprint(order.ID),
		})
	}
	return buttons
}
