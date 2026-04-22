package source

type ChannelSource interface {
	GetChannelIDs() ([]string, error)
}
