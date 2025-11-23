package cmd

import (
	"bytes"
	goerrors "errors"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/bauersimon/grnkdb/model"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type WebCommand struct {
	logger *zap.Logger

	DataPath     string `long:"data-path" default:"./public/data.json" description:"Data input path"`
	TemplatePath string `long:"template-path" default:"./web/html" description:"Template path"`
	HTMLPath     string `long:"html-path" default:"./public" description:"HTML output path"`
	Live         bool   `long:"live" short:"l" description:"Re-generate periodically"`
}

func NewWebCommand(logger *zap.Logger) flags.Commander {
	return &WebCommand{
		logger: logger,
	}
}

func (cmd *WebCommand) Execute(args []string) error {
	return cmd.web(cmd.DataPath, cmd.TemplatePath, cmd.HTMLPath, cmd.Live)
}

func (cmd *WebCommand) web(gameDataPath, templateDataPath, htmlDataPath string, loopGeneration bool) (err error) {
	for {
		err = webLoop(gameDataPath, templateDataPath, htmlDataPath)
		if !loopGeneration {
			break
		} else {
			if err != nil {
				cmd.logger.Error("web generation failed", zap.Error(err))
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
	games, err := model.JSONRead(bytes.NewBuffer(data))
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
