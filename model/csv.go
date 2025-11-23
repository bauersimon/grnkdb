package model

import (
	"io"
	"slices"
	"strings"

	"github.com/jszwec/csvutil"
	"github.com/pkg/errors"
)

// VideoCSVWrite writes video information as CSV format.
func VideoCSVWrite(writer io.Writer, videos []*Video) error {
	// Sort videos by VideoID for consistent output
	slices.SortStableFunc(videos, func(a, b *Video) int {
		return strings.Compare(a.VideoID, b.VideoID)
	})

	data, err := csvutil.Marshal(videos)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = writer.Write(data)
	return errors.WithStack(err)
}

// VideoCSVRead reads video information from CSV format.
func VideoCSVRead(reader io.Reader) ([]*Video, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var videos []*Video
	err = csvutil.Unmarshal(data, &videos)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Sort videos by VideoID for consistent output
	slices.SortStableFunc(videos, func(a, b *Video) int {
		return strings.Compare(a.VideoID, b.VideoID)
	})

	return videos, nil
}
