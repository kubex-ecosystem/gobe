// Package gdbase provides controllers for managing GDBase operations including Cloudflare tunneling.
package gdbase

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	"github.com/kubex-ecosystem/gobe/internal/module/logger"
	"github.com/kubex-ecosystem/gobe/internal/services/mcp/hooks"
	"github.com/kubex-ecosystem/gobe/internal/services/mcp/system"
	"gorm.io/gorm"

	l "github.com/kubex-ecosystem/logz"
)

var (
	gl = logger.GetLogger[l.Logger](nil)
)

// TunnelStatus represents the current tunnel state
type TunnelStatus struct {
	Mode    gdbasez.TunnelMode `json:"mode"`
	Public  string             `json:"public"`
	Running bool               `json:"running"`
	Network string             `json:"network,omitempty"`
	Target  string             `json:"target,omitempty"`
}

// TunnelRequest represents the request payload for tunnel operations
type TunnelRequest struct {
	Mode    string `json:"mode" binding:"required"`
	Network string `json:"network,omitempty"`
	Target  string `json:"target,omitempty"`
	Port    int    `json:"port,omitempty"`
	Token   string `json:"token,omitempty"`
	Timeout string `json:"timeout,omitempty"` // "10s"
}

// GDBaseController handles GDBase tunnel operations
type GDBaseController struct {
	dbConn       *gorm.DB
	mcpState     *hooks.Bitstate[uint64, system.SystemDomain]
	dockerCli    *client.Client
	tunnelState  *TunnelStatus
	tunnelMutex  sync.RWMutex
	activeHandle gdbasez.TunnelHandle
}

// NewGDBaseController creates a new GDBaseController instance
func NewGDBaseController(db *gorm.DB) *GDBaseController {
	if db == nil {
		gl.Log("warn", "Database connection is nil")
	}

	// Initialize Docker client
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		gl.Log("error", "Failed to create Docker client", err)
		dockerCli = nil
	}

	return &GDBaseController{
		dbConn:    db,
		dockerCli: dockerCli,
		tunnelState: &TunnelStatus{
			Running: false,
		},
	}
}

// PostGDBaseTunnelUp handles tunnel creation requests
func (g *GDBaseController) PostGDBaseTunnelUp(c *gin.Context) {
	if g.dockerCli == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Docker client not available",
		})
		return
	}

	var req TunnelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid JSON: " + err.Error(),
		})
		return
	}

	g.tunnelMutex.Lock()
	defer g.tunnelMutex.Unlock()

	// Check if tunnel is already running
	if g.tunnelState.Running {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Conflict",
			"message": "Tunnel is already running",
			"current": g.tunnelState,
		})
		return
	}

	mode := gdbasez.TunnelMode(req.Mode)
	ctx := c.Request.Context()

	switch mode {
	case gdbasez.TunnelQuick:
		if err := g.handleQuickTunnel(ctx, &req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"mode":    mode,
			"public":  g.tunnelState.Public,
			"network": req.Network,
			"target":  req.Target + ":" + strconv.Itoa(req.Port),
			"running": true,
		})

	case gdbasez.TunnelNamed:
		if err := g.handleNamedTunnel(ctx, &req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"mode":    mode,
			"public":  "Use your configured tunnel hostnames",
			"network": req.Network,
			"running": true,
		})

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Mode must be 'quick' or 'named'",
		})
	}
}

// PostGDBaseTunnelDown handles tunnel termination requests
func (g *GDBaseController) PostGDBaseTunnelDown(c *gin.Context) {
	g.tunnelMutex.Lock()
	defer g.tunnelMutex.Unlock()

	if !g.tunnelState.Running || g.activeHandle == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "No active tunnel to stop",
		})
		return
	}

	ctx := c.Request.Context()
	if err := g.activeHandle.Stop(ctx); err != nil {
		gl.Log("error", "Failed to stop tunnel", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to stop tunnel: " + err.Error(),
		})
		return
	}

	// Reset tunnel state
	g.tunnelState = &TunnelStatus{Running: false}
	g.activeHandle = nil

	gl.Log("info", "Tunnel stopped successfully")
	c.Status(http.StatusNoContent)
}

// GetGDBaseTunnelStatus returns the current tunnel status
func (g *GDBaseController) GetGDBaseTunnelStatus(c *gin.Context) {
	g.tunnelMutex.RLock()
	defer g.tunnelMutex.RUnlock()

	c.JSON(http.StatusOK, g.tunnelState)
}

// handleQuickTunnel creates a quick tunnel
func (g *GDBaseController) handleQuickTunnel(ctx context.Context, req *TunnelRequest) error {
	// Validate required fields for quick mode
	if req.Target == "" || req.Port <= 0 {
		return fmt.Errorf("quick mode requires target and port")
	}

	// Parse timeout (currently not used but could be implemented later)
	// timeout := 10 * time.Second
	// if req.Timeout != "" {
	// 	if d, err := time.ParseDuration(req.Timeout); err == nil && d > 0 {
	// 		timeout = d
	// 	}
	// }

	// Set default network if not provided
	networkName := req.Network
	if networkName == "" {
		networkName = "gdbase_net"
	}

	// Create tunnel options
	opts := gdbasez.NewCloudflaredOpts(
		gdbasez.TunnelQuick,
		networkName,
		req.Target,
		req.Port,
		"", // no token for quick mode
	)

	// Start tunnel
	handle, publicURL, err := opts.Start(ctx, g.dockerCli)
	if err != nil {
		return fmt.Errorf("failed to start quick tunnel: %w", err)
	}

	// Update state
	g.activeHandle = handle
	g.tunnelState = &TunnelStatus{
		Mode:    gdbasez.TunnelQuick,
		Public:  publicURL,
		Running: true,
		Network: networkName,
		Target:  req.Target + ":" + strconv.Itoa(req.Port),
	}

	gl.Log("info", "Quick tunnel started successfully", "url", publicURL)
	return nil
}

// handleNamedTunnel creates a named tunnel
func (g *GDBaseController) handleNamedTunnel(ctx context.Context, req *TunnelRequest) error {
	// Validate required fields for named mode
	if req.Token == "" {
		return fmt.Errorf("named mode requires token")
	}

	// Set default network if not provided
	networkName := req.Network
	if networkName == "" {
		networkName = "gdbase_net"
	}

	// Create tunnel options
	opts := gdbasez.NewCloudflaredOpts(
		gdbasez.TunnelNamed,
		networkName,
		"", // no target for named mode
		0,  // no port for named mode
		req.Token,
	)

	// Start tunnel
	handle, _, err := opts.Start(ctx, g.dockerCli)
	if err != nil {
		return fmt.Errorf("failed to start named tunnel: %w", err)
	}

	// Update state
	g.activeHandle = handle
	g.tunnelState = &TunnelStatus{
		Mode:    gdbasez.TunnelNamed,
		Public:  "Use your configured tunnel hostnames",
		Running: true,
		Network: networkName,
	}

	gl.Log("info", "Named tunnel started successfully")
	return nil
}
