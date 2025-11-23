package converter

import (
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/bauersimon/grnkdb/model"
	"github.com/bauersimon/grnkdb/steam"
	"github.com/bauersimon/grnkdb/util"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var steamStoreLinkRE = regexp.MustCompile(`steampowered\.com\/app\/(\d+)`)

// VideoToGameConverter converts video metadata to game information.
type VideoToGameConverter struct {
	steamClient *steam.Client
	windowSize  uint
	logger      *zap.Logger
}

var _ Interface = (*VideoToGameConverter)(nil)

// NewVideoToGameConverter creates a new video-to-game converter.
func NewVideoToGameConverter(steamClient *steam.Client, windowSize uint, logger *zap.Logger) *VideoToGameConverter {
	return &VideoToGameConverter{
		steamClient: steamClient,
		windowSize:  windowSize,
		logger:      logger,
	}
}

// Convert transforms video metadata into game information.
func (c *VideoToGameConverter) Convert(videos []*model.Video) (games []*model.Game, err error) {
	// Create cleaned copies of videos for processing without modifying originals
	cleanedVideos := make([]*model.Video, len(videos))
	for i, video := range videos {
		cleanedVideos[i] = &model.Video{
			VideoID:     video.VideoID,
			Title:       video.Title,
			Description: video.Description,
			Link:        video.Link,
			PublishedAt: video.PublishedAt,
			ChannelID:   video.ChannelID,
			Source:      video.Source,
		}
	}

	c.logger.Debug("cleaning up video meta")
	cleanupVideoMeta(cleanedVideos)

	c.logger.Info("converting videos to games", zap.Int("videos", len(videos)))
	for window := range util.SlidingWindowed(cleanedVideos, c.windowSize, max(uint(0), c.windowSize/2)) {
		g, err := c.convertVideosToGames(window)
		if err != nil {
			return nil, err
		}
		games = model.MergeGames(games, g)
	}

	return games, nil
}

// convertVideosToGames converts model.Video structs to games
func (c *VideoToGameConverter) convertVideosToGames(videos []*model.Video) (games []*model.Game, err error) {
	earliestVideoForGame := map[string]*model.Video{}
	for i, video := range videos {
		c.logger.Debug("extracting game information",
			zap.String("video", video.VideoID),
			zap.Int("progress", i+1),
			zap.Int("total", len(videos)))
		var newGameSpecifier string

		// Try to extract Steam links from description.
		if matches := steamStoreLinkRE.FindStringSubmatch(video.Description); len(matches) > 0 {
			name, err := c.steamClient.GameName(matches[1])
			if err != nil {
				c.logger.Error("cannot get name from steam",
					zap.String("video", video.VideoID),
					zap.Error(err))
			} else {
				newGameSpecifier = strings.ToLower(name)
				c.logger.Debug("found game information on steam",
					zap.String("video", video.VideoID),
					zap.String("game", name))
			}
		}

		// Brute force try to find similar pre- or suffixes among all found games.
		var oldGameSpecifier string
		if newGameSpecifier == "" {
			for preSuffix := range maps.Keys(earliestVideoForGame) {
				newPrefix := longestCommonPrefix(
					strings.ToLower(video.Title),
					strings.ToLower(preSuffix),
				)
				newSuffix := longestCommonSuffix(
					strings.ToLower(video.Title),
					strings.ToLower(preSuffix),
				)
				if commonWords[strings.ToLower(strings.TrimSpace(newPrefix))] {
					newPrefix = ""
				}
				if commonWords[strings.ToLower(strings.TrimSpace(newSuffix))] {
					newSuffix = ""
				}

				if len(strings.TrimSpace(newPrefix)) > 2 &&
					len(newPrefix) > len(newGameSpecifier) &&
					len(newPrefix) > len(newSuffix) {

					newGameSpecifier = newPrefix
					oldGameSpecifier = preSuffix
				} else if len(strings.TrimSpace(newSuffix)) > 2 &&
					len(newSuffix) > len(newGameSpecifier) {

					newGameSpecifier = newSuffix
					oldGameSpecifier = preSuffix
				}
			}
		}

		if newGameSpecifier != "" {
			if oldGameSpecifier != "" && newGameSpecifier != oldGameSpecifier { // Shorten the specifier.
				earliestVideo := earliestVideoForGame[oldGameSpecifier]
				delete(earliestVideoForGame, oldGameSpecifier)
				earliestVideoForGame[newGameSpecifier] = earliestVideo
			}
			if earlierVideo := earliestVideoForGame[newGameSpecifier]; earlierVideo != nil {
				if compareVideos(video, earlierVideo) < 0 { // Found earlier video.
					earliestVideoForGame[newGameSpecifier] = video
				}
			} else {
				earliestVideoForGame[newGameSpecifier] = video
			}
			c.logger.Debug("match", zap.String("video", video.Title))
		} else {
			earliestVideoForGame[video.Title] = video
			c.logger.Debug("no match", zap.String("video", video.Title))
		}
	}

	caser := cases.Title(language.German)
	for title, video := range earliestVideoForGame {
		games = append(games, &model.Game{
			Name: caser.String(strings.TrimSpace(strings.Trim(title, "-:\" \t"))),
			Content: []*model.Content{
				{
					Link:   video.Link,
					Start:  video.PublishedAt,
					Source: video.Source,
				},
			},
		})
	}
	slices.SortFunc(games, func(a, b *model.Game) int {
		return strings.Compare(a.Name, b.Name)
	})

	return games, nil
}

func compareVideos(a, b *model.Video) int {
	if a.PublishedAt.Before(b.PublishedAt) {
		return -1
	} else if b.PublishedAt.Before(a.PublishedAt) {
		return +1
	}

	return 0
}

// longestCommonPrefix finds the longest common prefix string amongst two input strings.
func longestCommonPrefix(str1, str2 string) string {
	minLen := min(len(str1), len(str2))

	i := 0
	for i < minLen {
		if str1[i] == str2[i] {
			i++
		} else {
			i = max(0, i-1)
			break
		}
	}

	return str1[:i]
}

// longestCommonSuffix finds the longest common suffix string amongst two input strings.
func longestCommonSuffix(str1, str2 string) string {
	return reverseString(longestCommonPrefix(reverseString(str1), reverseString(str2)))
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
