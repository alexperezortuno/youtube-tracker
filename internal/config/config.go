package config

import (
	"os"
	"strings"
)

type Config struct {
	YouTubeAPIKey   string
	RedisAddr       string
	PostgresURL     string
	ChannelIDs      []string
	ChannelFilePath string
}

func Load() Config {

	channelIDsEnv := os.Getenv("YOUTUBE_CHANNEL_IDS")
	channelFile := getEnv("YOUTUBE_CHANNEL_FILE", "channels.txt")

	var channelIDs []string

	if channelIDsEnv != "" {
		channelIDs = parseCSV(channelIDsEnv)
	} else {
		channelIDs = loadFromFile(channelFile)
	}

	return Config{
		YouTubeAPIKey:   os.Getenv("YOUTUBE_API_KEY"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		PostgresURL:     getEnv("POSTGRES_URL", "postgres://user:pass@localhost:5432/metrics?sslmode=disable&options=--search_path=metrics_db"),
		ChannelIDs:      channelIDs,
		ChannelFilePath: channelFile,
	}
}

func parseCSV(val string) []string {
	parts := strings.Split(val, ",")
	var result []string

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
