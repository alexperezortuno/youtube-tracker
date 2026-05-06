package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/youtube"
)

type Discovery struct {
	KeyManager *youtube.KeyManager
	Redis      *cache.RedisClient
	Logger     *slog.Logger
}

func NewDiscovery(km *youtube.KeyManager, rc *cache.RedisClient) *Discovery {
	return NewDiscoveryWithLogger(km, rc, nil)
}

func NewDiscoveryWithLogger(km *youtube.KeyManager, rc *cache.RedisClient, logger *slog.Logger) *Discovery {
	if logger == nil {
		logger = slog.Default()
	}
	return &Discovery{
		KeyManager: km,
		Redis:      rc,
		Logger:     logger,
	}
}

func (d *Discovery) FindLiveStreams(ctx context.Context, channelID string) error {

	d.Logger.Debug("discovery started", "channel_id", channelID)

	strURL := "https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=%s&eventType=live&type=video&key=%s"

	maxTries := d.KeyManager.Count()
	tries := 0

	for {
		apiKey, err := d.KeyManager.NextKey()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		url := fmt.Sprintf(strURL, channelID, apiKey)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			d.KeyManager.MarkError(apiKey, 0)
			return err
		}

		// always read body to avoid leaks
		bodyBytes, readErr := io.ReadAll(resp.Body)
		err = resp.Body.Close()
		if err != nil {
			return err
		}

		if readErr != nil {
			d.KeyManager.MarkError(apiKey, 0)
			return readErr
		}

		// SUCCESS
		if resp.StatusCode == http.StatusOK {

			var data map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &data); err != nil {
				return err
			}

			d.KeyManager.MarkSuccess(apiKey)

			itemsRaw, ok := data["items"]
			if !ok || itemsRaw == nil {
				return nil
			}

			items, ok := itemsRaw.([]interface{})
			if !ok {
				return fmt.Errorf("unexpected type for items: %T", itemsRaw)
			}

			for _, item := range items {

				itemMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				idObj, ok := itemMap["id"].(map[string]interface{})
				if !ok {
					continue
				}

				videoID, ok := idObj["videoId"].(string)
				if !ok || videoID == "" {
					continue
				}

				_ = d.Redis.AddStream(ctx, videoID)
			}

			return nil
		}

		// HANDLE YOUTUBE ERROR
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {

			var errResp map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &errResp); err == nil {

				reason := extractReason(errResp)
				d.Logger.Warn("YouTube API error",
					"reason", reason,
					"status", resp.StatusCode,
				)

				if reason == "quotaExceeded" || reason == "dailyLimitExceeded" {
					d.KeyManager.MarkError(apiKey, 403)
				}
			} else {
				d.KeyManager.MarkError(apiKey, resp.StatusCode)
			}

			tries++
			if tries >= maxTries {
				return fmt.Errorf("all API keys exhausted")
			}

			continue
		}

		// OTHER ERRORS
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func extractReason(errResp map[string]interface{}) string {

	errorObj, ok := errResp["error"].(map[string]interface{})
	if !ok {
		return ""
	}

	errorsArr, ok := errorObj["errors"].([]interface{})
	if !ok || len(errorsArr) == 0 {
		return ""
	}

	first, ok := errorsArr[0].(map[string]interface{})
	if !ok {
		return ""
	}

	reason, _ := first["reason"].(string)
	return reason
}
