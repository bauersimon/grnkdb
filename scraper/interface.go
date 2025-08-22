package scraper

import "github.com/bauersimon/grnkdb/model"

// Interface defines a generic scraper.
type Interface interface {
	// Scrape extracts game information from the specified channels.
	Scrape(channelIDs []string) ([]*model.Game, error)
	// Videos extracts video metadata from the specified channels.
	Videos(channelIDs []string) ([]*model.Video, error)
}
