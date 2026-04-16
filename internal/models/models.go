package models

type Stream struct {
	VideoID   string
	ChannelID string
}

type Metric struct {
	VideoID string
	Viewers int
	Likes   int
}
