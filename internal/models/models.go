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
}

type VideoDailyStat struct {
	VideoID   string
	Date      time.Time
	Views     int64
	Likes     int64
	Favorites *int
	Comments  *int
}
