package main

import (
	"os"

	"github.com/bauersimon/grnkdb/cmd"
	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync() // Ignore sync errors on exit.
	}()

	parser := flags.NewNamedParser("grnkdb", flags.Default)

	if _, err := parser.AddCommand(
		"convert",
		"Convert videos into games.",
		"Convert video from CSV into games as JSON.",
		cmd.NewConvertCommand(logger),
	); err != nil {
		panic(err)
	}

	if _, err := parser.AddCommand(
		"scrape",
		"Scrape videos.",
		"Scrape videos and store them as CSV.",
		cmd.NewScrapeCommand(logger),
	); err != nil {
		panic(err)
	}

	if _, err := parser.AddCommand(
		"web",
		"Render the website.",
		"Render the website game representation of JSON.",
		cmd.NewWebCommand(logger),
	); err != nil {
		panic(err)
	}

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		logger.Error("command failed", zap.Error(err))
		os.Exit(1)
	}
}
