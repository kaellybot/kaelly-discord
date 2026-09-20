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

func TestSlice_NonPositiveLimitPanics(t *testing.T) {
	// Documents current behavior: there is no guard, a zero limit divides by zero.
	assert.Panics(t, func() { Slice([]int{1, 2}, 0) })
}
