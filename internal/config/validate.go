package config

import (
	log "github.com/alexperezortuno/youtube-tracker/internal/logger"
)

func ValidateChannelIDs(ids []string) {
	logger := log.Default()
	for _, id := range ids {
		if len(id) < 10 {
			logger.Warn("suspicious channel ID", "channel_id", id)
		}
	}
}
