package model

import (
	"encoding/json"
	"io"
	"slices"
	"strings"

	"github.com/pkg/errors"
)

// JSONWrite writes game information in JSON format.
func JSONWrite(writer io.Writer, games []*Game) error {
	// Sort for consistent output.
	slices.SortStableFunc(games, func(a, b *Game) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, game := range games {
		slices.SortStableFunc(game.Content, func(a, b *Content) int {
			return strings.Compare(string(a.Source), string(b.Source))
		})
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(games); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// JSONRead reads game information from JSON format.
func JSONRead(reader io.Reader) ([]*Game, error) {
	var games []*Game

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&games); err != nil {
		return nil, errors.WithStack(err)
	}

	// Sort for consistent output.
	slices.SortStableFunc(games, func(a, b *Game) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, game := range games {
		slices.SortStableFunc(game.Content, func(a, b *Content) int {
			return strings.Compare(string(a.Source), string(b.Source))
		})
	}

	return games, nil
}
