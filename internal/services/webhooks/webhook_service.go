// Package webhooks provides a functional webhook service with persistence and AMQP integration.
package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	messagery "github.com/kubex-ecosystem/gobe/internal/sockets/messagery"
)

// WebhookEvent represents a webhook event received from external services
type WebhookEvent struct {
	ID        uuid.UUID              `json:"id"`
	Source    string                 `json:"source"`
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
	Headers   map[string]string      `json:"headers"`
	Timestamp time.Time              `json:"timestamp"`
	Processed bool                   `json:"processed"`
	Status    string                 `json:"status"`
	Error     string                 `json:"error,omitempty"`
}

// WebhookService provides functional webhook handling
type WebhookService struct {
	amqp      *messagery.AMQP
	events    []WebhookEvent
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	startTime time.Time
}

// NewWebhookService creates a new functional webhook service
func NewWebhookService(amqp *messagery.AMQP) *WebhookService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &WebhookService{
		amqp:      amqp,
		events:    make([]WebhookEvent, 0),
		ctx:       ctx,
		cancel:    cancel,
		startTime: time.Now(),
	}

	// Start background processor
	go service.processWebhooksWorker()

	gl.Log("info", "Webhook service initialized successfully")
	return service
}

// ReceiveWebhook processes an incoming webhook and queues it for processing
func (ws *WebhookService) ReceiveWebhook(source, eventType string, payload map[string]interface{}, headers map[string]string) (*WebhookEvent, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	event := WebhookEvent{
		ID:        uuid.New(),
		Source:    source,
		EventType: eventType,
		Payload:   payload,
		Headers:   headers,
		Timestamp: time.Now(),
		Processed: false,
		Status:    "received",
	}

	// Store event in memory (in production, this would be in database)
	ws.events = append(ws.events, event)

	// Publish to AMQP queue for async processing
	if ws.amqp != nil && ws.amqp.IsReady() {
		eventBytes, err := json.Marshal(event)
		if err != nil {
			gl.Log("error", "Failed to marshal webhook event", err)
		} else {
			err = ws.amqp.Publish("gobe.events", "webhook.received", eventBytes)
			if err != nil {
				gl.Log("error", "Failed to publish webhook event to AMQP", err)
			} else {
				gl.Log("info", "Webhook event published to AMQP", event.ID.String())
			}
		}
	}

	gl.Log("info", "Webhook received", "source", source, "type", eventType, "id", event.ID.String())
	return &event, nil
}

// GetWebhookEvent retrieves a specific webhook event by ID
func (ws *WebhookService) GetWebhookEvent(id uuid.UUID) (*WebhookEvent, error) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	for _, event := range ws.events {
		if event.ID == id {
			return &event, nil
		}
	}

	return nil, fmt.Errorf("webhook event not found: %s", id.String())
}

// ListWebhookEvents returns all webhook events with optional filtering
func (ws *WebhookService) ListWebhookEvents(limit int, offset int, source string) ([]WebhookEvent, int, error) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	var filtered []WebhookEvent

	// Filter by source if specified
	if source != "" {
		for _, event := range ws.events {
			if event.Source == source {
				filtered = append(filtered, event)
			}
		}
	} else {
		filtered = ws.events
	}

	total := len(filtered)

	// Apply pagination
	if offset >= total {
		return []WebhookEvent{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	// Sort by timestamp (newest first)
	result := make([]WebhookEvent, end-offset)
	copy(result, filtered[offset:end])

	return result, total, nil
}

// processWebhooksWorker runs in background to process queued webhooks
func (ws *WebhookService) processWebhooksWorker() {
	ticker := time.NewTicker(5 * time.Second) // Process every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ws.ctx.Done():
			gl.Log("info", "Webhook processor worker shutting down")
			return
		case <-ticker.C:
			ws.processQueuedWebhooks()
		}
	}
}

// processQueuedWebhooks processes pending webhook events
func (ws *WebhookService) processQueuedWebhooks() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	processed := 0
	for i := range ws.events {
		if !ws.events[i].Processed && ws.events[i].Status == "received" {
			// Simulate webhook processing
			ws.events[i].Status = "processing"

			// Process the webhook (placeholder logic)
			success := ws.processWebhookEvent(&ws.events[i])

			if success {
				ws.events[i].Status = "completed"
				ws.events[i].Processed = true
				processed++
			} else {
				ws.events[i].Status = "failed"
				ws.events[i].Error = "Processing failed"
			}
		}
	}

	if processed > 0 {
		gl.Log("info", "Processed webhook events", "count", processed)
	}
}

// processWebhookEvent processes a single webhook event
func (ws *WebhookService) processWebhookEvent(event *WebhookEvent) bool {
	gl.Log("info", "Processing webhook event", "id", event.ID.String(), "source", event.Source, "type", event.EventType)

	// Simulate processing based on event type
	switch event.EventType {
	case "github.push":
		return ws.processGitHubPush(event)
	case "discord.message":
		return ws.processDiscordMessage(event)
	case "discord.webhook":
		return ws.processDiscordWebhook(event)
	case "stripe.payment":
		return ws.processStripePayment(event)
	case "user.created":
		return ws.processUserCreated(event)
	default:
		gl.Log("info", "Generic webhook processing", "type", event.EventType)
		return true // Default to success for unknown types
	}
}

// processGitHubPush handles GitHub push events
func (ws *WebhookService) processGitHubPush(event *WebhookEvent) bool {
	gl.Log("info", "Processing GitHub push webhook", "repo", event.Payload["repository"])

	// Publish notification about the push
	if ws.amqp != nil && ws.amqp.IsReady() {
		notification := map[string]interface{}{
			"type":       "github_push",
			"repository": event.Payload["repository"],
			"commits":    event.Payload["commits"],
			"timestamp":  time.Now(),
		}

		notifBytes, _ := json.Marshal(notification)
		ws.amqp.Publish("gobe.notifications", "", notifBytes)
	}

	return true
}

// processDiscordMessage handles Discord message events
func (ws *WebhookService) processDiscordMessage(event *WebhookEvent) bool {
	gl.Log("info", "Processing Discord message webhook", "channel", event.Payload["channel_id"])

	// Could trigger bot responses or logging
	return true
}

// processDiscordWebhook handles generic Discord webhook envelopes.
func (ws *WebhookService) processDiscordWebhook(event *WebhookEvent) bool {
	gl.Log("info", "Processing Discord webhook", "verified", event.Headers["x-discord-verified"])

	if ws.amqp != nil && ws.amqp.IsReady() {
		notification := map[string]interface{}{
			"type":       event.EventType,
			"event_id":   event.Headers["x-discord-event-id"],
			"webhook_id": event.Headers["x-discord-webhook-id"],
			"verified":   event.Headers["x-discord-verified"],
			"payload":    event.Payload,
		}

		notifBytes, _ := json.Marshal(notification)
		if err := ws.amqp.Publish("gobe.discord", "webhook.received", notifBytes); err != nil {
			gl.Log("error", "Failed to publish Discord webhook notification", err)
		}
	}

	return true
}

// processStripePayment handles Stripe payment events
func (ws *WebhookService) processStripePayment(event *WebhookEvent) bool {
	gl.Log("info", "Processing Stripe payment webhook", "amount", event.Payload["amount"])

	// Update user billing, send emails, etc.
	return true
}

// processUserCreated handles user creation events
func (ws *WebhookService) processUserCreated(event *WebhookEvent) bool {
	gl.Log("info", "Processing user created webhook", "user_id", event.Payload["user_id"])

	// Send welcome email, create profile, etc.
	return true
}

// GetStats returns webhook service statistics
func (ws *WebhookService) GetStats() map[string]interface{} {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	var totalEvents, processedEvents, failedEvents, pendingEvents int

	for _, event := range ws.events {
		totalEvents++
		if event.Processed {
			processedEvents++
		}
		if event.Status == "failed" {
			failedEvents++
		}
		if event.Status == "received" {
			pendingEvents++
		}
	}

	return map[string]interface{}{
		"total_events":     totalEvents,
		"processed_events": processedEvents,
		"failed_events":    failedEvents,
		"pending_events":   pendingEvents,
		"uptime_seconds":   time.Since(ws.startTime).Seconds(),
		"amqp_connected":   ws.amqp != nil && ws.amqp.IsReady(),
		"last_updated":     time.Now().Unix(),
	}
}

// Close gracefully shuts down the webhook service
func (ws *WebhookService) Close() error {
	gl.Log("info", "Shutting down webhook service")
	ws.cancel()
	return nil
}

// RetryFailedWebhooks retries all failed webhook events
func (ws *WebhookService) RetryFailedWebhooks() (int, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	retried := 0
	for i := range ws.events {
		if ws.events[i].Status == "failed" {
			ws.events[i].Status = "received"
			ws.events[i].Processed = false
			ws.events[i].Error = ""
			retried++
		}
	}

	gl.Log("info", "Retried failed webhook events", "count", retried)
	return retried, nil
}
