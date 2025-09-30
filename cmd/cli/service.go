package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	gb "github.com/kubex-ecosystem/gobe"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	"github.com/kubex-ecosystem/gobe/internal/services/mcp"
	l "github.com/kubex-ecosystem/logz"
	"github.com/spf13/cobra"
)

func ServiceCmd() *cobra.Command {
	shortDesc := "Service management commands"
	longDesc := "Service management commands for GoBE or any other service"
	serviceCmd := &cobra.Command{
		Use:         "service",
		Short:       shortDesc,
		Long:        longDesc,
		Aliases:     []string{"svc", "serv", "backend", "server"},
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				gl.Log("error", "Failed to display help: ", err.Error())
			}
		},
	}

	serviceCmd.AddCommand([]*cobra.Command{
		startCommand(),
		stopCommand(),
		restartCommand(),
		statusCommand(),
		logsCommand(),
	}...)

	return serviceCmd
}

func startCommand() *cobra.Command {
	var name, port, bind, logFile, configFile string
	var isConfidential, debug, releaseMode bool

	shortDesc := "Start a minimal backend service"
	longDesc := "Start a minimal backend service with GoBE"

	var startCmd = &cobra.Command{
		Use:         "start",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				gl.SetDebug(true)
			}

			gbm, gbmErr := gb.NewGoBE(name, port, bind, logFile, configFile, isConfidential, l.GetLogger("GoBE"), debug, releaseMode)
			if gbmErr != nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", gbmErr.Error())
				return
			}
			if gbm == nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", "GoBE instance is nil")
				return
			}
			gbm.StartGoBE()
			gl.Log("success", "GoBE started successfully")
		},
	}

	startCmd.Flags().StringVarP(&name, "name", "n", "GoBE", "Name of the process")
	startCmd.Flags().StringVarP(&port, "port", "p", "8666", "Port to listen on")
	startCmd.Flags().StringVarP(&bind, "bind", "b", "0.0.0.0", "Bind address")
	startCmd.Flags().StringVarP(&logFile, "log-file", "l", "", "Log file path")
	startCmd.Flags().StringVarP(&configFile, "config-file", "c", "", "Configuration file path")
	startCmd.Flags().BoolVarP(&isConfidential, "confidential", "C", false, "Enable confidential mode")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	startCmd.Flags().BoolVarP(&releaseMode, "release", "r", false, "Enable release mode")

	return startCmd
}

func stopCommand() *cobra.Command {
	var name string
	var force, graceful bool
	var timeout int

	shortDesc := "Stop a running backend service"
	longDesc := "Stop a running backend service with GoBE with graceful shutdown support"

	var stopCmd = &cobra.Command{
		Use:         "stop",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			stopService(name, force, graceful, timeout)
		},
	}

	stopCmd.Flags().StringVarP(&name, "name", "n", "GoBE", "Name of the process")
	stopCmd.Flags().BoolVarP(&force, "force", "f", false, "Force stop the service")
	stopCmd.Flags().BoolVarP(&graceful, "graceful", "g", true, "Graceful shutdown")
	stopCmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "Timeout in seconds for graceful shutdown")

	return stopCmd
}

func restartCommand() *cobra.Command {
	var name string

	shortDesc := "Restart a running backend service"
	longDesc := "Restart a running backend service with GoBE"

	var restartCmd = &cobra.Command{
		Use:         "restart",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			gbm, gbmErr := gb.NewGoBE(name, "", "", "", "", false, l.GetLogger("GoBE"), false, false)
			if gbmErr != nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", gbmErr.Error())
				return
			}
			if gbm == nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", "GoBE instance is nil")
				return
			}
			gbm.StopGoBE()
			gl.Log("success", "GoBE stopped successfully")
			gbm.StartGoBE()
			gl.Log("success", "GoBE started successfully")
		},
	}

	restartCmd.Flags().StringVarP(&name, "name", "n", "GoBE", "Name of the process")

	return restartCmd
}

func statusCommand() *cobra.Command {
	var name, format string
	var detailed, json bool

	shortDesc := "Get the status of a running backend service"
	longDesc := "Get the status of a running backend service with GoBE including health checks, MCP tools, and system metrics"

	var statusCmd = &cobra.Command{
		Use:         "status",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			getServiceStatus(name, format, detailed, json)
		},
	}

	statusCmd.Flags().StringVarP(&name, "name", "n", "GoBE", "Name of the process")
	statusCmd.Flags().StringVarP(&format, "format", "f", "table", "Output format: table, json, yaml")
	statusCmd.Flags().BoolVarP(&detailed, "detailed", "v", false, "Show detailed status information")
	statusCmd.Flags().BoolVarP(&json, "json", "j", false, "Output in JSON format")

	return statusCmd
}

func logsCommand() *cobra.Command {
	var name, level, format string
	var follow, timestamps bool
	var lines, tail int

	shortDesc := "Get the logs of a running backend service"
	longDesc := "Get the logs of a running backend service with GoBE with filtering and tail support"

	var logsCmd = &cobra.Command{
		Use:         "logs",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			getLogs(name, level, format, follow, timestamps, lines, tail)
		},
	}

	logsCmd.Flags().StringVarP(&name, "name", "n", "GoBE", "Name of the process")
	logsCmd.Flags().StringVarP(&level, "level", "l", "", "Filter by log level (debug, info, warn, error)")
	logsCmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text, json")
	logsCmd.Flags().BoolVarP(&follow, "follow", "F", false, "Follow log output (like tail -f)")
	logsCmd.Flags().BoolVarP(&timestamps, "timestamps", "t", true, "Show timestamps")
	logsCmd.Flags().IntVar(&lines, "lines", 100, "Number of lines to show")
	logsCmd.Flags().IntVar(&tail, "tail", 0, "Show last N lines (0 = all)")

	return logsCmd
}

// ServiceStatus represents the current status of the GoBE service
type ServiceStatus struct {
	Name         string                 `json:"name"`
	Status       string                 `json:"status"`
	Uptime       string                 `json:"uptime,omitempty"`
	Version      string                 `json:"version"`
	Port         string                 `json:"port,omitempty"`
	HealthChecks map[string]string      `json:"health_checks"`
	MCPTools     []string               `json:"mcp_tools,omitempty"`
	SystemInfo   map[string]interface{} `json:"system_info"`
	LastCheck    time.Time              `json:"last_check"`
	ResponseTime string                 `json:"response_time,omitempty"`
}

func getServiceStatus(name, format string, detailed, jsonOutput bool) {
	gl.Log("info", fmt.Sprintf("Checking status for service: %s", name))

	status := &ServiceStatus{
		Name:         name,
		Status:       "unknown",
		Version:      getVersion(),
		HealthChecks: make(map[string]string),
		SystemInfo:   make(map[string]interface{}),
		LastCheck:    time.Now(),
	}

	// Try to connect to running service
	if isServiceRunning(name) {
		status.Status = "running"
		status.Port = getServicePort()
		status.Uptime = getServiceUptime()
		status.ResponseTime = checkServiceResponse()

		// Get health checks
		status.HealthChecks["api"] = checkAPIHealth()
		status.HealthChecks["database"] = checkDatabaseHealth()
		status.HealthChecks["mcp"] = checkMCPHealth()

		// Get MCP tools if detailed
		if detailed {
			status.MCPTools = getMCPTools()
		}
	} else {
		status.Status = "stopped"
		status.HealthChecks["service"] = "not running"
	}

	// Add system info
	status.SystemInfo["go_version"] = runtime.Version()
	status.SystemInfo["platform"] = runtime.GOOS + "/" + runtime.GOARCH
	status.SystemInfo["goroutines"] = runtime.NumGoroutine()

	if detailed {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		status.SystemInfo["memory_alloc"] = fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024)
		status.SystemInfo["memory_sys"] = fmt.Sprintf("%.2f MB", float64(m.Sys)/1024/1024)
	}

	// Output based on format
	if jsonOutput || format == "json" {
		outputJSON(status)
	} else if format == "yaml" {
		outputYAML(status)
	} else {
		outputTable(status, detailed)
	}
}

func isServiceRunning(name string) bool {
	// Try to check if service is running by attempting connection
	port := getServicePort()
	if port == "" {
		return false
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%s/health", port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func getServicePort() string {
	// Try to get port from config or environment
	if port := os.Getenv("GOBE_PORT"); port != "" {
		return port
	}
	return "8666" // Default port
}

func getServiceUptime() string {
	// Simple implementation - in real scenario would track start time
	return "unknown"
}

func checkServiceResponse() string {
	port := getServicePort()
	start := time.Now()

	client := &http.Client{Timeout: 5 * time.Second}
	_, err := client.Get(fmt.Sprintf("http://localhost:%s/health", port))

	duration := time.Since(start)
	if err != nil {
		return "timeout"
	}

	return duration.String()
}

func checkAPIHealth() string {
	port := getServicePort()
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(fmt.Sprintf("http://localhost:%s/health", port))
	if err != nil {
		return "unhealthy"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "healthy"
	}
	return "unhealthy"
}

func checkDatabaseHealth() string {
	// This would need actual database connection check
	// For now, return basic status
	return "unknown"
}

func checkMCPHealth() string {
	// Try to create MCP registry and check if tools are registered
	registry := mcp.NewRegistry()
	tools := registry.List()

	if len(tools) > 0 {
		return "healthy"
	}
	return "no tools registered"
}

func getMCPTools() []string {
	registry := mcp.NewRegistry()

	// Register builtin tools to get the list
	err := mcp.RegisterBuiltinTools(registry)
	if err != nil {
		return []string{"error getting tools"}
	}

	tools := registry.List()
	var toolNames []string
	for _, tool := range tools {
		toolNames = append(toolNames, tool.Name)
	}

	return toolNames
}

func getVersion() string {
	// In real implementation, this would come from build info
	return "1.3.5"
}

func outputJSON(status *ServiceStatus) {
	jsonData, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error marshaling JSON: %v", err))
		return
	}

	fmt.Println(string(jsonData))
}

func outputYAML(status *ServiceStatus) {
	// Simple YAML-like output
	fmt.Printf("name: %s\n", status.Name)
	fmt.Printf("status: %s\n", status.Status)
	fmt.Printf("version: %s\n", status.Version)
	if status.Port != "" {
		fmt.Printf("port: %s\n", status.Port)
	}
	if status.Uptime != "" {
		fmt.Printf("uptime: %s\n", status.Uptime)
	}
	if status.ResponseTime != "" {
		fmt.Printf("response_time: %s\n", status.ResponseTime)
	}

	fmt.Println("health_checks:")
	for k, v := range status.HealthChecks {
		fmt.Printf("  %s: %s\n", k, v)
	}

	if len(status.MCPTools) > 0 {
		fmt.Println("mcp_tools:")
		for _, tool := range status.MCPTools {
			fmt.Printf("  - %s\n", tool)
		}
	}

	fmt.Println("system_info:")
	for k, v := range status.SystemInfo {
		fmt.Printf("  %s: %v\n", k, v)
	}

	fmt.Printf("last_check: %s\n", status.LastCheck.Format(time.RFC3339))
}

func outputTable(status *ServiceStatus, detailed bool) {
	fmt.Println("🚀 GoBE Service Status")
	fmt.Println("=" + fmt.Sprintf("%*s", 50, ""))

	fmt.Printf("%-15s: %s\n", "Name", status.Name)
	fmt.Printf("%-15s: %s\n", "Status", getStatusIcon(status.Status)+status.Status)
	fmt.Printf("%-15s: %s\n", "Version", status.Version)

	if status.Port != "" {
		fmt.Printf("%-15s: %s\n", "Port", status.Port)
	}
	if status.Uptime != "" {
		fmt.Printf("%-15s: %s\n", "Uptime", status.Uptime)
	}
	if status.ResponseTime != "" {
		fmt.Printf("%-15s: %s\n", "Response Time", status.ResponseTime)
	}

	fmt.Println("\n🏥 Health Checks:")
	for check, result := range status.HealthChecks {
		fmt.Printf("  %-12s: %s%s\n", check, getHealthIcon(result), result)
	}

	if len(status.MCPTools) > 0 {
		fmt.Println("\n🔧 MCP Tools:")
		for _, tool := range status.MCPTools {
			fmt.Printf("  • %s\n", tool)
		}
	}

	if detailed {
		fmt.Println("\n💻 System Info:")
		for k, v := range status.SystemInfo {
			fmt.Printf("  %-12s: %v\n", k, v)
		}
	}

	fmt.Printf("\n⏱️  Last Check: %s\n", status.LastCheck.Format("2006-01-02 15:04:05"))
}

func getStatusIcon(status string) string {
	switch status {
	case "running":
		return "✅ "
	case "stopped":
		return "❌ "
	default:
		return "❓ "
	}
}

func getHealthIcon(health string) string {
	switch health {
	case "healthy":
		return "✅ "
	case "unhealthy", "not running":
		return "❌ "
	default:
		return "⚠️  "
	}
}

func stopService(name string, force, graceful bool, timeout int) {
	gl.Log("info", fmt.Sprintf("Stopping service: %s", name))

	// Check if service is running
	if !isServiceRunning(name) {
		gl.Log("warn", "Service is not running")
		return
	}

	if force {
		gl.Log("info", "Force stopping service...")
		// In real implementation, this would send SIGKILL
		forceStopService(name)
	} else if graceful {
		gl.Log("info", fmt.Sprintf("Gracefully stopping service (timeout: %ds)...", timeout))
		gracefulStopService(name, timeout)
	} else {
		// Normal stop
		gl.Log("info", "Stopping service...")
		normalStopService(name)
	}

	// Verify service stopped
	time.Sleep(1 * time.Second)
	if !isServiceRunning(name) {
		gl.Log("success", "Service stopped successfully")
	} else {
		gl.Log("error", "Service may still be running")
	}
}

func forceStopService(name string) {
	// In a real implementation, this would:
	// 1. Find the process ID
	// 2. Send SIGKILL
	// 3. Wait for process to die
	port := getServicePort()
	client := &http.Client{Timeout: 1 * time.Second}

	// Try to send shutdown signal via API
	_, err := client.Post(fmt.Sprintf("http://localhost:%s/admin/shutdown?force=true", port), "application/json", nil)
	if err != nil {
		gl.Log("warn", "Could not send shutdown signal via API")
	}
}

func gracefulStopService(name string, timeout int) {
	port := getServicePort()
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}

	// Send graceful shutdown signal
	_, err := client.Post(fmt.Sprintf("http://localhost:%s/admin/shutdown?graceful=true&timeout=%d", port, timeout), "application/json", nil)
	if err != nil {
		gl.Log("warn", "Could not send graceful shutdown signal via API")
		normalStopService(name)
		return
	}

	// Wait for service to stop gracefully
	for i := 0; i < timeout; i++ {
		time.Sleep(1 * time.Second)
		if !isServiceRunning(name) {
			gl.Log("info", fmt.Sprintf("Service stopped gracefully after %d seconds", i+1))
			return
		}
		if i%5 == 0 {
			gl.Log("info", fmt.Sprintf("Waiting for graceful shutdown... (%d/%d)", i, timeout))
		}
	}

	gl.Log("warn", "Graceful shutdown timeout reached, forcing stop...")
	forceStopService(name)
}

func normalStopService(name string) {
	port := getServicePort()
	client := &http.Client{Timeout: 10 * time.Second}

	// Send normal shutdown signal
	_, err := client.Post(fmt.Sprintf("http://localhost:%s/admin/shutdown", port), "application/json", nil)
	if err != nil {
		gl.Log("warn", "Could not send shutdown signal via API")
	}
}

func getLogs(name, level, format string, follow, timestamps bool, lines, tail int) {
	gl.Log("info", fmt.Sprintf("Retrieving logs for service: %s", name))

	// Try to get logs from running service first
	if isServiceRunning(name) {
		getLogsFromRunningService(name, level, format, follow, timestamps, lines, tail)
	} else {
		// Try to read from log file
		getLogsFromFile(name, level, format, follow, timestamps, lines, tail)
	}
}

func getLogsFromRunningService(name, level, format string, follow, timestamps bool, lines, tail int) {
	port := getServicePort()

	// Build query parameters
	params := fmt.Sprintf("?lines=%d", lines)
	if level != "" {
		params += "&level=" + level
	}
	if format == "json" {
		params += "&format=json"
	}
	if tail > 0 {
		params += fmt.Sprintf("&tail=%d", tail)
	}

	url := fmt.Sprintf("http://localhost:%s/admin/logs%s", port, params)

	if follow {
		followLogsFromService(url, timestamps)
	} else {
		getStaticLogsFromService(url, timestamps)
	}
}

func getStaticLogsFromService(url string, timestamps bool) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Failed to get logs from service: %v", err))
		return
	}
	defer resp.Body.Close()

	// Stream the response
	buffer := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			output := string(buffer[:n])
			if timestamps {
				// Add timestamp if not present
				fmt.Print(formatLogOutput(output, timestamps))
			} else {
				fmt.Print(output)
			}
		}
		if err != nil {
			break
		}
	}
}

func followLogsFromService(url string, timestamps bool) {
	gl.Log("info", "Following logs... Press Ctrl+C to stop")

	client := &http.Client{Timeout: 0} // No timeout for following
	resp, err := client.Get(url + "&follow=true")
	if err != nil {
		gl.Log("error", fmt.Sprintf("Failed to follow logs: %v", err))
		return
	}
	defer resp.Body.Close()

	// Stream the response continuously
	buffer := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			output := string(buffer[:n])
			fmt.Print(formatLogOutput(output, timestamps))
		}
		if err != nil {
			break
		}
	}
}

func getLogsFromFile(name, level, format string, follow, timestamps bool, lines, tail int) {
	// Try to find log file
	logFile := findLogFile(name)
	if logFile == "" {
		gl.Log("error", "No log file found and service is not running")
		return
	}

	gl.Log("info", fmt.Sprintf("Reading logs from file: %s", logFile))

	// Use tail command or read file directly
	if follow {
		gl.Log("info", "Following logs from file... Press Ctrl+C to stop")
		// In real implementation, this would use tail -f equivalent
		followLogFile(logFile, level, timestamps)
	} else {
		readLogFile(logFile, level, lines, tail, timestamps)
	}
}

func findLogFile(name string) string {
	// Try common log file locations
	logPaths := []string{
		fmt.Sprintf("/var/log/%s.log", name),
		fmt.Sprintf("./logs/%s.log", name),
		fmt.Sprintf("./%s.log", name),
		"./gobe.log",
		"/tmp/gobe.log",
	}

	for _, path := range logPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func followLogFile(filename, level string, timestamps bool) {
	// Simple implementation - in real scenario would use file watcher
	file, err := os.Open(filename)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Failed to open log file: %v", err))
		return
	}
	defer file.Close()

	// Seek to end of file
	file.Seek(0, 2)

	for {
		// Read new content
		buffer := make([]byte, 4096)
		n, err := file.Read(buffer)
		if n > 0 {
			output := string(buffer[:n])
			if level == "" || strings.Contains(strings.ToLower(output), strings.ToLower(level)) {
				fmt.Print(formatLogOutput(output, timestamps))
			}
		}
		if err != nil {
			time.Sleep(100 * time.Millisecond) // Wait for new content
		}
	}
}

func readLogFile(filename, level string, lines, tail int, timestamps bool) {
	content, err := os.ReadFile(filename)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Failed to read log file: %v", err))
		return
	}

	output := string(content)

	// Filter by level if specified
	if level != "" {
		filteredLines := []string{}
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(strings.ToLower(line), strings.ToLower(level)) {
				filteredLines = append(filteredLines, line)
			}
		}
		output = strings.Join(filteredLines, "\n")
	}

	// Apply tail if specified
	if tail > 0 {
		lines := strings.Split(output, "\n")
		if len(lines) > tail {
			lines = lines[len(lines)-tail:]
		}
		output = strings.Join(lines, "\n")
	}

	fmt.Print(formatLogOutput(output, timestamps))
}

func formatLogOutput(output string, timestamps bool) string {
	if !timestamps {
		return output
	}

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if line != "" && !strings.Contains(line, "[") {
			// Add timestamp if line doesn't seem to have one
			lines[i] = fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), line)
		}
	}

	return strings.Join(lines, "\n")
}
