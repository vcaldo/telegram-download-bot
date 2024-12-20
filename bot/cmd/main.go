package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/handlers"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/pipeline"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/redisutils"
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

	// Start the bot in a goroutine
	go func() {
		b.Start(ctx)
	}()

	updateChan := make(chan *redisutils.Download)

	ticker10 := time.NewTicker(10 * time.Second)
	defer ticker10.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker10.C:
				pipeline.CheckReadyToUpload(ctx, updateChan)
			case <-ticker10.C:
				pipeline.CheckReadyToRemove(ctx, updateChan)
			}
		}
	}()

	for update := range updateChan {
		pipeline.ProcessUpdate(ctx, b, update)
	}

	// Wait for the context to be done
	select {}
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !handlers.IsUserAllowed(ctx, b, update) {
		return
	}

	switch {
	// handle text message
	case update.Message != nil && update.Message.Text != "":
		handlers.HandleTextMessage(ctx, b, update)
		return
	// handle Documents
	case update.Message != nil && update.Message.Document != nil:
		handlers.HandleDocument(ctx, b, update)
		return
	}
}
