// Package system provides the controller for managing mcp system-level operations.
package system

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rafa-mori/gobe/internal/services"
	"github.com/rafa-mori/gobe/logger"
	"gorm.io/gorm"

	l "github.com/rafa-mori/logz"
)

var (
	gl      = logger.GetLogger[l.Logger](nil)
	sysServ services.ISystemService
)

type MetricsController struct {
	dbConn        *gorm.DB
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
