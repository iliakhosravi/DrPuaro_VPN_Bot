package controllers

import (
	"context"
	"strconv"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/models"
	"techybat.org/go-vpn/widgets/form"
)

func AddPackController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	var categories []models.Category
	db.Find(&categories)
	var catKeyboard = [][]tmodels.InlineKeyboardButton{}

	for _, category := range categories {
		catKeyboard = append(catKeyboard, []tmodels.InlineKeyboardButton{
			{
				Text:         category.Name,
				CallbackData: strconv.FormatUint(uint64(category.ID), 10),
			},
		})
	}

	fields := []form.Field{
		{
			Name:        "traffic",
			MessageText: "میزان حجم بسته بر حسب گیگ چقدر باشد؟ لطفا صرفا عدد صحیح مثبت وارد نمایید.",
			Validator:   models.PackValidator("traffic"),
		},
		{
			Name:        "period",
			MessageText: "دوره زمانی بسته بر حسب روز چقدر است؟ لطفا صرفا عدد صحیح مثبت وارد نمایید",
			Validator:   models.PackValidator("period"),
		},
		{
			Name:        "price",
			MessageText: "قیمت بسته برحسب تومان چقدر است؟ لطفا صرفا عدد صحیح مثبت وارد نمایید.",
			Validator:   models.PackValidator("price"),
		},
		{
			Name:        "category",
			MessageText: "یکی از دسته بندی های زیر را برای بسته موردنظر انتخاب کنید",
			Type:        form.ButtonField,
			Keyboard:    catKeyboard,
			Validator:   models.PackValidator("category_id"),
		},
	}
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	userID := update.CallbackQuery.From.ID
	form := form.CreateForm(cancelBtnText, fields, chatID, userID, packSubmitController, onCancelPack, nil)

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        "افزودن بسته (4 مرحله)",
		ReplyMarkup: nil,
	})

	form.Show(ctx, b, update)
}

func packSubmitController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)
	traffic, _ := strconv.Atoi(form.FindField("traffic").Value)
	period, _ := strconv.Atoi(form.FindField("period").Value)
	price, _ := strconv.Atoi(form.FindField("price").Value)
	categoryID, _ := strconv.ParseUint(form.FindField("category").Value, 10, 0)

	pack := models.Pack{
		Traffic:    traffic,
		Period:     period,
		Price:      price,
		CategoryID: uint(categoryID),
	}

	txtMsg := "افزودن بسته با موفقیت انجام شد"

	if err := pack.Store(database.GetDB()); err != nil {
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: form.ChatID,
		Text:   txtMsg,
	})

	ShowAdminDialog(ctx, b, update, form.ChatID)
}

func onCancelPack(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	var chatID int64
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Message.Chat.ID
	} else {
		chatID = update.Message.Chat.ID
	}
	ShowAdminDialog(ctx, b, update, chatID)
}
