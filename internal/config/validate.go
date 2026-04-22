package config

import "log"

func ValidateChannelIDs(ids []string) {
	for _, id := range ids {
		if len(id) < 10 {
			log.Printf("[WARN] suspicious channel id: %s", id)
		}
	}
}
