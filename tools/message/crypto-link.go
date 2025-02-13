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

func SendCryptoLink(ctx context.Context, b *bot.Bot, chatID any, order m.Order) {
	db := database.GetDB()
	if order.Type == m.PendingOrder && order.PayType == m.CryptoPay {
		cryptoLink, qrPath := order.CryptoLink(db)
		fileContent, _ := os.ReadFile(qrPath)
		_, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:    chatID,
			Caption:   fmt.Sprintf("برای واریز رمزارز '%s' در واحد '%s' می توانید تصویر را با کیف پول خود اسکن کرده یا از لینک زیر استفاده کنید \n%s\n🔗 لینک:\n`%s`\nاین لینک تنها تا 15 دقیقه دیگر معتبر است", bot.EscapeMarkdown(order.Pack.Currency.Name), bot.EscapeMarkdown(order.Pack.Currency.Unit), bot.EscapeMarkdown(order.Pack.Name()), bot.EscapeMarkdown(cryptoLink)),
			Photo:     &tmodels.InputFileUpload{Filename: "qrcode.jpg", Data: bytes.NewReader(fileContent)},
			ParseMode: tmodels.ParseModeMarkdown,
		})
		fmt.Println(err)
	}
}
