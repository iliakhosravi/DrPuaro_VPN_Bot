package customerController

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/shopspring/decimal"
	"techybat.org/go-vpn/middlewares/auth"
	m "techybat.org/go-vpn/models"
)

func BalanceHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	user := ctx.Value(auth.UserKey).(m.User)
	charge := decimal.NewFromUint64(user.Charge).Div(decimal.NewFromInt(1000000))

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: user.TelID,
		Text:   fmt.Sprintf("موجودی کیف پول شما %s دلار است.\nبرای شارژ کیف پول خود از منو اصلی می توانید شارژ اکانت را انتخاب کنید.", charge),
	})
}
