package customerController

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/middlewares/auth"
	m "techybat.org/go-vpn/models"
	msgTool "techybat.org/go-vpn/tools/message"
)

func TestConfigHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	user := ctx.Value(auth.UserKey).(m.User)

	if update.CallbackQuery != nil {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
	}

	// Already had their one trial.
	if user.HasUsedTest(db) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: user.TelID,
			Text:   "شما قبلا کانفیگ تست خود را دریافت کرده‌اید.\nبرای ادامه استفاده می‌توانید از منو اصلی یکی از بسته‌ها را خریداری کنید. 🙏",
		})
		return
	}

	pack, err := m.GetTestPack(db)
	if err != nil {
		fmt.Println("Error: no test pack available. err: ", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: user.TelID,
			Text:   "در حال حاضر کانفیگ تست موجود نیست. لطفا بعدا تلاش کنید یا با پشتیبانی در ارتباط باشید.",
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: user.TelID,
		Text:   "در حال ساخت کانفیگ تست شما... لطفا چند لحظه صبر کنید ⏳",
	})

	order, err := user.TakeTestPack(db, pack)
	if err != nil {
		fmt.Println("Error: unable to create test config. err: ", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: user.TelID,
			Text:   "خطایی در ساخت کانفیگ تست پیش آمد. لطفا با پشتیبانی در ارتباط باشید.",
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: user.TelID,
		Text: fmt.Sprintf(
			"کانفیگ تست شما با موفقیت ساخته شد ✅\n\nمشخصات:\nحجم: %s\nمدت: %s\n\nپس از اتمام تست، برای تهیه بسته کامل از منو اصلی اقدام کنید. 🌹",
			pack.TrafficString(),
			m.StringPeriod(pack.Period),
		),
	})

	msgTool.SendPanelSubLink(ctx, b, user.TelID, *order)
}
