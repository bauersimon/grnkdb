package cmd

import (
	goerrors "errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bauersimon/grnkdb/converter"
	"github.com/bauersimon/grnkdb/model"
	"github.com/bauersimon/grnkdb/steam"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert CSV files to games JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			inputDir, _ := cmd.Flags().GetString("input")
			output, _ := cmd.Flags().GetString("output")
			windowSize, _ := cmd.Flags().GetUint("window-size")

			videoConverter := converter.NewVideoToGameConverter(steam.NewClient(), windowSize, slog.Default())

			return convertCSVToGames(videoConverter, inputDir, output)
		},
	}
)

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().String("input", "./data", "input directory containing CSV files")
	convertCmd.Flags().String("output", "./public/data.json", "output JSON file path")
	convertCmd.Flags().Uint("window-size", 100, "conversion window size")
}

func convertCSVToGames(converter converter.Interface, inputDir, outputPath string) (err error) {
	csvFiles, err := filepath.Glob(filepath.Join(inputDir, "*.csv"))
	if err != nil {
		return errors.WithStack(err)
	}

	if len(csvFiles) == 0 {
		slog.Warn("no CSV files found", "directory", inputDir)
		return nil
	}

	var allVideos []*model.Video
	for _, csvFile := range csvFiles {
		slog.Debug("reading CSV file", "file", csvFile)

		file, err := os.Open(csvFile)
		if err != nil {
			return errors.Wrapf(err, "failed to open CSV file %s", csvFile)
		}
		defer func() {
			if errClose := file.Close(); errClose != nil {
				err = goerrors.Join(err, errors.WithStack(errClose))
			}
		}()

		videos, err := model.VideoCSVRead(file)
		if err != nil {
			return errors.Wrapf(err, "failed to read CSV file %s", csvFile)
		}

		allVideos = append(allVideos, videos...)
		slog.Debug("loaded videos", "file", filepath.Base(csvFile), "count", len(videos))
	}

	if len(allVideos) == 0 {
		slog.Warn("no videos found in CSV files")
		return nil
	}

	slog.Info("converting videos to games", "total_videos", len(allVideos))
	games, err := converter.Convert(allVideos)
	if err != nil {
		return err
	}

	slog.Info("conversion completed", "games", len(games))

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return errors.WithStack(err)
	}

	// Load existing data if it exists.
	var existingData []*model.Game
	if _, err := os.Stat(outputPath); err == nil {
		slog.Info("loading existing data", "file", outputPath)
		readFile, err := os.Open(outputPath)
		if err != nil {
			return errors.WithStack(err)
		}
		existingData, err = model.JSONRead(readFile)
		closeErr := readFile.Close()
		if err != nil || closeErr != nil {
			return goerrors.Join(errors.WithStack(err), errors.WithStack(closeErr))
		}
		slog.Info("loaded existing games", "count", len(existingData))
	}

	if existingData != nil {
		games = model.MergeGames(games, existingData)
		slog.Info("merged with existing data", "total_games", len(games))
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		err = goerrors.Join(file.Close(), err)
	}()

	if err := model.JSONWrite(file, games); err != nil {
		return err
	}

	slog.Info("conversion completed successfully", "output", outputPath, "games", len(games))
	return nil
}
