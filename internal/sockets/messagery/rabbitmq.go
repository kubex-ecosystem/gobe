// Package messagery provides AMQP messaging functionality.
package messagery

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	t "github.com/kubex-ecosystem/gdbase/types"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"

	crtSvc "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
)

type DBConfig = t.DBConfig

type AMQP struct {
	URL           string
	Conn          *amqp.Connection
	Chan          *amqp.Channel
	ready         atomic.Bool
	reconnecting  atomic.Bool
	mu            sync.RWMutex
	lastError     error
	lastErrorTime time.Time
	connAttempts  int64
	maxRetries    int
	retryInterval time.Duration
}

func NewAMQP() *AMQP {
	return &AMQP{
		maxRetries:    10,
		retryInterval: 5 * time.Second,
	}
}

func (a *AMQP) Connect(ctx context.Context, url string, logf func(string, ...any)) error {
	a.URL = url
	backoff := []time.Duration{500 * time.Millisecond, 1 * time.Second, 2 * time.Second, 5 * time.Second, 10 * time.Second, 30 * time.Second}
	var last error
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		conn, err := amqp.Dial(url)
		if err == nil {
			ch, err := conn.Channel()
			if err != nil {
				_ = conn.Close()
				last = err
			}
			if err == nil {
				a.Conn, a.Chan = conn, ch
				if err := a.declareTopology(); err != nil {
					_ = ch.Close()
					_ = conn.Close()
					last = err
				}
				a.ready.Store(true)
				logf("amqp conectado e pronto")
				go a.watch(ctx, logf) // reconectar em caso de close
				return nil
			}
		} else {
			last = err
		}
		d := backoff[min(i, len(backoff)-1)]
		logf("amqp falhou: %v; retry em %s", last, d)
		time.Sleep(d)
	}
}

func (a *AMQP) watch(ctx context.Context, logf func(string, ...any)) {
	errs := a.Conn.NotifyClose(make(chan *amqp.Error, 1))
	for {
		select {
		case <-ctx.Done():
			logf("amqp watcher context cancelled")
			return
		case e := <-errs:
			a.handleConnectionLoss(e, logf)
			// tenta reconectar em background
			go a.reconnectWithBackoff(ctx, logf)
			return
		}
	}
}

func (a *AMQP) handleConnectionLoss(err *amqp.Error, logf func(string, ...any)) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.ready.Store(false)
	a.lastError = err
	a.lastErrorTime = time.Now()

	if err != nil {
		logf("amqp connection lost: %v", err)
	} else {
		logf("amqp connection closed gracefully")
	}

	// Clean up existing connections
	if a.Chan != nil {
		_ = a.Chan.Close()
		a.Chan = nil
	}
	if a.Conn != nil {
		_ = a.Conn.Close()
		a.Conn = nil
	}
}

func (a *AMQP) reconnectWithBackoff(ctx context.Context, logf func(string, ...any)) {
	if !a.reconnecting.CompareAndSwap(false, true) {
		logf("reconnection already in progress")
		return
	}
	defer a.reconnecting.Store(false)

	backoff := []time.Duration{500 * time.Millisecond, 1 * time.Second, 2 * time.Second, 5 * time.Second, 10 * time.Second, 30 * time.Second}
	attempts := 0

	for attempts < a.maxRetries {
		select {
		case <-ctx.Done():
			logf("reconnection cancelled by context")
			return
		default:
		}

		atomic.AddInt64(&a.connAttempts, 1)
		attempts++

		logf("attempting amqp reconnection (attempt %d/%d)", attempts, a.maxRetries)

		err := a.Connect(ctx, a.URL, logf)
		if err == nil {
			logf("amqp reconnection successful after %d attempts", attempts)
			return
		}

		a.mu.Lock()
		a.lastError = err
		a.lastErrorTime = time.Now()
		a.mu.Unlock()

		delay := backoff[min(attempts-1, len(backoff)-1)]
		logf("amqp reconnection failed (attempt %d): %v, retrying in %v", attempts, err, delay)

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
			continue
		}
	}

	logf("amqp reconnection failed after %d attempts, giving up", a.maxRetries)
}

func (a *AMQP) declareTopology() error {
	if a.Chan == nil {
		return errors.New("channel is nil")
	}

	// Declare default exchanges and queues for the GoBE application
	exchanges := []struct {
		name       string
		kind       string
		durable    bool
		autoDelete bool
	}{
		{"gobe.events", "topic", true, false},
		{"gobe.logs", "direct", true, false},
		{"gobe.notifications", "fanout", true, false},
	}

	// Declare exchanges
	for _, exchange := range exchanges {
		err := a.Chan.ExchangeDeclare(
			exchange.name,
			exchange.kind,
			exchange.durable,
			exchange.autoDelete,
			false, // internal
			false, // noWait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange.name, err)
		}
		gl.Log("info", "Declared AMQP exchange", exchange.name, exchange.kind)
	}

	// Declare default queues
	queues := []struct {
		name       string
		durable    bool
		autoDelete bool
		exclusive  bool
	}{
		{"gobe.system.logs", true, false, false},
		{"gobe.system.events", true, false, false},
		{"gobe.mcp.tasks", true, false, false},
	}

	for _, queue := range queues {
		_, err := a.Chan.QueueDeclare(
			queue.name,
			queue.durable,
			queue.autoDelete,
			queue.exclusive,
			false, // noWait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queue.name, err)
		}
		gl.Log("info", "Declared AMQP queue", queue.name)
	}

	// Bind queues to exchanges
	bindings := []struct {
		queue    string
		exchange string
		key      string
	}{
		{"gobe.system.logs", "gobe.logs", "system"},
		{"gobe.system.events", "gobe.events", "system.*"},
		{"gobe.mcp.tasks", "gobe.events", "mcp.task.*"},
	}

	for _, binding := range bindings {
		err := a.Chan.QueueBind(
			binding.queue,
			binding.key,
			binding.exchange,
			false, // noWait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s: %w", binding.queue, binding.exchange, err)
		}
		gl.Log("info", "Bound AMQP queue", binding.queue, "to exchange", binding.exchange, "with key", binding.key)
	}

	return nil
}

func (a *AMQP) PublishReliable(exchange, key string, body []byte) error {
	return a.PublishReliableWithTimeout(exchange, key, body, 10*time.Second)
}

func (a *AMQP) PublishReliableWithTimeout(exchange, key string, body []byte, timeout time.Duration) error {
	if !a.ready.Load() {
		return errors.New("amqp not ready")
	}

	a.mu.RLock()
	channel := a.Chan
	a.mu.RUnlock()

	if channel == nil {
		return errors.New("amqp channel is nil")
	}

	if err := channel.Confirm(false); err != nil {
		return fmt.Errorf("failed to set confirm mode: %w", err)
	}

	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	err := channel.Publish(exchange, key, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Timestamp:    time.Now(),
		Body:         body,
	})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for confirmation with timeout
	select {
	case c := <-confirms:
		if !c.Ack {
			return errors.New("publish nack received")
		}
		return nil
	case <-time.After(timeout):
		return errors.New("publish confirmation timeout")
	}
}

func (a *AMQP) Publish(exchange, key string, body []byte) error {
	if !a.ready.Load() {
		return errors.New("amqp not ready")
	}

	a.mu.RLock()
	channel := a.Chan
	a.mu.RUnlock()

	if channel == nil {
		return errors.New("amqp channel is nil")
	}

	return channel.Publish(exchange, key, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Timestamp:    time.Now(),
		Body:         body,
	})
}

func GetRabbitMQURL(dbConfig *DBConfig) string {
	var host = ""
	var port = ""
	var username = ""
	var password = ""
	if dbConfig.Messagery.RabbitMQ.Host != "" {
		host = dbConfig.Messagery.RabbitMQ.Host
	} else {
		host = "localhost"
	}
	if dbConfig.Messagery.RabbitMQ.Port != "" {
		strPort, ok := dbConfig.Messagery.RabbitMQ.Port.(string)
		if ok {
			port = strPort
		} else {
			gl.Log("error", "RabbitMQ port is not a string")
			port = "5672"
		}
	} else {
		port = "5672"
	}
	if dbConfig.Messagery.RabbitMQ.Username != "" {
		username = dbConfig.Messagery.RabbitMQ.Username
	} else {
		username = "gobe"
	}
	if dbConfig.Messagery.RabbitMQ.Password != "" {
		password = dbConfig.Messagery.RabbitMQ.Password
	} else {
		rabbitPassKey, rabbitPassErr := crtSvc.GetOrGenPasswordKeyringPass("rabbitmq")
		if rabbitPassErr != nil {
			gl.Log("error", "Skipping RabbitMQ setup due to error generating password")
			gl.Log("debug", fmt.Sprintf("Error generating key: %v", rabbitPassErr))
			goto postRabbit
		}
		password = string(rabbitPassKey)
	}

	if host != "" && port != "" && username != "" && password != "" {
		return fmt.Sprintf("amqp://%s:%s@%s:%s/%s", username, password, host, port, "gobe")
	}
postRabbit:
	return ""
}

// ConnectionStats returns connection statistics
func (a *AMQP) ConnectionStats() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	stats := map[string]interface{}{
		"ready":               a.ready.Load(),
		"reconnecting":        a.reconnecting.Load(),
		"connection_attempts": atomic.LoadInt64(&a.connAttempts),
		"url":                 a.URL,
		"max_retries":         a.maxRetries,
		"retry_interval":      a.retryInterval.String(),
	}

	if !a.lastErrorTime.IsZero() {
		stats["last_error_time"] = a.lastErrorTime.Unix()
		stats["last_error_ago"] = time.Since(a.lastErrorTime).String()
		if a.lastError != nil {
			stats["last_error"] = a.lastError.Error()
		}
	}

	if a.Conn != nil && !a.Conn.IsClosed() {
		stats["connection_active"] = true
	} else {
		stats["connection_active"] = false
	}

	return stats
}

// IsReady returns whether the connection is ready
func (a *AMQP) IsReady() bool {
	return a.ready.Load()
}

// Close closes the AMQP connection gracefully
func (a *AMQP) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.ready.Store(false)

	var errs []error

	if a.Chan != nil {
		if err := a.Chan.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
		a.Chan = nil
	}

	if a.Conn != nil && !a.Conn.IsClosed() {
		if err := a.Conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
		a.Conn = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	return nil
}

// SetMaxRetries sets the maximum number of reconnection attempts
func (a *AMQP) SetMaxRetries(maxRetries int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.maxRetries = maxRetries
}

// SetRetryInterval sets the base retry interval for reconnections
func (a *AMQP) SetRetryInterval(interval time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.retryInterval = interval
}
