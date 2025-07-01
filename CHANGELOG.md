# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.2.0] - 2025-07-01

### Fixed
- **homestorage**: Fixed capacity violation in `InMemoryStorage.Upsert()` method
  - `Upsert()` now properly checks capacity limits when adding new keys
  - **BREAKING CHANGE**: `Upsert()` method signature changed from `func(string, T)` to `func(string, T) error`
  - Returns `ErrCapacityExceeded` when attempting to add new keys beyond configured capacity
  - Existing key updates continue to work without capacity restrictions
- **homestorage**: Added thread safety to `WeightedRandomSelector`
  - Added `sync.RWMutex` to prevent race conditions during concurrent access
  - All methods (`Add`, `AddMany`, `AddOrdered`, `AddItem`, `Get`) are now thread-safe
  - Fixed potential data corruption in `prioritySum` field during concurrent operations
- **homehttp**: Fixed incomplete retry logic causing ineffective backoff delays
  - `retryWaitMin` and `retryWaitMax` fields are now properly initialized with default values
  - Backoff strategies now receive correct min/max wait times instead of zero values
  - Added default retry wait times: 100ms minimum, 3s maximum
  - Fixed retry mechanism that was previously not applying proper delays between attempts
- **homehttp**: Fixed error logging bug in retry logic
  - Error logging now correctly logs the actual request error (`doErr`) instead of stale error
  - Improves debugging by showing the real cause of request failures during retries
- **homehttp**: Fixed basic auth token expiration behavior
  - Basic auth tokens no longer have arbitrary 1-hour expiration
  - `Token.IsValid()` method now properly handles tokens without expiration (zero time)
  - Basic auth credentials remain valid until explicitly invalidated, matching HTTP spec
- **homehttp**: Removed redundant `NoRetryStrategy` struct
  - Eliminated duplicate functionality already provided by `NoRetry` variable
  - Simplified codebase by using consistent functional approach for retry strategies

### Added
- **homestorage**: Test suites for concurrency and capacity violations
- **homehttp**: New client configuration options for retry timing
  - `WithRetryWaitTimes(min, max)` to set both minimum and maximum retry wait times
  - `WithMinRetryWait(duration)` to set minimum wait time between retries
  - `WithMaxRetryWait(duration)` to set maximum wait time between retries
  - Test suite for retry backoff timing validation
- **homehttp**: Enhanced test coverage for token validation
  - Added tests for basic auth token expiration behavior
  - Added tests for `Token.IsValid()` method with various scenarios
  - Improved test coverage for authentication mechanisms
- **homelogger**: Comprehensive test coverage and enhanced functionality
  - Added complete test suite covering all logger options and configurations
  - Added console/pretty output formatting options (`WithConsoleWriter`, `WithPrettyLogging`)
  - Added environment-specific logger presets (`NewDevelopment`, `NewProduction`)
  - Development logger with colored console output and debug level logging
  - Production logger optimized for structured JSON logging in production environments
- **homemath**: Expanded mathematical utilities and enhanced test coverage
  - Added `Abs()` function for absolute value calculation with signed numeric types
  - Added `Clamp()` function for constraining values within min/max bounds
  - Added `Sum()` and `SumSlice()` functions for efficient summation of numeric values
  - Added `MinMax()` and `MinMaxSlice()` functions for single-pass min/max calculation
  - Added random number generation functions with thread-safe initialization
    - `RandInt(n)` for random integers in range [0, n)
    - `RandIntRange(min, max)` for random integers in range [min, max]
    - `RandFloat64()` for random floats in range [0.0, 1.0)
    - `RandFloat64Range(min, max)` for random floats in range [min, max)
  - Added complete test coverage for all mathematical functions with edge cases
- **hometests**: Comprehensive testing utilities and infrastructure enhancements
  - Added context testing utilities
  - Added HTTP testing utilities
  - Enhanced `RandomPort` function to use homemath random utilities
  - Added HTTP mocking utilities (`MockRoundTripper`, request/response helpers)

## [v0.1.0] - 2024-02-04

### Added
- **homehttp**: HTTP client with retry mechanisms and backoff strategies
  - Configurable retry strategies (`NoRetry`, `RetryOn500x`, `MultiRetryStrategies`)
  - Multiple backoff strategies (`ConstantBackoff`, `LinearBackoff`, `NoBackoff`)
  - Middleware support for headers, authentication, and user-agent
  - Token provider interface with basic auth support
  - JSON request building utilities
- **homelogger**: Structured logging wrapper around zerolog
  - Configurable logger with functional options
  - Default logger configuration with caller, timestamp, and stack trace
  - Application name tagging support
  - No-op logger for testing
- **homemath**: Generic mathematical utilities
  - `Max()` and `Min()` functions for any ordered type
  - Zero-value handling for empty inputs
- **homestorage**: In-memory storage with configurable capacity
  - Thread-safe generic storage with CRUD operations
  - Configurable capacity limits with error handling
  - `WeightedRandomSelector` for priority-based random selection
  - Support for weighted item selection with various add methods
- **hometests**: Testing utilities
  - Environment variable override functions for test isolation
  - Environment file loading utilities with search functionality
  - Network testing helpers
- **Build & CI**: Development workflow
  - GitHub Actions CI with linting and testing
  - Codecov integration
  - golangci-lint
  - Task-based build system 
  - Dependabot for automatic dependency updates

### Documentation
- README with installation and usage instructions
- Package documentation with examples
- CI badges and code quality indicators
- MIT license