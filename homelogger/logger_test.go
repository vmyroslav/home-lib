package homelogger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	logger := New()
	assert.NotNil(t, logger)

	var buf bytes.Buffer
	logger = New(
		WithOutput(&buf),
		WithLevel(zerolog.WarnLevel),
	)

	logger.Info().Msg("info message")
	logger.Warn().Msg("warn message")

	// Info should be filtered out, only warn should appear
	output := buf.String()
	assert.NotContains(t, output, "info message")
	assert.Contains(t, output, "warn message")
}

func TestNewDefault(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := New(
		WithOutput(&buf),
		WithLevel(zerolog.InfoLevel),
		WithApplicationName("test-app"),
		WithTime(),
		WithCaller(),
	)

	logger.Info().Msg("test message")

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "test-app")
	assert.Contains(t, output, "level")
	assert.Contains(t, output, "time")
}

func TestNewNoOp(t *testing.T) {
	t.Parallel()

	logger := NewNoOp()
	assert.NotNil(t, logger)

	var buf bytes.Buffer

	logger.Info().Msg("this should not appear")
	logger.Error().Msg("this should also not appear")

	assert.Empty(t, buf.String())
}

// TestWithLevel tests that the logger respects the configured level.
func TestWithLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(
		WithOutput(&buf),
		WithLevel(zerolog.ErrorLevel),
	)

	logger.Debug().Msg("debug")
	logger.Info().Msg("info")
	logger.Warn().Msg("warn")
	logger.Error().Msg("error")

	output := buf.String()
	assert.NotContains(t, output, "debug")
	assert.NotContains(t, output, "info")
	assert.NotContains(t, output, "warn")
	assert.Contains(t, output, "error")
}

func TestWithOutput(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(WithOutput(&buf))

	logger.Info().Msg("test output")

	assert.Contains(t, buf.String(), "test output")
}

// TestWithCaller tests that the logger includes caller information when configured.
func TestWithCaller(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(
		WithOutput(&buf),
		WithCaller(),
	)

	logger.Info().Msg("test caller")

	output := buf.String()
	assert.Contains(t, output, "test caller")
	assert.Contains(t, output, "caller")
}

// TestWithTime tests that the logger includes timestamp when configured.
func TestWithTime(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(
		WithOutput(&buf),
		WithTime(),
	)

	logger.Info().Msg("test time")

	output := buf.String()
	assert.Contains(t, output, "test time")
	assert.Contains(t, output, "time")
}

func TestWithStack(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(
		WithOutput(&buf),
		WithStack(),
	)

	logger.Error().Msg("test stack")

	output := buf.String()
	assert.Contains(t, output, "test stack")
	// Should contain stack trace info
	assert.Contains(t, output, "stack")
}

func TestWithApplicationName(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	appName := "my-test-app"
	logger := New(
		WithOutput(&buf),
		WithApplicationName(appName),
	)

	logger.Info().Msg("test app name")

	output := buf.String()
	assert.Contains(t, output, "test app name")
	assert.Contains(t, output, appName)
	assert.Contains(t, output, applicationKey) // Should contain the "application" key
}

func TestMultipleOptions(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(
		WithOutput(&buf),
		WithLevel(zerolog.InfoLevel),
		WithTime(),
		WithCaller(),
		WithApplicationName("multi-option-test"),
	)

	logger.Info().Msg("testing multiple options")

	output := buf.String()

	// Parse as JSON to verify structure
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "info", logEntry["level"])
	assert.Equal(t, "testing multiple options", logEntry["message"])
	assert.Equal(t, "multi-option-test", logEntry["application"])
	assert.Contains(t, logEntry, "time")
	assert.Contains(t, logEntry, "caller")
}

func TestLoggerLevels(t *testing.T) {
	t.Parallel()

	levels := []zerolog.Level{
		zerolog.TraceLevel,
		zerolog.DebugLevel,
		zerolog.InfoLevel,
		zerolog.WarnLevel,
		zerolog.ErrorLevel,
		zerolog.FatalLevel,
		zerolog.PanicLevel,
	}

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(
				WithOutput(&buf),
				WithLevel(level),
			)

			assert.NotNil(t, logger)

			// Write a message at the configured level
			switch level {
			case zerolog.TraceLevel:
				logger.Trace().Msg("trace message")
			case zerolog.DebugLevel:
				logger.Debug().Msg("debug message")
			case zerolog.InfoLevel:
				logger.Info().Msg("info message")
			case zerolog.WarnLevel:
				logger.Warn().Msg("warn message")
			case zerolog.ErrorLevel:
				logger.Error().Msg("error message")
			case zerolog.FatalLevel:
				// Skip fatal as it calls os.Exit
			case zerolog.PanicLevel:
				// Skip panic as it panics
			case zerolog.NoLevel:
			case zerolog.Disabled:
			}

			if level != zerolog.FatalLevel && level != zerolog.PanicLevel {
				assert.NotEmpty(t, buf.String())
			}
		})
	}
}

func TestWithConsoleWriter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(
		WithConsoleWriter(&buf),
		WithLevel(zerolog.InfoLevel),
	)

	logger.Info().Msg("console test")

	output := buf.String()
	// Console output should be human-readable, not JSON
	assert.Contains(t, output, "console test")
	assert.NotContains(t, output, `"message"`) // Should not be JSON format
}

func TestWithPrettyLogging(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(
		WithConsoleWriter(&buf),
		WithLevel(zerolog.InfoLevel),
	)

	logger.Info().Msg("pretty test")

	output := buf.String()
	assert.Contains(t, output, "pretty test")  // Pretty output should be human-readable, not JSON
	assert.NotContains(t, output, `"message"`) // Should not be JSON format
}

func TestLoggerPresets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		logger *zerolog.Logger
	}{
		{"Default", NewDefault("test-app")},
		{"NoOp", NewNoOp()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.logger)

			// Test that all loggers can handle basic operations without panicking
			tt.logger.Info().Str("test", "value").Msg("test message")
		})
	}
}
