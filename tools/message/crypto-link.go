package msgTool

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"techybat.org/go-vpn/database"
	m "techybat.org/go-vpn/models"
)

type CryptoOrder interface {
	CryptoLink(db *gorm.DB) (string, string)
	CryptoAddrMemo(db *gorm.DB) (string, string)
	IsPending(db *gorm.DB) bool
	PaymentType() m.PayType
	CoinName(db *gorm.DB) string
	CoinUnit(db *gorm.DB) string
	ProductName() string
	GetAmount(db *gorm.DB) decimal.Decimal
}

func SendCryptoLink(ctx context.Context, b *bot.Bot, chatID any, order CryptoOrder) {
	db := database.GetDB()
	if order.IsPending(db) && order.PaymentType() == m.CryptoPay {
		cryptoLink, qrPath := order.CryptoLink(db)
		fileContent, _ := os.ReadFile(qrPath)
		addr, memo := order.CryptoAddrMemo(db)
		_, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID: chatID,
			Caption: fmt.Sprintf("برای واریز رمزارز '%s' در واحد '%s' به میزان %s می توانید تصویر را با کیف پول خود اسکن کرده یا از لینک زیر استفاده کنید \n%s\n🔗 لینک: \n`%s`\nآدرس کیف پول: `%s`\nMemo or Comment: `%s`\nاین لینک تنها تا 15 دقیقه دیگر معتبر است",
				bot.EscapeMarkdown(order.CoinName(db)),
				bot.EscapeMarkdown(order.CoinUnit(db)),
				bot.EscapeMarkdown(order.GetAmount(db).String()),
				bot.EscapeMarkdown(order.ProductName()),
				bot.EscapeMarkdown(cryptoLink),
				bot.EscapeMarkdown(addr),
				bot.EscapeMarkdown(memo)),
			Photo:     &tmodels.InputFileUpload{Filename: "qrcode.jpg", Data: bytes.NewReader(fileContent)},
			ParseMode: tmodels.ParseModeMarkdown,
		})
		fmt.Println(err)
	}
}
