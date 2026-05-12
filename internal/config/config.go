package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/alexperezortuno/youtube-tracker/internal/logger"
)

type Config struct {
	YouTubeAPIKeys  []string
	RedisAddr       string
	PostgresURL     string
	ChannelIDs      []string
	ChannelNames    []string
	ChannelFilePath string
}

func Load() Config {

	channelIDsEnv := os.Getenv("YOUTUBE_CHANNEL_IDS")
	channelFile := getEnv("YOUTUBE_CHANNEL_FILE", "channels.txt")
	channelFileNames := getEnv("YOUTUBE_CHANNEL_FILE_NAMES", "channel_names.txt")

	var channelIDs []string
	var channelNames []string

	if channelIDsEnv != "" {
		channelIDs = parseCSV(channelIDsEnv)
	} else {
		channelIDs = loadFromFile(channelFile)
		channelNames = loadFromFile(channelFileNames)
	}

	keys := parseCSV(os.Getenv("YOUTUBE_API_KEYS"))

	if len(keys) <= 0 {
		logger.Error("no YOUTUBE_API_KEYS provided")
	}

	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "user")
	dbPass := getEnv("DB_PASSWORD", "pass")
	dbName := getEnv("DB_NAME", "metrics_db")

	urlDb := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/metrics?sslmode=disable&options=--search_path=%s",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	urlRedis := fmt.Sprintf("%s:%s", redisHost, redisPort)
	logger.Debug("Redis URL: %s", urlRedis)

	return Config{
		YouTubeAPIKeys:  keys,
		RedisAddr:       getEnv("REDIS_ADDR", urlRedis),
		PostgresURL:     getEnv("POSTGRES_URL", urlDb),
		ChannelIDs:      channelIDs,
		ChannelFilePath: channelFile,
		ChannelNames:    channelNames,
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
