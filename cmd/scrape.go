package cmd

import (
	"go.uber.org/zap"
)

// ScrapeCommand is the parent command for all scraping operations
type ScrapeCommand struct {
	YouTube YouTubeCommand `command:"youtube" description:"Scrape YouTube channels and output CSV files"`
}

func NewScrapeCommand(logger *zap.Logger) *ScrapeCommand {
	return &ScrapeCommand{
		YouTube: YouTubeCommand{
			logger: logger,
		},
	}
}
