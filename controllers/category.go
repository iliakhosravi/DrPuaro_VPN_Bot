package controllers

import (
	"context"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/keyboard/reply"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/middlewares"
	"techybat.org/go-vpn/models"
)

type CategoryKey string

const CAT_KEY CategoryKey = "category-key"

var cancelBtnText string = "انصراف"

func AddCatController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	category := &models.Category{}
	var handlerID string
	handlerID = b.RegisterHandlerMatchFunc(checkUserMatch(update, cancelBtnText), passCategory(passHandlerID(catNameController, &handlerID), category))

	cancelReplyKeyboard := reply.New(
		b,
		reply.WithPrefix("cancel_order_keyboard"),
	).Button(cancelBtnText, b, bot.MatchTypeExact, passHandlerID(onCancelCat, &handlerID))

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		Text:        "نام دسته بندی چه باشد؟",
		ReplyMarkup: cancelReplyKeyboard,
	})

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        "افزودن دسته بندی (2 مرحله)",
		ReplyMarkup: nil,
	})

}

func catNameController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	handlerID := ctx.Value(middlewares.HandlerID).(*string)

	category := ctx.Value(CAT_KEY).(*models.Category)
	category.Name = update.Message.Text

	b.UnregisterHandler(*handlerID)
	*handlerID = b.RegisterHandlerMatchFunc(checkUserMatch(update, cancelBtnText), passCategory(passHandlerID(catDescController, handlerID), category))

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "توضیحات دسته بندی چه باشد؟",
	})
}

func catDescController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	handlerID := ctx.Value(middlewares.HandlerID).(*string)

	category := ctx.Value(CAT_KEY).(*models.Category)
	category.Description = update.Message.Text

	b.UnregisterHandler(*handlerID)

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
	handlerID := ctx.Value(middlewares.HandlerID).(*string)
	b.UnregisterHandler(*handlerID)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "پروسه لغو شد",
	})

	ShowAdminDialog(ctx, b, update)
}

func passHandlerID(next bot.HandlerFunc, handlerID *string) func(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	return func(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
		ctx = context.WithValue(ctx, middlewares.HandlerID, handlerID)

		next(ctx, b, update)
	}
}

func passCategory(next bot.HandlerFunc, category *models.Category) func(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	return func(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
		ctx = context.WithValue(ctx, CAT_KEY, category)

		next(ctx, b, update)
	}
}

func checkUserMatch(update *tmodels.Update, cancelBtnText string) bot.MatchFunc {
	return func(checkUpdate *tmodels.Update) bool {
		if checkUpdate.Message == nil || checkUpdate.Message.Text == cancelBtnText {
			return false
		}
		messageChatID := checkUpdate.Message.Chat.ID
		switch {
		case update.CallbackQuery != nil:
			return messageChatID == update.CallbackQuery.From.ID
		case update.Message != nil:
			return messageChatID == update.Message.Chat.ID
		default:
			return false
		}
	}
}
