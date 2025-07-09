package homesignal

import (
	"time"

	"github.com/vmyroslav/home-lib/homelogger"

	"github.com/rs/zerolog"
	"github.com/vmyroslav/home-lib/homeconfig"
)

// Config holds the configuration for a Scheduler.
type Config struct {
	Logger     *zerolog.Logger
	Period     time.Duration
	BufferSize uint16
}

// Option configures a Config.
type Option = homeconfig.Option[Config]

// NewConfig creates a new Config with the given options
func NewConfig(options ...Option) Config {
	cfg := Config{
		Period:     3 * time.Second,
		BufferSize: 5,
		Logger:     homelogger.NewNoOp(),
	}

	for _, option := range options {
		option.Apply(&cfg)
	}

	return cfg
}

// WithPeriod sets the scheduler period (time interval between each trigger).
func WithPeriod(period time.Duration) Option {
	return homeconfig.OptionFunc[Config](func(c *Config) {
		c.Period = period
	})
}

// WithBufferSize sets the size of the channel buffer for each subscription.
func WithBufferSize(size uint16) Option {
	return homeconfig.OptionFunc[Config](func(c *Config) {
		c.BufferSize = size
	})
}

// WithLogger sets the logger for the scheduler.
func WithLogger(logger *zerolog.Logger) Option {
	return homeconfig.OptionFunc[Config](func(c *Config) {
		c.Logger = logger
	})
}
