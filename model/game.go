package model

import (
	"slices"
	"strings"
	"time"
)

// Game represents a game.
type Game struct {
	// Name is the title of the game.
	Name string
	// Content is the content produced with this game.
	Content []*Content
}

// SourceType is a source type.
type SourceType string

var allSourceTypes []SourceType

func sourceType(s string) SourceType {
	newSourceType := SourceType(s)
	allSourceTypes = append(allSourceTypes, newSourceType)
	slices.SortStableFunc(allSourceTypes, func(a, b SourceType) int {
		return strings.Compare(string(a), string(b))
	})
	return newSourceType
}

var (
	// SourceYouTube is the YouTube source.
	SourceYouTube = sourceType("youtube")
)

// Content represents content.
type Content struct {
	// Link is a URL to the content.
	Link string
	// Start denotes when the content was first released.
	Start time.Time
	// Source is the source of the content.
	Source SourceType
}

// MergeGames merges two slices of games.
func MergeGames(a []*Game, b []*Game) []*Game {
	merged := append(a, b...)
	slices.SortStableFunc(merged, func(a *Game, b *Game) int {
		return strings.Compare(a.Name, b.Name)
	})

	return slices.CompactFunc(merged, func(a *Game, b *Game) bool {
		if a.Name != b.Name {
			return false
		}

		a.Content = append(a.Content, b.Content...)
		slices.SortStableFunc(a.Content, func(a *Content, b *Content) int {
			return strings.Compare(string(a.Source), string(b.Source))
		})

		a.Content = slices.CompactFunc(a.Content, func(a *Content, b *Content) bool {
			if a.Source != b.Source {
				return false
			}

			if b.Start.Before(a.Start) {
				a.Start = b.Start
				a.Link = b.Link
			}

			return true
		})

		return true
	})
}
