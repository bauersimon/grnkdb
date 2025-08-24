package youtube

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/api/youtube/v3"

	"github.com/bauersimon/grnkdb/model"
	"github.com/bauersimon/grnkdb/scraper"
	"github.com/pkg/errors"
)

// Scraper is a YouTube scraper.
type Scraper struct {
	service *youtube.Service

	pageLimit   uint
	pageResults uint

	logger *slog.Logger
}

var _ scraper.Interface = (*Scraper)(nil)

// NewScraper initializes a YouTube scraper.
func NewScraper(logger *slog.Logger, apiKey string, pageLimit uint, pageResults uint) (*Scraper, error) {
	service, err := initializeService(context.Background(), apiKey)
	if err != nil {
		return nil, err
	}

	return &Scraper{
		service: service,

		pageLimit:   pageLimit,
		pageResults: pageResults,

		logger: logger,
	}, nil
}

// Videos extracts video metadata from a single YouTube channel.
func (s *Scraper) Videos(channelID string) ([]*model.Video, error) {
	playlistItems, err := s.scrapeChannel(channelID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("converting playlist items to videos", "items", len(playlistItems))
	videos := make([]*model.Video, 0, len(playlistItems))
	for _, item := range playlistItems {
		video, err := convertPlaylistItemToVideo(item)
		if err != nil {
			s.logger.Warn("failed to convert playlist item to video", "videoId", item.Snippet.ResourceId.VideoId, "error", err)
			continue
		}
		videos = append(videos, video)
	}

	return videos, nil
}

func convertPlaylistItemToVideo(item *youtube.PlaylistItem) (*model.Video, error) {
	publishedAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse published date: %s", item.Snippet.PublishedAt)
	}

	return &model.Video{
		VideoID:     item.Snippet.ResourceId.VideoId,
		Title:       item.Snippet.Title,
		Description: item.Snippet.Description,
		Link:        fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.Snippet.ResourceId.VideoId),
		PublishedAt: publishedAt,
		ChannelID:   item.Snippet.ChannelId,
		Source:      model.SourceYouTube,
	}, nil
}

func (s *Scraper) scrapeChannel(id string) (videos []*youtube.PlaylistItem, err error) {
	s.logger.Info("scraping channel", "id", id)
	defer func() {
		s.logger.Info("scraping channel done", "id", id, "videos", len(videos))
	}()

	response, err := s.service.Channels.List([]string{"contentDetails"}).Id(id).Do()
	if err != nil {
		return nil, errors.WithStack(err)
	} else if len(response.Items) == 0 {
		return nil, errors.Errorf("channel not found %q", id)
	}

	uploadsPlaylistID := response.Items[0].ContentDetails.RelatedPlaylists.Uploads
	var nextPageToken string
	page := 0
	for {
		page++

		s.logger.Debug("scraping channel page", "page", page)
		call := s.service.PlaylistItems.List([]string{"snippet"}).
			PlaylistId(uploadsPlaylistID).
			MaxResults(int64(s.pageResults))
		if nextPageToken != "" {
			call = call.PageToken(nextPageToken)
		}

		playlistResult, err := call.Do()
		if err != nil {
			return videos, errors.Wrap(err, "error fetching playlist items")
		} else if len(playlistResult.Items) == 0 {
			break
		}
		s.logger.Debug("scraping channel page successful", "page", page, "videos", len(playlistResult.Items), "sample", playlistResult.Items[0].Snippet.Title)

		videos = append(videos, playlistResult.Items...)

		nextPageToken = playlistResult.NextPageToken
		if nextPageToken == "" {
			break
		} else if s.pageLimit != 0 && page > int(s.pageLimit)-1 {
			break
		}
	}

	return videos, nil
}
