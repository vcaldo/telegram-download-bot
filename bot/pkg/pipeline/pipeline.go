package pipeline

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/redisutils"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/upload"
)

func CheckReadyToUpload(ctx context.Context, updateChan chan<- *redisutils.Download) error {
	r, err := redisutils.NewAuthenticatedRedisClient(ctx)
	if err != nil {
		log.Printf("error creating redis client: %v", err)
	}
	defer r.Client.Close()

	// Get all downloads in compressed state
	uploads, err := r.GetDownloadState(ctx, redisutils.Compressed)
	if err != nil {
		log.Printf("error getting downloads by state: %v", err)
	}

	for _, id := range uploads {
		name, err := r.GetDownloadName(ctx, id)
		if err != nil {
			log.Printf("error getting download name: %v", err)
			continue
		}

		updateChan <- &redisutils.Download{
			ID:    id,
			Name:  name,
			State: redisutils.Compressed}
	}

	return nil
}

func ProcessUpdate(ctx context.Context, b *bot.Bot, update *redisutils.Download) error {
	r, err := redisutils.NewAuthenticatedRedisClient(ctx)
	if err != nil {
		return fmt.Errorf("error creating redis client: %v", err)
	}
	defer r.Client.Close()

	switch update.State {
	case redisutils.Compressed:
		log.Printf("Download ready to upload: %d", update.ID)

		// Set state to uploading
		update.State = redisutils.Uploading
		err := r.RegisterDownloadState(ctx, *update)
		if err != nil {
			log.Printf("error registering download state: %v", err)
		}

		err = upload.UploadDir(ctx, b, *update)
		if err != nil {
			log.Printf("error uploading files: %v", err)
			update.State = redisutils.Failed
			err := r.RegisterDownloadState(ctx, *update)
			if err != nil {
				log.Printf("error registering download state: %v", err)
			}
			return fmt.Errorf("failed to upload files: %v", err)
		}

		// Set state to uploaded
		update.State = redisutils.Uploaded
		err = r.RegisterDownloadState(ctx, *update)
		if err != nil {
			log.Printf("error registering download state: %v", err)
		}
	case redisutils.Uploaded:
		log.Printf("Download ready to be deleted: %d", update.ID)
	default:
		log.Printf("Download %d transitioned to: %s", update.ID, update.State)
		return nil
	}

	return nil
}
