package homestorage

import (
	"math"
	"reflect"
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

func TestAddTopPrioElement(t *testing.T) {
	t.Parallel()

	selector := NewWeightedRandomSelector[string]()
	selector.AddTopPrioElement("Apple")

	want := []Item[string]{{"Apple", math.MaxUint16}}

	if !reflect.DeepEqual(selector.items, want) {
		t.Errorf("Test AddTopPrioElement failed: expected %v, got %v", want, selector.items)
	}
}

func TestAddOrdered(t *testing.T) {
	t.Parallel()

	selector := NewWeightedRandomSelector[string]()
	selector.AddOrdered([]string{"Apple", "Banana"})

	want := []Item[string]{{"Apple", 0}, {"Banana", 1}}

	if !reflect.DeepEqual(selector.items, want) {
		t.Errorf("Test AddOrdered failed: expected %v, got %v", want, selector.items)
	}
}

func TestAddMany(t *testing.T) {
	t.Parallel()

	selector := NewWeightedRandomSelector[string]()
	selector.AddMany([]Item[string]{{"Apple", 1}, {"Banana", 2}})

	want := []Item[string]{{"Apple", 1}, {"Banana", 2}}

	if !reflect.DeepEqual(selector.items, want) {
		t.Errorf("Test AddMany failed: expected %v, got %v", want, selector.items)
	}
}

func TestAddItem(t *testing.T) {
	t.Parallel()

	selector := NewWeightedRandomSelector[string]()
	selector.AddItem(Item[string]{"Apple", 1})

	want := []Item[string]{{"Apple", 1}}

	if !reflect.DeepEqual(selector.items, want) {
		t.Errorf("Test AddItem failed: expected %v, got %v", want, selector.items)
	}
}
