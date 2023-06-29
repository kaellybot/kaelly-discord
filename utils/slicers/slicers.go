package slicers

func Slice[T any](items []T, limit int) [][]T {
	sliced := make([][]T, 0)

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
