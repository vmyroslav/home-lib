package homestorage

import (
	"fmt"
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

// TestWeightedRandomSelector_ConcurrentAccess tests potential race conditions in WeightedRandomSelector
func TestWeightedRandomSelector_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	selector := NewWeightedRandomSelector[string]()

	selector.Add("item1", 10)
	selector.Add("item2", 20)
	selector.Add("item3", 30)

	const (
		numGoroutines          = 100
		operationsPerGoroutine = 100
	)

	done := make(chan bool, numGoroutines*2)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < operationsPerGoroutine; j++ {
				selector.Add(fmt.Sprintf("concurrent_%d_%d", id, j), uint16(j%100))
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		go func(_ int) {
			defer func() { done <- true }()

			for j := 0; j < operationsPerGoroutine; j++ {
				// race condition: reading while others are modifying
				_, _ = selector.Get()
			}
		}(i)
	}

	for i := 0; i < numGoroutines*2; i++ {
		<-done
	}

	item, ok := selector.Get()
	if !ok {
		t.Error("Expected to get an item after concurrent operations")
	}

	t.Logf("Got item after concurrent operations: %s", item)
}

// TestWeightedRandomSelector_ConcurrentAddMany demonstrates race conditions in AddMany
func TestWeightedRandomSelector_ConcurrentAddMany(t *testing.T) {
	t.Parallel()

	selector := NewWeightedRandomSelector[int]()

	const numGoroutines = 50

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			items := make([]Item[int], 10)
			for j := 0; j < 10; j++ {
				items[j] = Item[int]{
					Value:          id*10 + j,
					PriorityWeight: uint16(j + 1),
				}
			}

			selector.AddMany(items)
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	value, ok := selector.Get()
	if !ok {
		t.Error("Expected to get a value after concurrent AddMany operations")
	}

	t.Logf("Final selector state - got value: %d, items count: %d", value, len(selector.items))
}

// TestWeightedRandomSelector_DataCorruption attempts to detect data corruption from race conditions
func TestWeightedRandomSelector_DataCorruption(t *testing.T) {
	t.Parallel()

	const (
		numGoroutines     = 20
		itemsPerGoroutine = 50
	)

	// this test tries to detect data corruption that might result from race conditions
	for iteration := 0; iteration < 10; iteration++ {
		selector := NewWeightedRandomSelector[string]()

		done := make(chan bool, numGoroutines)

		// Add items concurrently
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				for j := 0; j < itemsPerGoroutine; j++ {
					selector.Add(fmt.Sprintf("item_%d_%d", id, j), uint16(j+1))
				}
			}(i)
		}

		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// check for data corruption - prioritySum should match sum of all item weights
		expectedSum := uint32(0)
		actualSum := selector.prioritySum

		for _, item := range selector.items {
			expectedSum += uint32(item.PriorityWeight)
		}

		if actualSum != expectedSum {
			t.Errorf("Iteration %d: Data corruption detected - prioritySum mismatch. Expected: %d, Actual: %d, Items: %d",
				iteration, expectedSum, actualSum, len(selector.items))
		}

		// verify expected number of items
		expectedItems := numGoroutines * itemsPerGoroutine
		if len(selector.items) != expectedItems {
			t.Errorf("Iteration %d: Expected %d items, got %d", iteration, expectedItems, len(selector.items))
		}
	}
}
