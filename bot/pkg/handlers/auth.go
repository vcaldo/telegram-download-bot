package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func IsUserAllowed(ctx context.Context, b *bot.Bot, update *models.Update) bool {
	// switch {
	// case update.Message.From.ID < 0: // Bots aren't allowed to use the bot
	// 	log.Printf("user %v is a bot and isn't allowed to use the bot", update.Message.From.ID)
	// 	return false
	// case update.Message.Chat.Type == "channel": // Don't reply to messages on a channel
	// 	log.Printf("user %v is a channel and isn't allowed to use the bot", update.Message.From.ID)
	// 	return false
	// default:
	return validateUser(ctx, b, update)
	// }
}

func validateUser(ctx context.Context, b *bot.Bot, update *models.Update) bool {
	allowedUserIds := os.Getenv("ALLOWED_USER_IDS")
	allowedUserIdsSlice := strings.Split(allowedUserIds, ",")
	allowedUserIdsInt64 := make([]int64, len(allowedUserIdsSlice))
	for i, id := range allowedUserIdsSlice {
		var err error
		allowedUserIdsInt64[i], err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			return false
		}
	}

	for _, id := range allowedUserIdsInt64 {
		if id == update.Message.From.ID {
			log.Printf("user %v - %v %v is allowed to use the bot", update.Message.From.ID, update.Message.From.FirstName, update.Message.From.LastName)
			return true
		}
	}
	err := unauthorizedMessage(ctx, b, update)
	if err != nil {
		log.Printf("failed to send unauthorized message: %v", err)
	}
	log.Printf("user %v is not allowed to use the bot", update.Message.From.ID)
	return false
}

func unauthorizedMessage(ctx context.Context, b *bot.Bot, update *models.Update) error {
	_, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID: update.Message.Chat.ID,
		Photo:  &models.InputFileString{Data: "https://ih1.redbubble.net/image.3655810608.7816/flat,750x,075,f-pad,750x1000,f8f8f8.jpg"},
		Caption: fmt.Sprintf(
			"⚠️ Access Restricted ⚠️\n\n"+
				"This bot requires authorization for usage.\n"+
				"To request access, contact the administrator or the person who invited you, and provide your User ID\n"+
				"📋 User ID: %d\n\n"+
				"Thank you for your understanding.",
			update.Message.From.ID,
		),
	})
	if err != nil {
		return fmt.Errorf("failed to send unauthorized message: %v", err)
	}
	return nil
}
