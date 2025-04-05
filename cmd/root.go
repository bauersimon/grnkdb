package cmd

import (
	"log/slog"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "grnkdb",
		Short: "Gronkh database scraping.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			level := slog.LevelInfo
			if verbose {
				level = slog.LevelDebug
			} else if quiet {
				level = slog.LevelError
			}

			if logPath == "" {
				slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
					Level: level,
				})))
			} else {
				logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0700)
				if err != nil {
					return errors.WithStack(err)
				}
				slog.SetDefault(slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
					Level: level,
				})))
			}
			slog.SetLogLoggerLevel(level)

			return nil
		},
	}

	verbose bool
	quiet   bool
	logPath string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "log only errors")
	rootCmd.MarkFlagsMutuallyExclusive("verbose", "quiet")
	rootCmd.PersistentFlags().StringVar(&logPath, "log-path", "", "JSON log file path")
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
