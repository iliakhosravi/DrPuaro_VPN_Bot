package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/joho/godotenv"
	"techybat.org/go-vpn/controllers"
	"techybat.org/go-vpn/database"
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
		bot.WithDefaultHandler(controllers.MainController),
		bot.WithCallbackQueryDataHandler("mainpack_", bot.MatchTypePrefix, controllers.BuyController),
	}

	b, err := bot.New(telegramBotToken, opts...)

	if err != nil {
		fmt.Println("Error in creating bot: ", err)
		cancel()
	}

	b.Start(ctx)
}
