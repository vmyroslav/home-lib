package homestorage

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryStorage_Add(t *testing.T) {
	t.Parallel()

	type args[T any] struct {
		key   string
		value T
	}

	type testCase[T any] struct {
		name    string
		s       *InMemoryStorage[T]
		args    []args[T]
		wantErr error
	}

	tests := []testCase[string]{
		{
			name: "Add one element",
			s:    NewInMemoryStorage[string](),
			args: []args[string]{
				{
					key:   "key",
					value: "value",
				},
			},
			wantErr: nil,
		},
		{
			name: "Add multiple elements",
			s:    NewInMemoryStorage[string](WithCapacity(100)),
			args: []args[string]{
				{
					key:   "key",
					value: "value",
				},
				{
					key:   "key2",
					value: "value2",
				},
				{
					key:   "key3",
					value: "value3",
				},
			},
			wantErr: nil,
		},
		{
			name: "Add duplicated element",
			s:    NewInMemoryStorage[string](WithCapacity(100)),
			args: []args[string]{
				{
					key:   "key",
					value: "value",
				},
				{
					key:   "key",
					value: "value2",
				},
			},
			wantErr: ErrAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, arg := range tt.args {
				err := tt.s.Add(arg.key, arg.value)
				if err != nil {
					assert.EqualError(t, err, tt.wantErr.Error())
					continue
				}

				require.NoError(t, err)

				got, err := tt.s.Get(arg.key)
				require.NoError(t, err)

				assert.Equalf(t, arg.value, got, "Get(%v)", arg.key)
			}
		})
	}
}

func TestInMemoryStorage_Add_ExceedCapacity(t *testing.T) {
	t.Parallel()

	s := NewInMemoryStorage[int64](WithCapacity(1))

	_ = s.Add("key", 1)
	err := s.Add("key2", 2)

	assert.ErrorIs(t, err, ErrCapacityExceeded)
}

func TestInMemoryStorage_All(t *testing.T) {
	t.Parallel()

	type args[T any] struct {
		key   string
		value T
	}

	type testCase[T any] struct {
		name string
		s    *InMemoryStorage[T]
		args []args[T]
	}

	tests := []testCase[string]{
		{
			name: "Add one element",
			s:    NewInMemoryStorage[string](),
			args: []args[string]{
				{
					key:   "key",
					value: "value",
				},
			},
		},
		{
			name: "Add multiple elements",
			s:    NewInMemoryStorage[string](WithCapacity(100)),
			args: []args[string]{
				{
					key:   "key",
					value: "value",
				},
				{
					key:   "key2",
					value: "value2",
				},
				{
					key:   "key3",
					value: "value3",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, arg := range tt.args {
				err := tt.s.Add(arg.key, arg.value)
				require.NoError(t, err)
			}

			assert.Equal(t, len(tt.args), len(tt.s.All()))
		})
	}
}

func TestInMemoryStorage_ConcurrentAdd(t *testing.T) {
	t.Parallel()

	s := NewInMemoryStorage[int](WithCapacity(100))

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			err := s.Add(fmt.Sprintf("key%d", i), i)
			require.NoError(t, err)
		}(i)
	}

	wg.Wait()

	assert.Equal(t, uint64(100), s.Count())
}

func TestInMemoryStorage_Upsert(t *testing.T) {
	t.Parallel()

	s := NewInMemoryStorage[int](WithCapacity(100))

	_ = s.Add("key", 1)
	s.Upsert("key", 2)  // update existing key
	s.Upsert("key2", 3) // add new key

	got, err := s.Get("key")

	require.NoError(t, err)
	assert.Equal(t, 2, got)

	got, err = s.Get("key2")
	require.NoError(t, err)
	assert.Equal(t, 3, got)
}

func TestInMemoryStorage_Replace(t *testing.T) {
	t.Parallel()

	type args[T any] struct {
		key   string
		value T
	}

	type testCase[T any] struct {
		name    string
		s       *InMemoryStorage[T]
		args    []args[T]
		wantErr error
	}

	s := NewInMemoryStorage[string](WithCapacity(100))
	_ = s.Add("key", "value")
	_ = s.Add("key-2", "value-2")
	_ = s.Add("key-3", "value-3")

	tests := []testCase[string]{
		{
			name: "Replace non-existing element",
			s:    NewInMemoryStorage[string](),
			args: []args[string]{
				{
					key:   "key",
					value: "value",
				},
			},
			wantErr: ErrNotFound,
		},
		{
			name: "Replace existing element",
			s:    s,
			args: []args[string]{
				{
					key:   "key",
					value: "new-value",
				},
			},
			wantErr: nil,
		},
		{
			name: "Replace existing element for multiple times",
			s:    s,
			args: []args[string]{
				{
					key:   "key-3",
					value: "new-value-1",
				},
				{
					key:   "key-3",
					value: "new-value-2",
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			for _, arg := range tt.args {
				err := tt.s.Replace(arg.key, arg.value)
				assert.ErrorIs(t, err, tt.wantErr)

				if tt.wantErr == nil {
					got, err := tt.s.Get(arg.key)
					require.NoError(t, err)
					assert.Equal(t, arg.value, got)
				}
			}
		})
	}
}

func TestInMemoryStorage_Delete(t *testing.T) {
	t.Parallel()

	s := NewInMemoryStorage[int64](WithCapacity(100))

	_ = s.Add("key", 1)
	_ = s.Add("key2", 2)
	_ = s.Add("key3", 3)

	type testCase[T any] struct {
		name    string
		s       *InMemoryStorage[T]
		key     string
		wantErr error
	}

	tests := []testCase[int64]{
		{
			name:    "Delete non-existing element",
			s:       NewInMemoryStorage[int64](),
			key:     "key",
			wantErr: ErrNotFound,
		},
		{
			name:    "Delete existing element",
			s:       s,
			key:     "key",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.s.Delete(tt.key)
			assert.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				_, err := tt.s.Get(tt.key)
				assert.ErrorIs(t, err, ErrNotFound)
			}
		})
	}
}

func TestInMemoryStorage_MustDelete(t *testing.T) {
	t.Parallel()

	s := NewInMemoryStorage[int64](WithCapacity(100))

	_ = s.Add("key", 1)
	_ = s.Add("key2", 2)

	s.MustDelete("key")
	s.MustDelete("key3")

	_, err := s.Get("key")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestInMemoryStorage_Clear(t *testing.T) {
	t.Parallel()

	s := NewInMemoryStorage[int64]()

	_ = s.Add("key", 1)
	_ = s.Add("key2", 2)

	s.Clear()

	_, err := s.Get("key")
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, uint64(0), s.Count())
}

func TestInMemoryStorage_Count(t *testing.T) {
	t.Parallel()

	type testCase[T any] struct {
		name     string
		i        *InMemoryStorage[T]
		elements []T
		want     uint64
	}

	tests := []testCase[string]{
		{
			name:     "Empty homestorage",
			i:        NewInMemoryStorage[string](),
			elements: []string{},
			want:     0,
		},
		{
			name:     "One element in homestorage",
			i:        NewInMemoryStorage[string](),
			elements: []string{"one"},
			want:     1,
		},
		{
			name:     "Five elements in homestorage",
			i:        NewInMemoryStorage[string](),
			elements: []string{"one", "two", "three", "four", "five"},
			want:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, element := range tt.elements {
				err := tt.i.Add(element, element)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			assert.Equalf(t, tt.want, tt.i.Count(), "Count()")
		})
	}
}
