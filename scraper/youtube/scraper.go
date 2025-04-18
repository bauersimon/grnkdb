package youtube

import (
	"context"
	"log/slog"

	"google.golang.org/api/youtube/v3"

	"github.com/bauersimon/grnkdb/model"
	"github.com/bauersimon/grnkdb/scraper"
	"github.com/bauersimon/grnkdb/steam"
	"github.com/bauersimon/grnkdb/util"
	"github.com/pkg/errors"
)

// Scraper is a YouTube scraper.
type Scraper struct {
	service *youtube.Service

	pageLimit   uint
	pageResults uint
	channelIDs  []string
	windowSize  uint

	logger     *slog.Logger
	loggerRoot *slog.Logger
}

var _ scraper.Interface = (*Scraper)(nil)

// NewScraper initializes a YouTube scraper.
func NewScraper(logger *slog.Logger, apiKey string, pageLimit uint, pageResults uint, windowSize uint, channelIDs []string) (*Scraper, error) {
	service, err := initializeService(context.Background(), apiKey)
	if err != nil {
		return nil, err
	}

	return &Scraper{
		service: service,

		pageLimit:   pageLimit,
		pageResults: pageResults,
		channelIDs:  channelIDs,
		windowSize:  windowSize,

		logger:     logger.With("module", "scraper.youtube.Scraper"),
		loggerRoot: logger,
	}, nil
}

// Scrape extracts game information.
func (s *Scraper) Scrape() ([]*model.Game, error) {
	var videos []*youtube.PlaylistItem
	for _, channelID := range s.channelIDs {
		v, err := s.scrapeChannel(channelID)
		if err != nil {
			return nil, err
		}
		videos = append(videos, v...)
	}

	s.logger.Info("converting videos", "videos", len(videos))
	var games []*model.Game
	steamClient := steam.NewClient()
	logger := s.loggerRoot.With("module", "scraper.youtube.Convert")
	for window := range util.SlidingWindowed(videos, s.windowSize, max(uint(0), s.windowSize/2)) {
		g, err := convertVideosToGames(logger, steamClient, window)
		if err != nil {
			return nil, err
		}
		games = model.MergeGames(games, g)
	}

	return games, nil
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

func parseGames(items []*youtube.PlaylistItem) (games []*model.Game) {
	for _, item := range items {
		games = append(games, &model.Game{
			Name: item.Snippet.Title,
		})
	}

	return games
}
