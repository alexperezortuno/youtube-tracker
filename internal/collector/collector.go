package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/alexperezortuno/youtube-tracker/internal/models"
)

type Collector struct {
	APIKey string
}

func (c *Collector) FetchMetrics(ctx context.Context, videoIDs []string) ([]models.Metric, error) {

	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?part=liveStreamingDetails,statistics&id=%s&key=%s",
		strings.Join(videoIDs, ","), c.APIKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var result []models.Metric

	items := data["items"].([]interface{})
	for _, item := range items {
		m := item.(map[string]interface{})

		id := m["id"].(string)

		stats := m["statistics"].(map[string]interface{})
		live := m["liveStreamingDetails"].(map[string]interface{})

		viewers := 0
		if v, ok := live["concurrentViewers"]; ok {
			fmt.Sscanf(v.(string), "%d", &viewers)
		}

		likes := 0
		if l, ok := stats["likeCount"]; ok {
			fmt.Sscanf(l.(string), "%d", &likes)
		}

		result = append(result, models.Metric{
			VideoID: id,
			Viewers: viewers,
			Likes:   likes,
		})
	}

	return result, nil
}
