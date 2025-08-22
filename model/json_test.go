package model

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONWrite(t *testing.T) {
	type testCase struct {
		Name string

		Games []*Game

		Expected string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			var writer bytes.Buffer
			err := JSONWrite(&writer, tc.Games)
			require.NoError(t, err)
			assert.JSONEq(t, tc.Expected, writer.String())
		})
	}

	validate(t, &testCase{
		Name: "Single Game",

		Games: []*Game{
			&Game{
				Name: "Minecraft",
				Content: []*Content{
					&Content{
						Source: SourceType("youtube"),
						Link:   "some link",
						Start:  time.Date(2025, 7, 26, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},

		Expected: `[
			{
				"Name": "Minecraft",
				"Content": [
					{
						"Link": "some link",
						"Start": "2025-07-26T00:00:00Z",
						"Source": "youtube"
					}
				]
			}
		]`,
	})

	validate(t, &testCase{
		Name: "Multiple Games",

		Games: []*Game{
			&Game{
				Name: "Minecraft",
				Content: []*Content{
					&Content{
						Source: SourceType("youtube"),
						Link:   "some link",
						Start:  time.Date(2025, 7, 26, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			&Game{
				Name: "Adrift",
				Content: []*Content{
					&Content{
						Source: SourceType("youtube"),
						Link:   "other link",
						Start:  time.Date(2025, 8, 26, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},

		Expected: `[
			{
				"Name": "Adrift",
				"Content": [
					{
						"Link": "other link",
						"Start": "2025-08-26T00:00:00Z",
						"Source": "youtube"
					}
				]
			},
			{
				"Name": "Minecraft",
				"Content": [
					{
						"Link": "some link",
						"Start": "2025-07-26T00:00:00Z",
						"Source": "youtube"
					}
				]
			}
		]`,
	})

	validate(t, &testCase{
		Name: "Multiple Source Types",

		Games: []*Game{
			&Game{
				Name: "Minecraft",
				Content: []*Content{
					&Content{
						Source: SourceType("youtube"),
						Link:   "youtube link",
						Start:  time.Date(2025, 7, 28, 0, 0, 0, 0, time.UTC),
					},
					&Content{
						Source: SourceType("twitch"),
						Link:   "twitch link",
						Start:  time.Date(2025, 7, 26, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},

		Expected: `[
			{
				"Name": "Minecraft",
				"Content": [
					{
						"Link": "twitch link",
						"Start": "2025-07-26T00:00:00Z",
						"Source": "twitch"
					},
					{
						"Link": "youtube link",
						"Start": "2025-07-28T00:00:00Z",
						"Source": "youtube"
					}
				]
			}
		]`,
	})
}

func TestJSONRead(t *testing.T) {
	type testCase struct {
		Name string

		JSON string

		Expected []*Game
		Error    string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := JSONRead(strings.NewReader(tc.JSON))
			if tc.Error != "" {
				assert.ErrorContains(t, err, tc.Error)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.Expected, actual)
			}
		})
	}

	validate(t, &testCase{
		Name: "Single Game",

		JSON: `[
			{
				"Name": "Minecraft",
				"Content": [
					{
						"Link": "some link",
						"Start": "2025-07-26T00:00:00Z",
						"Source": "youtube"
					}
				]
			}
		]`,

		Expected: []*Game{
			&Game{
				Name: "Minecraft",
				Content: []*Content{
					&Content{
						Source: SourceType("youtube"),
						Link:   "some link",
						Start:  time.Date(2025, 7, 26, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	})

	validate(t, &testCase{
		Name: "Multiple Games",

		JSON: `[
			{
				"Name": "Adrift",
				"Content": [
					{
						"Link": "other link",
						"Start": "2025-08-26T00:00:00Z",
						"Source": "youtube"
					}
				]
			},
			{
				"Name": "Minecraft",
				"Content": [
					{
						"Link": "some link",
						"Start": "2025-07-26T00:00:00Z",
						"Source": "youtube"
					}
				]
			}
		]`,

		Expected: []*Game{
			&Game{
				Name: "Adrift",
				Content: []*Content{
					&Content{
						Source: SourceType("youtube"),
						Link:   "other link",
						Start:  time.Date(2025, 8, 26, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			&Game{
				Name: "Minecraft",
				Content: []*Content{
					&Content{
						Source: SourceType("youtube"),
						Link:   "some link",
						Start:  time.Date(2025, 7, 26, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	})

	validate(t, &testCase{
		Name: "Multiple Source Types",

		JSON: `[
			{
				"Name": "Minecraft",
				"Content": [
					{
						"Link": "twitch link",
						"Start": "2025-07-26T00:00:00Z",
						"Source": "twitch"
					},
					{
						"Link": "youtube link",
						"Start": "2025-07-28T00:00:00Z",
						"Source": "youtube"
					}
				]
			}
		]`,

		Expected: []*Game{
			&Game{
				Name: "Minecraft",
				Content: []*Content{
					&Content{
						Source: SourceType("twitch"),
						Link:   "twitch link",
						Start:  time.Date(2025, 7, 26, 0, 0, 0, 0, time.UTC),
					},
					&Content{
						Source: SourceType("youtube"),
						Link:   "youtube link",
						Start:  time.Date(2025, 7, 28, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	})

	validate(t, &testCase{
		Name: "Invalid JSON",

		JSON: `[invalid json}`,

		Error: "invalid character",
	})

	validate(t, &testCase{
		Name: "Empty Array",

		JSON: `[]`,

		Expected: []*Game{},
	})
}

func TestVideoCSVWrite(t *testing.T) {
	type testCase struct {
		Name string

		Videos []*Video

		Expected []string
		Error    string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			var writer bytes.Buffer
			err := VideoCSVWrite(&writer, tc.Videos)
			if tc.Error != "" {
				assert.ErrorContains(t, err, tc.Error)
			} else {
				require.NoError(t, err)
				assert.Equal(t, strings.Join(tc.Expected, "\n")+"\n", writer.String())
			}
		})
	}

	validate(t, &testCase{
		Name: "Single Video",

		Videos: []*Video{
			{
				VideoID:     "dQw4w9WgXcQ",
				Title:       "Never Gonna Give You Up",
				Description: "Rick Astley's official music video",
				Link:        "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
				PublishedAt: time.Date(2009, 10, 25, 9, 57, 33, 0, time.UTC),
				ChannelID:   "UCuAXFkgsw1L7xaCfnd5JJOw",
			},
		},

		Expected: []string{
			"Link,PublishedAt,Title,Description,ChannelID,VideoID",
			"https://www.youtube.com/watch?v=dQw4w9WgXcQ,2009-10-25T09:57:33Z,Never Gonna Give You Up,Rick Astley's official music video,UCuAXFkgsw1L7xaCfnd5JJOw,dQw4w9WgXcQ",
		},
	})

	validate(t, &testCase{
		Name: "Multiple Videos",

		Videos: []*Video{
			{
				VideoID:     "dQw4w9WgXcQ",
				Title:       "Never Gonna Give You Up",
				Description: "Rick Astley's official music video",
				Link:        "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
				PublishedAt: time.Date(2009, 10, 25, 9, 57, 33, 0, time.UTC),
				ChannelID:   "UCuAXFkgsw1L7xaCfnd5JJOw",
			},
			{
				VideoID:     "abc123",
				Title:       "Test Video",
				Description: "A test video description",
				Link:        "https://www.youtube.com/watch?v=abc123",
				PublishedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				ChannelID:   "UCtest123",
			},
		},

		Expected: []string{
			"Link,PublishedAt,Title,Description,ChannelID,VideoID",
			"https://www.youtube.com/watch?v=abc123,2023-01-01T12:00:00Z,Test Video,A test video description,UCtest123,abc123",
			"https://www.youtube.com/watch?v=dQw4w9WgXcQ,2009-10-25T09:57:33Z,Never Gonna Give You Up,Rick Astley's official music video,UCuAXFkgsw1L7xaCfnd5JJOw,dQw4w9WgXcQ",
		},
	})

	validate(t, &testCase{
		Name:   "Empty Videos",
		Videos: []*Video{},
		Expected: []string{
			"Link,PublishedAt,Title,Description,ChannelID,VideoID",
		},
	})

	validate(t, &testCase{
		Name: "Video with Empty Fields",

		Videos: []*Video{
			{
				VideoID:     "test123",
				Title:       "",
				Description: "",
				Link:        "https://www.youtube.com/watch?v=test123",
				PublishedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				ChannelID:   "",
			},
		},

		Expected: []string{
			"Link,PublishedAt,Title,Description,ChannelID,VideoID",
			"https://www.youtube.com/watch?v=test123,2023-01-01T00:00:00Z,,,,test123",
		},
	})
}

func TestVideoCSVRead(t *testing.T) {
	type testCase struct {
		Name string

		CSV []string

		Expected []*Video
		Error    string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := VideoCSVRead(strings.NewReader(strings.Join(tc.CSV, "\n")))
			if tc.Error != "" {
				assert.ErrorContains(t, err, tc.Error)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.Expected, actual)
			}
		})
	}

	validate(t, &testCase{
		Name: "Single Video",

		CSV: []string{
			"Link,PublishedAt,Title,Description,ChannelID,VideoID",
			"https://www.youtube.com/watch?v=dQw4w9WgXcQ,2009-10-25T09:57:33Z,Never Gonna Give You Up,Rick Astley's official music video,UCuAXFkgsw1L7xaCfnd5JJOw,dQw4w9WgXcQ",
		},

		Expected: []*Video{
			{
				VideoID:     "dQw4w9WgXcQ",
				Title:       "Never Gonna Give You Up",
				Description: "Rick Astley's official music video",
				Link:        "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
				PublishedAt: time.Date(2009, 10, 25, 9, 57, 33, 0, time.UTC),
				ChannelID:   "UCuAXFkgsw1L7xaCfnd5JJOw",
			},
		},
	})

	validate(t, &testCase{
		Name: "Multiple Videos",

		CSV: []string{
			"Link,PublishedAt,Title,Description,ChannelID,VideoID",
			"https://www.youtube.com/watch?v=abc123,2023-01-01T12:00:00Z,Test Video,A test video description,UCtest123,abc123",
			"https://www.youtube.com/watch?v=dQw4w9WgXcQ,2009-10-25T09:57:33Z,Never Gonna Give You Up,Rick Astley's official music video,UCuAXFkgsw1L7xaCfnd5JJOw,dQw4w9WgXcQ",
		},

		Expected: []*Video{
			{
				VideoID:     "abc123",
				Title:       "Test Video",
				Description: "A test video description",
				Link:        "https://www.youtube.com/watch?v=abc123",
				PublishedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				ChannelID:   "UCtest123",
			},
			{
				VideoID:     "dQw4w9WgXcQ",
				Title:       "Never Gonna Give You Up",
				Description: "Rick Astley's official music video",
				Link:        "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
				PublishedAt: time.Date(2009, 10, 25, 9, 57, 33, 0, time.UTC),
				ChannelID:   "UCuAXFkgsw1L7xaCfnd5JJOw",
			},
		},
	})

	validate(t, &testCase{
		Name: "Header Only",
		CSV: []string{
			"Link,PublishedAt,Title,Description,ChannelID,VideoID",
		},
		Expected: []*Video{},
	})

	validate(t, &testCase{
		Name: "Video with Empty Fields",

		CSV: []string{
			"Link,PublishedAt,Title,Description,ChannelID,VideoID",
			"https://www.youtube.com/watch?v=test123,2023-01-01T00:00:00Z,,,,test123",
		},

		Expected: []*Video{
			{
				VideoID:     "test123",
				Title:       "",
				Description: "",
				Link:        "https://www.youtube.com/watch?v=test123",
				PublishedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				ChannelID:   "",
			},
		},
	})
}
