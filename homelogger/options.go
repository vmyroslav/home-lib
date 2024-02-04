package homelogger

import (
	"io"

	"github.com/rs/zerolog"
)

// Option sets a parameter for the logger.
type Option interface {
	apply(logger *zerolog.Logger)
}

type optionFn func(cfg *zerolog.Logger)

func (fn optionFn) apply(cfg *zerolog.Logger) {
	fn(cfg)
}

// WithLevel sets the logger level.
func WithLevel(level zerolog.Level) Option {
	return optionFn(func(logger *zerolog.Logger) {
		l := logger.Level(level)
		*logger = l
	})
}

// WithCaller adds the caller to the log messages.
func WithCaller() Option {
	return optionFn(func(logger *zerolog.Logger) {
		l := logger.With().Caller().Logger()
		*logger = l
	})
}

// WithOutput sets the output writer.
func WithOutput(output io.Writer) Option {
	return optionFn(func(logger *zerolog.Logger) {
		l := logger.Output(output)
		*logger = l
	})
}

// WithTime adds the time to the log messages.
func WithTime() Option {
	return optionFn(func(logger *zerolog.Logger) {
		l := logger.With().Timestamp().Logger()
		*logger = l
	})
}

// WithStack adds the stack trace to the log messages.
func WithStack() Option {
	return optionFn(func(logger *zerolog.Logger) {
		l := logger.With().Stack().Logger()
		*logger = l
	})
}

// WithApplicationName adds the application name to the log messages.
func WithApplicationName(n string) Option {
	return optionFn(func(logger *zerolog.Logger) {
		l := logger.With().Str(applicationKey, n).Logger()
		*logger = l
	})
}
