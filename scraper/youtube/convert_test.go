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
		Name: "Prefix",

		Videos: []*youtube.PlaylistItem{
			&youtube.PlaylistItem{
				Snippet: &youtube.PlaylistItemSnippet{
					Title:       "Let's Play Minecraft #001 [Deutsch] [HD] - Alles auf Anfang",
					PublishedAt: "2010-10-19T19:00:17Z",
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
							p, _ := time.Parse(time.RFC3339, "2010-10-19T19:00:17Z")
							return p
						}(),
						Link: "https://www.youtube.com/watch?v=DM52HxaLK-Y",
					},
				},
			},
		},
	})

	validate(t, &testCase{
		Name: "Suffix",

		Videos: []*youtube.PlaylistItem{
			&youtube.PlaylistItem{
				Snippet: &youtube.PlaylistItemSnippet{
					Title:       "Der Mann mit dem Hut ist wieder da! 🛕 INDIANA JONES AND THE GREAT CIRCLE #01",
					PublishedAt: "2025-12-13T19:00:17Z",
					ResourceId: &youtube.ResourceId{
						VideoId: "XONCCUxHGxo",
					},
				},
			},
			&youtube.PlaylistItem{
				Snippet: &youtube.PlaylistItemSnippet{
					Title:       "Schwarze Hemden, niedrige Lebenserwartung 🛕 INDIANA JONES AND THE GREAT CIRCLE #02",
					PublishedAt: "2025-12-14T19:00:17Z",
					ResourceId: &youtube.ResourceId{
						VideoId: "wVsDQx0SY1M",
					},
				},
			},
			&youtube.PlaylistItem{
				Snippet: &youtube.PlaylistItemSnippet{
					Title:       "Indiana Jones und das Geheimnis der Schwerkraft 🛕 INDIANA JONES AND THE GREAT CIRCLE #03",
					PublishedAt: "2025-12-15T19:00:17Z",
					ResourceId: &youtube.ResourceId{
						VideoId: "9Ack9uoQRIM",
					},
				},
			},
		},

		Expected: []*model.Game{
			&model.Game{
				Name: "Indiana Jones And The Great Circle",
				Content: []*model.Content{
					&model.Content{
						Source: model.SourceYouTube,
						Start: func() time.Time {
							p, _ := time.Parse(time.RFC3339, "2025-12-13T19:00:17Z")
							return p
						}(),
						Link: "https://www.youtube.com/watch?v=XONCCUxHGxo",
					},
				},
			},
		},
	})

	validate(t, &testCase{
		Name: "Steam",

		Videos: []*youtube.PlaylistItem{
			&youtube.PlaylistItem{
				Snippet: &youtube.PlaylistItemSnippet{
					Title:       "Der Mann mit dem Hut ist wieder da! 🛕 INDIANA JONES AND THE GREAT CIRCLE #01",
					PublishedAt: "2025-12-13T19:00:17Z",
					ResourceId: &youtube.ResourceId{
						VideoId: "XONCCUxHGxo",
					},
					Description: "https://store.steampowered.com/app/2677660",
				},
			},
		},
		Expected: []*model.Game{
			&model.Game{
				Name: "Indiana Jones And The Great Circle",
				Content: []*model.Content{
					&model.Content{
						Source: model.SourceYouTube,
						Start: func() time.Time {
							p, _ := time.Parse(time.RFC3339, "2025-12-13T19:00:17Z")
							return p
						}(),
						Link: "https://www.youtube.com/watch?v=XONCCUxHGxo",
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

func TestReverseString(t *testing.T) {
	assert.Equal(t, "aaa", reverseString("aaa"))
	assert.Equal(t, "", reverseString(""))
	assert.Equal(t, "abc", reverseString("cba"))
	assert.Equal(t, "abcd", reverseString("dcba"))
}
