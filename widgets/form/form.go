package form

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
)

type FormKey string
type FieldType int

const (
	CANCEL_DATA = "cancel"
	SKIP_DATA   = "skip"
)
const FORM_KEY FormKey = "form-key"
const (
	TextField FieldType = iota
	ButtonField
	CustomTextField
)

type Form struct {
	CancelButtonText string
	SkipButtonText   string
	SkipMessageText  string
	Fields           []Field
	FieldIndex       int
	HandlerID        string
	ChatID           int64
	UserID           int64
	SubmitHandler    bot.HandlerFunc
	CancelHandler    bot.HandlerFunc
	ButtonPrefix     string
	DefaultValidator Validator
	Description      string
	ManageHandlerID  string
	ManageBtnPrefix  string
	lastMsgID        int
}

type Field struct {
	Name          string
	MessageText   string
	Value         string
	Type          FieldType
	Keyboard      [][]tmodels.InlineKeyboardButton
	Validator     Validator
	Filter        Filter
	IsSkippable   bool
	CustomHandler CustomHandler
	ParseMode     tmodels.ParseMode
}

type Filter func(value string) string
type Validator func(value string) (bool, string)
type CustomHandler func(ctx context.Context, b *bot.Bot, update *tmodels.Update, form Form, setter FieldSetter) (bool, error)
type FieldSetter func(value string) (bool, string)

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
		ManageBtnPrefix:  bot.RandomString(16),
		DefaultValidator: defaultValidator,
		SkipButtonText:   "Skip",
		SkipMessageText:  "Input Skipped",
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
	if form.Description != "" {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: form.ChatID,
			Text:   form.Description,
		})
	}

	form.ManageHandlerID = b.RegisterHandler(bot.HandlerTypeCallbackQueryData, form.ManageBtnPrefix, bot.MatchTypePrefix, form.onManage)

	form.loadNextField(ctx, b, update)
}

func (form *Form) fieldHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	field := form.currentField()
	var value string
	switch field.Type {
	case ButtonField:
		value = strings.TrimPrefix(update.CallbackQuery.Data, form.ButtonPrefix)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
	case TextField:
		value = update.Message.Text
	}

	if ok, errMsg := field.SetValue(value); !ok {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: form.ChatID,
			Text:   errMsg,
		})
		return
	}

	form.loadNextField(ctx, b, update)
}

func (form *Form) customFieldHandler(next CustomHandler) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
		field := form.currentField()
		ok, err := next(ctx, b, update, *form, field.SetValue)
		if !ok {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: form.ChatID,
				Text:   fmt.Sprintf("%s", err),
			})
			return
		}
		form.loadNextField(ctx, b, update)
	}
}

func (form *Form) loadNextField(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	b.UnregisterHandler(form.HandlerID)

	if !form.isFirstField() {
		b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
			ChatID:      form.ChatID,
			MessageID:   form.lastMsgID,
			ReplyMarkup: nil,
		})
	}

	if !form.hasField() {
		b.UnregisterHandler(form.ManageHandlerID)
		ctx = context.WithValue(ctx, FORM_KEY, form)
		form.SubmitHandler(ctx, b, update)
		return
	}

	nextField := form.nextField()
	switch nextField.Type {
	case ButtonField:
		form.HandlerID = b.RegisterHandler(bot.HandlerTypeCallbackQueryData, form.ButtonPrefix, bot.MatchTypePrefix, form.fieldHandler)
	case TextField:
		form.HandlerID = b.RegisterHandlerMatchFunc(form.checkUserMatch(), form.fieldHandler)
	case CustomTextField:
		form.HandlerID = b.RegisterHandlerMatchFunc(form.checkUserMatch(), form.customFieldHandler(nextField.CustomHandler))
	}

	params := &bot.SendMessageParams{
		ChatID:    form.ChatID,
		Text:      nextField.MessageText,
		ParseMode: nextField.ParseMode,
	}

	params.ReplyMarkup = form.buildKB()

	if msg, err := b.SendMessage(ctx, params); err == nil {
		form.lastMsgID = msg.ID
	}
}

func (form *Form) buildKB() *tmodels.InlineKeyboardMarkup {
	field := form.currentField()
	if field.Type == ButtonField {
		for i := 0; i < len(field.Keyboard); i++ {
			for j := 0; j < len(field.Keyboard[i]); j++ {
				field.Keyboard[i][j].CallbackData = form.ButtonPrefix + field.Keyboard[i][j].CallbackData
			}
		}
	} else {
		field.Keyboard = [][]tmodels.InlineKeyboardButton{}
	}
	manageRow := []tmodels.InlineKeyboardButton{
		{
			Text:         form.CancelButtonText,
			CallbackData: form.ManageBtnPrefix + CANCEL_DATA,
		},
	}
	if field.IsSkippable {
		manageRow = append(manageRow, tmodels.InlineKeyboardButton{
			Text:         form.SkipButtonText,
			CallbackData: form.ManageBtnPrefix + SKIP_DATA,
		})
	}

	field.Keyboard = append(field.Keyboard, manageRow)
	return &tmodels.InlineKeyboardMarkup{
		InlineKeyboard: field.Keyboard,
	}
}

func (form *Form) onManage(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:      form.ChatID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ReplyMarkup: nil,
	})

	data := strings.TrimPrefix(update.CallbackQuery.Data, form.ManageBtnPrefix)
	switch data {
	case CANCEL_DATA:
		form.onCancel(ctx, b, update)
	case SKIP_DATA:
		form.onSkip(ctx, b, update)
	}
}

func (form *Form) onCancel(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	b.UnregisterHandler(form.HandlerID)
	b.UnregisterHandler(form.ManageHandlerID)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: form.ChatID,
		Text:   "پروسه لغو شد",
	})

	form.CancelHandler(ctx, b, update)
}

func (form *Form) onSkip(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: form.ChatID,
		Text:   form.SkipMessageText,
	})

	form.loadNextField(ctx, b, update)
}

func (form *Form) checkUserMatch() bot.MatchFunc {
	return func(checkUpdate *tmodels.Update) bool {
		if checkUpdate.Message == nil || checkUpdate.Message.Text == form.CancelButtonText || checkUpdate.Message.Text == form.SkipButtonText {
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

func (form *Form) hasField() bool {
	return form.FieldIndex < len(form.Fields)
}

func (form *Form) nextField() *Field {
	if !form.hasField() {
		return nil
	}

	field := &form.Fields[form.FieldIndex]
	form.FieldIndex++
	return field
}

func (form *Form) currentField() *Field {
	if form.isFirstField() {
		return &form.Fields[0]
	}

	return &form.Fields[form.FieldIndex-1]
}

func (form *Form) isFirstField() bool {
	return form.FieldIndex == 0
}

func (form Form) CurrentField() Field {
	if form.isFirstField() {
		return form.Fields[0]
	}

	return form.Fields[form.FieldIndex-1]
}
