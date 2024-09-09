package admin_controller

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/middlewares/auth"
	m "techybat.org/go-vpn/models"
	"techybat.org/go-vpn/widgets/form"
)

func AddTrustedUser(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	user := ctx.Value(auth.UserKey).(m.User)
	fields := []form.Field{
		{
			Name:        "tg_id",
			MessageText: "لطفا آیدی عددی کاربری که می خواهید معتمد شود را وارد نمایید",
		},
	}
	f := form.CreateForm("انصراف", fields, user.TelID, user.TelID, onSubmitTrusted, onCancelTrusted, nil)
	f.Show(ctx, b, update)
}

func onSubmitTrusted(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	f := ctx.Value(form.FORM_KEY).(*form.Form)
	tgID, _ := strconv.ParseInt(f.FindField("tg_id").Value, 0, 0)
	db := database.GetDB()

	newUser := m.User{}
	res := db.Find(&newUser, "tel_id = ?", tgID)

	newUser.TelID = tgID
	if newUser.Type == m.NoramlUser || res.RowsAffected == 0 {
		newUser.Type = m.TrustedUser
	}

	txtMsg := "کاربر با موفقیت به کاربران معتمد پیوست."
	if res := db.Save(&newUser); res.Error != nil {
		fmt.Println("Unable to save trusted user. err: ", res.Error)
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: f.ChatID,
		Text:   txtMsg,
	})
}

func onCancelTrusted(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
}
