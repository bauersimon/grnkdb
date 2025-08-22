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
			csvDataPath, _ := cmd.Flags().GetString("data-path")
			youtubeApiKey, _ := cmd.Flags().GetString("youtube-api-key")
			youtubePageResults, _ := cmd.Flags().GetUint("youtube-page-results")
			youtubePageLimit, _ := cmd.Flags().GetUint("youtube-page-limit")
			youtubeWindowSize, _ := cmd.Flags().GetUint("youtube-window-size")
			youtubeChannelIDs, _ := cmd.Flags().GetStringSlice("youtube-channel-ids")

			return scrape(csvDataPath, youtubeApiKey, youtubePageResults, youtubePageLimit, youtubeWindowSize, youtubeChannelIDs)
		},
	}
)

func init() {
	rootCmd.AddCommand(scrapeCmd)

	scrapeCmd.Flags().String("data-path", "./public/data.json", "data output path")
	scrapeCmd.Flags().String("youtube-api-key", "", "YouTube API key")
	_ = scrapeCmd.MarkFlagRequired("youtube-api-key")
	scrapeCmd.Flags().Uint("youtube-page-results", 50, "YouTube results per request")
	scrapeCmd.Flags().Uint("youtube-page-limit", 0, "YouTube page limit (disabled: 0)")
	scrapeCmd.Flags().Uint("youtube-window-size", 100, "YouTube conversion window size")
	scrapeCmd.Flags().StringSlice("youtube-channel-ids", []string{"UCYJ61XIK64sp6ZFFS8sctxw"}, "comma-separated list of channel IDs to scrape")
}

func scrape(csvDataPath, youtubeApiKey string, youtubePageResults, youtubePageLimit, youtubeWindowSize uint, youtubeChannelIDs []string) (err error) {
	youtube, err := youtube.NewScraper(slog.Default(), youtubeApiKey, youtubePageLimit, youtubePageResults, youtubeWindowSize)
	if err != nil {
		return err
	}
	games, err := youtube.Scrape(youtubeChannelIDs)
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
		existingData, err = model.JSONRead(readFile)
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

	return model.JSONWrite(file, games)
}
