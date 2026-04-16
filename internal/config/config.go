package config

import "os"

type Config struct {
	YouTubeAPIKey string
	RedisAddr     string
	PostgresURL   string
}

func Load() Config {
	return Config{
		YouTubeAPIKey: os.Getenv("YOUTUBE_API_KEY"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		PostgresURL:   os.Getenv("POSTGRES_URL"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
