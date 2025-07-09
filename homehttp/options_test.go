package homehttp

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestClientOptionWithAppName(t *testing.T) {
	config := &clientConfig{}
	option := WithAppName("TestApp")
	option.Apply(config)

	assert.Equal(t, "TestApp", config.AppName)
}

func TestClientOptionWithTimeout(t *testing.T) {
	config := &clientConfig{}
	option := WithTimeout(time.Second * 5)
	option.Apply(config)

	assert.Equal(t, time.Second*5, config.Timeout)
}

func TestClientOptionWithLogger(t *testing.T) {
	config := &clientConfig{}
	logger := zerolog.Nop()
	option := WithLogger(&logger)
	option.Apply(config)

	assert.Equal(t, &logger, config.Logger)
}
