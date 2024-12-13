package handlers

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/utils"
)

func HandleTextMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Switch case for handling different types of text messages
	switch {
	// handle magnet link
	case regexp.MustCompile(`magnet:\?xt=urn:.+`).MatchString(update.Message.Text):
		log.Printf("processing magnet link: %s", update.Message.Text)
		torrent, err := utils.AddTorrentFromMagnet(ctx, update.Message.Text)
		if err != nil {
			log.Printf("failed to add torrent: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Failed to add torrent",
			})
			return
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Torrent added: %s", *torrent.Name),
		})
	}
}

func HandleDocument(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Switch case for handling different types of documents
	switch {
	// handle torrent file
	case update.Message.Document.MimeType == "application/x-bittorrent":
		log.Printf("processing torrent file: %s", update.Message.Document.FileName)
		torrent, err := utils.AddTorrentFromFile(ctx, b, update.Message.Document.FileID, update.Message.Document.FileName)
		if err != nil {
			log.Printf("failed to add torrent: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Failed to add torrent",
			})
			return
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Torrent added: %s", *torrent.Name),
		})
	}
}
