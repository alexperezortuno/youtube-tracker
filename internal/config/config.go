package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	YouTubeAPIKeys  []string
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

	keys := parseCSV(os.Getenv("YOUTUBE_API_KEYS"))

	if len(keys) <= 0 {
		log.Fatal("no YOUTUBE_API_KEYS provided")
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
	log.Println("Redis URL:", urlRedis)

	return Config{
		YouTubeAPIKeys:  keys,
		RedisAddr:       getEnv("REDIS_ADDR", urlRedis),
		PostgresURL:     getEnv("POSTGRES_URL", urlDb),
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
