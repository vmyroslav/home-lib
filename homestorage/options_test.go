package homestorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithCapacity(t *testing.T) {
	t.Parallel()

	limits := []uint64{1, 5, 10, 100, 1000, 10000}

	for _, limit := range limits {
		storage := NewInMemoryStorage[string](WithCapacity(limit))

		assert.Equal(t, limit, storage.limit)
	}
}
