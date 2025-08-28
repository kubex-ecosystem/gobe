// Package approval implements the approval request management.
package approval

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rafa-mori/gobe/internal/config"
	"github.com/rafa-mori/gobe/internal/observers/events"

	"github.com/google/uuid"
)

type Manager struct {
	config           config.ApprovalConfig
	eventStream      *events.Stream
	pendingApprovals map[string]*Request
	mu               sync.RWMutex
}

type Request struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	Platform  string                 `json:"platform"`
	Details   map[string]interface{} `json:"details"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	Status    Status                 `json:"status"`
}

type Response struct {
	RequestID  string    `json:"request_id"`
	Approved   bool      `json:"approved"`
	ApproverID string    `json:"approver_id"`
	Timestamp  time.Time `json:"timestamp"`
}

type Status int

const (
	StatusPending Status = iota
	StatusApproved
	StatusRejected
	StatusExpired
)

func NewManager(config config.ApprovalConfig, eventStream *events.Stream) *Manager {
	return &Manager{
		config:           config,
		eventStream:      eventStream,
		pendingApprovals: make(map[string]*Request),
	}
}

func (m *Manager) RequestApproval(ctx context.Context, req Request) (*Response, error) {
	req.ID = uuid.New().String()
	req.CreatedAt = time.Now()
	req.ExpiresAt = time.Now().Add(time.Duration(m.config.ApprovalTimeoutMinutes) * time.Minute)
	req.Status = StatusPending

	m.mu.Lock()
	m.pendingApprovals[req.ID] = &req
	m.mu.Unlock()

	// Broadcast approval request to frontend
	m.eventStream.Broadcast(events.Event{
		Type: "approval_request",
		Data: req,
	})

	// Wait for approval with timeout
	return m.waitForApproval(ctx, req.ID)
}

func (m *Manager) ProcessApproval(requestID string, approved bool, approverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	req, exists := m.pendingApprovals[requestID]
	if !exists {
		return fmt.Errorf("approval request not found: %s", requestID)
	}

	if time.Now().After(req.ExpiresAt) {
		req.Status = StatusExpired
		return fmt.Errorf("approval request expired")
	}

	if approved {
		req.Status = StatusApproved
	} else {
		req.Status = StatusRejected
	}

	response := Response{
		RequestID:  requestID,
		Approved:   approved,
		ApproverID: approverID,
		Timestamp:  time.Now(),
	}

	// Notify frontend of approval result
	m.eventStream.Broadcast(events.Event{
		Type: "approval_result",
		Data: response,
	})

	return nil
}

func (m *Manager) waitForApproval(ctx context.Context, requestID string) (*Response, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(time.Duration(m.config.ApprovalTimeoutMinutes) * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("approval timeout")
		case <-ticker.C:
			m.mu.RLock()
			req, exists := m.pendingApprovals[requestID]
			if !exists {
				m.mu.RUnlock()
				return nil, fmt.Errorf("approval request not found")
			}

			if req.Status != StatusPending {
				response := &Response{
					RequestID: requestID,
					Approved:  req.Status == StatusApproved,
					Timestamp: time.Now(),
				}
				m.mu.RUnlock()
				return response, nil
			}
			m.mu.RUnlock()
		}
	}
}

func (m *Manager) GetPendingApprovals() []*Request {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var pending []*Request
	for _, req := range m.pendingApprovals {
		if req.Status == StatusPending && time.Now().Before(req.ExpiresAt) {
			pending = append(pending, req)
		}
	}

	return pending
}
