package transmission

import (
	"context"
	"os"
	"time"

	"github.com/hekmon/transmissionrpc"
)

type Client struct {
	client *transmissionrpc.Client
}

func NewClient(ctx context.Context, address, username, password string) (*Client, error) {
	client, err := transmissionrpc.New(address, username, password, &transmissionrpc.AdvancedConfig{
		HTTPTimeout: 60 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

func NewTransmissionClient(ctx context.Context) (*Client, error) {
	url := os.Getenv("TRANSMISSION_URL")
	user := os.Getenv("TRANSMISSION_USER")
	pass := os.Getenv("TRANSMISSION_PASS")
	return NewClient(ctx, url, user, pass)
}

func (c *Client) AddTorrent(ctx context.Context, torrentURL string) (*transmissionrpc.Torrent, error) {
	torrent, err := c.client.TorrentAdd(&transmissionrpc.TorrentAddPayload{
		Filename: &torrentURL,
	})
	if err != nil {
		return nil, err
	}
	return torrent, nil
}

func (c *Client) AddTorrentFromFile(ctx context.Context, torrentPath string) (*transmissionrpc.Torrent, error) {
	torrent, err := c.client.TorrentAddFile(torrentPath)
	if err != nil {
		return nil, err
	}
	return torrent, nil
}

func (c *Client) GetSessionStats(ctx context.Context) (*transmissionrpc.SessionStats, error) {
	stats, err := c.client.SessionStats()
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (c *Client) GetCompletedDownloads(ctx context.Context) ([]*transmissionrpc.Torrent, error) {
	torrents, err := c.client.TorrentGet([]string{"id", "name", "percentDone"}, nil)
	if err != nil {
		return nil, err
	}

	var completed []*transmissionrpc.Torrent
	for _, t := range torrents {
		if *t.PercentDone == 1.0 {
			completed = append(completed, t)
		}
	}
	return completed, nil
}

func (c *Client) RemoveTorrents(ctx context.Context, id []int64) error {
	err := c.client.TorrentRemove(&transmissionrpc.TorrentRemovePayload{
		IDs:             id,
		DeleteLocalData: true,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetTorrentInfo(ctx context.Context, id int64) (*transmissionrpc.Torrent, error) {
	torrent, err := c.client.TorrentGet([]string{"id", "name", "totalSize"}, []int64{id})
	if err != nil {
		return nil, err
	}
	return torrent[0], nil
}
