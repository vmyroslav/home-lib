# HomeLib

[![Build Status](https://github.com/vmyroslav/home-lib/actions/workflows/ci.yml/badge.svg)](https://github.com/vmyroslav/home-lib/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/vmyroslav/home-lib/branch/main/graph/badge.svg?token=8F5APGAZT6)](https://codecov.io/gh/vmyroslav/home-lib)
[![Go Report Card](https://goreportcard.com/badge/github.com/vmyroslav/home-lib)](https://goreportcard.com/report/github.com/vmyroslav/home-lib)
[![Godoc](https://pkg.go.dev/badge/github.com/vmyroslav/home-lib?utm_source=godoc)](https://pkg.go.dev/github.com/vmyroslav/home-lib)

## Table of Contents

- [Prerequisites](#Prerequisites)
- [Installation](#Installation)
- [Description](#Description)
    - [Configuration](#Configuration)
    - [HTTP](#HTTP)
    - [Math](#Math)
    - [Storage](#Storage)
    - [Logging](#Logging)
    - [Signaling](#Signaling)
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

### Configuration
Generic option pattern implementation for type-safe configuration across all packages.

```go
import "github.com/vmyroslav/home-lib/homeconfig"

// Define your configuration struct
type ServerConfig struct {
    Port    int
    Host    string
    Timeout time.Duration
}

// Create option functions using the generic pattern
func WithPort(port int) homeconfig.Option[ServerConfig] {
    return homeconfig.OptionFunc[ServerConfig](func(c *ServerConfig) {
        c.Port = port
    })
}

func WithHost(host string) homeconfig.Option[ServerConfig] {
    return homeconfig.OptionFunc[ServerConfig](func(c *ServerConfig) {
        c.Host = host
    })
}

// Use options to configure your service
func NewServer(opts ...homeconfig.Option[ServerConfig]) *Server {
    config := &ServerConfig{
        Port:    8080,  // defaults
        Host:    "localhost",
        Timeout: 30 * time.Second,
    }
    
    // Apply all options
    homeconfig.ApplyOptions(config, opts...)
    
    return &Server{config: config}
}

// Usage
server := NewServer(
    WithPort(3000),
    WithHost("0.0.0.0"),
)
```

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

### Signaling
Thread-safe signaling system with non-blocking operations and concurrent delivery for applications.

```go
import (
    "context"
    "fmt"
    "time"
    
    "github.com/vmyroslav/home-lib/homesignal"
    "github.com/vmyroslav/home-lib/homelogger"
)

// Create configuration with functional options
cfg := homesignal.NewConfig(
    homesignal.WithPeriod(100 * time.Millisecond),
    homesignal.WithBufferSize(10),
    homesignal.WithLogger(homelogger.NewNoOp()),
)

// Create a scheduler for string signals - choose implementation:
scheduler := homesignal.NewBrokerScheduler[string](cfg)      // concurrent delivery
// OR
// scheduler := homesignal.NewSequentialScheduler[string](cfg) // sequential delivery

// Subscribe to receive signals
subscription := scheduler.Subscribe()

// Start the scheduler with a signal factory function
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    err := scheduler.Start(ctx, func() string {
        return fmt.Sprintf("tick-%d", time.Now().Unix())
    })
    if err != nil {
        fmt.Printf("Scheduler error: %v\n", err)
    }
}()

// Receive signals from subscription
for {
    signal, ok := subscription.Next()
    if !ok {
        break // channel closed
    }
    fmt.Printf("Received: %s\n", signal)
}

// Clean up
scheduler.Stop()

// Example with JobSignal for direct signaling
jobSignal := homesignal.NewJobSignal[int]("worker-1", 5)

// Send signals (non-blocking, drops if buffer full)
jobSignal.Send(42)

// Send with context (non-blocking, drops if context canceled or buffer full)
ctx, cancel = context.WithTimeout(context.Background(), time.Second)
defer cancel()
jobSignal.SendWithContext(ctx, 100)

// Receive signals
value, ok := jobSignal.Next()
if ok {
    fmt.Printf("Received job: %d\n", value)
}

// Close when done
jobSignal.Close()
```

#### Scheduler Implementations

Choose the implementation that best fits your use case:

- **BrokerScheduler** (recommended):
  - Concurrent signal delivery prevents head-of-line blocking
  - Slow subscribers don't affect fast subscribers
  - Ideal for high-throughput applications with mixed subscriber speeds
  - Use `NewBrokerScheduler[T](cfg)`
  
- **SequentialScheduler**:
  - Sends signals to all subscribers sequentially in main loop
  - Lower resource usage but slower subscribers can delay others
  - Suitable for simple applications with predictable subscriber behavior
  - Use `NewSequentialScheduler[T](cfg)`

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