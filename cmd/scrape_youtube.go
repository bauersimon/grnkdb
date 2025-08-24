package cmd

import (
	goerrors "errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bauersimon/grnkdb/model"
	"github.com/bauersimon/grnkdb/scraper"
	"github.com/bauersimon/grnkdb/scraper/youtube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	scrapeYoutubeCmd = &cobra.Command{
		Use:   "youtube [channel-id...]",
		Short: "Scrape YouTube channels and output CSV files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, _ := cmd.Flags().GetString("api-key")
			outputDir, _ := cmd.Flags().GetString("output")
			pageResults, _ := cmd.Flags().GetUint("page-results")
			pageLimit, _ := cmd.Flags().GetUint("page-limit")

			youtubeScraper, err := youtube.NewScraper(slog.Default(), apiKey, pageLimit, pageResults)
			if err != nil {
				return err
			}

			return scrapeYoutube(youtubeScraper, outputDir, args)
		},
	}
)

func init() {
	scrapeYoutubeCmd.Flags().String("api-key", "", "YouTube API key")
	_ = scrapeYoutubeCmd.MarkFlagRequired("api-key")
	scrapeYoutubeCmd.Flags().String("output", "./data", "output directory for CSV files")
	scrapeYoutubeCmd.Flags().Uint("page-results", 50, "YouTube results per request")
	scrapeYoutubeCmd.Flags().Uint("page-limit", 0, "YouTube page limit (disabled: 0)")
}

func scrapeYoutube(scraper scraper.Interface, outputDir string, channelIDs []string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return errors.WithStack(err)
	}

	var scrapeErrors []error

	for _, channelID := range channelIDs {
		slog.Info("scraping channel", "channel", channelID)

		videos, err := scraper.Videos(channelID)
		if err != nil {
			scrapeErr := errors.Wrapf(err, "failed to scrape channel %s", channelID)
			slog.Error("channel scraping failed", "channel", channelID, "error", scrapeErr)
			scrapeErrors = append(scrapeErrors, scrapeErr)
			continue
		}

		outputFile := filepath.Join(outputDir, fmt.Sprintf("%s.csv", channelID))
		file, err := os.Create(outputFile)
		if err != nil {
			fileErr := errors.Wrapf(err, "failed to create output file for channel %s", channelID)
			slog.Error("file creation failed", "channel", channelID, "error", fileErr)
			scrapeErrors = append(scrapeErrors, fileErr)
			continue
		}

		if err := model.VideoCSVWrite(file, videos); err != nil {
			_ = file.Close()
			csvErr := errors.Wrapf(err, "failed to write CSV for channel %s", channelID)
			slog.Error("CSV writing failed", "channel", channelID, "error", csvErr)
			scrapeErrors = append(scrapeErrors, csvErr)
			continue
		}

		if err := file.Close(); err != nil {
			closeErr := errors.Wrapf(err, "failed to close file for channel %s", channelID)
			slog.Error("file closing failed", "channel", channelID, "error", closeErr)
			scrapeErrors = append(scrapeErrors, closeErr)
			continue
		}

		slog.Info("wrote CSV file", "file", outputFile, "videos", len(videos))
	}

	if len(scrapeErrors) > 0 {
		return errors.Wrap(goerrors.Join(scrapeErrors...), "encountered errors")
	}

	return nil
}
