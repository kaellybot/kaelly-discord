package slicers

import "slices"

func Slice[T any](items []T, limit int) [][]T {
	sliced := make([][]T, 0)

	// A non-positive limit means "do not chunk": every item is kept together
	// instead of dividing by zero on i%limit.
	if limit <= 0 {
		if len(items) > 0 {
			sliced = append(sliced, slices.Clone(items))
		}

		return sliced
	}

	var currentSlice []T
	for i, item := range items {
		if i != 0 && i%limit == 0 {
			sliced = append(sliced, currentSlice)
			currentSlice = make([]T, 0)
		}
		currentSlice = append(currentSlice, item)
	}

	if len(currentSlice) > 0 {
		sliced = append(sliced, currentSlice)
	}

	return sliced
}
