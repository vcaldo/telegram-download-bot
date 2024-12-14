package fileutils

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vcaldo/telegram-download-bot/bot/pkg/redisutils"
)

func CompressDownload(ctx context.Context, download *redisutils.Download) error {
	log.Printf("Compressing download: %s\n", download.Name)
	r, err := redisutils.NewAuthenticatedRedisClient(ctx)
	if err != nil {
		return fmt.Errorf("new redis client failed: %w", err)
	}

	// Set download state to compressing
	log.Printf("Setting download state to compressing: %s\n", download.Name)
	download.State = redisutils.Compressing
	if err := r.RegisterDownloadState(ctx, *download); err != nil {
		return fmt.Errorf("set download state failed: %w", err)
	}

	source := filepath.Join(redisutils.CompletedDownloadsPath, download.Name)
	destination := filepath.Join(redisutils.UploadsReadyPath, download.Name)
	if err := compressAndSplitDownload(ctx, source, destination); err != nil {
		return fmt.Errorf("compress and split failed: %w", err)

	}

	// Set download state to compressed
	log.Printf("Setting download state to compressed: %s\n", download.Name)
	download.State = redisutils.Compressed
	if err := r.RegisterDownloadState(ctx, *download); err != nil {
		return fmt.Errorf("set download state failed: %w", err)
	}
	return nil
}

func compressAndSplitDownload(ctx context.Context, source, destination string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}

	// Prepare 7za command with volume size parameter
	cmd := exec.Command("7zz",
		"a",                               // add to archive
		"-y",                              // assume Yes on all queries
		"-mx0",                            // no compression
		"-v2000m",                         // split into 2gb volumes
		"-t7z",                            // use 7z format
		fmt.Sprintf("%s.7z", destination), // output file
		source,                            // input file/directory
	)

	// Capture command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compression failed: %v\nOutput: %s", err, output)
	}

	log.Printf("Compression completed: %s\n", output)
	return nil
}
