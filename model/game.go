package model

// Game represents a game.
type Game struct {
	// Name is the title of the game.
	Name string
	// Content is the content produced with this game.
	Content []*Content
}

// SourceType is a source type.
type SourceType string

// Content represents content.
type Content struct {
	// Link is a URL to the content.
	Link string
	// Source is the source of the content.
	Source SourceType
}
