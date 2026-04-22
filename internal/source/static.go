package source

import "github.com/alexperezortuno/youtube-tracker/internal/config"

type StaticSource struct {
	Config config.Config
}

func (s *StaticSource) GetChannelIDs() ([]string, error) {
	return s.Config.ChannelIDs, nil
}
