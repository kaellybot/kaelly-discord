package slicers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlice_ChunksInOrderWithShorterTail(t *testing.T) {
	cases := []struct {
		name     string
		items    []int
		limit    int
		expected [][]int
	}{
		{name: "nil input", items: nil, limit: 3, expected: [][]int{}},
		{name: "empty input", items: []int{}, limit: 3, expected: [][]int{}},
		{name: "fewer than limit", items: []int{1, 2}, limit: 3, expected: [][]int{{1, 2}}},
		{name: "exactly limit", items: []int{1, 2, 3}, limit: 3, expected: [][]int{{1, 2, 3}}},
		{name: "exact multiple", items: []int{1, 2, 3, 4, 5, 6}, limit: 3, expected: [][]int{{1, 2, 3}, {4, 5, 6}}},
		{name: "remainder in last chunk", items: []int{1, 2, 3, 4, 5, 6, 7}, limit: 3,
			expected: [][]int{{1, 2, 3}, {4, 5, 6}, {7}}},
		{name: "limit of one", items: []int{1, 2, 3}, limit: 1, expected: [][]int{{1}, {2}, {3}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, Slice(tc.items, tc.limit))
		})
	}
}

func TestSlice_ChunksDoNotAliasEachOther(t *testing.T) {
	sliced := Slice([]string{"a", "b", "c", "d"}, 2)

	sliced[0] = append(sliced[0], "x")

	assert.Equal(t, []string{"c", "d"}, sliced[1], "growing one chunk must not overwrite the next one")
}

func TestSlice_NonPositiveLimitKeepsEverythingInOneChunk(t *testing.T) {
	// A non-positive limit used to divide by zero on i%limit. Chosen no-op:
	// skip the chunking rather than drop items, so the returned chunks still
	// concatenate back to the input.
	assert.Equal(t, [][]int{{1, 2, 3}}, Slice([]int{1, 2, 3}, 0))
	assert.Equal(t, [][]int{{1, 2, 3}}, Slice([]int{1, 2, 3}, -3))
	assert.Equal(t, [][]int{}, Slice([]int{}, 0))
	assert.Equal(t, [][]int{}, Slice[int](nil, -1))
}

func TestSlice_NonPositiveLimitDoesNotAliasTheInput(t *testing.T) {
	items := []int{1, 2, 3}

	sliced := Slice(items, 0)
	sliced[0][0] = 99

	assert.Equal(t, []int{1, 2, 3}, items, "the single chunk must be a copy")
}
