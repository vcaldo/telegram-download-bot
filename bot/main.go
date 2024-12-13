package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
		bot.WithServerURL(os.Getenv("LOCAL_TELEGRAM_BOT_API_URL")),
		bot.WithCheckInitTimeout(60 * time.Second),
		bot.WithHTTPClient(60*time.Second, &http.Client{
			Timeout: 20 * time.Minute, // It could take up to 20 minutes to upload a 2gb file
		}),
	}

	token := os.Getenv("BOT_TOKEN")
	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
}
