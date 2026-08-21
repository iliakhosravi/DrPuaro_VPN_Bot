package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/models"
	"techybat.org/go-vpn/vars"
)

type Key int

const (
	UndefinedKey Key = iota
	UserKey
)

// TrustedMiddleware blocks normal (non-trusted) users when ONLY_TRUSTED_USERS
// is enabled. It is no longer wired into the default middleware chain in
// cmd/bot/start.go — every user can now start the bot, gated instead by
// ChannelMembershipMiddleware below. Kept here in case it's needed again.
func TrustedMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
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
		if vars.Get("ONLY_TRUSTED_USERS") == "true" && user.Type == models.NoramlUser {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: user.TelID,
				Text:   "عدم دسترسی کاربر",
			})
			return
		}
		next(ctx, b, update)
	}
}

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
		if strings.EqualFold(vars.Get("DEFAULT_ADMIN_USERNAME"), user.Username) {
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

// ChannelMembershipMiddleware requires the user to be a member of the
// channel configured via REQUIRED_CHANNEL_USERNAME (e.g. "@drpuaro_net")
// before letting the update reach any handler. Anyone can now start the
// bot; this is the only gate.
//
// If REQUIRED_CHANNEL_USERNAME is empty (not configured), the check is
// skipped entirely (fail-open) so an incomplete deployment never fully
// locks users out of the bot.
//
// IMPORTANT: for GetChatMember to work against a channel, this bot must
// itself be a member (in practice: an admin) of that channel — otherwise
// Telegram returns an error, which we also treat as fail-open (see below).
func ChannelMembershipMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
		channel := vars.Get("REQUIRED_CHANNEL_USERNAME")
		if channel == "" {
			next(ctx, b, update)
			return
		}

		var tuser *tmodels.User
		switch {
		case update.Message != nil:
			tuser = update.Message.From
		case update.CallbackQuery != nil:
			tuser = &update.CallbackQuery.From
		}

		if tuser == nil || tuser.ID == 0 {
			next(ctx, b, update)
			return
		}

		member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
			ChatID: channel,
			UserID: tuser.ID,
		})
		if err != nil {
			// Fail open: don't lock legitimate users out of the bot because
			// of a misconfiguration (bot not added to the channel yet) or a
			// transient Telegram API error.
			fmt.Printf("Error: cannot check channel membership for user %d. Details: %v\n", tuser.ID, err)
			next(ctx, b, update)
			return
		}

		if isChannelMember(member) {
			next(ctx, b, update)
			return
		}

		if update.CallbackQuery != nil {
			b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "هنوز عضو کانال نشده‌اید. لطفا ابتدا عضو شوید و دوباره تلاش کنید.",
				ShowAlert:       true,
			})
			return
		}

		sendJoinChannelPrompt(ctx, b, update.Message.Chat.ID)
	}
}

func isChannelMember(member *tmodels.ChatMember) bool {
	if member == nil {
		return false
	}
	switch member.Type {
	case tmodels.ChatMemberTypeLeft, tmodels.ChatMemberTypeBanned:
		return false
	default:
		return true
	}
}

func sendJoinChannelPrompt(ctx context.Context, b *bot.Bot, chatID int64) {
	link := vars.Get("REQUIRED_CHANNEL_LINK")
	nodes := []dialog.Node{
		{
			ID:   "join-channel",
			Text: "برای استفاده از ربات، ابتدا باید عضو کانال ما شوید:",
			Keyboard: [][]dialog.Button{
				{
					{Text: "📢 عضویت در کانال", URL: link},
				},
				{
					{ID: "check-channel-membership", Text: "✅ بررسی عضویت", CallbackHandler: CheckMembershipHandler},
				},
			},
		},
	}
	dialog.New(nodes, dialog.Inline()).Show(ctx, b, chatID, "join-channel")
}

// CheckMembershipHandler is wired as the CallbackHandler on the "بررسی
// عضویت" (check membership) button. By the time it runs, the global
// ChannelMembershipMiddleware has already re-checked membership for this
// same update — if the user still wasn't a member, the middleware would
// have stopped the update with an alert instead of reaching this handler.
// So this only needs to confirm success and ask the user to restart.
func CheckMembershipHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	if update.CallbackQuery == nil {
		return
	}
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "عضویت شما تایید شد ✅",
		ShowAlert:       true,
	})
	if update.CallbackQuery.Message.Message != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Text:   "عضویت شما تایید شد ✅ لطفا دستور /start را دوباره ارسال کنید.",
		})
	}
}
