package model

import (
	"encoding/csv"
	goerrors "errors"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const timeDateFormat = "02.01.2006"

func header(knownSourceTypes []SourceType) (h []string) {
	h = []string{"titel"}
	for _, sourceType := range knownSourceTypes {
		h = append(h, []string{
			string(sourceType) + "-link",
			string(sourceType) + "-start",
		}...)
	}
	return h
}

// CSVWrite writes game information as CSV format.
func CSVWrite(writer io.Writer, games []*Game) error {
	return internalCSVWrite(writer, games, allSourceTypes)
}

func internalCSVWrite(writer io.Writer, games []*Game, knownSourceTypes []SourceType) (err error) {
	csvWriter := csv.NewWriter(writer)
	defer func() {
		csvWriter.Flush()
		err = goerrors.Join(err, errors.WithStack(csvWriter.Error()))
	}()

	if err := csvWriter.Write(header(knownSourceTypes)); err != nil {
		return errors.WithStack(err)
	}

	slices.SortStableFunc(games, func(a, b *Game) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, game := range games {
		row := []string{game.Name}
		slices.SortStableFunc(game.Content, func(a, b *Content) int {
			return strings.Compare(string(a.Source), string(b.Source))
		})
		for _, content := range game.Content {
			if slices.IndexFunc(knownSourceTypes, func(a SourceType) bool {
				return a == content.Source
			}) == -1 {
				return errors.Errorf("unknown source type %q", content.Source)
			}

			row = append(row, []string{
				content.Link,
				content.Start.Format(timeDateFormat),
			}...)
		}

		if err := csvWriter.Write(row); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// CSVRead reads game information from CSV format.
func CSVRead(reader io.Reader) (games []*Game, err error) {
	return internalCSVRead(reader, allSourceTypes)
}

func internalCSVRead(reader io.Reader, knownSourceTypes []SourceType) (games []*Game, err error) {
	csvReader := csv.NewReader(reader)

	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, errors.WithStack(err)
	} else if len(rows) == 0 {
		return nil, errors.New("empty CSV data")
	}

	headers := rows[0]
	var sourceFromHeaders []SourceType
	for i := 1; i < len(headers); i = i + 2 {
		s := strings.SplitN(headers[i], "-", 2)[0]
		if slices.IndexFunc(knownSourceTypes, func(st SourceType) bool {
			return string(st) == s
		}) == -1 {
			return nil, errors.Errorf("unknown source type %q", s)
		}
		sourceFromHeaders = append(sourceFromHeaders, SourceType(s))
	}

	for _, row := range rows[1:] {
		game := &Game{
			Name: row[0],
		}
		for i, source := range sourceFromHeaders {
			started, _ := time.Parse(timeDateFormat, row[i*2+2])
			game.Content = append(game.Content, &Content{
				Source: source,
				Link:   row[i*2+1],
				Start:  started,
			})
		}
		slices.SortStableFunc(game.Content, func(a, b *Content) int {
			return strings.Compare(string(a.Source), string(b.Source))
		})

		games = append(games, game)
	}

	slices.SortStableFunc(games, func(a, b *Game) int {
		return strings.Compare(a.Name, b.Name)
	})
	return games, nil
}
