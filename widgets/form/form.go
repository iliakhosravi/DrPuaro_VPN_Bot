package form

import (
	"context"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/keyboard/reply"
	"techybat.org/go-vpn/tools/escaper"
)

type FormKey string

const FORM_KEY FormKey = "form-key"

type Form struct {
	CancelButtonText string
	Fields           []Field
	FieldIndex       int
	HandlerID        string
	ChatID           int64
	UserID           int64
	SubmitHandler    bot.HandlerFunc
	CancelHandler    bot.HandlerFunc
}

type Field struct {
	Name        string
	MessageText string
	Value       string
}

func CreateForm(cancelBtnText string, fields []Field, chatID, userID int64, submitHandler, cancelHandler bot.HandlerFunc) *Form {
	form := &Form{
		CancelButtonText: cancelBtnText,
		Fields:           fields,
		FieldIndex:       0,
		ChatID:           chatID,
		UserID:           userID,
		SubmitHandler:    submitHandler,
		CancelHandler:    cancelHandler,
	}

	return form
}

func (form *Form) Show(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	form.HandlerID = b.RegisterHandlerMatchFunc(form.checkUserMatch(), form.fieldHandler)

	cancelReplyKeyboard := reply.New(
		b,
		reply.WithPrefix("cancel_order_keyboard"),
	).Button(form.CancelButtonText, b, bot.MatchTypeExact, form.onCancel)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      form.ChatID,
		Text:        form.Fields[0].MessageText,
		ReplyMarkup: cancelReplyKeyboard,
	})
}

func (form *Form) fieldHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	field := &form.Fields[form.FieldIndex]
	field.Value = escaper.EscapeToMark(update.Message.Text)

	form.FieldIndex++
	b.UnregisterHandler(form.HandlerID)

	if form.FieldIndex == len(form.Fields) {
		ctx = context.WithValue(ctx, FORM_KEY, form)
		form.SubmitHandler(ctx, b, update)
		return
	}

	form.HandlerID = b.RegisterHandlerMatchFunc(form.checkUserMatch(), form.fieldHandler)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   form.Fields[form.FieldIndex].MessageText,
	})
}

func (form *Form) onCancel(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	b.UnregisterHandler(form.HandlerID)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "پروسه لغو شد",
	})

	form.CancelHandler(ctx, b, update)
}

func (form *Form) checkUserMatch() bot.MatchFunc {
	return func(checkUpdate *tmodels.Update) bool {
		if checkUpdate.Message == nil || checkUpdate.Message.Text == form.CancelButtonText {
			return false
		}
		messageChatID := checkUpdate.Message.Chat.ID
		return messageChatID == form.ChatID && form.UserID == checkUpdate.Message.From.ID
	}
}

func (form Form) FindField(name string) Field {
	for _, field := range form.Fields {
		if field.Name == name {
			return field
		}
	}
	return Field{}
}
