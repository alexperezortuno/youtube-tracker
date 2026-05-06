package source

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/logger"
)

type ChannelWatcher struct {
	FilePath    string
	LastModTime time.Time
}

func NewChannelWatcher(path string) *ChannelWatcher {
	return NewChannelWatcherWithLogger(path)
}

func NewChannelWatcherWithLogger(path string) *ChannelWatcher {
	return &ChannelWatcher{
		FilePath: path,
	}
}

func (w *ChannelWatcher) HasChanged() bool {
	info, err := os.Stat(w.FilePath)
	if err != nil {
		logger.Error("error reading file error: %v", err)
		return false
	}

	modTime := info.ModTime()

	if w.LastModTime.IsZero() {
		w.LastModTime = modTime
		return false
	}

	if modTime.After(w.LastModTime) {
		w.LastModTime = modTime
		return true
	}

	return false
}

func (w *ChannelWatcher) Reload() []string {
	file, err := os.Open(w.FilePath)
	if err != nil {
		logger.Error("error reading file error: %v", err)
		return nil
	}
	defer file.Close()

	var result []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		result = append(result, line)
	}

	return result
}
