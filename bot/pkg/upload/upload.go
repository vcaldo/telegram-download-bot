package upload

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/redisutils"
)

func UploadDir(ctx context.Context, b *bot.Bot, download redisutils.Download) error {
	chatId, ok := os.LookupEnv("CHAT_ID")
	if !ok {
		panic("CHAT_ID env var is required")
	}

	chatIdInt, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		panic("CHAT_ID must be a valid int64")
	}

	path := filepath.Join(redisutils.UploadsReadyPath, download.Name)
	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			file, err := os.Open(filepath.Join(path, file.Name()))
			if err != nil {
				return fmt.Errorf("failed to open file: %v", err)
			}
			defer file.Close()

			log.Printf("Uploading file: %s", file.Name())
			err = uploadFile(ctx, b, chatIdInt, file)
			if err != nil {
				return fmt.Errorf("failed to upload file: %v", err)
			}
			time.Sleep(5 * time.Second)
		}
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatIdInt,
		Text:   download.Name,
	})
	if err != nil {
		fmt.Errorf("failed to send message: %v", err)

	}
	return nil
}

func uploadFile(ctx context.Context, b *bot.Bot, chatId int64, file *os.File) error {
	fileReader := &models.InputFileUpload{
		Filename: file.Name(),
		Data:     file,
	}

	_, err := b.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID:   chatId,
		Document: fileReader,
	})

	if err != nil {
		return fmt.Errorf("failed to send document: %v", err)
	}
	return nil
}
