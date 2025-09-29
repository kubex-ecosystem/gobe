package types

import "time"

type SystemHealth struct {
	Status     string
	Version    string
	Uptime     time.Duration
	Host       string
	Mem        string
	Disk       string
	CPU        string
	Goroutines string
	GoBE       any
	MCP        any
	Analyzer   any
}
