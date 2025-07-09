// Package homesignal provides a thread-safe signaling system.
//
// # Core Components
//
// JobSignal provides individual signal management with buffered channels and non-blocking operations.
// Signals are dropped when buffers are full, preventing blocking behavior:
//
// The Scheduler interface provides periodic signal distribution to multiple subscribers
// with two implementations optimized for different use cases:
//
// # BrokerScheduler
//
// High-performance concurrent scheduler that prevents head-of-line blocking by sending
// signals to all subscribers concurrently. Slow subscribers don't affect fast ones.
//
// # SequentialScheduler
//
// Simple sequential scheduler that sends signals to all subscribers within the main loop.
// Uses fewer resources but slow subscribers can delay signal delivery to others.
//
// # Thread Safety
//
// All operations are thread-safe and designed for concurrent use. Signal operations
// are non-blocking and will silently drop signals that cannot be delivered, preventing
// deadlocks.
package homesignal
