package converter

import (
	"testing"
	"time"

	"github.com/bauersimon/grnkdb/model"
	"github.com/bauersimon/grnkdb/steam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestConvertVideosToGames(t *testing.T) {
	type testCase struct {
		Name string

		Videos []*model.Video

		Expected []*model.Game
		Error    string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)

			converter := NewVideoToGameConverter(steam.NewClient(), 100, logger)
			actual, err := converter.Convert(tc.Videos)
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

		Videos: []*model.Video{
			{
				Title:       "Let's Play Minecraft #001 [Deutsch] [HD] - Alles auf Anfang",
				PublishedAt: time.Date(2010, 10, 19, 19, 0, 17, 0, time.UTC),
				VideoID:     "DM52HxaLK-Y",
				Link:        "https://www.youtube.com/watch?v=DM52HxaLK-Y",
				Source:      model.SourceYouTube,
			},
			{
				Title:       "Let's Play Minecraft #002 [Deutsch] [HD] - Inselkoller & Nachtwache",
				PublishedAt: time.Date(2010, 10, 20, 19, 0, 17, 0, time.UTC),
				VideoID:     "tAaCTvht5Co",
				Link:        "https://www.youtube.com/watch?v=tAaCTvht5Co",
				Source:      model.SourceYouTube,
			},
			{
				Title:       "Let's Play Minecraft #003 [Deutsch] [HD] - Majestätische Landschaften",
				PublishedAt: time.Date(2010, 10, 21, 19, 0, 17, 0, time.UTC),
				VideoID:     "ednqMErMOsM",
				Link:        "https://www.youtube.com/watch?v=ednqMErMOsM",
				Source:      model.SourceYouTube,
			},
		},

		Expected: []*model.Game{
			&model.Game{
				Name: "Minecraft",
				Content: []*model.Content{
					&model.Content{
						Source: model.SourceYouTube,
						Start:  time.Date(2010, 10, 19, 19, 0, 17, 0, time.UTC),
						Link:   "https://www.youtube.com/watch?v=DM52HxaLK-Y",
					},
				},
			},
		},
	})

	validate(t, &testCase{
		Name: "Suffix",

		Videos: []*model.Video{
			{
				Title:       "Der Mann mit dem Hut ist wieder da! 🛕 INDIANA JONES AND THE GREAT CIRCLE #01",
				PublishedAt: time.Date(2025, 12, 13, 19, 0, 17, 0, time.UTC),
				VideoID:     "XONCCUxHGxo",
				Link:        "https://www.youtube.com/watch?v=XONCCUxHGxo",
				Source:      model.SourceYouTube,
			},
			{
				Title:       "Schwarze Hemden, niedrige Lebenserwartung 🛕 INDIANA JONES AND THE GREAT CIRCLE #02",
				PublishedAt: time.Date(2025, 12, 14, 19, 0, 17, 0, time.UTC),
				VideoID:     "wVsDQx0SY1M",
				Link:        "https://www.youtube.com/watch?v=wVsDQx0SY1M",
				Source:      model.SourceYouTube,
			},
			{
				Title:       "Indiana Jones und das Geheimnis der Schwerkraft 🛕 INDIANA JONES AND THE GREAT CIRCLE #03",
				PublishedAt: time.Date(2025, 12, 15, 19, 0, 17, 0, time.UTC),
				VideoID:     "9Ack9uoQRIM",
				Link:        "https://www.youtube.com/watch?v=9Ack9uoQRIM",
				Source:      model.SourceYouTube,
			},
		},

		Expected: []*model.Game{
			&model.Game{
				Name: "Indiana Jones And The Great Circle",
				Content: []*model.Content{
					&model.Content{
						Source: model.SourceYouTube,
						Start:  time.Date(2025, 12, 13, 19, 0, 17, 0, time.UTC),
						Link:   "https://www.youtube.com/watch?v=XONCCUxHGxo",
					},
				},
			},
		},
	})

	validate(t, &testCase{
		Name: "Steam",

		Videos: []*model.Video{
			{
				Title:       "Der Mann mit dem Hut ist wieder da! 🛕 INDIANA JONES AND THE GREAT CIRCLE #01",
				PublishedAt: time.Date(2025, 12, 13, 19, 0, 17, 0, time.UTC),
				VideoID:     "XONCCUxHGxo",
				Link:        "https://www.youtube.com/watch?v=XONCCUxHGxo",
				Description: "https://store.steampowered.com/app/2677660",
				Source:      model.SourceYouTube,
			},
		},
		Expected: []*model.Game{
			&model.Game{
				Name: "Indiana Jones And The Great Circle",
				Content: []*model.Content{
					&model.Content{
						Source: model.SourceYouTube,
						Start:  time.Date(2025, 12, 13, 19, 0, 17, 0, time.UTC),
						Link:   "https://www.youtube.com/watch?v=XONCCUxHGxo",
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
