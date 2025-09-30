package testsmcp

import (
	"context"
	"fmt"
	"testing"

	"github.com/kubex-ecosystem/gobe/internal/services/mcp"
)

func TestRegistry_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		spec    mcp.ToolSpec
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid tool registration",
			spec: mcp.ToolSpec{
				Name:        "test.tool",
				Title:       "Test Tool",
				Description: "A test tool",
				Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
					return "test result", nil
				},
			},
			wantErr: false,
		},
		{
			name: "empty name should fail",
			spec: mcp.ToolSpec{
				Name:        "",
				Title:       "Test Tool",
				Description: "A test tool",
				Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
					return "test result", nil
				},
			},
			wantErr: true,
			errMsg:  "tool name cannot be empty",
		},
		{
			name: "nil handler should fail",
			spec: mcp.ToolSpec{
				Name:        "test.tool",
				Title:       "Test Tool",
				Description: "A test tool",
				Handler:     nil,
			},
			wantErr: true,
			errMsg:  "tool handler cannot be nil for tool: test.tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := mcp.NewRegistry()
			err := registry.Register(tt.spec)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Register() expected error but got none")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Register() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Register() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestRegistry_List(t *testing.T) {
	t.Parallel()

	registry := mcp.NewRegistry()

	// Initially empty
	tools := registry.List()
	if len(tools) != 0 {
		t.Errorf("List() expected empty registry, got %d tools", len(tools))
	}

	// Add a tool
	spec := mcp.ToolSpec{
		Name:        "test.tool",
		Title:       "Test Tool",
		Description: "A test tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "test result", nil
		},
	}

	err := registry.Register(spec)
	if err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	// Should have one tool
	tools = registry.List()
	if len(tools) != 1 {
		t.Errorf("List() expected 1 tool, got %d", len(tools))
	}

	// Handler should be nil in response (security)
	if tools[0].Handler != nil {
		t.Errorf("List() returned tool with handler, should be nil for security")
	}

	// Other fields should be preserved
	if tools[0].Name != spec.Name {
		t.Errorf("List() tool name = %v, want %v", tools[0].Name, spec.Name)
	}
}

func TestRegistry_Exec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		toolName   string
		args       map[string]interface{}
		setupTool  bool
		wantResult interface{}
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "successful execution",
			toolName:   "test.tool",
			args:       map[string]interface{}{"input": "test"},
			setupTool:  true,
			wantResult: "test result",
			wantErr:    false,
		},
		{
			name:      "empty tool name",
			toolName:  "",
			args:      map[string]interface{}{},
			setupTool: false,
			wantErr:   true,
			errMsg:    "tool name cannot be empty",
		},
		{
			name:      "tool not found",
			toolName:  "nonexistent.tool",
			args:      map[string]interface{}{},
			setupTool: false,
			wantErr:   true,
			errMsg:    "tool not found: nonexistent.tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := mcp.NewRegistry()

			if tt.setupTool {
				spec := mcp.ToolSpec{
					Name:        "test.tool",
					Title:       "Test Tool",
					Description: "A test tool",
					Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
						return "test result", nil
					},
				}
				err := registry.Register(spec)
				if err != nil {
					t.Fatalf("Register() unexpected error = %v", err)
				}
			}

			ctx := context.Background()
			result, err := registry.Exec(ctx, tt.toolName, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Exec() expected error but got none")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Exec() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Exec() unexpected error = %v", err)
				}
				if result != tt.wantResult {
					t.Errorf("Exec() result = %v, want %v", result, tt.wantResult)
				}
			}
		})
	}
}

func TestRegistry_GetTool(t *testing.T) {
	t.Parallel()

	registry := mcp.NewRegistry()

	// Tool not found
	tool, exists := registry.GetTool("nonexistent")
	if exists {
		t.Errorf("GetTool() expected tool to not exist")
	}
	if tool != nil {
		t.Errorf("GetTool() expected nil tool, got %v", tool)
	}

	// Add a tool
	spec := mcp.ToolSpec{
		Name:        "test.tool",
		Title:       "Test Tool",
		Description: "A test tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "test result", nil
		},
	}

	err := registry.Register(spec)
	if err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	// Tool found
	tool, exists = registry.GetTool("test.tool")
	if !exists {
		t.Errorf("GetTool() expected tool to exist")
	}
	if tool == nil {
		t.Fatalf("GetTool() expected non-nil tool")
	}

	// Handler should be nil for security
	if tool.Handler != nil {
		t.Errorf("GetTool() returned tool with handler, should be nil for security")
	}

	// Other fields should be preserved
	if tool.Name != spec.Name {
		t.Errorf("GetTool() tool name = %v, want %v", tool.Name, spec.Name)
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	registry := mcp.NewRegistry()

	// Test concurrent registrations and executions
	const numGoroutines = 10

	// Register tools concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			spec := mcp.ToolSpec{
				Name:        fmt.Sprintf("tool.%d", id),
				Title:       fmt.Sprintf("Tool %d", id),
				Description: fmt.Sprintf("Test tool %d", id),
				Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
					return fmt.Sprintf("result %d", id), nil
				},
			}
			registry.Register(spec)
		}(i)
	}

	// List tools concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			registry.List()
		}()
	}

	// Execute tools concurrently (some may fail if tools aren't registered yet)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			ctx := context.Background()
			registry.Exec(ctx, fmt.Sprintf("tool.%d", id), map[string]interface{}{})
		}(i)
	}

	// This test primarily checks for race conditions
	// If there are race conditions, the test will fail with -race flag
}
