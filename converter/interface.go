package converter

import "github.com/bauersimon/grnkdb/model"

// Interface defines a generic video-to-game converter.
type Interface interface {
	// Convert transforms video metadata into game information.
	Convert(videos []*model.Video) ([]*model.Game, error)
}
