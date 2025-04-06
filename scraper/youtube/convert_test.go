package youtube

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/bauersimon/grnkdb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/youtube/v3"
)

func TestConvertVideosToGames(t *testing.T) {
	type testCase struct {
		Name string

		Videos []*youtube.PlaylistItem

		Expected []*model.Game
		Error    string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			var logContent strings.Builder
			logger := slog.New(slog.NewTextHandler(&logContent, &slog.HandlerOptions{Level: slog.LevelDebug}))
			defer func() {
				if t.Failed() {
					t.Logf("logs:\n%s", logContent.String())
				}
			}()

			actual, err := convertVideosToGames(logger, tc.Videos)
			if tc.Error != "" {
				assert.ErrorContains(t, err, tc.Error)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.Expected, actual)
			}
		})
	}

	validate(t, &testCase{
		Name: "Single Let's Play",

		Videos: []*youtube.PlaylistItem{
			&youtube.PlaylistItem{
				Snippet: &youtube.PlaylistItemSnippet{
					Title:       "Let's Play Minecraft #001 [Deutsch] [HD] - Alles auf Anfang",
					PublishedAt: "2020-10-19T19:00:17Z",
					ResourceId: &youtube.ResourceId{
						VideoId: "DM52HxaLK-Y",
					},
				},
			},
			&youtube.PlaylistItem{
				Snippet: &youtube.PlaylistItemSnippet{
					Title:       "Let's Play Minecraft #002 [Deutsch] [HD] - Inselkoller & Nachtwache",
					PublishedAt: "2010-10-20T19:00:17Z",
					ResourceId: &youtube.ResourceId{
						VideoId: "tAaCTvht5Co",
					},
				},
			},
			&youtube.PlaylistItem{
				Snippet: &youtube.PlaylistItemSnippet{
					Title:       "Let's Play Minecraft #003 [Deutsch] [HD] - Majestätische Landschaften",
					PublishedAt: "2010-10-21T19:00:17Z",
					ResourceId: &youtube.ResourceId{
						VideoId: "ednqMErMOsM",
					},
				},
			},
		},

		Expected: []*model.Game{
			&model.Game{
				Name: "Minecraft",
				Content: []*model.Content{
					&model.Content{
						Source: model.SourceYouTube,
						Start: func() time.Time {
							p, _ := time.Parse(time.RFC3339, "2020-10-19T19:00:17Z")
							return p
						}(),
						Link: "https://www.youtube.com/watch?v=DM52HxaLK-Y",
					},
				},
			},
		},
	})
}

func TestLongestCommonPrefix(t *testing.T) {
	tests := []struct {
		name     string
		str1     string
		str2     string
		expected string
	}{
		{
			name:     "Basic example",
			str1:     "golang",
			str2:     "golem",
			expected: "go",
		},
		{
			name:     "No common prefix",
			str1:     "apple",
			str2:     "banana",
			expected: "",
		},
		{
			name:     "One string is prefix of another",
			str1:     "go",
			str2:     "golang",
			expected: "go",
		},
		{
			name:     "Empty string input",
			str1:     "",
			str2:     "golang",
			expected: "",
		},
		{
			name:     "Both empty strings",
			str1:     "",
			str2:     "",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := longestCommonPrefix(test.str1, test.str2)
			assert.Equal(t, test.expected, actual)
		})
	}
}
