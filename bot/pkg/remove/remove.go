package remove

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/vcaldo/telegram-download-bot/bot/pkg/redisutils"
	"github.com/vcaldo/telegram-download-bot/bot/pkg/transmission"
)

func RemoveDownload(ctx context.Context, download *redisutils.Download) error {
	r, err := redisutils.NewAuthenticatedRedisClient(ctx)
	if err != nil {
		return fmt.Errorf("error creating redis client: %v", err)
	}
	defer r.Client.Close()

	// Remove uploaded directory
	err = os.RemoveAll(filepath.Join(redisutils.UploadsReadyPath, download.Name))
	if err != nil {
		log.Printf("error removing download: %v", err)
		download.State = redisutils.Failed
		err := r.RegisterDownloadState(ctx, *download)
		if err != nil {
			log.Printf("error registering download state: %v", err)
		}
		return fmt.Errorf("failed to remove download: %v", err)
	}

	// Remove from Transmission
	c, err := transmission.NewTransmissionClient(ctx)
	if err != nil {
		log.Printf("error creating transmission client: %v", err)
		download.State = redisutils.Failed
		err := r.RegisterDownloadState(ctx, *download)
		if err != nil {
			log.Printf("error registering download state: %v", err)
		}
		return fmt.Errorf("failed to create transmission client: %v", err)
	}

	err = c.RemoveTorrents(ctx, []int64{download.ID})
	if err != nil {
		log.Printf("error removing torrent: %v", err)
		download.State = redisutils.Failed
		err := r.RegisterDownloadState(ctx, *download)
		if err != nil {
			log.Printf("error registering download state: %v", err)
		}
		return fmt.Errorf("failed to remove torrent: %v", err)
	}

	// Set state to removed
	download.State = redisutils.Removed
	err = r.RegisterDownloadState(ctx, *download)
	if err != nil {
		return fmt.Errorf("error registering download state: %v", err)
	}
	return nil
}
