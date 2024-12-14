package main

import (
	"context"
	"time"

	"github.com/vcaldo/telegram-download-bot/bot/pkg/pipeline"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/redisutils"
	"github.com/vcaldo/telegram-download-bot/splitter/pkg/fileutils"
)

func main() {
	ctx := context.Background()
	downloadChan := make(chan *redisutils.Download)

	// Goroutine to constantly check for completed downloads
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pipeline.CheckCompletedDownloads(ctx, downloadChan)
			}
		}
	}()

	// Goroutine to process downloads one at a time
	go func() {
		for download := range downloadChan {
			fileutils.CompressDownload(ctx, download)
		}
	}()

	// Keep the main function running
	select {}
}
