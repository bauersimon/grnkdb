package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMergeGames(t *testing.T) {
	type testCase struct {
		Name string

		A []*Game
		B []*Game

		Expected []*Game
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, MergeGames(tc.A, tc.B))
		})
	}

	validate(t, &testCase{
		Name: "Different",

		A: []*Game{
			&Game{
				Name: "foo",
			},
		},
		B: []*Game{
			&Game{
				Name: "bar",
			},
		},

		Expected: []*Game{
			&Game{
				Name: "bar",
			},
			&Game{
				Name: "foo",
			},
		},
	})

	t.Run("Same", func(t *testing.T) {
		validate(t, &testCase{
			Name: "First",

			A: []*Game{
				&Game{
					Name: "foo",
					Content: []*Content{
						&Content{
							Source: SourceYouTube,
							Link:   "A",
							Start:  time.Date(2020, 10, 8, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			B: []*Game{
				&Game{
					Name: "foo",
					Content: []*Content{
						&Content{
							Source: SourceYouTube,
							Link:   "B",
							Start:  time.Date(2020, 10, 9, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},

			Expected: []*Game{
				&Game{
					Name: "foo",
					Content: []*Content{
						&Content{
							Source: SourceYouTube,
							Link:   "A",
							Start:  time.Date(2020, 10, 8, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
		})
		validate(t, &testCase{
			Name: "Second",

			A: []*Game{
				&Game{
					Name: "foo",
					Content: []*Content{
						&Content{
							Source: SourceYouTube,
							Link:   "A",
							Start:  time.Date(2020, 10, 10, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			B: []*Game{
				&Game{
					Name: "foo",
					Content: []*Content{
						&Content{
							Source: SourceYouTube,
							Link:   "B",
							Start:  time.Date(2020, 10, 9, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},

			Expected: []*Game{
				&Game{
					Name: "foo",
					Content: []*Content{
						&Content{
							Source: SourceYouTube,
							Link:   "B",
							Start:  time.Date(2020, 10, 9, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
		})
	})
}
