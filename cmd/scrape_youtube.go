package cmd

import (
	goerrors "errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bauersimon/grnkdb/model"
	"github.com/bauersimon/grnkdb/scraper"
	"github.com/bauersimon/grnkdb/scraper/youtube"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type YouTubeCommand struct {
	logger *zap.Logger

	APIKey      string `long:"api-key" description:"YouTube API key"`
	Output      string `long:"output" default:"./data" description:"Output directory for CSV files"`
	PageResults uint   `long:"page-results" default:"50" description:"YouTube results per request"`
	PageLimit   uint   `long:"page-limit" default:"0" description:"YouTube page limit (disabled: 0)"`

	Args struct {
		ChannelIDs []string `positional-arg-name:"channel-id" required:"yes" description:"YouTube channel IDs to scrape"`
	} `positional-args:"yes"`
}

func (cmd *YouTubeCommand) Execute(args []string) error {
	youtubeScraper, err := youtube.NewScraper(cmd.logger, cmd.APIKey, cmd.PageLimit, cmd.PageResults)
	if err != nil {
		return err
	}

	return cmd.scrapeYoutube(youtubeScraper, cmd.Output, cmd.Args.ChannelIDs)
}

func (cmd *YouTubeCommand) scrapeYoutube(scraper scraper.Interface, outputDir string, channelIDs []string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return errors.WithStack(err)
	}

	var scrapeErrors []error

	for _, channelID := range channelIDs {
		cmd.logger.Info("scraping channel", zap.String("channel", channelID))

		videos, err := scraper.Videos(channelID)
		if err != nil {
			scrapeErr := errors.Wrapf(err, "failed to scrape channel %s", channelID)
			cmd.logger.Error("channel scraping failed",
				zap.String("channel", channelID),
				zap.Error(scrapeErr))
			scrapeErrors = append(scrapeErrors, scrapeErr)
			continue
		}

		outputFile := filepath.Join(outputDir, fmt.Sprintf("%s.csv", channelID))
		file, err := os.Create(outputFile)
		if err != nil {
			fileErr := errors.Wrapf(err, "failed to create output file for channel %s", channelID)
			cmd.logger.Error("file creation failed",
				zap.String("channel", channelID),
				zap.Error(fileErr))
			scrapeErrors = append(scrapeErrors, fileErr)
			continue
		}

		if err := model.VideoCSVWrite(file, videos); err != nil {
			_ = file.Close()
			csvErr := errors.Wrapf(err, "failed to write CSV for channel %s", channelID)
			cmd.logger.Error("CSV writing failed",
				zap.String("channel", channelID),
				zap.Error(csvErr))
			scrapeErrors = append(scrapeErrors, csvErr)
			continue
		}

		if err := file.Close(); err != nil {
			closeErr := errors.Wrapf(err, "failed to close file for channel %s", channelID)
			cmd.logger.Error("file closing failed",
				zap.String("channel", channelID),
				zap.Error(closeErr))
			scrapeErrors = append(scrapeErrors, closeErr)
			continue
		}

		cmd.logger.Info("wrote CSV file",
			zap.String("file", outputFile),
			zap.Int("videos", len(videos)))
	}

	if len(scrapeErrors) > 0 {
		return errors.Wrap(goerrors.Join(scrapeErrors...), "encountered errors")
	}

	return nil
}
