package util

import "iter"

// Windowed walks the slice in a windowed-manner.
func Windowed[A any](slice []A, windowSize uint) iter.Seq[[]A] {
	return SlidingWindowed(slice, windowSize, windowSize)
}

// SlidingWindowed walks the slice in a sliding-windowed-manner.
func SlidingWindowed[A any](slice []A, windowSize uint, sliding uint) iter.Seq[[]A] {
	return func(yield func([]A) bool) {
		for i := 0; i < len(slice); i = i + int(sliding) {
			upper := min(i+int(windowSize), len(slice))
			if !yield(slice[i:upper]) {
				return
			}
		}
	}
}
