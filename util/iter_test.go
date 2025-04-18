package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWindowed(t *testing.T) {
	type testCase struct {
		Name string

		Slice      []int
		WindowSize uint

		Expected [][]int
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			var actual [][]int
			for window := range Windowed(tc.Slice, tc.WindowSize) {
				actual = append(actual, window)
			}

			assert.Equal(t, tc.Expected, actual)
		})
	}

	validate(t, &testCase{
		Name: "1",

		Slice:      []int{1, 2, 3},
		WindowSize: 1,

		Expected: [][]int{
			{1},
			{2},
			{3},
		},
	})
	validate(t, &testCase{
		Name: "2",

		Slice:      []int{1, 2, 3},
		WindowSize: 2,

		Expected: [][]int{
			{1, 2},
			{3},
		},
	})
	validate(t, &testCase{
		Name: "3",

		Slice:      []int{1, 2, 3},
		WindowSize: 3,

		Expected: [][]int{
			{1, 2, 3},
		},
	})
}
