package model

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSVWrite(t *testing.T) {
	type testCase struct {
		Name string

		KnownSourceTypes []SourceType
		Games            []*Game

		Expected []string
		Error    string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			var writer bytes.Buffer
			err := internalCSVWrite(&writer, tc.Games, tc.KnownSourceTypes)
			if tc.Error != "" {
				assert.ErrorContains(t, err, tc.Error)
			} else {
				require.NoError(t, err)
				assert.Equal(t, strings.Join(tc.Expected, "\n")+"\n", writer.String())
			}
		})
	}

	validate(t, &testCase{
		Name: "Single Game",

		KnownSourceTypes: []SourceType{SourceType("youtube")},
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

		Expected: []string{
			"titel,youtube-link,youtube-start",
			"Minecraft,some link,26.07.2025",
		},
	})

	validate(t, &testCase{
		Name: "Multiple Games",

		KnownSourceTypes: []SourceType{SourceType("youtube")},
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

		Expected: []string{
			"titel,youtube-link,youtube-start",
			"Adrift,other link,26.08.2025",
			"Minecraft,some link,26.07.2025",
		},
	})

	validate(t, &testCase{
		Name: "Unknown Source Type",

		KnownSourceTypes: []SourceType{SourceType("youtube")},
		Games: []*Game{
			&Game{
				Name: "Minecraft",
				Content: []*Content{
					&Content{
						Source: SourceType("Twitch"),
						Link:   "some link",
						Start:  time.Date(2025, 7, 26, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},

		Error: "unknown source type",
	})
}

func TestCSVRead(t *testing.T) {
	type testCase struct {
		Name string

		KnownSourceTypes []SourceType
		CSV              []string

		Expected []*Game
		Error    string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := internalCSVRead(strings.NewReader(strings.Join(tc.CSV, "\n")), tc.KnownSourceTypes)
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

		KnownSourceTypes: []SourceType{SourceType("youtube")},
		CSV: []string{
			"titel,youtube-link,youtube-start",
			"Minecraft,some link,26.07.2025",
		},

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

		KnownSourceTypes: []SourceType{SourceType("youtube")},
		CSV: []string{
			"titel,youtube-link,youtube-start",
			"Adrift,other link,26.08.2025",
			"Minecraft,some link,26.07.2025",
		},

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
		Name: "Unknown Source Type",

		KnownSourceTypes: []SourceType{SourceType("youtube")},
		CSV: []string{
			"titel,twitch-link,twitch-start",
			"Minecraft,some link,26.07.2025",
		},

		Error: "unknown source type",
	})

	validate(t, &testCase{
		Name: "Multiple Source Types",

		KnownSourceTypes: []SourceType{SourceType("youtube"), SourceType("twitch")},
		CSV: []string{
			"titel,twitch-link,twitch-start,youtube-link,youtube-start",
			"Minecraft,some link,26.07.2025,other link,28.07.2025",
		},

		Expected: []*Game{
			&Game{
				Name: "Minecraft",
				Content: []*Content{
					&Content{
						Source: SourceType("twitch"),
						Link:   "some link",
						Start:  time.Date(2025, 7, 26, 0, 0, 0, 0, time.UTC),
					},
					&Content{
						Source: SourceType("youtube"),
						Link:   "other link",
						Start:  time.Date(2025, 7, 28, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	})

	validate(t, &testCase{
		Name: "Not all Source Types",

		KnownSourceTypes: []SourceType{SourceType("youtube"), SourceType("twitch")},
		CSV: []string{
			"titel,twitch-link,twitch-start",
			"Minecraft,some link,26.07.2025",
		},

		Expected: []*Game{
			&Game{
				Name: "Minecraft",
				Content: []*Content{
					&Content{
						Source: SourceType("twitch"),
						Link:   "some link",
						Start:  time.Date(2025, 7, 26, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	})
}
