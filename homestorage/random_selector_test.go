package homestorage

import (
	"testing"
)

func TestWeightedRandomSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		items      []Item[string]
		pickCounts int
		want       bool
	}{
		{
			name: "Single Item",
			items: []Item[string]{
				{"Apple", 1},
			},
			pickCounts: 1,
			want:       true,
		},
		{
			name: "Multiple Items",
			items: []Item[string]{
				{"Apple", 1},
				{"Banana", 2},
			},
			pickCounts: 2,
			want:       true,
		},
		{
			name: "All zero priority items",
			items: []Item[string]{
				{"Apple", 0},
				{"Banana", 0},
				{"Orange", 0},
			},
			pickCounts: 2,
			want:       true,
		},
		{
			name:       "No Items",
			items:      []Item[string]{},
			pickCounts: 1,
			want:       false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			selector := NewWeightedRandomSelector[string]()
			for _, item := range tc.items {
				selector.Add(item.Value, item.PriorityWeight)
			}

			for i := 0; i < tc.pickCounts; i++ {
				_, got := selector.Get()
				if got != tc.want {
					t.Errorf("Test %v failed: expected %v, got %v", tc.name, tc.want, got)
				}
			}
		})
	}
}
