// Package homelogger provides a structured logging wrapper around zerolog with functional options.
//
// This package simplifies the creation and configuration of zerolog loggers,
// making it easy to create consistent, loggers across applications.
//
// # Basic Usage
//
// Create a logger with default settings:
//
//	logger := homelogger.NewDefault("my-app")
//	logger.Info().Msg("Application started")
//
// Create a logger with custom configuration:
//
//	logger := homelogger.New(
//		homelogger.WithLevel(zerolog.DebugLevel),
//		homelogger.WithApplicationName("my-service"),
//		homelogger.WithCaller(),
//		homelogger.WithTime(),
//	)
//
// Create a no-op logger for testing:
//
//	logger := homelogger.NewNoOp()
package homelogger
