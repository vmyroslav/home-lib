package homelogger

import (
	"io"

	"github.com/rs/zerolog"

	"github.com/vmyroslav/home-lib/homeconfig"
)

// Option sets a parameter for the logger.
type Option = homeconfig.Option[zerolog.Logger]

// WithLevel sets the logger level.
func WithLevel(level zerolog.Level) Option {
	return homeconfig.OptionFunc[zerolog.Logger](func(logger *zerolog.Logger) {
		l := logger.Level(level)
		*logger = l
	})
}

// WithCaller adds the caller to the log messages.
func WithCaller() Option {
	return homeconfig.OptionFunc[zerolog.Logger](func(logger *zerolog.Logger) {
		l := logger.With().Caller().Logger()
		*logger = l
	})
}

// WithOutput sets the output writer.
func WithOutput(output io.Writer) Option {
	return homeconfig.OptionFunc[zerolog.Logger](func(logger *zerolog.Logger) {
		l := logger.Output(output)
		*logger = l
	})
}

// WithTime adds the time to the log messages.
func WithTime() Option {
	return homeconfig.OptionFunc[zerolog.Logger](func(logger *zerolog.Logger) {
		l := logger.With().Timestamp().Logger()
		*logger = l
	})
}

// WithStack adds the stack trace to the log messages.
func WithStack() Option {
	return homeconfig.OptionFunc[zerolog.Logger](func(logger *zerolog.Logger) {
		l := logger.With().Stack().Logger()
		*logger = l
	})
}

// WithApplicationName adds the application name to the log messages.
func WithApplicationName(n string) Option {
	return homeconfig.OptionFunc[zerolog.Logger](func(logger *zerolog.Logger) {
		l := logger.With().Str(applicationKey, n).Logger()
		*logger = l
	})
}

// WithConsoleWriter enables pretty, human-readable console output with colors.
func WithConsoleWriter(out io.Writer) Option {
	return homeconfig.OptionFunc[zerolog.Logger](func(logger *zerolog.Logger) {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        out,
			TimeFormat: "15:04:05",
		}
		l := logger.Output(consoleWriter)
		*logger = l
	})
}
