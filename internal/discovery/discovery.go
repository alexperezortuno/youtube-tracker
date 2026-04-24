package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/youtube"
)

type Discovery struct {
	KeyManager *youtube.KeyManager
	Redis      *cache.RedisClient
}

func (d *Discovery) FindLiveStreams(ctx context.Context, channelID string) error {

	log.Printf("[DISCOVERY] channel=%s", channelID)

	strURL := "https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=%s&eventType=live&type=video&key=%s"

	maxTries := d.KeyManager.Count()
	tries := 0

	for {

		apiKey := d.KeyManager.NextKey()

		url := fmt.Sprintf(strURL, channelID, apiKey)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			d.KeyManager.MarkError(apiKey)
			return err
		}

		// always read body to avoid leaks
		bodyBytes, readErr := io.ReadAll(resp.Body)
		err = resp.Body.Close()
		if err != nil {
			return err
		}

		if readErr != nil {
			d.KeyManager.MarkError(apiKey)
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
				log.Printf("[YOUTUBE ERROR] reason=%s", reason)

				if reason == "quotaExceeded" || reason == "dailyLimitExceeded" {
					d.KeyManager.MarkError(apiKey)
				}
			} else {
				d.KeyManager.MarkError(apiKey)
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
