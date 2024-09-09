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

func SendSubLink(ctx context.Context, b *bot.Bot, chatID any, order m.Order) {
	db := database.GetDB()
	if order.Type == m.ActiveOrder {
		subLink, qrPath := order.Config(db).SubLink(db)
		fileContent, _ := os.ReadFile(qrPath)
		_, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:    chatID,
			Caption:   fmt.Sprintf("%s\n🔗 لینک:\n`%s`\nبرای کپی کردن لینک روی آن کلیک کنید", bot.EscapeMarkdown(order.Pack.Name()), bot.EscapeMarkdown(subLink)),
			Photo:     &tmodels.InputFileUpload{Filename: "qrcode.jpg", Data: bytes.NewReader(fileContent)},
			ParseMode: tmodels.ParseModeMarkdown,
		})
		fmt.Println(err)
	}
}
