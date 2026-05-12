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
	"github.com/alexperezortuno/youtube-tracker/internal/models"
	"github.com/alexperezortuno/youtube-tracker/internal/youtube"
	"github.com/go-rod/rod"
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

func (d *Discovery) SaveLiveStreamsByExtractor(ctx context.Context, results []models.Result, discoverInterval int) error {
	logger.Debug("extractor discovery started channel_ids %v", results)

	// Save ids in Redis
	for _, result := range results {
		_ = d.Redis.AddStream(ctx, result.VideoID, discoverInterval)
	}
	return nil
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

func (d *Discovery) GetLiveVideoID(page *rod.Page, channelIdentifier string) (*models.Result, error) {
	// Build URL
	var url string
	if strings.HasPrefix(channelIdentifier, "http") {
		url = channelIdentifier
	} else if strings.HasPrefix(channelIdentifier, "@") {
		url = fmt.Sprintf("https://www.youtube.com/%s", channelIdentifier)
	} else {
		url = fmt.Sprintf("https://www.youtube.com/channel/%s", channelIdentifier)
	}

	fmt.Printf("Searching: %s\n", url)

	// Navigate to page
	if err := page.Navigate(url); err != nil {
		return nil, fmt.Errorf("error navigating: %w", err)
	}

	// Wait for initial load
	if err := page.WaitLoad(); err != nil {
		if strings.Contains(err.Error(), "Execution context was destroyed") {
			logger.Warn("context destroyed waiting for initial load for %s, continue: %v", channelIdentifier, err)
		} else {
			return nil, fmt.Errorf("error waiting load: %w", err)
		}
	}

	time.Sleep(3 * time.Second)

	// Scroll to load more videos
	if _, err := page.Eval(`() => window.scrollBy(0, 600)`); err != nil {
		if strings.Contains(err.Error(), "Execution context was destroyed") {
			logger.Warn("context destroyed waiting for initial load for %s, continue: %v", channelIdentifier, err)
		} else {
			return nil, fmt.Errorf("error scrolling: %w", err)
		}
	}

	time.Sleep(2 * time.Second)

	// Correct method with Eval()
	result, err := page.Eval(`() => {
            // Search all load classes .ytBadgeShapeText
            const badges = document.querySelectorAll('.ytBadgeShapeText');
            for (let badge of badges) {
                if (badge.textContent.trim().toUpperCase() === 'EN VIVO' || badge.textContent.trim().toUpperCase() === 'LIVE') {
                    // Upload to parent link
                    let parent = badge.closest('a[href*="/watch?v="]');
                    if (parent) {
                        const match = parent.href.match(/watch\?v=([^&]+)/);
                        if (match) return match[1];
                    }
                }
            }
            return '';
        }`)

	if err != nil {
		if strings.Contains(err.Error(), "Execution context was destroyed") {
			logger.Warn("destroyed context evaluating live video for %s: %v", channelIdentifier, err)
			return nil, nil
		}

		return nil, fmt.Errorf("error evaluating live video: %w", err)
	}

	// Convert result to string
	videoID := result.Value.String()
	if videoID != "" && videoID != "<nil>" {
		return &models.Result{
			Channel: channelIdentifier,
			VideoID: videoID,
			URL:     fmt.Sprintf("https://youtube.com/watch?v=%s", videoID),
		}, nil
	}

	return nil, nil
}

func extractVideoID(url string) string {
	// Buscar watch?v= en la URL
	if idx := strings.Index(url, "watch?v="); idx != -1 {
		start := idx + 8
		end := strings.Index(url[start:], "&")
		if end == -1 {
			end = len(url[start:])
		}
		return url[start : start+end]
	}

	// Buscar /live/ o /shorts/
	if idx := strings.Index(url, "/live/"); idx != -1 {
		start := idx + 6
		end := strings.Index(url[start:], "?")
		if end == -1 {
			end = len(url[start:])
		}
		return url[start : start+end]
	}

	return ""
}
