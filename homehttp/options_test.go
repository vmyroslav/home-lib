package homehttp

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestClientOptionWithAppName(t *testing.T) {
	t.Parallel()

	config := &clientConfig{}
	option := WithAppName("TestApp")
	option.Apply(config)

	assert.Equal(t, "TestApp", config.AppName)
}

func TestClientOptionWithTimeout(t *testing.T) {
	t.Parallel()

	config := &clientConfig{}
	option := WithTimeout(time.Second * 5)
	option.Apply(config)

	assert.Equal(t, time.Second*5, config.Timeout)
}

func TestClientOptionWithLogger(t *testing.T) {
	t.Parallel()

	config := &clientConfig{}
	logger := zerolog.Nop()
	option := WithLogger(&logger)
	option.Apply(config)

	assert.Equal(t, &logger, config.Logger)
}

func TestPerHostRateLimitOptions(t *testing.T) {
	t.Parallel()

	t.Run("WithPerHostTokenBucketRateLimit creates strategy", func(t *testing.T) {
		strategy := TokenBucketRateLimit(10, 5, WithScope(RateLimitScopeHost))
		assert.NotNil(t, strategy)

		rateLimitStrat, ok := strategy.(*rateLimitStrategy)
		assert.True(t, ok, "expected rateLimitStrategy")
		assert.NotNil(t, rateLimitStrat.limiter)
		assert.Equal(t, RateLimitScopeHost, rateLimitStrat.limiter.scope)
	})

	t.Run("WithPerHostTokenBucketRateLimit wrapper matches manual scope", func(t *testing.T) {
		cfg1 := &clientConfig{RateLimitStrategy: NoRateLimitStrategy()}
		WithPerHostTokenBucketRateLimit(10, 5).Apply(cfg1)

		cfg2 := &clientConfig{RateLimitStrategy: NoRateLimitStrategy()}
		WithTokenBucketRateLimit(10, 5, WithScope(RateLimitScopeHost)).Apply(cfg2)

		strat1, ok1 := cfg1.RateLimitStrategy.(*rateLimitStrategy)
		strat2, ok2 := cfg2.RateLimitStrategy.(*rateLimitStrategy)

		assert.True(t, ok1 && ok2, "both should be rateLimitStrategy")
		assert.Equal(t, strat2.limiter.scope, strat1.limiter.scope)
		assert.Equal(t, RateLimitScopeHost, strat1.limiter.scope)
	})

	t.Run("WithPerHostFixedWindowRateLimit creates strategy", func(t *testing.T) {
		strategy := FixedWindowRateLimit(100, time.Minute, WithScope(RateLimitScopeHost))
		assert.NotNil(t, strategy)

		rateLimitStrat, ok := strategy.(*rateLimitStrategy)
		assert.True(t, ok, "expected rateLimitStrategy")
		assert.NotNil(t, rateLimitStrat.limiter)
		assert.Equal(t, RateLimitScopeHost, rateLimitStrat.limiter.scope)
	})

	t.Run("WithPerHostFixedWindowRateLimit with additional options", func(t *testing.T) {
		strategy := FixedWindowRateLimit(100, time.Minute,
			WithScope(RateLimitScopeHost),
			WithBehavior(RateLimitBehaviorError),
		)

		rateLimitStrat, ok := strategy.(*rateLimitStrategy)
		assert.True(t, ok, "expected rateLimitStrategy")
		assert.Equal(t, RateLimitBehaviorError, rateLimitStrat.behavior)
		assert.Equal(t, RateLimitScopeHost, rateLimitStrat.limiter.scope)
	})
}
