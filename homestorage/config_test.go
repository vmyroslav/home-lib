package homestorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := newDefaultConfig()
	assert.Truef(t, cfg.capacity > 0, "capacity should be greater than 0")
}
