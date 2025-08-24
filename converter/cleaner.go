package converter

import (
	"regexp"
	"strings"

	"github.com/bauersimon/grnkdb/model"
	"github.com/forPelevin/gomoji"
)

var commonWords = map[string]bool{}

func init() {
	words := "alles,der,die,das,ein,the,gronkh"
	for _, word := range strings.Split(words, ",") {
		commonWords[word] = true
	}
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
