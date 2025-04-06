package youtube

import (
	"fmt"
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/bauersimon/grnkdb/model"
	"google.golang.org/api/youtube/v3"
)

func convertVideosToGames(logger *slog.Logger, videos []*youtube.PlaylistItem) (games []*model.Game, err error) {
	logger.Debug("cleaning up video meta")
	cleanupVideoMeta(videos)

	logger.Debug("extracting unique games via video prefix")
	earliestVideoForPrefix := map[string]*youtube.PlaylistItem{}
	for _, video := range videos {
		matched := false
		for prefix := range maps.Keys(earliestVideoForPrefix) {
			if newPrefix := longestCommonPrefix(video.Snippet.Title, prefix); newPrefix != "" {
				matched = true

				if newPrefix != prefix {
					earliestVideo := earliestVideoForPrefix[prefix]
					delete(earliestVideoForPrefix, prefix)
					earliestVideoForPrefix[newPrefix] = earliestVideo
					prefix = newPrefix
				}

				if compareVideos(video, earliestVideoForPrefix[prefix]) > 0 {
					earliestVideoForPrefix[prefix] = video
				}
			}
		}
		if !matched {
			earliestVideoForPrefix[video.Snippet.Title] = video
		}
	}

	for title, video := range earliestVideoForPrefix {
		published, _ := time.Parse(time.RFC3339, video.Snippet.PublishedAt)
		games = append(games, &model.Game{
			Name: strings.Trim(title, "-"),
			Content: []*model.Content{
				&model.Content{
					Link:   fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.Snippet.ResourceId.VideoId),
					Start:  published,
					Source: model.SourceYouTube,
				},
			},
		})
	}
	slices.SortFunc(games, func(a, b *model.Game) int {
		return strings.Compare(a.Name, b.Name)
	})

	return games, nil
}

var cleanupVideoTitleREs = []*regexp.Regexp{
	regexp.MustCompile(`\s*Let's Play\s*`),
	regexp.MustCompile(`\s*#\d+\s*`),
	regexp.MustCompile(`\s*\[[^\[]*\]\s*`),
}

func cleanupVideoMeta(videos []*youtube.PlaylistItem) {
	for _, video := range videos {
		for _, regex := range cleanupVideoTitleREs {
			for _, match := range regex.FindAllString(video.Snippet.Title, -1) {
				video.Snippet.Title = strings.Replace(video.Snippet.Title, match, "", 1)
			}
		}
	}
}

func compareVideos(a, b *youtube.PlaylistItem) int {
	dateA, errA := time.Parse(time.RFC3339, a.Snippet.PublishedAt)
	dateB, errB := time.Parse(time.RFC3339, b.Snippet.PublishedAt)
	if errA != nil && errB != nil {
		return 0
	} else if errA != nil {
		return -1
	} else if errB != nil {
		return +1
	}

	if dateA.Before(dateB) {
		return -1
	} else if dateB.Before(dateA) {
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
