package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"techybat.org/go-vpn/controllers/admin_menu"
	main2 "techybat.org/go-vpn/controllers/main_controller"
	"techybat.org/go-vpn/panel"
	msgTool "techybat.org/go-vpn/tools/message"
	"techybat.org/go-vpn/vars"

	"github.com/go-telegram/bot"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	configCrons "techybat.org/go-vpn/crons/config"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/middlewares/auth"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the VPN bot",
	Long:  `Start the Telegram bot for managing VPN configurations and orders.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		database.Setup()

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		Start(ctx, cancel)
	},
}

func Start(ctx context.Context, cancel context.CancelFunc) {
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
	// c.AddFunc("@every 30m", func() { configCrons.NotifyAll(ctx, b) })
	c.AddFunc("@every 30m", panel.Setup)
	c.AddFunc("@every 1h", func() { msgTool.SendBackup(ctx, b, vars.Get("STORAGE_CHANNEL_ID")) })

	c.Start()

	b.Start(ctx)
}
