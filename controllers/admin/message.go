package admin_controller

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/progress"
	"techybat.org/go-vpn/database"
	m "techybat.org/go-vpn/models"
	"techybat.org/go-vpn/widgets/form"
)

func SendMsgHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.From.ID
	userID := update.CallbackQuery.From.ID
	db := database.GetDB()
	fields := []form.Field{
		{
			Name:          "message",
			MessageText:   "لطفا پیام خود را ارسال نمایید",
			Type:          form.CustomTextField,
			CustomHandler: onInputMsg,
		},
		{
			Name:        "user_id",
			MessageText: "آیدی کاربر را ارسال کنید",
			Validator:   m.UserIDValidator(db),
		},
		{
			Name:        "check",
			MessageText: "آیا از ارسال این پیام به  این کاربر اطمینان دارید؟",
			Type:        form.ButtonField,
			Keyboard: [][]tmodels.InlineKeyboardButton{
				{
					{
						Text:         "بله",
						CallbackData: "true",
					},
				},
			},
		},
	}
	form := form.CreateForm("انصراف", fields, chatID, userID, onSendMsg, onCancelSendMsg, nil)
	form.Show(ctx, b, update)
}

func onCancelSendMsg(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
}

func onSendMsg(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)
	msgID, _ := strconv.Atoi(form.FindField("message").Value)
	userID := form.FindField("user_id").Value
	_, err := b.CopyMessage(ctx, &bot.CopyMessageParams{
		FromChatID: fmt.Sprintf("%d", form.ChatID),
		ChatID:     userID,
		MessageID:  msgID,
	})
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: form.ChatID,
			Text:   "خطایی پیش آمده. پیام ارسال نشد.",
		})
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: form.ChatID,
			Text:   "پیام شما به کاربر ارسال شد.",
		})
	}
}

func SendToAllMsgHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.From.ID
	userID := update.CallbackQuery.From.ID
	fields := []form.Field{
		{
			Name:          "message",
			MessageText:   "لطفا پیام خود را ارسال نمایید",
			Type:          form.CustomTextField,
			CustomHandler: onInputMsg,
		},
		{
			Name:        "check",
			MessageText: "آیا از ارسال این پیام به همه کاربران اطمینان دارید؟",
			Type:        form.ButtonField,
			Keyboard: [][]tmodels.InlineKeyboardButton{
				{
					{
						Text:         "بله",
						CallbackData: "true",
					},
				},
			},
		},
	}
	form := form.CreateForm("انصراف", fields, chatID, userID, onSendToAllMsg, onCancelSendAllMsg, nil)
	form.Show(ctx, b, update)
}

func onSendToAllMsg(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)
	msgID, _ := strconv.Atoi(form.FindField("message").Value)

	p := progress.New(progress.WithRenderTextFunc(renderCopy), progress.WithCancel("انصراف", true, nil))
	p.Show(ctx, b, form.ChatID)

	go copyMessagesToUsers(ctx, b, p, form.ChatID, msgID)
}

func copyMessagesToUsers(ctx context.Context, b *bot.Bot, p *progress.Progress, chatID int64, msgID int) {
	db := database.GetDB()
	var users []m.User
	db.Find(&users)
	for i, user := range users {
		if i%10 == 0 {
			var progressValue float64 = float64(i) * 100.0 / float64(len(users))
			p.SetValue(ctx, b, progressValue)
		}
		b.CopyMessage(ctx, &bot.CopyMessageParams{
			FromChatID: fmt.Sprintf("%d", chatID),
			ChatID:     user.TelID,
			MessageID:  msgID,
		})
	}
	p.SetValue(ctx, b, 100.0)
	p.Delete(ctx, b)
	p.Done(ctx, b)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "پیام همگانی ارسال شد",
	})
}

func renderCopy(value float64) string {
	s := fmt.Sprintf("درحال ارسال پیام به کاربران: %.2f%%", value)

	return bot.EscapeMarkdown(s)
}

func onInputMsg(ctx context.Context, b *bot.Bot, update *tmodels.Update, form form.Form, setter form.FieldSetter) (bool, error) {
	ok, msg := setter(strconv.Itoa(update.Message.ID))
	return ok, fmt.Errorf(msg)
}

func onCancelSendAllMsg(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
}
