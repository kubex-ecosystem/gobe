package services

import (
	"fmt"
	"net/http"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// All this file structures need to be migrated to the GDBase module later.
// GDBase is a module that provides all data modeling, persistence, and retrieval functionalities.
// It is a core module for Data Services, Business Logic, and this kind of stuff.

type SystemMetrics struct {
	CPU         CPUMetrics     `json:"cpu"`
	Memory      MemoryMetrics  `json:"memory"`
	Disk        DiskMetrics    `json:"disk"`
	Network     NetworkMetrics `json:"network"`
	Uptime      int64          `json:"uptime"`
	LoadAverage []float64      `json:"loadAverage"`
	Processes   int            `json:"processes"`
}

type CPUMetrics struct {
	Usage       float64 `json:"usage"`
	Cores       int     `json:"cores"`
	Temperature float64 `json:"temperature,omitempty"`
}

type MemoryMetrics struct {
	Used       float64 `json:"used"`
	Total      float64 `json:"total"`
	Percentage float64 `json:"percentage"`
}

type DiskMetrics struct {
	Used       float64 `json:"used"`
	Total      float64 `json:"total"`
	Percentage float64 `json:"percentage"`
}

type NetworkMetrics struct {
	BytesIn    uint64 `json:"bytesIn"`
	BytesOut   uint64 `json:"bytesOut"`
	PacketsIn  uint64 `json:"packetsIn"`
	PacketsOut uint64 `json:"packetsOut"`
}

type ISystemService interface {
	GetCurrentMetrics() (*SystemMetrics, error)
}

type SystemService struct{}

func NewSystemService() ISystemService {
	return &SystemService{}
}

func (s *SystemService) GetCurrentMetrics() (*SystemMetrics, error) {
	// CPU Metrics
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU metrics: %w", err)
	}

	cpuCount := runtime.NumCPU()

	// Memory Metrics
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory metrics: %w", err)
	}

	// Disk Metrics (root partition)
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk metrics: %w", err)
	}

	// Network Metrics
	netIO, err := net.IOCounters(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get network metrics: %w", err)
	}

	// Load Average
	loadInfo, err := load.Avg()
	if err != nil {
		return nil, fmt.Errorf("failed to get load average: %w", err)
	}

	// Process count
	processes, err := process.Pids()
	if err != nil {
		return nil, fmt.Errorf("failed to get process count: %w", err)
	}

	// System uptime
	var sysinfo syscall.Sysinfo_t
	syscall.Syscall(syscall.SYS_SYSINFO, uintptr(unsafe.Pointer(&sysinfo)), 0, 0)

	metrics := &SystemMetrics{
		CPU: CPUMetrics{
			Usage: cpuPercent[0],
			Cores: cpuCount,
		},
		Memory: MemoryMetrics{
			Used:       float64(memInfo.Used) / 1024 / 1024 / 1024,  // GB
			Total:      float64(memInfo.Total) / 1024 / 1024 / 1024, // GB
			Percentage: memInfo.UsedPercent,
		},
		Disk: DiskMetrics{
			Used:       float64(diskInfo.Used) / 1024 / 1024 / 1024,  // GB
			Total:      float64(diskInfo.Total) / 1024 / 1024 / 1024, // GB
			Percentage: diskInfo.UsedPercent,
		},
		Network: NetworkMetrics{
			BytesIn:    netIO[0].BytesRecv,
			BytesOut:   netIO[0].BytesSent,
			PacketsIn:  netIO[0].PacketsRecv,
			PacketsOut: netIO[0].PacketsSent,
		},
		Uptime:      int64(sysinfo.Uptime),
		LoadAverage: []float64{loadInfo.Load1, loadInfo.Load5, loadInfo.Load15},
		Processes:   len(processes),
	}

	return metrics, nil
}

func (s *SystemService) RegisterRoutes(routerGroup gin.IRoutes) {
	routerGroup.GET("/system/metrics", func(c *gin.Context) {
		metrics, err := s.GetCurrentMetrics()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, metrics)
	})
}
