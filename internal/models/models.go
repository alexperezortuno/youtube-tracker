package models

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
}
