package homestorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := newDefaultConfig()
	assert.Greater(t, cfg.capacity, uint64(0))
}
