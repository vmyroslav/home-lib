# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- **homestorage**: Fixed apacity violation in `InMemoryStorage.Upsert()` method
  - `Upsert()` now properly checks capacity limits when adding new keys
  - **BREAKING CHANGE**: `Upsert()` method signature changed from `func(string, T)` to `func(string, T) error`
  - Returns `ErrCapacityExceeded` when attempting to add new keys beyond configured capacity
  - Existing key updates continue to work without capacity restrictions
- **homestorage**: Added thread safety to `WeightedRandomSelector`
  - Added `sync.RWMutex` to prevent race conditions during concurrent access
  - All methods (`Add`, `AddMany`, `AddOrdered`, `AddItem`, `Get`) are now thread-safe
  - Fixed potential data corruption in `prioritySum` field during concurrent operations

### Added
- **homestorage**: Test suites for concurrency and capacity violations

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
- Comprehensive README with installation and usage instructions
- Package documentation with examples
- CI badges and code quality indicators
- MIT license