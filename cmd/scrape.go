package cmd

import (
	"fmt"

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

	csvData string
)

func init() {
	rootCmd.AddCommand(scrapeCmd)

	scrapeCmd.Flags().StringVar(&csvData, "data-path", "data.csv", "Data")
}

func scrape() error {
	fmt.Println(csvData)

	return nil
}
