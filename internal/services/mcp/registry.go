// Package mcp provides the registry for managing MCP tools dynamically.
package mcp

import (
	"context"
	"fmt"
	"sync"

	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
)

// ToolSpec defines the specification for an MCP tool
type ToolSpec struct {
	Name        string                 `json:"name"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Auth        string                 `json:"auth,omitempty"`
	Args        map[string]interface{} `json:"args,omitempty"`
	Handler     ToolHandler            `json:"-"`
}

// ToolHandler defines the function signature for tool handlers
type ToolHandler func(context.Context, map[string]interface{}) (interface{}, error)

// Registry interface for managing MCP tools
type Registry interface {
	Register(spec ToolSpec) error
	List() []ToolSpec
	Exec(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error)
	GetTool(name string) (*ToolSpec, bool)
}

// registry implements the Registry interface
type registry struct {
	mu    sync.RWMutex
	tools map[string]ToolSpec
}

// NewRegistry creates a new MCP tools registry
func NewRegistry() Registry {
	gl.Log("debug", "Creating new MCP registry")
	return &registry{
		tools: make(map[string]ToolSpec),
	}
}

// Register adds a new tool to the registry
func (r *registry) Register(spec ToolSpec) error {
	if spec.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if spec.Handler == nil {
		return fmt.Errorf("tool handler cannot be nil for tool: %s", spec.Name)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[spec.Name]; exists {
		gl.Log("warn", "Tool already exists, overwriting", spec.Name)
	}

	r.tools[spec.Name] = spec
	gl.Log("info", "Tool registered successfully", spec.Name, spec.Description)

	return nil
}

// List returns all registered tools
func (r *registry) List() []ToolSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]ToolSpec, 0, len(r.tools))
	for _, tool := range r.tools {
		// Remove handler from response for security
		toolCopy := tool
		toolCopy.Handler = nil
		tools = append(tools, toolCopy)
	}

	gl.Log("debug", "Listing tools", len(tools))
	return tools
}

// Exec executes a tool by name with the provided arguments
func (r *registry) Exec(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	if toolName == "" {
		return nil, fmt.Errorf("tool name cannot be empty")
	}

	r.mu.RLock()
	tool, exists := r.tools[toolName]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	gl.Log("info", "Executing tool", toolName, len(args))

	result, err := tool.Handler(ctx, args)
	if err != nil {
		gl.Log("error", "Tool execution failed", toolName, err)
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	gl.Log("debug", "Tool executed successfully", toolName)
	return result, nil
}

// GetTool returns a specific tool by name
func (r *registry) GetTool(name string) (*ToolSpec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, false
	}

	// Return a copy without the handler for security
	toolCopy := tool
	toolCopy.Handler = nil
	return &toolCopy, true
}