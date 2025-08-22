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