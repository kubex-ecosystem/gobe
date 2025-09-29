// Package system provides the controller for managing mcp system-level operations.
package system

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/app/security/execsafe"
	services "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	"github.com/kubex-ecosystem/gobe/internal/contracts/types"
	"github.com/kubex-ecosystem/gobe/internal/module/logger"
	"github.com/kubex-ecosystem/gobe/internal/services/mcp"
	"github.com/kubex-ecosystem/gobe/internal/services/mcp/hooks"
	"github.com/kubex-ecosystem/gobe/internal/services/mcp/system"
	"gorm.io/gorm"

	l "github.com/kubex-ecosystem/logz"
)

var (
	gl          = logger.GetLogger[l.Logger](nil)
	sysServ     services.ISystemService
	mcpRegistry mcp.Registry
)

type MetricsController struct {
	dbConn        *gorm.DB
	mcpState      *hooks.Bitstate[uint64, system.SystemDomain]
	systemService services.ISystemService
	registry      mcp.Registry
	apiWrapper    *types.APIWrapper[interface{}]
}

func NewMetricsController(db *gorm.DB) *MetricsController {
	if db == nil {
		gl.Log("warn", "Database connection is nil")
	}

	// Initialize registry if not already done
	if mcpRegistry == nil {
		mcpRegistry = mcp.NewRegistry()
		gl.Log("info", "Initialized new MCP registry")

		// Register built-in tools
		err := mcp.RegisterBuiltinTools(mcpRegistry)
		if err != nil {
			gl.Log("error", "Failed to register built-in tools", err)
		}
	}

	return &MetricsController{
		dbConn:        db,
		systemService: sysServ,
		registry:      mcpRegistry,
		apiWrapper:    types.NewAPIWrapper[interface{}](),
	}
}

func (c *MetricsController) GetGeneralSystemMetrics(ctx *gin.Context) {
	if c.systemService == nil {
		if sysServ == nil {
			sysServ = services.NewSystemService()
		}
		if sysServ == nil {
			gl.Log("error", "System service is nil")
			return
		}
		c.systemService = sysServ
	}

	// mcp := getMCPInstance()
	// cpu, mem := collectCpuMem()
	// mcpstate.UpdateSystemStateFromMetrics(mcp.SystemState, cpu, mem)

	metrics, err := c.systemService.GetCurrentMetrics()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"data":      metrics,
		"timestamp": time.Now().Unix(),
	})
}

//	type IMCPServer interface {
//		RegisterTools()
//		RegisterResources()
//		HandleAnalyzeMessage(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error)
//		HandleSendMessage(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error)
//		HandleCreateTask(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error)
//		HandleSystemInfo(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error)
//		HandleShellCommand(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error)
//		GetCPUInfo() (string, error)
//		GetMemoryInfo() (string, error)
//		GetDiskInfo() (string, error)
//	}

// RegisterRoutes registers the routes for the MetricsController.
func (c *MetricsController) RegisterRoutes(router *gin.RouterGroup) {
	if router == nil {
		gl.Log("error", "Router group is nil, cannot register routes")
		return
	}

	gl.Log("info", "Routes registered for MetricsController")
	if c.systemService == nil {
		gl.Log("warn", "System service is nil, initializing a new instance")
		c.systemService = services.NewSystemService()
	}
	if c.systemService == nil {
		gl.Log("error", "Failed to initialize system service")
		return
	}
	// Register the system service routes
	ssrvc, ok := c.systemService.(*services.SystemService)
	if !ok {
		gl.Log("error", "Failed to assert system service")
		return
	}
	ssrvc.RegisterRoutes(router)
	gl.Log("info", "System service routes registered")

}

// SetSystemService allows setting the system service externally.
func SetSystemService(service services.ISystemService) {
	if service == nil {
		gl.Log("warn", "Attempted to set a nil system service")
		return
	}
	sysServ = service
}

// GetSystemService returns the current system service instance.
func GetSystemService() services.ISystemService {
	if sysServ == nil {
		gl.Log("warn", "System service is not initialized, creating a new instance")
		sysServ = services.NewSystemService()
	}
	return sysServ
}

func (c *MetricsController) SendMessage(ctx *gin.Context) {
	gl.Log("info", "Sending message via AMQP")

	var request struct {
		Exchange string                 `json:"exchange" binding:"required"`
		Key      string                 `json:"key" binding:"required"`
		Message  map[string]interface{} `json:"message" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		gl.Log("error", "Failed to bind send message request", err)
		c.apiWrapper.JSONResponseWithError(ctx, fmt.Errorf("invalid request format: %w", err))
		return
	}

	// Add timestamp to message
	request.Message["timestamp"] = time.Now().Unix()
	request.Message["source"] = "gobe-mcp"

	messageBody, err := json.Marshal(request.Message)
	if err != nil {
		gl.Log("error", "Failed to marshal message", err)
		c.apiWrapper.JSONResponseWithError(ctx, fmt.Errorf("failed to serialize message: %w", err))
		return
	}

	// Note: In a real implementation, you would use the AMQP connection here
	// For now, we'll simulate the message sending
	gl.Log("info", "Message would be sent to exchange", request.Exchange, "with key", request.Key)

	c.apiWrapper.JSONResponseWithSuccess(ctx, "message queued successfully", "", map[string]interface{}{
		"exchange":   request.Exchange,
		"key":        request.Key,
		"message_id": fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		"status":     "queued",
		"size_bytes": len(messageBody),
	})
}

func (c *MetricsController) SystemInfo(ctx *gin.Context) {
	gl.Log("info", "Getting system information")

	hostname, _ := os.Hostname()
	wd, _ := os.Getwd()

	systemInfo := map[string]interface{}{
		"hostname":      hostname,
		"working_dir":   wd,
		"go_version":    runtime.Version(),
		"go_os":         runtime.GOOS,
		"go_arch":       runtime.GOARCH,
		"num_cpu":       runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
		"pid":           os.Getpid(),
		"timestamp":     time.Now().Unix(),
	}

	c.apiWrapper.JSONResponseWithSuccess(ctx, "system info retrieved successfully", "", systemInfo)
}

func (c *MetricsController) ShellCommand(ctx *gin.Context) {
	gl.Log("info", "Executing shell command")

	var request struct {
		Command string `json:"command" binding:"required"`
		Text    string `json:"text"`    // raw text to parse command from
		UserID  string `json:"user_id"` // for audit
		Channel string `json:"channel"` // for audit
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		gl.Log("error", "Failed to bind shell command request", err)
		c.apiWrapper.JSONResponseWithError(ctx, fmt.Errorf("invalid request format: %w", err))
		return
	}

	// Initialize execsafe registry with allowed commands
	registry := c.getExecSafeRegistry()

	var parsed *execsafe.Parsed
	var err error

	// Parse command from text if provided, otherwise use direct command
	if request.Text != "" {
		parsed, err = execsafe.ParseUserCommand(request.Text)
	} else {
		// Direct command format for API calls
		parsed = &execsafe.Parsed{
			Name: request.Command,
			Args: []string{}, // no args for direct command
		}
	}

	if err != nil {
		gl.Log("warn", "Failed to parse command", err)
		c.apiWrapper.JSONResponseWithError(ctx, fmt.Errorf("command parsing failed: %w", err))
		return
	}

	// Execute command using execsafe
	start := time.Now()
	result, err := execsafe.RunSafe(ctx.Request.Context(), registry, parsed.Name, parsed.Args)
	duration := time.Since(start)

	// Prepare audit entry
	auditEntry := map[string]interface{}{
		"user_id":     request.UserID,
		"channel":     request.Channel,
		"command":     parsed.Name,
		"args_json":   toJSON(parsed.Args),
		"exit_code":   0,
		"duration_ms": duration.Milliseconds(),
		"stdout_len":  0,
		"stderr_len":  0,
		"truncated":   false,
		"created_at":  time.Now().Unix(),
	}

	response := map[string]interface{}{
		"command":   parsed.Name,
		"args":      parsed.Args,
		"duration":  duration.String(),
		"timestamp": time.Now().Unix(),
	}

	if err != nil {
		if result != nil {
			auditEntry["exit_code"] = result.ExitCode
			auditEntry["stdout_len"] = len(result.Stdout)
			auditEntry["stderr_len"] = len(result.Stderr)
			auditEntry["truncated"] = result.Truncated

			response["exit_code"] = result.ExitCode
			response["stdout"] = result.Stdout
			response["stderr"] = result.Stderr
			response["truncated"] = result.Truncated
		}

		gl.Log("warn", "Command execution failed", parsed.Name, err)
		response["error"] = err.Error()
		response["status"] = "failed"

		// Log audit entry
		c.logAuditEntry(auditEntry)

		c.apiWrapper.JSONResponseWithSuccess(ctx, "command executed with errors", "", response)
		return
	}

	// Success case
	auditEntry["exit_code"] = result.ExitCode
	auditEntry["stdout_len"] = len(result.Stdout)
	auditEntry["stderr_len"] = len(result.Stderr)
	auditEntry["truncated"] = result.Truncated

	response["exit_code"] = result.ExitCode
	response["stdout"] = result.Stdout
	response["stderr"] = result.Stderr
	response["truncated"] = result.Truncated
	response["status"] = "success"

	// Log audit entry
	c.logAuditEntry(auditEntry)

	c.apiWrapper.JSONResponseWithSuccess(ctx, "command executed successfully", "", response)
}

func (c *MetricsController) GetCPUInfo(ctx *gin.Context) {
	gl.Log("info", "Getting CPU information")

	cpuInfo := map[string]interface{}{
		"num_cpu":      runtime.NumCPU(),
		"go_max_procs": runtime.GOMAXPROCS(0),
		"goroutines":   runtime.NumGoroutine(),
		"architecture": runtime.GOARCH,
		"os":           runtime.GOOS,
		"go_version":   runtime.Version(),
		"timestamp":    time.Now().Unix(),
	}

	c.apiWrapper.JSONResponseWithSuccess(ctx, "CPU info retrieved successfully", "", cpuInfo)
}

func (c *MetricsController) GetMemoryInfo(ctx *gin.Context) {
	gl.Log("info", "Getting memory information")

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	memoryInfo := map[string]interface{}{
		"alloc_bytes":       memStats.Alloc,
		"alloc_mb":          float64(memStats.Alloc) / 1024 / 1024,
		"total_alloc_bytes": memStats.TotalAlloc,
		"total_alloc_mb":    float64(memStats.TotalAlloc) / 1024 / 1024,
		"sys_bytes":         memStats.Sys,
		"sys_mb":            float64(memStats.Sys) / 1024 / 1024,
		"num_gc":            memStats.NumGC,
		"gc_cpu_fraction":   memStats.GCCPUFraction,
		"heap_alloc_bytes":  memStats.HeapAlloc,
		"heap_alloc_mb":     float64(memStats.HeapAlloc) / 1024 / 1024,
		"heap_sys_bytes":    memStats.HeapSys,
		"heap_sys_mb":       float64(memStats.HeapSys) / 1024 / 1024,
		"timestamp":         time.Now().Unix(),
	}

	c.apiWrapper.JSONResponseWithSuccess(ctx, "memory info retrieved successfully", "", memoryInfo)
}

func (c *MetricsController) GetDiskInfo(ctx *gin.Context) {
	gl.Log("info", "Getting disk information")

	diskInfo := map[string]interface{}{
		"timestamp": time.Now().Unix(),
	}

	// Get disk usage for current working directory
	wd, err := os.Getwd()
	if err == nil {
		var stat syscall.Statfs_t
		err := syscall.Statfs(wd, &stat)
		if err == nil {
			totalBytes := stat.Blocks * uint64(stat.Bsize)
			freeBytes := stat.Bavail * uint64(stat.Bsize)
			usedBytes := totalBytes - freeBytes

			diskInfo["working_directory"] = map[string]interface{}{
				"path":          wd,
				"total_bytes":   totalBytes,
				"total_gb":      float64(totalBytes) / 1024 / 1024 / 1024,
				"free_bytes":    freeBytes,
				"free_gb":       float64(freeBytes) / 1024 / 1024 / 1024,
				"used_bytes":    usedBytes,
				"used_gb":       float64(usedBytes) / 1024 / 1024 / 1024,
				"usage_percent": float64(usedBytes) / float64(totalBytes) * 100,
			}
		} else {
			diskInfo["error"] = "Failed to get disk stats: " + err.Error()
		}
	} else {
		diskInfo["error"] = "Failed to get working directory: " + err.Error()
	}

	c.apiWrapper.JSONResponseWithSuccess(ctx, "disk info retrieved successfully", "", diskInfo)
}

func (c *MetricsController) RegisterTools(ctx *gin.Context) {
	gl.Log("info", "Registering new MCP tool")

	var request struct {
		Name        string                 `json:"name" binding:"required"`
		Title       string                 `json:"title" binding:"required"`
		Description string                 `json:"description" binding:"required"`
		Auth        string                 `json:"auth"`
		Args        map[string]interface{} `json:"args"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		gl.Log("error", "Failed to bind register tool request", err)
		c.apiWrapper.JSONResponseWithError(ctx, fmt.Errorf("invalid request format: %w", err))
		return
	}

	// Note: For security reasons, we cannot allow arbitrary tool registration with custom handlers
	// This endpoint would be used for registering metadata of external tools
	gl.Log("info", "Tool registration requested", request.Name, request.Title)

	c.apiWrapper.JSONResponseWithSuccess(ctx, "tool registration completed", "", map[string]interface{}{
		"name":        request.Name,
		"title":       request.Title,
		"description": request.Description,
		"status":      "registered_metadata_only",
		"note":        "Handler execution requires server restart for security",
		"timestamp":   time.Now().Unix(),
	})
}

func (c *MetricsController) RegisterResources(ctx *gin.Context) {
	gl.Log("info", "Registering MCP resources")

	var request struct {
		Resources []map[string]interface{} `json:"resources" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		gl.Log("error", "Failed to bind register resources request", err)
		c.apiWrapper.JSONResponseWithError(ctx, fmt.Errorf("invalid request format: %w", err))
		return
	}

	registeredResources := make([]map[string]interface{}, 0)
	for _, resource := range request.Resources {
		if name, ok := resource["name"].(string); ok {
			gl.Log("info", "Registering resource", name)
			resource["status"] = "registered"
			resource["timestamp"] = time.Now().Unix()
			registeredResources = append(registeredResources, resource)
		}
	}

	c.apiWrapper.JSONResponseWithSuccess(ctx, "resources registered successfully", "", map[string]interface{}{
		"registered_count": len(registeredResources),
		"resources":        registeredResources,
		"timestamp":        time.Now().Unix(),
	})
}

func (c *MetricsController) HandleAnalyzeMessage(ctx *gin.Context) {
	gl.Log("info", "Analyzing message")

	var request struct {
		Message string                 `json:"message" binding:"required"`
		Options map[string]interface{} `json:"options"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		gl.Log("error", "Failed to bind analyze message request", err)
		c.apiWrapper.JSONResponseWithError(ctx, fmt.Errorf("invalid request format: %w", err))
		return
	}

	// Basic message analysis
	analysis := map[string]interface{}{
		"message":         request.Message,
		"length":          len(request.Message),
		"word_count":      len(strings.Fields(request.Message)),
		"character_count": len([]rune(request.Message)),
		"has_uppercase":   strings.ToUpper(request.Message) != request.Message && strings.ToLower(request.Message) != request.Message,
		"has_numbers":     containsNumbers(request.Message),
		"has_special":     containsSpecialChars(request.Message),
		"timestamp":       time.Now().Unix(),
	}

	// Add sentiment placeholder (would integrate with actual sentiment analysis)
	analysis["sentiment"] = map[string]interface{}{
		"score":      0.0,
		"label":      "neutral",
		"confidence": 0.5,
	}

	c.apiWrapper.JSONResponseWithSuccess(ctx, "message analyzed successfully", "", analysis)
}

func (c *MetricsController) HandleCreateTask(ctx *gin.Context) {
	gl.Log("info", "Creating new task")

	var request struct {
		Title       string                 `json:"title" binding:"required"`
		Description string                 `json:"description"`
		Priority    string                 `json:"priority"` // low, medium, high
		DueDate     string                 `json:"due_date"` // ISO format
		Tags        []string               `json:"tags"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		gl.Log("error", "Failed to bind create task request", err)
		c.apiWrapper.JSONResponseWithError(ctx, fmt.Errorf("invalid request format: %w", err))
		return
	}

	// Validate priority
	validPriorities := []string{"low", "medium", "high"}
	if request.Priority == "" {
		request.Priority = "medium"
	} else {
		priorityValid := false
		for _, validPriority := range validPriorities {
			if request.Priority == validPriority {
				priorityValid = true
				break
			}
		}
		if !priorityValid {
			request.Priority = "medium"
		}
	}

	// Generate task ID
	taskID := fmt.Sprintf("task_%d", time.Now().UnixNano())

	task := map[string]interface{}{
		"id":          taskID,
		"title":       request.Title,
		"description": request.Description,
		"priority":    request.Priority,
		"due_date":    request.DueDate,
		"tags":        request.Tags,
		"metadata":    request.Metadata,
		"status":      "created",
		"created_at":  time.Now().Unix(),
		"updated_at":  time.Now().Unix(),
	}

	// Note: In a real implementation, this would be stored in a database
	gl.Log("info", "Task created", taskID, request.Title)

	c.apiWrapper.JSONResponseWithSuccess(ctx, "task created successfully", "", task)
}

// ListTools returns all registered MCP tools
func (c *MetricsController) ListTools(ctx *gin.Context) {
	if c.registry == nil {
		gl.Log("error", "MCP registry is not initialized")
		c.apiWrapper.JSONResponseWithError(ctx, fmt.Errorf("registry not available"))
		return
	}

	tools := c.registry.List()

	c.apiWrapper.JSONResponseWithSuccess(ctx, "tools listed successfully", "", map[string]interface{}{
		"tools": tools,
		"count": len(tools),
	})
}

// ExecTool executes an MCP tool by name
func (c *MetricsController) ExecTool(ctx *gin.Context) {
	if c.registry == nil {
		gl.Log("error", "MCP registry is not initialized")
		c.apiWrapper.JSONResponseWithError(ctx, fmt.Errorf("registry not available"))
		return
	}

	var request struct {
		Tool string                 `json:"tool" binding:"required"`
		Args map[string]interface{} `json:"args"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		gl.Log("error", "Failed to bind exec request", err)
		c.apiWrapper.JSONResponseWithError(ctx, fmt.Errorf("invalid request format: %w", err))
		return
	}

	if request.Args == nil {
		request.Args = make(map[string]interface{})
	}

	result, err := c.registry.Exec(ctx.Request.Context(), request.Tool, request.Args)
	if err != nil {
		gl.Log("error", "Tool execution failed", request.Tool, err)
		c.apiWrapper.JSONResponseWithError(ctx, fmt.Errorf("tool execution failed: %w", err))
		return
	}

	c.apiWrapper.JSONResponseWithSuccess(ctx, "tool executed successfully", "", map[string]interface{}{
		"tool":   request.Tool,
		"result": result,
	})
}

// Helper functions
func containsNumbers(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func containsSpecialChars(s string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;':,.<>?"
	for _, r := range s {
		for _, special := range specialChars {
			if r == special {
				return true
			}
		}
	}
	return false
}

// getExecSafeRegistry initializes and returns an execsafe registry with allowed commands
func (c *MetricsController) getExecSafeRegistry() *execsafe.Registry {
	registry := execsafe.NewRegistry()

	// Basic file operations
	registry.Register("ls", execsafe.CommandSpec{
		Binary:       "ls",
		Timeout:      5 * time.Second,
		MaxOutputKB:  256,
		ArgsValidate: execsafe.OneOfFlags("-la", "-l", "-a", "-h", "--help"),
	})

	registry.Register("pwd", execsafe.CommandSpec{
		Binary:      "pwd",
		Timeout:     3 * time.Second,
		MaxOutputKB: 64,
	})

	// System information
	registry.Register("date", execsafe.CommandSpec{
		Binary:      "date",
		Timeout:     3 * time.Second,
		MaxOutputKB: 64,
	})

	registry.Register("whoami", execsafe.CommandSpec{
		Binary:      "whoami",
		Timeout:     3 * time.Second,
		MaxOutputKB: 64,
	})

	registry.Register("uname", execsafe.CommandSpec{
		Binary:       "uname",
		Timeout:      3 * time.Second,
		MaxOutputKB:  256,
		ArgsValidate: execsafe.OneOfFlags("-a", "-r", "-v", "-m", "-s"),
	})

	// Process information
	registry.Register("ps", execsafe.CommandSpec{
		Binary:       "ps",
		Timeout:      5 * time.Second,
		MaxOutputKB:  512,
		ArgsValidate: execsafe.OneOfFlags("-aux", "-ef", "-u", "-f"),
	})

	// Memory and disk
	registry.Register("free", execsafe.CommandSpec{
		Binary:       "free",
		Timeout:      3 * time.Second,
		MaxOutputKB:  64,
		ArgsValidate: execsafe.OneOfFlags("-h", "-m", "-g"),
	})

	registry.Register("df", execsafe.CommandSpec{
		Binary:       "df",
		Timeout:      5 * time.Second,
		MaxOutputKB:  256,
		ArgsValidate: execsafe.OneOfFlags("-h", "-k", "-m"),
	})

	// System status
	registry.Register("uptime", execsafe.CommandSpec{
		Binary:      "uptime",
		Timeout:     3 * time.Second,
		MaxOutputKB: 64,
	})

	return registry
}

// toJSON helper function to convert slice to JSON string
func toJSON(v interface{}) string {
	if v == nil {
		return "[]"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// logAuditEntry logs audit information to the logger (placeholder for DB)
func (c *MetricsController) logAuditEntry(entry map[string]interface{}) {
	// TODO: Store in database when audit table is available
	gl.Log("info", "AUDIT: cmd=%s user=%s channel=%s exit=%v duration=%vms",
		entry["command"],
		entry["user_id"],
		entry["channel"],
		entry["exit_code"],
		entry["duration_ms"])
}
