# HomeLib

[![Build Status](https://github.com/vmyroslav/home-lib/actions/workflows/ci.yml/badge.svg)](https://github.com/vmyroslav/home-lib/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/vmyroslav/home-lib/branch/main/graph/badge.svg?token=8F5APGAZT6)](https://codecov.io/gh/vmyroslav/home-lib)
[![Go Report Card](https://goreportcard.com/badge/github.com/vmyroslav/home-lib)](https://goreportcard.com/report/github.com/vmyroslav/home-lib)
[![Godoc](https://pkg.go.dev/badge/github.com/vmyroslav/home-lib?utm_source=godoc)](https://pkg.go.dev/github.com/vmyroslav/home-lib)

## Table of Contents

- [Prerequisites](#Prerequisites)
- [Installation](#Installation)
- [Description](#Description)
    - [HTTP](#HTTP)
    - [Math](#Math)
    - [Storage](#Storage)
    - [Logging](#Logging)
    - [Tests](#Tests)

## Prerequisites
- `Go >= 1.24`

## Installation
```bash
go get github.com/vmyroslav/home-lib
```

## Description
This is a collection of packages that I use in my personal projects. 
I decided to make them public so that I can use them in other projects.
I do not guarantee that they will be supported and updated, so I do not recommend using them in production.

### HTTP
HTTP client with retry logic, customizable backoff strategies, and comprehensive timeout handling.

```go
client := homehttp.NewClient(
    homehttp.WithRetry(3, homehttp.LinearBackoff(100*time.Millisecond)),
    homehttp.WithTimeout(5*time.Second),
)
resp, err := client.Get(ctx, "https://api.example.com/data")
```

### Math
Mathematical utilities with generic constraints and thread-safe random number generation.

```go
// Min/Max operations
max := homemath.Max(1, 5, 3, 2) // returns 5
min := homemath.Min(1, 5, 3, 2) // returns 1

// Sum operations
sum := homemath.Sum(1, 2, 3, 4) // returns 10

// Thread-safe random numbers
randomInt := homemath.RandInt(100)        // 0-99
randomRange := homemath.RandIntRange(1, 6) // 1-6 (dice roll)
```

### Storage
Key-value storage with multiple backends (in-memory, file-based) and advanced features like weighted random selection.

```go
// In-memory storage
storage := homestorage.NewMemoryStorage()
storage.Set("key", "value")
value, exists := storage.Get("key")

// Weighted random selector
selector := homestorage.NewWeightedRandomSelector[string]()
selector.Add("item1", 3)
selector.Add("item2", 1)
selected := selector.Select() // "item1" is 3x more likely
```

### Logging
Structured logging with configurable levels and JSON output.

```go
logger := homelogger.New(homelogger.LevelInfo)
logger.Info("Processing request", homelogger.Field("user_id", 123))
logger.Error("Failed to process", homelogger.Field("error", err))
```

### Tests
Testing utilities for HTTP servers, context management, environment variables, and assertions.

```go
// HTTP testing
server, url := hometests.JSONServer(t, 200, map[string]string{"status": "ok"})
defer server.Close()

// Context testing
ctx, cancel := hometests.ContextWithTimeout(t, 5*time.Second)
defer cancel()

// Environment testing
cleanup := hometests.EnvOverride(t, "API_KEY", "test-key")
defer cleanup()
```