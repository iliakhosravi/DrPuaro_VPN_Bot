package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/models"
)

type Key int

const (
	UndefinedKey Key = iota
	UserKey
)

func UserMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
		var user models.User
		var tuser *tmodels.User
		switch {
		case update.Message != nil:
			tuser = update.Message.From
		case update.CallbackQuery != nil:
			tuser = &update.CallbackQuery.From
		default:
			tuser = &tmodels.User{}
		}
		user.CreateOrFindUserByTelegram(database.GetDB(), tuser)
		ctx = context.WithValue(ctx, UserKey, user)
		next(ctx, b, update)
	}
}

func AdminMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
		user := ctx.Value(UserKey).(models.User)
		if strings.EqualFold(os.Getenv("DEFAULT_ADMIN_USERNAME"), user.Username) {
			if err := user.MakeAdmin(database.GetDB()); err != nil {
				fmt.Printf("Error: Cannot set default admin. Details:%v\n", err)
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: user.TelID,
					Text:   "خطایی پیش آمده",
				})
				return
			}
		}
		if user.Type != models.AdminUser {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: user.TelID,
				Text:   "عدم دسترسی",
			})
			return
		}

		next(ctx, b, update)
	}
}
