package discovery

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/logger"
	"github.com/alexperezortuno/youtube-tracker/internal/youtube"
)

type youtubeRSSFeed struct {
	Entries []youtubeRSSEntry `xml:"entry"`
}

type youtubeRSSEntry struct {
	VideoID string `xml:"http://www.youtube.com/xml/schemas/2015 videoId"`
	Title   string `xml:"title"`
	PubDate string `xml:"published"`
}

type Discovery struct {
	KeyManager *youtube.KeyManager
	Redis      *cache.RedisClient
}

func NewDiscovery(km *youtube.KeyManager, rc *cache.RedisClient) *Discovery {
	return NewDiscoveryWithLogger(km, rc)
}

func NewDiscoveryWithLogger(km *youtube.KeyManager, rc *cache.RedisClient) *Discovery {
	return &Discovery{
		KeyManager: km,
		Redis:      rc,
	}
}

func (d *Discovery) FindLiveStreams(ctx context.Context, channelID string, discoverInterval int) error {

	logger.Debug("discovery started channel_id %s", channelID)

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

				_ = d.Redis.AddStream(ctx, videoID, discoverInterval)
			}

			return nil
		}

		// HANDLE YOUTUBE ERROR
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {

			var errResp map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &errResp); err == nil {

				reason := extractReason(errResp)
				logger.Warn("YouTube API error | reason %s | status %d", reason, resp.StatusCode)

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

func (d *Discovery) FindLiveStreamsByRSS(ctx context.Context, channelID string, discoverInterval int) error {
	logger.Debug("rss discovery started channel_id %s", channelID)

	videoIDs, err := d.fetchRecentVideoIDsFromRSS(ctx, channelID, discoverInterval)
	if err != nil {
		return err
	}

	if len(videoIDs) == 0 {
		return nil
	}

	for _, videoID := range videoIDs {
		_ = d.Redis.AddStream(ctx, videoID, discoverInterval)
	}

	return nil
}

func (d *Discovery) fetchRecentVideoIDsFromRSS(ctx context.Context, channelID string, discoverInterval int) ([]string, error) {
	url := fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", channelID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if closeErr != nil {
		return nil, closeErr
	}

	if readErr != nil {
		return nil, readErr
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected RSS status code: %d", resp.StatusCode)
	}

	var feed youtubeRSSFeed
	if err := xml.Unmarshal(bodyBytes, &feed); err != nil {
		return nil, err
	}

	videoIDs := make([]string, 0, len(feed.Entries))
	seen := make(map[string]struct{}, len(feed.Entries))

	for _, entry := range feed.Entries {
		videoID := strings.TrimSpace(entry.VideoID)
		if videoID == "" {
			continue
		}

		if _, exists := seen[videoID]; exists {
			continue
		}

		pubDate, err := time.Parse(time.RFC3339, entry.PubDate)
		if err != nil {
			logger.Warn("invalid pubDate format for video %s: %v", videoID, err)
			continue
		}

		if time.Since(pubDate) > time.Duration(discoverInterval+60)*time.Minute {
			continue
		}

		seen[videoID] = struct{}{}
		videoIDs = append(videoIDs, videoID)
	}

	return videoIDs, nil
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
