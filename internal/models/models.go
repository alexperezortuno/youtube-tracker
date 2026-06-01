package models

import "time"

type Stream struct {
	VideoID      string
	ChannelID    string
	VideoTitle   string
	ChannelTitle string
}

type Metric struct {
	VideoID      string
	ChannelTitle string
	VideoTitle   string
	Viewers      int
	Likes        int
	Favorites    *int
	Comments     *int
	ChannelID    *string
	PublishedAt  *string
}

type VideoDailyStat struct {
	VideoID     string
	Date        time.Time
	Views       int64
	Likes       int64
	Favorites   *int
	Comments    *int
	ChannelID   *string
	PublishedAt *string
}

type Result struct {
	Channel string
	VideoID string
	URL     string
}

type Channel struct {
	ID         string
	Name       string
	Active     bool
	Category   *string
	Language   *string
	Country    *string
	FollowedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
