package admin_controller

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"gorm.io/gorm"
	comp "techybat.org/go-vpn/components"
	"techybat.org/go-vpn/database"
	m "techybat.org/go-vpn/models"
	bp "techybat.org/go-vpn/widgets/buttonpage"
	"techybat.org/go-vpn/widgets/form"
)

func AddInlineKBHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	userID := update.CallbackQuery.From.ID
	fields := []form.Field{
		{
			Name:        "name",
			MessageText: "نام کیبورد (دکمه ای که در کیبورد منو نمایش داده می شود) را وارد نمایید.",
		},
		{
			Name:        "text",
			MessageText: "توضیحاتی که در پیام برای کیبورد نمایش داده خواهد شد را وارد نمایید.",
		},
	}
	form := form.CreateForm("انصراف", fields, chatID, userID, onAddInlineKB, onCancelInlineKB, nil)
	form.Show(ctx, b, update)
}

func onAddInlineKB(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)
	db := database.GetDB()
	kb := &m.InlineKeyboard{
		Name: form.FindField("name").Value,
		Text: form.FindField("text").Value,
	}

	txtMsg := "کیبورد با موفقیت افزوده شد."
	if result := db.Save(kb); result.Error != nil {
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: form.ChatID,
		Text:   txtMsg,
	})
	comp.BuildMainKeyboard(b)
}

func onCancelInlineKB(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
}

func RemoveInlineKBHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.From.ID
	btns := createInlineKBBtns(database.GetDB())
	bpage := bp.CreateButtonPage(bot.EscapeMarkdown("کدام یک از کیبورد های زیر را می خواهید حذف کنید؟"), btns, 5, true)
	_, err := bpage.Show(ctx, b, chatID)
	if err != nil {
		fmt.Println("Error, unable to show inline keyboards to remove. err: ", err)
	}
}

func onRemoveInlineKB(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	kbID := update.CallbackQuery.Data
	db := database.GetDB()
	var kb m.InlineKeyboard
	db.Find(&kb, kbID)

	txtMsg := fmt.Sprintf("کیبورد شیشه ای با نام %s حذف شد.", kb.Name)
	if res := db.Delete(&kb); res.Error != nil {
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   txtMsg,
	})
	comp.BuildMainKeyboard(b)
}

func createInlineKBBtns(db *gorm.DB) []dialog.Button {
	var kbs []m.InlineKeyboard
	db.Find(&kbs)

	btns := []dialog.Button{}
	for _, kb := range kbs {
		btns = append(btns, dialog.Button{
			ID:              fmt.Sprintf("%d", kb.ID),
			Text:            kb.Name,
			CallbackHandler: onRemoveInlineKB,
			CallbackData:    fmt.Sprintf("%d", kb.ID),
		})
	}

	return btns
}
