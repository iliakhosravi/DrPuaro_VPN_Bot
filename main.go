package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"techybat.org/go-vpn/controllers/admin_menu"
	main2 "techybat.org/go-vpn/controllers/main_controller"
	"techybat.org/go-vpn/panel"
	"techybat.org/go-vpn/sub"
	"techybat.org/go-vpn/vars"

	"github.com/go-telegram/bot"
	"github.com/robfig/cron/v3"
	configCrons "techybat.org/go-vpn/crons/config"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/middlewares/auth"
)

func main() {
	// err := godotenv.Load(".env")

	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

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

	// c.AddFunc("@every 30s", func() { configCrons.NotifyAll(ctx, b) })
	c.AddFunc("@every 5m", func() { configCrons.NotifyAll(ctx, b) })
	c.AddFunc("@every 30m", panel.Setup)

	c.Start()

	// go sub.ServeHttp(ctx)
	go sub.ServeHttps(ctx)

	b.Start(ctx)
}
