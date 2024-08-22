package admin_controller

import (
	"context"

	"github.com/go-telegram/bot"
	tm "github.com/go-telegram/bot/models"
	"techybat.org/go-vpn/database"
	m "techybat.org/go-vpn/models"
	"techybat.org/go-vpn/widgets/form"
)

func AddCardHandler(ctx context.Context, b *bot.Bot, update *tm.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	userID := update.CallbackQuery.From.ID
	fields := []form.Field{
		{
			Name:        "fullname",
			MessageText: "لطفا نام صاحب کارت را کامل وارد نمایید",
		},
		{
			Name:        "card_number",
			MessageText: "لطفا شماره کارت را وارد نمایید",
		},
	}
	form := form.CreateForm("انصراف", fields, chatID, userID, onSubmitAddCard, onCancelAddCard, nil)
	form.Show(ctx, b, update)
}

func onSubmitAddCard(ctx context.Context, b *bot.Bot, update *tm.Update) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)

	card := m.Card{
		Fullname: form.FindField("fullname").Value,
		Number:   form.FindField("card_number").Value,
	}

	db := database.GetDB()
	txtMsg := "افزودن کارت با موفقیت انجام شد. این کارت به عنوان کارت پیش فرض شما برای پرداخت به مشتریان نمایش داده خواهد شد."
	if err := m.AddNewCard(db, &card); err != nil {
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: form.ChatID,
		Text:   txtMsg,
	})
}

func onCancelAddCard(ctx context.Context, b *bot.Bot, update *tm.Update) {
}
