package youtube

import (
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/bauersimon/grnkdb/model"
	"github.com/bauersimon/grnkdb/steam"
	"github.com/forPelevin/gomoji"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var commonWords = map[string]bool{}

func init() {
	words := "alles,der,die,das,ein,the,gronkh"
	for _, word := range strings.Split(words, ",") {
		commonWords[word] = true
	}
}

var steamStoreLinkRE = regexp.MustCompile(`steampowered\.com\/app\/(\d+)`)

// convertVideosToGames converts model.Video structs to games
func convertVideosToGames(logger *slog.Logger, steamClient *steam.Client, videos []*model.Video) (games []*model.Game, err error) {
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
		}
	}

	logger.Debug("cleaning up video meta")
	cleanupVideoMeta(cleanedVideos)

	earliestVideoForGame := map[string]*model.Video{}
	for i, video := range cleanedVideos {
		logger.Debug("extracting game information", "video", video.VideoID, "progress", i+1, "total", len(cleanedVideos))
		var newGameSpecifier string

		// Try to extract Steam links from description.
		if matches := steamStoreLinkRE.FindStringSubmatch(video.Description); len(matches) > 0 {
			name, err := steamClient.GameName(matches[1])
			if err != nil {
				logger.Error("cannot get name from steam", "video", video.VideoID, "error", err.Error())
			} else {
				newGameSpecifier = strings.ToLower(name)
				logger.Debug("found game information on steam", "video", video.VideoID, "game", name)
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
			logger.Debug("match", "video", video.Title)
		} else {
			earliestVideoForGame[video.Title] = video
			logger.Debug("no match", "video", video.Title)
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

type cleaner struct {
	// match holds a regular expression to match.
	// Will match the whole string if "nil".
	match *regexp.Regexp
	// replace holds a replace function.
	// Matches will be replaced with "" if "nil".
	replace func(in string, match [][]string) string
}

func (c cleaner) process(s string) string {
	if c.match == nil && c.replace == nil {
		return ""
	}

	var match [][]string
	if c.match != nil {
		match = c.match.FindAllStringSubmatch(s, -1)
		if len(match) == 0 {
			return s
		}
	}

	if c.replace == nil {
		for _, m := range match {
			s = strings.ReplaceAll(s, m[0], "")
		}
		return s
	}

	return c.replace(s, match)
}

var cleanups = []*cleaner{
	// Common tags.
	{regexp.MustCompile(`(?i)Let's (Play|Test)`), nil},
	{regexp.MustCompile(`(?i)\(?Ende\)?`), nil},
	{regexp.MustCompile(`(?i)\(?Demo\)?`), nil},
	{regexp.MustCompile(`(?i)\(?Angespielt\)?`), nil},
	{regexp.MustCompile(`(?i)\(?Preview\)?`), nil},
	{regexp.MustCompile(`(?i)\(LPT[^\)]*\)`), nil},
	{regexp.MustCompile(`M\.?e\.?t\.?t\.?`), nil},
	// Episode numbers.
	{regexp.MustCompile(`#\d+`), nil},
	{regexp.MustCompile(`\D\d\d\d\:`), nil},
	{regexp.MustCompile(`\d+/\d+`), nil},
	{regexp.MustCompile(`Folge\s+\d+`), nil},
	{regexp.MustCompile(`S\d+E\d+`), nil},
	// Anything in squared brackets.
	{regexp.MustCompile(`\[[^\[]*\]`), nil},
	// All non-character or non-digit characters.
	{regexp.MustCompile(`[^\p{L}\p{N}\s\:]+`), nil},
	{
		replace: func(in string, match [][]string) string {
			return gomoji.ReplaceEmojisWith(in, ' ')
		},
	},
	{
		replace: func(in string, match [][]string) string {
			return whitespaceRE.ReplaceAllString(in, " ")
		},
	},
}

var whitespaceRE = regexp.MustCompile(`\s+`)

func cleanupVideoMeta(videos []*model.Video) {
	for _, video := range videos {
		for _, c := range cleanups {
			video.Title = c.process(video.Title)
		}
	}
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
