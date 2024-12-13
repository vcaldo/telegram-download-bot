package pipeline

import (
	"context"
	"fmt"
	"log"

	"github.com/vcaldo/telegram-download-bot/bot/pkg/redisutils"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/transmission"
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
		log.Printf("New download completion detected: %s", d.Name)
		// Check if download exists in Redis
		// exists, err := r.Client.DownloadExists(ctx, d.ID)
		// if err != nil {
		// 	log.Printf("error checking redis: %v", err)
		// 	continue
		// }

		// // Store in Redis and push to channel if new
		// if !exists {
		// 	log.Printf("New download completion detected: %s", d.Name)
		// 	if err := r.Client.RegisterDownloadState(ctx, d); err != nil {
		// 		log.Printf("error storing in redis: %v", err)
		// 		continue
		// 	}
		// 	updateChan <- d
		// }
	}
	return nil
}

func ProcessUpdate(ctx context.Context, update *redisutils.Download) error {
	r, err := redisutils.NewAuthenticatedRedisClient(ctx)
	if err != nil {
		return fmt.Errorf("error creating redis client: %v", err)
	}
	defer r.Client.Close()

	return nil
}
