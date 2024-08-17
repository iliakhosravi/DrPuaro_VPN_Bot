package form

import (
	"context"
	"strings"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/keyboard/reply"
)

type FormKey string
type FieldType int

const FORM_KEY FormKey = "form-key"
const (
	TextField FieldType = iota
	ButtonField
)

type Form struct {
	CancelButtonText string
	Fields           []Field
	FieldIndex       int
	HandlerID        string
	ChatID           int64
	UserID           int64
	SubmitHandler    bot.HandlerFunc
	CancelHandler    bot.HandlerFunc
	ButtonPrefix     string
	DefaultValidator Validator
}

type Field struct {
	Name        string
	MessageText string
	Value       string
	Type        FieldType
	Keyboard    [][]tmodels.InlineKeyboardButton
	Validator   Validator
	Filter      Filter
}

type Filter func(value string) string
type Validator func(value string) (bool, string)

func DefaultValidator(value string) (bool, string) {
	return value != "", "error"
}

func CreateForm(cancelBtnText string, fields []Field, chatID, userID int64, submitHandler, cancelHandler bot.HandlerFunc, defaultValidator Validator) *Form {
	if defaultValidator == nil {
		defaultValidator = DefaultValidator
	}

	form := &Form{
		CancelButtonText: cancelBtnText,
		Fields:           fields,
		FieldIndex:       0,
		ChatID:           chatID,
		UserID:           userID,
		SubmitHandler:    submitHandler,
		CancelHandler:    cancelHandler,
		ButtonPrefix:     bot.RandomString(16),
		DefaultValidator: defaultValidator,
	}

	form.initFields()

	return form
}

func (form *Form) initFields() {
	for i, field := range form.Fields {
		if field.Validator == nil {
			form.Fields[i].Validator = form.DefaultValidator
		}
	}
}

func (form *Form) Show(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	form.HandlerID = b.RegisterHandlerMatchFunc(form.checkUserMatch(form.Fields[0].Type == ButtonField), form.fieldHandler)

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
	var value string
	if field.Type == ButtonField {
		value = strings.TrimPrefix(update.CallbackQuery.Data, form.ButtonPrefix)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
	} else {
		value = update.Message.Text
	}
	if ok, errMsg := field.SetValue(value); !ok {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: form.ChatID,
			Text:   errMsg,
		})
		return
	}

	form.FieldIndex++
	b.UnregisterHandler(form.HandlerID)
	if field.Type == ButtonField {
		b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
			ChatID:      form.ChatID,
			MessageID:   update.CallbackQuery.Message.Message.ID,
			ReplyMarkup: nil,
		})
	}

	if form.FieldIndex == len(form.Fields) {
		ctx = context.WithValue(ctx, FORM_KEY, form)
		form.SubmitHandler(ctx, b, update)
		return
	}

	nextField := &form.Fields[form.FieldIndex]
	if nextField.Type == ButtonField {
		form.HandlerID = b.RegisterHandler(bot.HandlerTypeCallbackQueryData, form.ButtonPrefix, bot.MatchTypePrefix, form.fieldHandler)
	} else {
		form.HandlerID = b.RegisterHandlerMatchFunc(form.checkUserMatch(form.Fields[form.FieldIndex].Type == ButtonField), form.fieldHandler)
	}

	params := &bot.SendMessageParams{
		ChatID: form.ChatID,
		Text:   nextField.MessageText,
	}

	if nextField.Type == ButtonField {
		params.ReplyMarkup = nextField.buildKB(form.ButtonPrefix)
	}

	b.SendMessage(ctx, params)
}

func (field *Field) buildKB(prefix string) *tmodels.InlineKeyboardMarkup {
	if field.Type != ButtonField {
		return nil
	}
	for i := 0; i < len(field.Keyboard); i++ {
		for j := 0; j < len(field.Keyboard[i]); j++ {
			field.Keyboard[i][j].CallbackData = prefix + field.Keyboard[i][j].CallbackData
		}
	}
	return &tmodels.InlineKeyboardMarkup{
		InlineKeyboard: field.Keyboard,
	}
}

func (form *Form) onCancel(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	b.UnregisterHandler(form.HandlerID)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "پروسه لغو شد",
	})

	form.CancelHandler(ctx, b, update)
}

func (form *Form) checkUserMatch(isCallback bool) bot.MatchFunc {
	return func(checkUpdate *tmodels.Update) bool {
		if isCallback {
			if checkUpdate.CallbackQuery == nil {
				return false
			}
			messageChatID := checkUpdate.CallbackQuery.Message.Message.Chat.ID
			return messageChatID == form.ChatID && form.UserID == checkUpdate.CallbackQuery.From.ID
		}

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

func (field *Field) SetValue(value string) (bool, string) {
	ok, errMsg := field.Validator(value)
	if !ok {
		return ok, errMsg
	}
	if field.Filter != nil {
		field.Value = field.Filter(value)
	} else {
		field.Value = value
	}
	return ok, errMsg
}
