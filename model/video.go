package model

import "time"

// Video represents a video from a content platform.
type Video struct {
	// Link is the full URL to the video.
	Link string `csv:"Link"`
	// PublishedAt is when the video was published.
	PublishedAt time.Time `csv:"PublishedAt"`

	// Title is the video title.
	Title string `csv:"Title"`
	// Description is the video description.
	Description string `csv:"Description"`

	// ChannelID is the identifier of the channel/creator.
	ChannelID string `csv:"ChannelID"`
	// VideoID is the unique identifier for the video on the platform.
	VideoID string `csv:"VideoID"`
}
