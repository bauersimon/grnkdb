package cmd

import (
	goerrors "errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bauersimon/grnkdb/model"
	"github.com/bauersimon/grnkdb/scraper/youtube"
	"github.com/pkg/errors"
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
	youtubeWindowSize  uint
	youtubeChannelIDs  []string
)

func init() {
	rootCmd.AddCommand(scrapeCmd)

	scrapeCmd.Flags().StringVar(&csvDataPath, "data-path", "./public/data.csv", "data output path")
	scrapeCmd.Flags().StringVar(&youtubeApiKey, "youtube-api-key", "", "YouTube API key")
	scrapeCmd.MarkFlagRequired("youtube-api-key")
	scrapeCmd.Flags().UintVar(&youtubePageResults, "youtube-page-results", 50, "YouTube results per request")
	scrapeCmd.Flags().UintVar(&youtubePageLimit, "youtube-page-limit", 0, "YouTube page limit (disabled: 0)")
	scrapeCmd.Flags().UintVar(&youtubeWindowSize, "youtube-window-size", 100, "YouTube conversion window size")
	scrapeCmd.Flags().StringSliceVar(&youtubeChannelIDs, "youtube-channel-ids", []string{"UCYJ61XIK64sp6ZFFS8sctxw"}, "comma-separated list of channel IDs to scrape")
}

func scrape() (err error) {
	youtube, err := youtube.NewScraper(slog.Default(), youtubeApiKey, youtubePageLimit, youtubeWindowSize, youtubePageResults, youtubeChannelIDs)
	if err != nil {
		return err
	}
	games, err := youtube.Scrape()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(csvDataPath), 0700); err != nil {
		return errors.WithStack(err)
	}

	var existingData []*model.Game
	if _, err := os.Stat(csvDataPath); err == nil {
		readFile, err := os.Open(csvDataPath)
		if err != nil {
			return errors.WithStack(err)
		}
		existingData, err = model.CSVRead(readFile)
		closeErr := readFile.Close()
		if err != nil || closeErr != nil {
			return goerrors.Join(errors.WithStack(err), errors.WithStack(closeErr))
		}
	}
	if existingData != nil {
		games = model.MergeGames(games, existingData)
	}

	file, err := os.Create(csvDataPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		err = goerrors.Join(file.Close(), err)
	}()

	return model.CSVWrite(file, games)
}
