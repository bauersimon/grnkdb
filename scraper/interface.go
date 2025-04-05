package scraper

import "github.com/bauersimon/grnkdb/model"

// Interface defines a generic scraper.
type Interface interface {
	// Scrape extracts game information.
	Scrape() ([]*model.Game, error)
}
