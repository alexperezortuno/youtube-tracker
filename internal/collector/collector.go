package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/models"
	"github.com/alexperezortuno/youtube-tracker/internal/youtube"
)

// CONFIG

const (
	maxBatchSize   = 50
	defaultWorkers = 5
	maxRetries     = 3
	baseBackoff    = 500 * time.Millisecond
)

// TYPES

type Collector struct {
	KeyManager *youtube.KeyManager
	HTTPClient *http.Client
	Workers    int
	RateLimit  <-chan time.Time // ticker channel
}

type youtubeResponse struct {
	Items []struct {
		ID string `json:"id"`

		Snippet struct {
			Title        string `json:"title"`
			ChannelTitle string `json:"channelTitle"`
			PublishedAt  string `json:"publishedAt"`
			ChannelID    string `json:"channelId"`
		} `json:"snippet"`

		Statistics struct {
			LikeCount     string `json:"likeCount"`
			ViewCount     string `json:"viewCount"`
			FavoriteCount string `json:"favoriteCount"`
			CommentCount  string `json:"commentCount"`
		} `json:"statistics"`

		LiveStreamingDetails struct {
			ConcurrentViewers string `json:"concurrentViewers"`
		} `json:"liveStreamingDetails"`
	} `json:"items"`
}

// CONSTRUCTOR

func NewCollector(apiKey *youtube.KeyManager, rps int, workers int) *Collector {
	if workers <= 0 {
		workers = defaultWorkers
	}

	return &Collector{
		KeyManager: apiKey,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		Workers:   workers,
		RateLimit: time.Tick(time.Second / time.Duration(rps)), // requests per second
	}
}

func parseResponse(data youtubeResponse) ([]models.Stream, []models.Metric) {

	var streams []models.Stream
	var metrics []models.Metric

	for _, item := range data.Items {

		viewers := parseInt(item.LiveStreamingDetails.ConcurrentViewers)
		likes := parseInt(item.Statistics.LikeCount)

		streams = append(streams, models.Stream{
			VideoID:      item.ID,
			VideoTitle:   item.Snippet.Title,
			ChannelTitle: item.Snippet.ChannelTitle,
		})

		metrics = append(metrics, models.Metric{
			VideoID:      item.ID,
			VideoTitle:   item.Snippet.Title,
			ChannelTitle: item.Snippet.ChannelTitle,
			Viewers:      viewers,
			Likes:        likes,
		})
	}

	return streams, metrics
}

func parseDailyResponse(data youtubeResponse) []models.Metric {
	var metrics []models.Metric

	for _, item := range data.Items {

		viewers := parseInt(item.Statistics.ViewCount)
		likes := parseInt(item.Statistics.LikeCount)

		metrics = append(metrics, models.Metric{
			VideoID:      item.ID,
			VideoTitle:   item.Snippet.Title,
			ChannelTitle: item.Snippet.ChannelTitle,
			Viewers:      viewers,
			Likes:        likes,
			Favorites:    new(parseInt(item.Statistics.FavoriteCount)),
			Comments:     new(parseInt(item.Statistics.CommentCount)),
			ChannelID:    new(item.Snippet.ChannelID),
			PublishedAt:  new(item.Snippet.PublishedAt),
		})
	}

	return metrics
}

// PUBLIC METHOD

func (c *Collector) FetchMetrics(ctx context.Context, videoIDs []string) ([]models.Metric, error) {

	if len(videoIDs) == 0 {
		return nil, nil
	}

	// 1. chunking
	batches := chunk(videoIDs, maxBatchSize)

	jobs := make(chan []string, len(batches))
	results := make(chan []models.Metric, len(batches))

	var wg sync.WaitGroup

	// 2. workers
	for i := 0; i < c.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range jobs {
				_, metrics := c.processBatchWithRetry(ctx, batch)
				if metrics != nil {
					results <- metrics
				}
			}
		}()
	}

	// 3. enqueue jobs
	for _, b := range batches {
		jobs <- b
	}
	close(jobs)

	// 4. wait workers
	wg.Wait()
	close(results)

	// 5. collect results
	var final []models.Metric
	for res := range results {
		final = append(final, res...)
	}

	return final, nil
}

// INTERNAL

func (c *Collector) processBatchWithRetry(ctx context.Context, batch []string) ([]models.Stream, []models.Metric) {

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {

		<-c.RateLimit // rate limiting

		streams, metrics, err := c.processBatch(ctx, batch)
		if err == nil {
			return streams, metrics
		}

		lastErr = err

		// backoff exponencial
		sleep := baseBackoff * time.Duration(1<<attempt)
		time.Sleep(sleep)
	}

	fmt.Printf("[WARN] batch failed after retries: %v\n", lastErr)
	return nil, nil
}

func (c *Collector) processBatch(ctx context.Context, videoIDs []string) ([]models.Stream, []models.Metric, error) {

	maxTries := c.KeyManager.Count()
	tries := 0

	strURL := "https://www.googleapis.com/youtube/v3/videos?part=liveStreamingDetails,statistics,snippet&id=%s&key=%s"
	ids := strings.Join(videoIDs, ",")

	for {

		apiKey := c.KeyManager.NextKey()

		url := fmt.Sprintf(strURL, ids, apiKey)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, nil, err
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			c.KeyManager.MarkError(apiKey)
			return nil, nil, err
		}

		// always read body to avoid leaks
		bodyBytes, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr != nil {
			c.KeyManager.MarkError(apiKey)
			return nil, nil, readErr
		}

		// SUCCESS
		if resp.StatusCode == http.StatusOK {

			var data youtubeResponse
			if err := json.Unmarshal(bodyBytes, &data); err != nil {
				return nil, nil, err
			}

			c.KeyManager.MarkSuccess(apiKey)

			streams, metrics := parseResponse(data)
			return streams, metrics, nil
		}

		// HANDLE YOUTUBE ERROR
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {

			// manage quotaExceeded
			var errResp map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &errResp); err == nil {

				if reason := extractReason(errResp); reason != "" {
					log.Printf("[YOUTUBE ERROR] reason=%s", reason)

					if reason == "quotaExceeded" {
						c.KeyManager.MarkError(apiKey)
					}
				}
			} else {
				c.KeyManager.MarkError(apiKey)
			}

			tries++
			if tries >= maxTries {
				return nil, nil, fmt.Errorf("all API keys exhausted")
			}

			continue
		}

		// OTHER ERRORS
		return nil, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (c *Collector) Fetch(ctx context.Context, videoIDs []string) ([]models.Stream, []models.Metric, error) {

	batches := chunk(videoIDs, maxBatchSize)

	var allStreams []models.Stream
	var allMetrics []models.Metric

	for _, batch := range batches {

		streams, metrics, err := c.processBatch(ctx, batch)
		if err != nil {
			continue
		}

		allStreams = append(allStreams, streams...)
		allMetrics = append(allMetrics, metrics...)
	}

	return allStreams, allMetrics, nil
}

func (c *Collector) FetchDaily(ctx context.Context, videoIDs []string) ([]models.Metric, error) {

	batches := chunk(videoIDs, maxBatchSize)
	var allMetrics []models.Metric

	for _, batch := range batches {

		metrics, err := c.processDailyBatch(ctx, batch)
		if err != nil {
			continue
		}

		allMetrics = append(allMetrics, metrics...)
	}

	return allMetrics, nil
}

func (c *Collector) processDailyBatch(ctx context.Context, videoIDs []string) ([]models.Metric, error) {
	maxTries := c.KeyManager.Count()
	tries := 0

	strURL := "https://www.googleapis.com/youtube/v3/videos?part=statistics,snippet&id=%s&key=%s"
	ids := strings.Join(videoIDs, ",")

	for {
		apiKey := c.KeyManager.NextKey()

		url := fmt.Sprintf(strURL, ids, apiKey)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			c.KeyManager.MarkError(apiKey)
			return nil, err
		}

		// always read body to avoid leaks
		bodyBytes, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr != nil {
			c.KeyManager.MarkError(apiKey)
			return nil, readErr
		}

		// SUCCESS
		if resp.StatusCode == http.StatusOK {

			var data youtubeResponse
			if err := json.Unmarshal(bodyBytes, &data); err != nil {
				return nil, err
			}

			c.KeyManager.MarkSuccess(apiKey)

			metrics := parseDailyResponse(data)
			return metrics, nil
		}

		// HANDLE YOUTUBE ERROR
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {

			// manage quotaExceeded
			var errResp map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &errResp); err == nil {

				if reason := extractReason(errResp); reason != "" {
					log.Printf("[YOUTUBE ERROR] reason=%s", reason)

					if reason == "quotaExceeded" {
						c.KeyManager.MarkError(apiKey)
					}
				}
			} else {
				c.KeyManager.MarkError(apiKey)
			}

			tries++
			if tries >= maxTries {
				return nil, fmt.Errorf("all API keys exhausted")
			}

			continue
		}

		// OTHER ERRORS
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// HELPERS

func parseMetrics(data youtubeResponse) []models.Metric {

	var result []models.Metric

	for _, item := range data.Items {

		viewers := parseInt(item.LiveStreamingDetails.ConcurrentViewers)
		likes := parseInt(item.Statistics.LikeCount)

		result = append(result, models.Metric{
			VideoID: item.ID,
			Viewers: viewers,
			Likes:   likes,
		})
	}

	return result
}

func parseInt(s string) int {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	if err != nil {
		return 0
	}
	return v
}

func chunk(ids []string, size int) [][]string {
	var chunks [][]string

	for i := 0; i < len(ids); i += size {
		end := i + size
		if end > len(ids) {
			end = len(ids)
		}
		chunks = append(chunks, ids[i:end])
	}

	return chunks
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
