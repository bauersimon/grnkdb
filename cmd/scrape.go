package cmd

import (
	"github.com/spf13/cobra"
)

var (
	scrapeCmd = &cobra.Command{
		Use:   "scrape",
		Short: "Scrape video data from various platforms",
	}
)

func init() {
	rootCmd.AddCommand(scrapeCmd)
	scrapeCmd.AddCommand(scrapeYoutubeCmd)
}
