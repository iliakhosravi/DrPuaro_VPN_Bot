package customerController

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"techybat.org/go-vpn/middlewares/auth"
	m "techybat.org/go-vpn/models"
)

func BalanceHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	user := ctx.Value(auth.UserKey).(m.User)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: user.TelID,
		Text:   fmt.Sprintf("موجودی کیف پول شما %d تومان است.\nبرای شارژ کیف پول خود از منو اصلی می توانید شارژ اکانت را انتخاب کنید.", user.Charge),
	})
}
