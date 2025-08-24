package scraper

import "github.com/bauersimon/grnkdb/model"

// Interface defines a generic scraper.
type Interface interface {
	// Videos extracts video metadata from a single channel.
	Videos(channelID string) ([]*model.Video, error)
}
