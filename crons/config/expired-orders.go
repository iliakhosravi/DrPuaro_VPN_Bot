package config_crons

import (
	"context"
	"time"

	"github.com/go-telegram/bot"
	"techybat.org/go-vpn/database"
	m "techybat.org/go-vpn/models"
)

func CheckExpiredOrders(ctx context.Context, b *bot.Bot) {
	db := database.GetDB()
	var expiredOrders []m.Order
	db.Joins("JOIN invoices on invoices.id = orders.invoice_id").
		Where("invoices.expires_at <= ?", time.Now()).
		Where("orders.type = ?", string(m.PendingOrder)).
		Preload("User").
		Find(&expiredOrders)

	for _, order := range expiredOrders {
		order.ChangeType(db, m.CancelledOrder)
		txtMsg := "مهلت پرداخت سفارش شما منقضی شده:\n\n"
		txtMsg += order.FullStr(db)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: order.User.TelID,
			Text:   txtMsg,
		})
	}
}
