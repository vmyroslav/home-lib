package homemath

import (
	"math/rand"
	"sync"
	"time"

	"golang.org/x/exp/constraints"
)

var (
	rng  *rand.Rand
	once sync.Once
	mu   sync.Mutex
)

func initRNG() {
	//nolint:gosec
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func Max[T constraints.Ordered](s ...T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}

	m := s[0]
	for _, v := range s {
		if m < v {
			m = v
		}
	}

	return m
}

func Min[T constraints.Ordered](s ...T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}

	m := s[0]
	for _, v := range s {
		if m > v {
			m = v
		}
	}

	return m
}

// Sum returns the sum of all values in s.
func Sum[T constraints.Integer | constraints.Float](s ...T) T {
	var sum T
	for _, v := range s {
		sum += v
	}

	return sum
}

// SumSlice returns the sum of all values in the slice.
// This is more efficient than Sum for large slices as it avoids variadic overhead.
func SumSlice[T constraints.Integer | constraints.Float](s []T) T {
	var sum T
	for _, v := range s {
		sum += v
	}

	return sum
}

// RandInt returns a random integer in the range [0, n).
// Returns 0 if n <= 0.
func RandInt(n int) int {
	once.Do(initRNG)

	if n <= 0 {
		return 0
	}

	mu.Lock()
	defer mu.Unlock()

	return rng.Intn(n)
}

// RandIntRange returns a random integer in the range [min, max].
// Returns min if min >= max.
func RandIntRange(minVal, maxVal int) int {
	once.Do(initRNG)

	if minVal >= maxVal {
		return minVal
	}

	mu.Lock()
	defer mu.Unlock()

	return rng.Intn(maxVal-minVal+1) + minVal
}
