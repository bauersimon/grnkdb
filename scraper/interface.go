package scraper

import "github.com/bauersimon/grnkdb/model"

// Interface defines a generic scraper.
type Interface interface {
	// Videos extracts video metadata from the specified channels.
	Videos(channelIDs []string) ([]*model.Video, error)
}
