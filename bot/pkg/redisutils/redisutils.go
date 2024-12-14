package redisutils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

const (
	Downloaded             = "downloaded"
	Compressing            = "compressing"
	Compressed             = "compressed"
	Uploading              = "uploading"
	Uploaded               = "uploaded"
	Removed                = "removed"
	NameKey                = "name"
	StateKey               = "state"
	ComplatedDownloadsPath = "/downloads/complete"
	UploadsReadyPath       = "/downloads/uploads"
	DownloadsHash          = "downloads"
)

type Download struct {
	ID    int64
	Name  string
	State string
}

type RedisClient struct {
	Client *redis.Client
}

func NewAuthenticatedRedisClient(ctx context.Context) (*RedisClient, error) {
	addr := os.Getenv("REDIS_ADDR")
	password := ""
	db := 0
	return newRedisClient(ctx, addr, password, db)
}

func newRedisClient(ctx context.Context, addr, password string, db int) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return &RedisClient{Client: rdb}, nil
}

func (r *RedisClient) DownloadExists(ctx context.Context, id int64) (bool, error) {
	val, err := r.Client.HExists(ctx, fmt.Sprintf("%s:%d", DownloadsHash, id), StateKey).Result()
	if err != nil {
		return false, fmt.Errorf("redis check failed: %w", err)
	}

	return val, nil
}

func (r *RedisClient) RegisterDownloadState(ctx context.Context, d Download) error {
	log.Printf("Storing download in Redis: %s", d.Name)
	err := r.Client.HSet(ctx, fmt.Sprintf("%s:%d", DownloadsHash, d.ID), []string{
		NameKey, d.Name,
		StateKey, d.State}).Err()
	if err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	switch d.State {
	case Downloaded:
		err = r.Client.SAdd(ctx, Downloaded, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
	case Compressing:
		err = r.Client.SAdd(ctx, Compressing, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
		err = r.Client.SRem(ctx, Downloaded, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
	case Compressed:
		err = r.Client.SAdd(ctx, Compressed, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
		err = r.Client.SRem(ctx, Compressing, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
	case Uploading:
		err = r.Client.SAdd(ctx, Uploading, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
		err = r.Client.SRem(ctx, Compressed, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
	case Uploaded:
		err = r.Client.SAdd(ctx, Uploaded, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
		err = r.Client.SRem(ctx, Uploading, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
	case Removed:
		err = r.Client.SAdd(ctx, Removed, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
		err = r.Client.SRem(ctx, Uploaded, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
	default:
		log.Println("Unknown state")
	}

	return nil
}

func (r *RedisClient) GetDownloadState(ctx context.Context, state string) ([]int64, error) {
	val, err := r.Client.SMembers(ctx, state).Result()
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	// Convert string slice to int64 slice
	var ids []int64
	for _, v := range val {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (r *RedisClient) GetDownloadName(ctx context.Context, id int64) (string, error) {
	val, err := r.Client.HGet(ctx, fmt.Sprintf("%s:%d", DownloadsHash, id), NameKey).Result()
	if err != nil {
		return "", fmt.Errorf("redis get failed: %w", err)
	}

	return val, nil
}
