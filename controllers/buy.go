package controllers

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"github.com/sinasadeghi83/go-telegram-bot-ui/keyboard/reply"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/middlewares/auth"
	"techybat.org/go-vpn/models"
	"techybat.org/go-vpn/widgets/buttonpage"
	"techybat.org/go-vpn/widgets/form"
)

func BuyController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	user := ctx.Value(auth.UserKey).(models.User)
	// answering callback query first to let Telegram know that we received the callback query,
	// and we're handling it. Otherwise, Telegram might retry sending the update repetitively
	// as it thinks the callback query doesn't reach to our application. learn more by
	// reading the footnote of the https://core.telegram.org/bots/api#callbackquery type.
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	var card models.Card
	txtMsg := "جهت پرداخت مبلغ ذکر شده را به شماره کارت زیر واریز کرده و سپس تصویری از فیش واریزی را در یک پیام ارسال کنید. پس از این مرحله خرید شما در وضعیت نیاز به تایید قرار گرفته و با تایید نهایی از سوی ادمین کانفیگ به صورت خودکار برای شما ارسال خواهد شد."
	if result := db.First(&card); result.RowsAffected > 0 {
		txtMsg = fmt.Sprintf("%s\nشماره کارت: %s\nبه نام: %s", txtMsg, card.Number, card.Fullname)
	} else {
		txtMsg = "خطایی پیش آمده"
	}

	packId, _ := strconv.Atoi(update.CallbackQuery.Data)
	var pack models.Pack
	var packMsg string
	if result := db.First(&pack, packId); result.RowsAffected == 0 {
		txtMsg = "خطایی پیش آمده"
		packMsg = "این بسته در دسترس نمی باشد"
	} else {
		packMsg = fmt.Sprintf("شما بسته %s را برای خرید انتخاب کرده اید:", pack.String())
	}
	cancelBtnText := "انصراف از خرید"

	checkUser := func(checkUpdate *tmodels.Update) bool {
		if checkUpdate.Message == nil || checkUpdate.Message.Text == cancelBtnText {
			return false
		}
		return checkUpdate.Message.Chat.ID == update.CallbackQuery.From.ID
	}
	handlerID := b.RegisterHandlerMatchFunc(checkUser, RetrieveUserReceipt)

	order := models.Order{
		UserID:    user.ID,
		PackID:    pack.ID,
		Type:      models.SentOrder,
		HandlerID: handlerID,
	}

	if err := order.CreateOrder(db); err != nil {
		txtMsg = "خطایی پیش آمده"
		b.UnregisterHandler(handlerID)
	}

	cancelReplyKeyboard := reply.New(
		b,
		reply.WithPrefix("cancel_order_keyboard"),
	).Button(cancelBtnText, b, bot.MatchTypeExact, onCancelbuy)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		Text:        txtMsg,
		ReplyMarkup: cancelReplyKeyboard,
	})

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        packMsg,
		ReplyMarkup: nil,
	})

}

func RetrieveUserReceipt(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()

	user := ctx.Value(auth.UserKey).(models.User)

	msg, err := b.ForwardMessage(ctx, &bot.ForwardMessageParams{
		ChatID:     os.Getenv("STORAGE_CHANNEL_ID"),
		FromChatID: fmt.Sprintf("%d", update.Message.Chat.ID),
		MessageID:  update.Message.ID,
	})

	if err != nil {
		fmt.Println("unable to forward receipt to storage channel\nerror:", err)
		return
	}

	var txtMsg string
	var order models.Order

	if err := user.FirstSentOrder(db, &order); err != nil {
		txtMsg = "خطایی پیش آمده"
	}

	receipt, err := order.AddReceiptByMsg(db, msg)
	if err != nil {
		txtMsg = "خطایی پیش آمده"
	} else {
		txtMsg = fmt.Sprintf("درخواست شما با موفقیت ثبت شد. کد رسید: %d", receipt.ID)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   txtMsg,
	})

	b.UnregisterHandler(order.HandlerID)

	ShowMainDialog(ctx, b, update)
}

func onCancelbuy(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	user := ctx.Value(auth.UserKey).(models.User)

	txtMsg := "خرید شما با موفقیت لغو شد"
	var order models.Order

	if err := user.FirstSentOrder(db, &order); err != nil {
		txtMsg = "خطایی پیش آمده"
	}

	if err := order.CancelOrder(db); err != nil {
		txtMsg = "خطایی پیش آمده"
	}

	b.UnregisterHandler(order.HandlerID)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   txtMsg,
	})

	ShowMainDialog(ctx, b, update)
}

func VerifyBuyController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	var orders []models.Order
	db.Where(models.Order{Type: models.PendingOrder}).Preload("Pack").Find(&orders)

	orderBtns := []dialog.Button{}

	for _, order := range orders {
		orderBtns = append(orderBtns, dialog.Button{
			ID:              fmt.Sprintf("Order%d", order.ID),
			Text:            order.Pack.String(),
			CallbackHandler: onWatchOrder,
			CallbackData:    strconv.FormatUint(uint64(order.ID), 10),
		})
	}

	buttonPage := buttonpage.CreateButtonPage("لیست درخواست های ارسالی:", orderBtns, 5, true)

	buttonPage.Show(ctx, b, update.CallbackQuery.Message.Message.Chat.ID)
}

func onWatchOrder(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()

	orderID, _ := strconv.ParseUint(update.CallbackQuery.Data, 10, 0)
	var order models.Order
	db.Where(orderID).Preload("User").Preload("Pack").Preload("Pack.Category").Find(&order)

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	_, err := b.ForwardMessage(ctx, &bot.ForwardMessageParams{
		ChatID:     chatID,
		FromChatID: os.Getenv("STORAGE_CHANNEL_ID"),
		MessageID:  order.GetReceipt(db).MessageID,
	})

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "خطایی پیش آمده یا درخواست ارسالی دردسترس نمی باشد",
		})
	}

	txtMsg := fmt.Sprintf("شماره رسید:%d\nنام کاربر:%s\nآیدی کاربر:%d\nنام کاربری:%s\n\nبسته درخواستی:%s\nگروه بسته درخواستی:%s\n",
		order.GetReceipt(db).ID,
		order.User.Fullname(),
		order.User.TelID,
		order.User.Username,
		order.Pack.String(),
		order.Pack.Category.Name,
	)

	fields := []form.Field{
		{
			Name:        "order",
			MessageText: txtMsg,
			Type:        form.ButtonField,
			Keyboard: [][]tmodels.InlineKeyboardButton{
				{
					{
						Text:         "تایید✅",
						CallbackData: "true_" + update.CallbackQuery.Data,
					},
					{
						Text:         "رد🚫",
						CallbackData: "false_" + update.CallbackQuery.Data,
					},
				},
			},
		},

		{
			Name:        "carry_msg",
			MessageText: "در جواب چه پیامی به کاربر ارسال شود؟",
			Type:        form.TextField,
		},
	}
	form := form.CreateForm("انصراف از ادامه پروسه", fields, chatID, update.CallbackQuery.From.ID, onSubmitOrder, onCancelOrder, nil)
	form.Description = "شروع"
	form.Show(ctx, b, update)
}

func onSubmitOrder(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)
	db := database.GetDB()
	orderField := form.FindField("order").Value
	splitedOrderField := strings.Split(orderField, "_")
	ok, strOrderID := splitedOrderField[0], splitedOrderField[1]
	orderID, _ := strconv.ParseUint(strOrderID, 10, 0)
	var order models.Order
	db.Preload("User").Preload("Pack").First(&order, orderID)

	var txtMsg, orderResult string
	if ok == "true" {
		err := order.Verify(db, form.FindField("carry_msg").Value)
		if err != nil {
			txtMsg = "خطایی پیش آمده"
		} else {
			txtMsg = fmt.Sprintf("سفارش کاربر تایید شد. کد درخواست: %d", orderID)
			orderResult = "<b>تایید شده✅</b>"
		}
	} else {
		err := order.Dismiss(db)
		if err != nil {
			txtMsg = "خطایی پیش آمده"
		} else {
			txtMsg = fmt.Sprintf("سفارش کاربر رد شد. کد درخواست: %d", orderID)
			orderResult = "<b>رد شده❌</b>"
		}
	}

	carryMsg := fmt.Sprintf("سفارش شما بررسی شد.\nبسته انتخابی:%s\nشماره سفارش: %d\nنتیجه:%s\nتوضیحات ادمین:%s", order.Pack.String(), orderID, orderResult, form.FindField("carry_msg").Value)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    order.User.TelID,
		Text:      carryMsg,
		ParseMode: tmodels.ParseModeHTML,
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: form.ChatID,
		Text:   txtMsg,
	})

	ShowAdminDialog(ctx, b, update, form.ChatID)
}

func onCancelOrder(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	var chatID int64
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Message.Chat.ID
	} else {
		chatID = update.Message.Chat.ID
	}
	ShowAdminDialog(ctx, b, update, chatID)
}
