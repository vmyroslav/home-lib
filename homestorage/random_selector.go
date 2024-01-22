package homestorage

import (
	"math"
	"math/rand"
	"time"
)

type Item[T any] struct {
	Value          T
	PriorityWeight uint16
}

type WeightedRandomSelector[T any] struct {
	items       []Item[T]
	prioritySum uint32
}

// NewWeightedRandomSelector creates a new instance of WeightedRandomSelector for a specific type.
func NewWeightedRandomSelector[T any]() *WeightedRandomSelector[T] {
	return &WeightedRandomSelector[T]{}
}

// AddItem adds a new item to the selector.
func (wrs *WeightedRandomSelector[T]) AddItem(item Item[T]) {
	wrs.items = append(wrs.items, item)
	wrs.prioritySum += uint32(item.PriorityWeight)
}

// Add adds a new item to the selector with a specific priority.
func (wrs *WeightedRandomSelector[T]) Add(value T, priority uint16) {
	wrs.items = append(wrs.items, Item[T]{Value: value, PriorityWeight: priority})
	wrs.prioritySum += uint32(priority)
}

// AddMany adds multiple items to the selector.
func (wrs *WeightedRandomSelector[T]) AddMany(items []Item[T]) {
	for _, item := range items {
		wrs.items = append(wrs.items, item)
		wrs.prioritySum += uint32(item.PriorityWeight)
	}
}

// AddOrdered adds multiple items to the selector with their priorities based on their order.
func (wrs *WeightedRandomSelector[T]) AddOrdered(values []T) {
	for i, value := range values {
		wrs.items = append(wrs.items, Item[T]{Value: value, PriorityWeight: uint16(i)})
		wrs.prioritySum += uint32(i)
	}
}

// AddTopPrioElement adds a new item to the selector with the highest (math.MaxUint16) priority.
func (wrs *WeightedRandomSelector[T]) AddTopPrioElement(value T) {
	highestPriority := uint16(math.MaxUint16)
	wrs.AddItem(Item[T]{Value: value, PriorityWeight: highestPriority})
}

// Get picks an item randomly, considering the item's priority as its weight.
func (wrs *WeightedRandomSelector[T]) Get() (T, bool) {
	var zero T

	if len(wrs.items) == 0 {
		return zero, false
	}

	if wrs.prioritySum == 0 {
		// If total sum of priorities is 0, select an item randomly without considering the priorities
		return wrs.items[rand.Intn(len(wrs.items))].Value, true //nolint:gosec
	}

	rs := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	pick := uint32(rs.Intn(int(wrs.prioritySum)))

	current := uint32(0)
	for _, item := range wrs.items {
		current += uint32(item.PriorityWeight)
		if pick < current {
			return item.Value, true
		}
	}

	return zero, false
}
