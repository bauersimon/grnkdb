package cmd

import (
	goerrors "errors"
	"os"
	"path/filepath"

	"github.com/bauersimon/grnkdb/converter"
	"github.com/bauersimon/grnkdb/model"
	"github.com/bauersimon/grnkdb/steam"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ConvertCommand struct {
	logger *zap.Logger

	Input      string `long:"input" default:"./data" description:"Input directory containing CSV files"`
	Output     string `long:"output" default:"./public/data.json" description:"Output JSON file path"`
	WindowSize uint   `long:"window-size" default:"100" description:"Conversion window size"`
}

func NewConvertCommand(logger *zap.Logger) flags.Commander {
	return &ConvertCommand{
		logger: logger,
	}
}

func (cmd *ConvertCommand) Execute(args []string) error {
	videoConverter := converter.NewVideoToGameConverter(steam.NewClient(), cmd.WindowSize, cmd.logger)

	return cmd.convertCSVToGames(videoConverter, cmd.Input, cmd.Output)
}

func (cmd *ConvertCommand) convertCSVToGames(converter converter.Interface, inputDir, outputPath string) (err error) {
	csvFiles, err := filepath.Glob(filepath.Join(inputDir, "*.csv"))
	if err != nil {
		return errors.WithStack(err)
	}

	if len(csvFiles) == 0 {
		cmd.logger.Warn("no CSV files found", zap.String("directory", inputDir))

		return nil
	}

	var allVideos []*model.Video
	for _, csvFile := range csvFiles {
		cmd.logger.Debug("reading CSV file", zap.String("file", csvFile))

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
		cmd.logger.Debug("loaded videos",
			zap.String("file", filepath.Base(csvFile)),
			zap.Int("count", len(videos)))
	}

	if len(allVideos) == 0 {
		cmd.logger.Warn("no videos found in CSV files")

		return nil
	}

	cmd.logger.Info("converting videos to games", zap.Int("videos", len(allVideos)))
	games, err := converter.Convert(allVideos)
	if err != nil {
		return err
	}

	cmd.logger.Info("conversion completed", zap.Int("games", len(games)))

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return errors.WithStack(err)
	}

	// Load existing data if it exists.
	var existingData []*model.Game
	if _, err := os.Stat(outputPath); err == nil {
		cmd.logger.Info("loading existing data", zap.String("file", outputPath))
		readFile, err := os.Open(outputPath)
		if err != nil {
			return errors.WithStack(err)
		}
		existingData, err = model.JSONRead(readFile)
		closeErr := readFile.Close()
		if err != nil || closeErr != nil {
			return goerrors.Join(errors.WithStack(err), errors.WithStack(closeErr))
		}
		cmd.logger.Info("loaded existing games", zap.Int("count", len(existingData)))
	}

	if existingData != nil {
		games = model.MergeGames(games, existingData)
		cmd.logger.Info("merged with existing data", zap.Int("games", len(games)))
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

	cmd.logger.Info("conversion completed successfully",
		zap.String("output", outputPath),
		zap.Int("games", len(games)))
	return nil
}
