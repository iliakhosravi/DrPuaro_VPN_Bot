package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/dialog"
	"github.com/joho/godotenv"
)

var (
	dialogNodes = []dialog.Node{
		{ID: "start", Text: "Start Node", Keyboard: [][]dialog.Button{{{Text: "Go to node 2", NodeID: "2"}, {Text: "Go to node 3", NodeID: "3"}}, {{Text: "Go Telegram UI", URL: "https://github.com/go-telegram/ui"}}}},
		{ID: "2", Text: "node 2 without keyboard"},
		{ID: "3", Text: "node 3", Keyboard: [][]dialog.Button{{{Text: "Go to start", NodeID: "start"}, {Text: "Go to node 4", NodeID: "4"}}}},
		{ID: "4", Text: "node 4", Keyboard: [][]dialog.Button{{{Text: "Back to 3", NodeID: "3"}}}},
	}
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	opts := []bot.Option{
		bot.WithDefaultHandler(handlerDialogInline),
	}

	b, err := bot.New(telegramBotToken, opts...)

	if err != nil {
		fmt.Println("Error in creating bot: ", err)
		cancel()
	}

	b.Start(ctx)
}

func handlerDialogInline(ctx context.Context, b *bot.Bot, update *models.Update) {
	p := dialog.New(dialogNodes, dialog.Inline())

	p.Show(ctx, b, update.Message.Chat.ID, "start")
}
