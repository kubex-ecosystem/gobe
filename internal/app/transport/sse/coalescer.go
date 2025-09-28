// Package sse provides a coalescer to aggregate small chunks into user friendly payloads.
package sse

import (
	"errors"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	// ErrCoalescerClosed indicates writes after Close.
	ErrCoalescerClosed = errors.New("sse: coalescer closed")
)

// FlushFunc receives coalesced payloads ready to be streamed to clients.
type FlushFunc func(data string)

// Coalescer aggregates micro chunks into user friendly payloads before flushing to the stream.
type Coalescer struct {
	mu      sync.Mutex
	buffer  strings.Builder
	timer   *time.Timer
	timeout time.Duration
	maxSize int
	flush   FlushFunc
	closed  bool
}

const (
	defaultTimeout = 75 * time.Millisecond
	defaultMaxSize = 100
)

// NewCoalescer builds a new Coalescer with sensible defaults.
func NewCoalescer(timeout time.Duration, maxSize int, flush FlushFunc) *Coalescer {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if maxSize <= 0 {
		maxSize = defaultMaxSize
	}
	if flush == nil {
		flush = func(string) {}
	}
	return &Coalescer{
		timeout: timeout,
		maxSize: maxSize,
		flush:   flush,
	}
}

// Add appends a chunk to the buffer, flushing when punctuation, newline or limits hit.
func (c *Coalescer) Add(chunk string) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrCoalescerClosed
	}
	if chunk == "" {
		c.mu.Unlock()
		return nil
	}

	c.buffer.WriteString(chunk)

	shouldFlush := c.shouldFlushLocked(chunk) || c.buffer.Len() >= c.maxSize
	var payload string
	if shouldFlush {
		payload = c.flushLocked()
	} else {
		c.ensureTimerLocked()
	}
	c.mu.Unlock()

	if payload != "" {
		c.flush(payload)
	}
	return nil
}

// Flush forces the buffer to emit if there is pending data.
func (c *Coalescer) Flush() {
	c.mu.Lock()
	payload := c.flushLocked()
	c.mu.Unlock()

	if payload != "" {
		c.flush(payload)
	}
}

// Close flushes remaining data and prevents new writes.
func (c *Coalescer) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	payload := c.flushLocked()
	c.mu.Unlock()

	if payload != "" {
		c.flush(payload)
	}
}

func (c *Coalescer) ensureTimerLocked() {
	if c.timer == nil {
		c.timer = time.AfterFunc(c.timeout, c.onTimeout)
		return
	}
	c.timer.Reset(c.timeout)
}

func (c *Coalescer) onTimeout() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	payload := c.flushLocked()
	c.mu.Unlock()

	if payload != "" {
		c.flush(payload)
	}
}

func (c *Coalescer) flushLocked() string {
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
	if c.buffer.Len() == 0 {
		return ""
	}
	payload := c.buffer.String()
	c.buffer.Reset()
	return payload
}

func (c *Coalescer) shouldFlushLocked(chunk string) bool {
	trimmed := strings.TrimSpace(chunk)
	if trimmed == "" {
		return false
	}
	r, _ := utf8.DecodeLastRuneInString(trimmed)
	switch r {
	case '.', '!', '?', '\n', '\r':
		return true
	default:
		return false
	}
}
