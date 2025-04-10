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
			return web()
		},
	}

	gameDataPath     string
	templateDataPath string
	htmlDataPath     string
	loopGeneration   bool
)

func init() {
	rootCmd.AddCommand(webCmd)

	webCmd.Flags().StringVar(&gameDataPath, "data-path", "data.csv", "data input path")
	webCmd.Flags().StringVar(&templateDataPath, "template-path", "./web/html", "template path")
	webCmd.Flags().StringVar(&htmlDataPath, "html-path", "./public", "html output path")
	webCmd.Flags().BoolVarP(&loopGeneration, "live", "l", false, "re-generate periodically")
}

func web() (err error) {
	for {
		err = webLoop()
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

func webLoop() (err error) {
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
