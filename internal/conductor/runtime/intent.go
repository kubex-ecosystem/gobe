package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Intent struct {
	ID        uuid.UUID              `json:"id"`
	Method    string                 `json:"method"`
	Path      string                 `json:"path"`
	Headers   map[string]string      `json:"headers"`
	Body      map[string]interface{} `json:"body"`
	Caller    string                 `json:"caller"`
	Timestamp time.Time              `json:"timestamp"`
	Schedule  string                 `json:"schedule"`
	Command   string                 `json:"command"`
	Context   context.Context        `json:"-"`
}

func NewIntentFromRequest(r *http.Request) (*Intent, error) {
	intent := &Intent{
		ID:        GenerateUUID(),
		Method:    r.Method,
		Path:      r.URL.Path,
		Headers:   make(map[string]string),
		Caller:    r.RemoteAddr,
		Timestamp: time.Now(),
		Schedule:  r.Header.Get("X-Schedule"),
		Command:   r.Header.Get("X-Command"),
		Context:   r.Context(),
	}
	for k, v := range r.Header {
		if len(v) > 0 {
			intent.Headers[k] = v[0]
		}
	}
	if r.Body != nil {
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&intent.Body)
	}
	return intent, nil
}

func GenerateUUID() uuid.UUID {
	uid, _ := uuid.NewRandom()
	return uid
}
