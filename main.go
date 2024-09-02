package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"techybat.org/go-vpn/controllers"
	configCrons "techybat.org/go-vpn/crons/config"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/middlewares/auth"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Setup()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	opts := []bot.Option{
		bot.WithMiddlewares(auth.UserMiddleware),
		bot.WithDefaultHandler(controllers.MainController),
		bot.WithMessageTextHandler("/admin", bot.MatchTypeExact, auth.AdminMiddleware(controllers.AdminController)),
	}

	b, err := bot.New(telegramBotToken, opts...)

	if err != nil {
		fmt.Println("Error in creating bot: ", err)
		cancel()
	}

	c := cron.New()

	// c.AddFunc("@every 30s", func() { configCrons.NotifyAll(ctx, b) })
	c.AddFunc("@every 30m", func() { configCrons.NotifyAll(ctx, b) })

	c.Start()

	b.Start(ctx)
}
