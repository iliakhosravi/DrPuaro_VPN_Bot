package controllers

import (
	"context"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"techybat.org/go-vpn/middlewares/auth"
	"techybat.org/go-vpn/models"
)

func AdminController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	user := ctx.Value(auth.UserKey).(models.User)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: user.TelID,
		Text:   "Welcome to admin panel!",
	})
}
