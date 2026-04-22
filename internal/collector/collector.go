package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/models"
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
	APIKey     string
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
		} `json:"snippet"`

		Statistics struct {
			LikeCount string `json:"likeCount"`
		} `json:"statistics"`

		LiveStreamingDetails struct {
			ConcurrentViewers string `json:"concurrentViewers"`
		} `json:"liveStreamingDetails"`
	} `json:"items"`
}

// CONSTRUCTOR

func NewCollector(apiKey string, rps int, workers int) *Collector {
	if workers <= 0 {
		workers = defaultWorkers
	}

	return &Collector{
		APIKey: apiKey,
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

	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?part=liveStreamingDetails,statistics,snippet&id=%s&key=%s",
		strings.Join(videoIDs, ","), c.APIKey,
	)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var data youtubeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, nil, err
	}

	streams, metrics := parseResponse(data)

	return streams, metrics, nil
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
	fmt.Sscanf(s, "%d", &v)
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
