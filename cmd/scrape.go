package cmd

import (
	"fmt"
	"log/slog"

	"github.com/bauersimon/grnkdb/scraper/youtube"
	"github.com/spf13/cobra"
)

var (
	scrapeCmd = &cobra.Command{
		Use:   "scrape",
		Short: "Scrape the internet",
		RunE: func(cmd *cobra.Command, args []string) error {
			return scrape()
		},
	}

	csvDataPath        string
	youtubeApiKey      string
	youtubePageResults uint
	youtubePageLimit   uint
	youtubeChannelIDs  []string
)

func init() {
	rootCmd.AddCommand(scrapeCmd)

	scrapeCmd.Flags().StringVar(&csvDataPath, "data-path", "data.csv", "data output path")
	scrapeCmd.Flags().StringVar(&youtubeApiKey, "youtube-api-key", "", "YouTube API key")
	scrapeCmd.MarkFlagRequired("youtube-api-key")
	scrapeCmd.Flags().UintVar(&youtubePageResults, "youtube-page-results", 50, "YouTube results per request")
	scrapeCmd.Flags().UintVar(&youtubePageLimit, "youtube-page-limit", 0, "YouTube page limit (disabled: 0)")
	scrapeCmd.Flags().StringSliceVar(&youtubeChannelIDs, "youtube-channel-ids", []string{"UCYJ61XIK64sp6ZFFS8sctxw"}, "comma-separated list of channel IDs to scrape")
}

func scrape() error {
	youtube, err := youtube.NewScraper(slog.Default(), youtubeApiKey, youtubePageLimit, youtubePageResults, youtubeChannelIDs)
	if err != nil {
		return err
	}
	games, err := youtube.Scrape()
	if err != nil {
		return err
	}

	fmt.Println("scraped", len(games))

	return nil
}
