package homelogger

import (
	"os"

	"github.com/rs/zerolog"
)

const applicationKey = "application"

func New(options ...Option) *zerolog.Logger {
	logger := zerolog.New(os.Stdout)

	for _, o := range options {
		o.Apply(&logger)
	}

	return &logger
}

func NewDefault(appName string) *zerolog.Logger {
	options := []Option{
		WithLevel(zerolog.InfoLevel),
		WithOutput(os.Stdout),
		WithCaller(),
		WithTime(),
		WithStack(),
		WithApplicationName(appName),
	}

	return New(options...)
}

func NewNoOp() *zerolog.Logger {
	logger := zerolog.Nop()

	return &logger
}
