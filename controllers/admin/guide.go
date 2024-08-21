package admin_controller

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"techybat.org/go-vpn/database"
	m "techybat.org/go-vpn/models"
	"techybat.org/go-vpn/widgets/form"
)

func AddGuideHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	userID := update.CallbackQuery.From.ID

	fields := []form.Field{
		{
			Name:        "title",
			Type:        form.TextField,
			MessageText: "لطفا عنوان دکمه را وارد نمایید.",
		},
		{
			Name:        "type",
			Type:        form.ButtonField,
			MessageText: "نوع دکمه را مشخص کنید",
			Keyboard: [][]tmodels.InlineKeyboardButton{
				{
					{
						Text:         "لینک",
						CallbackData: string(m.LINK_GUIDE),
					},
					{
						Text:         "فوروارد پیام",
						CallbackData: string(m.FWD_MSG_GUIDE),
					},
				},
			},
		},
		{
			Name:          "value",
			Type:          form.CustomTextField,
			MessageText:   "مقدار مربوطه را ارسال نمایید",
			CustomHandler: onInputValue,
		},
	}
	form := form.CreateForm("انصراف", fields, chatID, userID, onSubmitAddGuideHandler, onCancelAddGuideHandler, nil)
	form.Show(ctx, b, update)
}

func onSubmitAddGuideHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)
	guideType := m.GuideType(form.FindField("type").Value)
	guide := &m.Guide{
		Title: form.FindField("title").Value,
		Type:  guideType,
	}

	guideValue := form.FindField("value").Value

	switch guideType {
	case m.LINK_GUIDE:
		guide.Link = guideValue
	case m.FWD_MSG_GUIDE:
		guide.FwdMsgID, _ = strconv.Atoi(guideValue)
	}

	txtMsg := "دکمه راهنما با موفقیت افزوده شد"
	db := database.GetDB()
	if result := db.Save(guide); result.Error != nil {
		fmt.Println("Error on add guide to db: ", result.Error)
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: form.ChatID,
		Text:   txtMsg,
	})
}

func onInputValue(ctx context.Context, b *bot.Bot, update *tmodels.Update, form form.Form, setter form.FieldSetter) (bool, error) {
	typeField := form.FindField("type")
	switch m.GuideType(typeField.Value) {
	case m.LINK_GUIDE:
		if !m.UrlValidator(update.Message.Text) {
			return false, fmt.Errorf("لینک نامعتبر است. لطفا یک لینک معتبر ارسال نمایید.")
		}
		if ok, msg := setter(update.Message.Text); !ok {
			return false, fmt.Errorf(msg)
		}
		return true, nil
	case m.FWD_MSG_GUIDE:
		msg, err := b.ForwardMessage(ctx, &bot.ForwardMessageParams{
			FromChatID: fmt.Sprintf("%d", form.ChatID),
			ChatID:     os.Getenv("STORAGE_CHANNEL_ID"),
			MessageID:  update.Message.ID,
		})
		if err != nil {
			return false, err
		}
		if ok, msg := setter(fmt.Sprintf("%d", msg.ID)); !ok {
			return false, fmt.Errorf(msg)
		}
		return true, nil
	}
	return false, fmt.Errorf("Unsupported guide type to enter:%s", typeField.Value)
}

func onCancelAddGuideHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {

}
