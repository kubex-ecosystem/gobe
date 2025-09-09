// Package system provides the controller for managing mcp system-level operations.
package system

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	services "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	"github.com/kubex-ecosystem/gobe/internal/module/logger"
	"github.com/kubex-ecosystem/gobe/internal/services/mcp/hooks"
	"github.com/kubex-ecosystem/gobe/internal/services/mcp/system"
	"gorm.io/gorm"

	l "github.com/kubex-ecosystem/logz"
)

var (
	gl      = logger.GetLogger[l.Logger](nil)
	sysServ services.ISystemService
)

type MetricsController struct {
	dbConn        *gorm.DB
	mcpState      *hooks.Bitstate[uint64, system.SystemDomain]
	systemService services.ISystemService
}

func NewMetricsController(db *gorm.DB) *MetricsController {
	if db == nil {
		// gl.Log("error", "Database connection is nil")
		gl.Log("warn", "Database connection is nil")
		// return nil
	}

	// We allow the system service to be nil, as it can be set later.
	return &MetricsController{
		dbConn:        db,
		systemService: sysServ,
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
	// Placeholder for message sending logic
	ctx.JSON(http.StatusOK, gin.H{"message": "SendMessage endpoint not implemented"})
}

func (c *MetricsController) SystemInfo(ctx *gin.Context) {
	// Placeholder for system info retrieval logic
	ctx.JSON(http.StatusOK, gin.H{"message": "SystemInfo endpoint not implemented"})
}

func (c *MetricsController) ShellCommand(ctx *gin.Context) {
	// Placeholder for shell command execution logic
	ctx.JSON(http.StatusOK, gin.H{"message": "ShellCommand endpoint not implemented"})
}

func (c *MetricsController) GetCPUInfo(ctx *gin.Context) {
	// Placeholder for CPU info retrieval logic
	ctx.JSON(http.StatusOK, gin.H{"message": "GetCPUInfo endpoint not implemented"})
}

func (c *MetricsController) GetMemoryInfo(ctx *gin.Context) {
	// Placeholder for memory info retrieval logic
	ctx.JSON(http.StatusOK, gin.H{"message": "GetMemoryInfo endpoint not implemented"})
}

func (c *MetricsController) GetDiskInfo(ctx *gin.Context) {
	// Placeholder for disk info retrieval logic
	ctx.JSON(http.StatusOK, gin.H{"message": "GetDiskInfo endpoint not implemented"})
}

func (c *MetricsController) RegisterTools(ctx *gin.Context) {
	// Placeholder for tool registration logic
	ctx.JSON(http.StatusOK, gin.H{"message": "RegisterTools endpoint not implemented"})
}

func (c *MetricsController) RegisterResources(ctx *gin.Context) {
	// Placeholder for resource registration logic
	ctx.JSON(http.StatusOK, gin.H{"message": "RegisterResources endpoint not implemented"})
}

func (c *MetricsController) HandleAnalyzeMessage(ctx *gin.Context) {
	// Placeholder for handling analyze message logic
	ctx.JSON(http.StatusOK, gin.H{"message": "HandleAnalyzeMessage endpoint not implemented"})
}

func (c *MetricsController) HandleCreateTask(ctx *gin.Context) {
	// Placeholder for task creation logic
	ctx.JSON(http.StatusOK, gin.H{"message": "HandleCreateTask endpoint not implemented"})
}
