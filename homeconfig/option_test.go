package homeconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// testConfig is a test configuration struct
type testConfig struct {
	Name    string
	Value   int
	Enabled bool
}

func TestOptionFunc_Apply(t *testing.T) {
	t.Parallel()

	config := &testConfig{}

	option := OptionFunc[testConfig](func(c *testConfig) {
		c.Name = "test"
		c.Value = 42
		c.Enabled = true
	})

	option.Apply(config)

	assert.Equal(t, "test", config.Name)
	assert.Equal(t, 42, config.Value)
	assert.True(t, config.Enabled)
}

func TestOptionFunc_Multiple(t *testing.T) {
	t.Parallel()

	config := &testConfig{}

	nameOption := OptionFunc[testConfig](func(c *testConfig) {
		c.Name = "example"
	})

	valueOption := OptionFunc[testConfig](func(c *testConfig) {
		c.Value = 100
	})

	enabledOption := OptionFunc[testConfig](func(c *testConfig) {
		c.Enabled = true
	})

	nameOption.Apply(config)
	valueOption.Apply(config)
	enabledOption.Apply(config)

	assert.Equal(t, "example", config.Name)
	assert.Equal(t, 100, config.Value)
	assert.True(t, config.Enabled)
}

func TestApplyOptions(t *testing.T) {
	t.Parallel()

	config := &testConfig{}

	options := []Option[testConfig]{
		OptionFunc[testConfig](func(c *testConfig) {
			c.Name = "applied"
		}),
		OptionFunc[testConfig](func(c *testConfig) {
			c.Value = 200
		}),
		OptionFunc[testConfig](func(c *testConfig) {
			c.Enabled = false
		}),
	}

	ApplyOptions(config, options...)

	assert.Equal(t, "applied", config.Name)
	assert.Equal(t, 200, config.Value)
	assert.False(t, config.Enabled)
}

func TestApplyOptions_EmptyOptions(t *testing.T) {
	t.Parallel()

	config := &testConfig{
		Name:    "initial",
		Value:   1,
		Enabled: true,
	}

	ApplyOptions(config)

	assert.Equal(t, "initial", config.Name)
	assert.Equal(t, 1, config.Value)
	assert.True(t, config.Enabled)
}

func TestApplyOptions_OverwriteValues(t *testing.T) {
	t.Parallel()

	config := &testConfig{
		Name:    "initial",
		Value:   1,
		Enabled: true,
	}

	option1 := OptionFunc[testConfig](func(c *testConfig) {
		c.Name = "first"
	})

	option2 := OptionFunc[testConfig](func(c *testConfig) {
		c.Name = "second"
	})

	ApplyOptions(config, option1, option2)

	assert.Equal(t, "second", config.Name)
	assert.Equal(t, 1, config.Value)
	assert.True(t, config.Enabled)
}

// Example of creating helper functions using the generic pattern
func WithName(name string) Option[testConfig] {
	return OptionFunc[testConfig](func(c *testConfig) {
		c.Name = name
	})
}

func WithValue(value int) Option[testConfig] {
	return OptionFunc[testConfig](func(c *testConfig) {
		c.Value = value
	})
}

func WithEnabled(enabled bool) Option[testConfig] {
	return OptionFunc[testConfig](func(c *testConfig) {
		c.Enabled = enabled
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Parallel()

	config := &testConfig{}

	ApplyOptions(config,
		WithName("helper"),
		WithValue(999),
		WithEnabled(true),
	)

	assert.Equal(t, "helper", config.Name)
	assert.Equal(t, 999, config.Value)
	assert.True(t, config.Enabled)
}
