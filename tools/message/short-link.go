package msgTool

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"techybat.org/go-vpn/database"
	m "techybat.org/go-vpn/models"
)

func SendShortLink(ctx context.Context, b *bot.Bot, chatID any, order m.Order) {
	db := database.GetDB()
	if order.Type == m.ActiveOrder {
		shortLink, qrPath := order.Config(db).ShortLink(db)
		fileContent, _ := os.ReadFile(qrPath)
		_, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:    chatID,
			Caption:   fmt.Sprintf("%s\n🔗 لینک:\n`%s`\nبرای کپی کردن آن روی آن کلیک کنید", bot.EscapeMarkdown(order.Pack.Name()), bot.EscapeMarkdown(shortLink)),
			Photo:     &tmodels.InputFileUpload{Filename: "qrcode.jpg", Data: bytes.NewReader(fileContent)},
			ParseMode: tmodels.ParseModeMarkdown,
		})
		fmt.Println(err)
	}
}
