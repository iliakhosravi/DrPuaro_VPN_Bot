package controllers

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/keyboard/reply"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/middlewares/auth"
	"techybat.org/go-vpn/models"
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
