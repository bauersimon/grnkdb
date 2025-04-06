package model

import "time"

// Game represents a game.
type Game struct {
	// Name is the title of the game.
	Name string
	// Content is the content produced with this game.
	Content []*Content
}

// SourceType is a source type.
type SourceType string

var (
	// SourceYouTube is the YouTube source.
	SourceYouTube = SourceType("youtube")
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
