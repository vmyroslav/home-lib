package homestorage

import (
	"errors"
	"sync"
)

var (
	ErrNotFound         = errors.New("element not found")
	ErrAlreadyExists    = errors.New("element already exists")
	ErrCapacityExceeded = errors.New("storage capacity exceeded")
)

// InMemoryStorage is a simple thread-safe in-memory storage that you can use for testing, mocking, etc.
type InMemoryStorage[T any] struct {
	storage  map[string]T
	capacity uint64

	mutex sync.RWMutex
}

// NewInMemoryStorage returns a new instance of InMemoryStorage with the given options.
// The default capacity is 1024.
func NewInMemoryStorage[T any](opts ...Option) *InMemoryStorage[T] {
	cfg := newDefaultConfig()

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return &InMemoryStorage[T]{
		storage:  make(map[string]T),
		capacity: cfg.capacity,
		mutex:    sync.RWMutex{},
	}
}

// All returns all elements from the storage.
func (i *InMemoryStorage[T]) All() []T {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	values := make([]T, 0, len(i.storage))
	for _, value := range i.storage {
		values = append(values, value)
	}

	return values
}

// Add adds a new element to the storage.
// If the element with the given key already exists, ErrAlreadyExists is returned.
// If the storage is full, ErrCapacityExceeded is returned.
func (i *InMemoryStorage[T]) Add(key string, value T) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if len(i.storage) >= int(i.capacity) {
		return ErrCapacityExceeded
	}

	if _, ok := i.storage[key]; ok {
		return ErrAlreadyExists
	}

	i.storage[key] = value

	return nil
}

// Get returns an element from the storage by the given key.
// If the element is not found, ErrNotFound is returned.
func (i *InMemoryStorage[T]) Get(key string) (T, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	var defaultVal T

	value, ok := i.storage[key]
	if !ok {
		return defaultVal, ErrNotFound
	}

	return value, nil
}

// Upsert updates an element in the storage by the given key.
// If the element is not found, it is added to the storage.
func (i *InMemoryStorage[T]) Upsert(key string, value T) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.storage[key] = value
}

func (i *InMemoryStorage[T]) Replace(key string, value T) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if _, ok := i.storage[key]; !ok {
		return ErrNotFound
	}

	i.storage[key] = value

	return nil
}

func (i *InMemoryStorage[T]) Remove(key string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if _, ok := i.storage[key]; !ok {
		return ErrNotFound
	}

	delete(i.storage, key)

	return nil
}

// Clear removes all elements from the storage.
func (i *InMemoryStorage[T]) Clear() {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.storage = make(map[string]T)
}

// Count returns the number of elements in the storage.
func (i *InMemoryStorage[T]) Count() uint64 {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	return uint64(len(i.storage))
}
