// Package events provides a WebSocket-based event streaming service.
package events

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"

	gl "github.com/kubex-ecosystem/logz/logger"
)

type Stream struct {
	clients      map[string]*Client
	register     chan *Client
	unregister   chan *Client
	broadcast    chan Event
	messageQueue chan MessageProcessingJob
	mu           sync.RWMutex
}

type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan Event
}

type Event struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

type MessageProcessingJob struct {
	ID        string      `json:"id"`
	Platform  string      `json:"platform"`
	Message   interface{} `json:"message"`
	Priority  Priority    `json:"priority"`
	CreatedAt time.Time   `json:"created_at"`
}

type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityUrgent
)

func NewStream() *Stream {
	return &Stream{
		clients:      make(map[string]*Client),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		broadcast:    make(chan Event, 256),
		messageQueue: make(chan MessageProcessingJob, 100),
	}
}

func (s *Stream) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client.ID] = client
			s.mu.Unlock()
			go s.handleClient(client)

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client.ID]; ok {
				delete(s.clients, client.ID)
				close(client.Send)
			}
			s.mu.Unlock()

		case event := <-s.broadcast:
			s.mu.RLock()
			for _, client := range s.clients {
				select {
				case client.Send <- event:
				default:
					close(client.Send)
					delete(s.clients, client.ID)
				}
			}
			s.mu.RUnlock()

		case job := <-s.messageQueue:
			go s.processMessageJob(job)
		}
	}
}

func (s *Stream) RegisterClient(client *Client) {
	s.register <- client
}

func (s *Stream) UnregisterClient(client *Client) {
	s.unregister <- client
}

func (s *Stream) Broadcast(event Event) {
	event.Timestamp = time.Now()
	s.broadcast <- event
}

func (s *Stream) ProcessMessage(job MessageProcessingJob) {
	job.CreatedAt = time.Now()
	s.messageQueue <- job
}

func (s *Stream) handleClient(client *Client) {
	defer func() {
		s.UnregisterClient(client)
		client.Conn.Close()
	}()

	for event := range client.Send {
		client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := client.Conn.WriteJSON(event); err != nil {
			gl.Log("error", "WebSocket write error: "+err.Error())
			return
		}
	}
}

func (s *Stream) processMessageJob(job MessageProcessingJob) {
	// Notify frontend that processing started
	s.Broadcast(Event{
		Type: "message_processing_started",
		Data: map[string]interface{}{
			"job_id":   job.ID,
			"platform": job.Platform,
			"priority": job.Priority,
		},
	})

	// Simulate processing time
	time.Sleep(1 * time.Second)

	// Notify processing completed
	s.Broadcast(Event{
		Type: "message_processing_completed",
		Data: map[string]interface{}{
			"job_id":   job.ID,
			"platform": job.Platform,
			"result":   "Mensagem processada com sucesso",
		},
	})
}

func (s *Stream) Close() {
	close(s.broadcast)
	close(s.messageQueue)
	close(s.register)
	close(s.unregister)
}
