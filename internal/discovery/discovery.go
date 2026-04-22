package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
)

type Discovery struct {
	APIKey string
	Redis  *cache.RedisClient
}

func (d *Discovery) FindLiveStreams(ctx context.Context, channelID string) error {
	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=%s&eventType=live&type=video&key=%s",
		channelID, d.APIKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("[INFO] error: %v", err)
		}
	}(resp.Body)

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	itemsRaw, ok := data["items"]
	if !ok || itemsRaw == nil {
		// No hay items en respuesta; no es fatal
		return nil
	}

	items, ok := itemsRaw.([]interface{})
	if !ok {
		return fmt.Errorf("unexpected type for items: %T", itemsRaw)
	}

	for _, item := range items {
		videoID := item.(map[string]interface{})["id"].(map[string]interface{})["videoId"].(string)
		d.Redis.AddStream(ctx, videoID)
	}

	return nil
}
