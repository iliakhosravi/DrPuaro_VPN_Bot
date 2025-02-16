package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"gorm.io/gorm/clause"
	"techybat.org/go-vpn/controllers/admin_menu"
	main2 "techybat.org/go-vpn/controllers/main_controller"
	"techybat.org/go-vpn/models"
	"techybat.org/go-vpn/panel"
	"techybat.org/go-vpn/sub"
	msgTool "techybat.org/go-vpn/tools/message"
	"techybat.org/go-vpn/vars"

	"github.com/go-telegram/bot"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	configCrons "techybat.org/go-vpn/crons/config"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/middlewares/auth"

	tmodels "github.com/go-telegram/bot/models"
	cryptodb "github.com/sinasadeghi83/go-crypto-paywall/db"
	"github.com/sinasadeghi83/go-crypto-paywall/listeners"
	paym "github.com/sinasadeghi83/go-crypto-paywall/models"
	"github.com/sinasadeghi83/go-crypto-paywall/queue"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Setup()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	telegramBotToken := vars.Get("TELEGRAM_BOT_TOKEN")

	opts := []bot.Option{
		bot.WithMiddlewares(auth.UserMiddleware, auth.TrustedMiddleware),
		bot.WithDefaultHandler(main2.MainController),
		bot.WithMessageTextHandler("/admin", bot.MatchTypeExact, auth.AdminMiddleware(admin_menu.AdminController)),
	}

	b, err := bot.New(telegramBotToken, opts...)

	if err != nil {
		fmt.Println("Error in creating bot: ", err)
		cancel()
	}

	c := cron.New()

	c.AddFunc("@every 5m", func() { configCrons.NotifyAll(ctx, b) })
	c.AddFunc("@every 5m", func() { configCrons.CheckExpiredOrders(ctx, b) })
	// c.AddFunc("@every 30m", func() { configCrons.NotifyAll(ctx, b) })
	c.AddFunc("@every 30m", panel.Setup)
	c.AddFunc("@every 1h", func() { msgTool.SendBackup(ctx, b, vars.Get("STORAGE_CHANNEL_ID")) })

	c.Start()

	if vars.Get("env") == "prod" {
		go sub.ServeHttps(ctx)
	} else {
		go sub.ServeHttp(ctx)
	}

	db := database.GetDB()
	cryptodb.Setup(db)
	queue.Setup(ctx, vars.Get("REDIS"), func(t paym.Transaction, i paym.Invoice, cw paym.CryptoWallet) {
		var order models.Order
		res := db.Preload(clause.Associations).Find(&order, "invoice_id = ?", i.ID)
		if res.RowsAffected > 0 {
			order.Verify(db, "واریز با موفقیت دریافت شد(پیام سیستمی)", "")

			carryMsg := fmt.Sprintf("سفارش شما به طور سیستمی تایید شد.\n%s", order.UserStr(db))

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    order.User.TelID,
				Text:      carryMsg,
				ParseMode: tmodels.ParseModeHTML,
			})

			msgTool.SendSubLink(ctx, b, order.User.TelID, order)
		}

		var cOrder models.ChargeOrder
		res = db.Preload(clause.Associations).Find(&cOrder, "invoice_id = ?", i.ID)
		if res.RowsAffected > 0 {
			cOrder.AcceptCharge(db)

			carryMsg := fmt.Sprintf("سفارش شما به طور سیستمی تایید شد.\n%s", cOrder.FullStr())

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    cOrder.User.TelID,
				Text:      carryMsg,
				ParseMode: tmodels.ParseModeHTML,
			})
		}
	})

	fmt.Println("Listening on payments...")

	go listeners.ListenPayments(ctx)

	fmt.Println("Bot is going to start...")
	b.Start(ctx)
}
