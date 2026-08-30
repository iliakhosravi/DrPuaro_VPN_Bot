package admin_controller

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"techybat.org/go-vpn/database"
	m "techybat.org/go-vpn/models"
	bp "techybat.org/go-vpn/widgets/buttonpage"
)

func TestOrdersHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	var orders []m.Order
	db.Where("is_test = ?", true).
		Preload("Pack").
		Preload("Pack.Category").
		Preload("User").
		Order("created_at desc").
		Find(&orders)

	buttons := createOrderButtons(orders)

	title := bot.EscapeMarkdown(fmt.Sprintf("لیست کانفیگ های تست رایگان (%d مورد):", len(orders)))
	buttonPage := bp.CreateButtonPage(title, buttons, 5, true)
	if _, err := buttonPage.Show(ctx, b, chatID); err != nil {
		fmt.Println("Error showing test orders handler:", err)
	}
}
