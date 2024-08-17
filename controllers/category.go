package controllers

import (
	"context"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/models"
	"techybat.org/go-vpn/widgets/form"
)

var cancelBtnText string = "انصراف"

func AddCatController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	fields := []form.Field{
		{
			Name:        "cat_name",
			MessageText: "نام دسته بندی چه باشد؟",
		},
		{
			Name:        "cat_description",
			MessageText: "توضیحات دسته بندی چه باشد؟",
		},
	}
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	userID := update.CallbackQuery.From.ID
	form := form.CreateForm(cancelBtnText, fields, chatID, userID, catSubmitController, onCancelCat)

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        "افزودن دسته بندی (2 مرحله)",
		ReplyMarkup: nil,
	})

	form.Show(ctx, b, update)
}

func catSubmitController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)

	category := models.Category{
		Name:        form.FindField("cat_name").Value,
		Description: form.FindField("cat_description").Value,
	}

	txtMsg := "افزودن دسته بندی با موفقیت انجام شد"
	if err := category.Store(database.GetDB()); err != nil {
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   txtMsg,
	})

	ShowAdminDialog(ctx, b, update)
}

func onCancelCat(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	ShowAdminDialog(ctx, b, update)
}
