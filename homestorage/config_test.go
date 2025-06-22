package homestorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := newDefaultConfig()
	assert.Positive(t, cfg.capacity, "capacity should be positive")
}
