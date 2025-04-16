package youtube

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/bauersimon/grnkdb/model"
	"github.com/forPelevin/gomoji"
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/api/youtube/v3"
)

var commonWords = map[string]bool{}

func init() {
	words := "alles,der,die,das,ein,the"
	for _, word := range strings.Split(words, ",") {
		commonWords[word] = true
	}
}

var steamStoreLinkRE = regexp.MustCompile(`steampowered\.com\/app\/(\d+)`)

var steamNameCache = map[string]string{}

func getSteamGameName(appID string) (game string, err error) {
	if game = steamNameCache[appID]; game != "" {
		return game, nil
	}
	defer func() {
		if err == nil {
			steamNameCache[appID] = game
		}
	}()

	url := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%s", appID)
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.WithStack(err)
	}

	var apiResponse map[string]struct {
		Success bool `json:"success"`
		Data    struct {
			Name string `json:"name"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", errors.WithStack(err)
	}

	if appData, exists := apiResponse[appID]; exists && appData.Success {
		return appData.Data.Name, nil
	}

	return "", errors.Errorf("unknown game ID %q", appID)
}

func convertVideosToGames(logger *slog.Logger, videos []*youtube.PlaylistItem) (games []*model.Game, err error) {
	logger.Debug("cleaning up video meta")
	cleanupVideoMeta(videos)

	logger.Debug("extracting unique games")
	earliestVideoForGame := map[string]*youtube.PlaylistItem{}
	for _, video := range videos {
		var newGameSpecifier string

		// Try to extract Steam links from description.
		if matches := steamStoreLinkRE.FindStringSubmatch(video.Snippet.Description); len(matches) > 0 {
			name, err := getSteamGameName(matches[1])
			if err != nil {
				logger.Error("cannot get name from steam", "video", video.Snippet.ResourceId.VideoId, "error", err.Error())
			} else {
				newGameSpecifier = strings.ToLower(name)
			}
		}

		// Brute force try to find similar pre- or suffixes among all found games.
		var oldGameSpecifier string
		if newGameSpecifier == "" {
			for preSuffix := range maps.Keys(earliestVideoForGame) {
				newPrefix := longestCommonPrefix(
					strings.ToLower(video.Snippet.Title),
					strings.ToLower(preSuffix),
				)
				newSuffix := longestCommonSuffix(
					strings.ToLower(video.Snippet.Title),
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
			logger.Debug("match", "video", video.Snippet.Title)
		} else {
			earliestVideoForGame[video.Snippet.Title] = video
			logger.Debug("no match", "video", video.Snippet.Title)
		}
	}

	caser := cases.Title(language.German)
	for title, video := range earliestVideoForGame {
		published, _ := time.Parse(time.RFC3339, video.Snippet.PublishedAt)
		games = append(games, &model.Game{
			Name: caser.String(strings.TrimSpace(strings.Trim(title, "-:\" \t"))),
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

type cleaner struct {
	// match holds a regular expression to match.
	// Will match the whole string if "nil".
	match *regexp.Regexp
	// replace holds a replace function.
	// First match occurence will be replaced with "" if "nil".
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
	{regexp.MustCompile(`Let's Play`), nil},
	{regexp.MustCompile(`#\d+`), nil},
	{regexp.MustCompile(`\((ENDE|Ende)\)`), nil},
	{regexp.MustCompile(`\(?(DEMO|Demo)\)?`), nil},
	{regexp.MustCompile(`\((ANGESPIELT|Angespielt)\)`), nil},
	{regexp.MustCompile(`\d+/\d+`), nil},
	{regexp.MustCompile(`\[[^\[]*\]`), nil},
	{regexp.MustCompile(`\(LPT[^\)]*\)`), nil},
	{regexp.MustCompile(`Folge\s+\d+`), nil},
	{regexp.MustCompile(`S\d+E\d+`), nil},
	{regexp.MustCompile(`•`), nil},
	{regexp.MustCompile(`M\.e\.t\.t\.`), nil},
	{regexp.MustCompile(`Mett`), nil},
	{regexp.MustCompile(`"`), nil},
	{regexp.MustCompile(`Simulator`), nil},
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

func cleanupVideoMeta(videos []*youtube.PlaylistItem) {
	for _, video := range videos {
		for _, c := range cleanups {
			video.Snippet.Title = c.process(video.Snippet.Title)
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
