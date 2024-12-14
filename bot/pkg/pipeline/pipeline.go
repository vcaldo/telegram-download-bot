package pipeline

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/redisutils"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/remove"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/transmission"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/upload"
)

func CheckCompletedDownloads(ctx context.Context, updateChan chan<- *redisutils.Download) error {
	c, err := transmission.NewTransmissionClient(ctx)
	if err != nil {
		log.Printf("error creating transmission client: %v", err)
		return err
	}

	r, err := redisutils.NewAuthenticatedRedisClient(ctx)
	if err != nil {
		log.Printf("error creating redis client: %v", err)
	}
	defer r.Client.Close()

	completedDownloads, err := c.GetCompletedDownloads(ctx)
	if err != nil {
		log.Printf("error getting completed downloads: %v", err)
		return err
	}

	for _, download := range completedDownloads {
		d := redisutils.Download{
			ID:    *download.ID,
			Name:  *download.Name,
			State: redisutils.Downloaded,
		}

		// Check if download exists in Redis
		exists, err := r.DownloadExists(ctx, d.ID)
		if err != nil {
			log.Printf("error checking redis: %v", err)
			continue
		}

		// Store in Redis and push to channel if new
		if !exists {
			log.Printf("New download completion detected: %.24s...", d.Name)
			if err := r.RegisterDownloadState(ctx, d); err != nil {
				log.Printf("error storing in redis: %v", err)
				continue
			}
			updateChan <- &d
		}
	}
	return nil
}

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
			return fmt.Errorf("error getting download name: %v", err)
		}

		updateChan <- &redisutils.Download{
			ID:    id,
			Name:  name,
			State: redisutils.Compressed}
	}
	return nil
}

func CheckReadyToRemove(ctx context.Context, updateChan chan<- *redisutils.Download) error {
	r, err := redisutils.NewAuthenticatedRedisClient(ctx)
	if err != nil {
		log.Printf("error creating redis client: %v", err)
	}
	defer r.Client.Close()

	// Get all downloads in uploaded state
	uploads, err := r.GetDownloadState(ctx, redisutils.Uploaded)
	if err != nil {
		log.Printf("error getting downloads by state: %v", err)
	}

	for _, id := range uploads {
		name, err := r.GetDownloadName(ctx, id)
		if err != nil {
			log.Printf("error getting download name: %v", err)
			return fmt.Errorf("error getting download name: %v", err)
		}
		log.Printf("Download %d - %sready to be deleted", id, name)
		updateChan <- &redisutils.Download{
			ID:    id,
			Name:  name,
			State: redisutils.Uploaded,
		}
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
			return fmt.Errorf("error registering download state: %v", err)
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
			return fmt.Errorf("error registering download state: %v", err)
		}
	case redisutils.Uploaded:
		log.Printf("Download ready to be deleted: %d", update.ID)

		// Set state to removing
		update.State = redisutils.Removing
		err := r.RegisterDownloadState(ctx, *update)
		if err != nil {
			log.Printf("error registering download state: %v", err)
		}

		err = remove.RemoveDownload(ctx, update)
		if err != nil {
			log.Printf("error removing download: %v", err)
			update.State = redisutils.Failed
			err := r.RegisterDownloadState(ctx, *update)
			if err != nil {
				log.Printf("error registering download state: %v", err)
			}
			return fmt.Errorf("failed to remove download: %v", err)
		}
	default:
		// log.Printf("Download %d transitioned to: %s", update.ID, update.State)
		return nil
	}
	return nil
}
