package cmd

import (
	"bytes"
	goerrors "errors"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/bauersimon/grnkdb/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	webCmd = &cobra.Command{
		Use:   "web",
		Short: "Generate website",
		RunE: func(cmd *cobra.Command, args []string) error {
			gameDataPath, _ := cmd.Flags().GetString("data-path")
			templateDataPath, _ := cmd.Flags().GetString("template-path")
			htmlDataPath, _ := cmd.Flags().GetString("html-path")
			loopGeneration, _ := cmd.Flags().GetBool("live")

			return web(gameDataPath, templateDataPath, htmlDataPath, loopGeneration)
		},
	}
)

func init() {
	rootCmd.AddCommand(webCmd)

	webCmd.Flags().String("data-path", "./public/data.csv", "data input path")
	webCmd.Flags().String("template-path", "./web/html", "template path")
	webCmd.Flags().String("html-path", "./public", "html output path")
	webCmd.Flags().BoolP("live", "l", false, "re-generate periodically")
}

func web(gameDataPath, templateDataPath, htmlDataPath string, loopGeneration bool) (err error) {
	for {
		err = webLoop(gameDataPath, templateDataPath, htmlDataPath)
		if !loopGeneration {
			break
		} else {
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(time.Second)
		}
	}

	return err
}

func webLoop(gameDataPath, templateDataPath, htmlDataPath string) (err error) {
	t, err := template.ParseGlob(filepath.Join(templateDataPath, "*.html"))
	if err != nil {
		return errors.WithStack(err)
	}

	if err := os.MkdirAll(htmlDataPath, 0755); err != nil {
		return errors.WithStack(err)
	}
	file, err := os.Create(filepath.Join(htmlDataPath, "index.html"))
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		err = goerrors.Join(err, errors.WithStack(file.Close()))
	}()

	data, err := os.ReadFile(gameDataPath)
	if err != nil {
		return errors.WithStack(err)
	}
	games, err := model.CSVRead(bytes.NewBuffer(data))
	if err != nil {
		return errors.WithStack(err)
	}

	if err := t.Execute(file, games); err != nil {
		return errors.WithStack(err)
	}

	if out, err := exec.Command("tailwindcss", []string{
		"-i", filepath.Join(templateDataPath, "style.css"),
		"-o", filepath.Join(htmlDataPath, "style.css"),
	}...).CombinedOutput(); err != nil {
		return errors.WithStack(errors.WithMessage(err, string(out)))
	}

	return nil
}
