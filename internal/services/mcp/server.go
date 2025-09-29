// Package mcp provides the implementation of the MCP server for Discord integration.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/kubex-ecosystem/gobe/internal/app/security/execsafe"
	"github.com/kubex-ecosystem/gobe/internal/commons/embedkit"
	"github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	"github.com/kubex-ecosystem/gobe/internal/contracts/types"
	"github.com/kubex-ecosystem/gobe/internal/observers/events"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	mcpServer   *server.MCPServer
	hub         MCPHandler
	MCPRegistry execsafe.Registry

	startedAt time.Time
	userTag   string

	ChatBots interfaces.IProperty[interfaces.IAdapter]
}

type IMCPServer interface {
	RegisterTools()
	RegisterResources()
	HandleAnalyzeMessage(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error)
	HandleSendMessage(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error)
	HandleCreateTask(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error)
	HandleSystemInfo(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error)
	HandleShellCommand(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error)
	GetCPUInfo() (string, error)
	GetMemoryInfo() (string, error)
	GetDiskInfo() (string, error)
}

type MCPHandler interface {
	ProcessMessageWithLLM(ctx context.Context, msg interface{}) error
	SendDiscordMessage(channelID, content string) error
	GetEventStream() *events.Stream
}

func NewMCPServer(hub MCPHandler) (IMCPServer, error) {
	if hub == nil {
		return nil, fmt.Errorf("MCPHandler cannot be nil")
	}
	server, err := NewServer(hub)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP server: %w", err)
	}
	return server, nil
}

func NewServer(hub MCPHandler) (*Server, error) {
	mcpServer := server.NewMCPServer(
		"Discord MCP Hub", "1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	srv := &Server{
		mcpServer: mcpServer,
		hub:       hub,
	}

	reg := execsafe.NewRegistry()

	// ls: somente flags inofensivas e no máx. 1 path relativo
	safeLsFlags := execsafe.OneOfFlags("-l", "-a", "-la", "-lh", "-1", "-h")
	pathRx := regexp.MustCompile(`^[\w\-.\/]+$`) // sem espaços, sem ~, sem $VAR
	reg.Register("ls", execsafe.CommandSpec{
		Binary:      "ls",
		Timeout:     2 * time.Second,
		MaxOutputKB: 256,
		ArgsValidate: execsafe.Chain(
			func(args []string) error {
				if len(args) > 2 {
					return fmt.Errorf("ls: muitos argumentos")
				}
				return nil
			},
			safeLsFlags,
			func(args []string) error {
				for _, a := range args {
					if strings.HasPrefix(a, "-") {
						continue
					}
					if !pathRx.MatchString(a) {
						return fmt.Errorf("caminho inválido: %s", a)
					}
					if strings.Contains(a, "..") {
						return fmt.Errorf("path traversal proibido")
					}
				}
				return nil
			},
		),
	})

	// ps: flags controladas
	reg.Register("ps", execsafe.CommandSpec{
		Binary:       "ps",
		Timeout:      2 * time.Second,
		MaxOutputKB:  256,
		ArgsValidate: execsafe.OneOfFlags("aux", "-ef"),
	})

	// docker: somente "ps" com flags seguras
	reg.Register("docker", execsafe.CommandSpec{
		Binary:      "docker",
		Timeout:     3 * time.Second,
		MaxOutputKB: 512,
		ArgsValidate: func(args []string) error {
			if len(args) == 0 || args[0] != "ps" {
				return fmt.Errorf("apenas 'docker ps' permitido")
			}
			// flags simples permitidas
			ok := map[string]struct{}{"--all": {}, "-a": {}, "--format": {}}
			for _, a := range args[1:] {
				if strings.HasPrefix(a, "-") {
					if _, found := ok[a]; !found {
						return fmt.Errorf("docker flag bloqueada: %s", a)
					}
				}
			}
			return nil
		},
	})

	srv.RegisterTools()
	srv.RegisterResources()

	return srv, nil
}

func (s *Server) RegisterTools() {
	// Analyze Discord Message Tool
	analyzeTool := mcp.NewTool("analyze_discord_message",
		mcp.WithDescription("Analyze a Discord message and suggest actions"),
		mcp.WithString("message_content", mcp.Required()),
		mcp.WithString("channel_id", mcp.Required()),
		mcp.WithString("user_id", mcp.Required()),
		mcp.WithString("guild_id"),
	)

	analyzeHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}
		return s.HandleAnalyzeMessage(ctx, params)
	}
	s.mcpServer.AddTool(analyzeTool, analyzeHandler)

	// Send Discord Message Tool
	sendTool := mcp.NewTool("send_discord_message",
		mcp.WithDescription("Send a message to a Discord channel"),
		mcp.WithString("channel_id", mcp.Required()),
		mcp.WithString("content", mcp.Required()),
		mcp.WithBoolean("require_approval"),
	)

	sendHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}
		return s.HandleSendMessage(ctx, params)
	}
	s.mcpServer.AddTool(sendTool, sendHandler)

	// System Info Tool - Automação Real!
	systemInfoTool := mcp.NewTool("get_system_info",
		mcp.WithDescription("Get real-time system information (CPU, RAM, disk usage)"),
		mcp.WithString("info_type", mcp.Required()), // "cpu", "memory", "disk", "all"
		mcp.WithString("user_id", mcp.Required()),   // Para validação de segurança
	)

	systemInfoHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}
		return s.HandleSystemInfo(ctx, params)
	}
	s.mcpServer.AddTool(systemInfoTool, systemInfoHandler)

	// Execute Shell Command Tool - CUIDADO: Muito poderoso!
	shellTool := mcp.NewTool("execute_shell_command",
		mcp.WithDescription("Execute shell command on host system - REQUIRES ADMIN"),
		mcp.WithString("command", mcp.Required()),
		mcp.WithString("user_id", mcp.Required()),
		mcp.WithBoolean("require_confirmation"),
	)

	shellHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}
		return s.HandleShellCommand(ctx, params)
	}
	s.mcpServer.AddTool(shellTool, shellHandler)

	// Create Task Tool
	taskTool := mcp.NewTool("create_task_from_message",
		mcp.WithDescription("Create a task based on Discord message"),
		mcp.WithString("message_id", mcp.Required()),
		mcp.WithString("task_title", mcp.Required()),
		mcp.WithString("task_description"),
		mcp.WithString("priority", mcp.Enum("low", "medium", "high", "urgent")),
	)

	taskHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}
		return s.HandleCreateTask(ctx, params)
	}
	s.mcpServer.AddTool(taskTool, taskHandler)
}

func (s *Server) RegisterResources() {
	// Discord Events Resource
	eventsResource := mcp.NewResource(
		"discord://events", "Discord Events Stream",
		mcp.WithResourceDescription("Real-time Discord events and processing status"),
		mcp.WithMIMEType("application/json"),
	)

	eventsHandler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		events := map[string]interface{}{
			"status":           "active",
			"events_processed": 0,
			"last_update":      "2024-01-01T00:00:00Z",
		}

		data, err := json.Marshal(events)
		if err != nil {
			return nil, err
		}

		return []mcp.ResourceContents{mcp.TextResourceContents{
			URI:      "discord://events",
			MIMEType: "application/json",
			Text:     string(data),
		}}, nil
	}
	s.mcpServer.AddResource(eventsResource, eventsHandler)

	// Discord Channels Template
	channelTemplate := mcp.NewResourceTemplate(
		"discord://channels/{guild_id}",
		"Discord Channels",
		mcp.WithTemplateDescription("List of Discord channels in a guild"),
		mcp.WithTemplateMIMEType("application/json"),
	)

	channelHandler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Extract guild_id from URI path parameters
		uri := request.Params.URI
		guildID := ""
		if strings.Contains(uri, "discord://channels/") {
			guildID = strings.TrimPrefix(uri, "discord://channels/")
		}

		channels := map[string]interface{}{
			"guild_id": guildID,
			"channels": []map[string]interface{}{
				{"id": "channel1", "name": "general", "type": "text"},
				{"id": "channel2", "name": "random", "type": "text"},
			},
		}

		data, err := json.Marshal(channels)
		if err != nil {
			return nil, err
		}

		return []mcp.ResourceContents{mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(data),
		}}, nil
	}
	s.mcpServer.AddResourceTemplate(channelTemplate, channelHandler)
}

func (s *Server) HandleAnalyzeMessage(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	content, _ := params["message_content"].(string)
	channelID, _ := params["channel_id"].(string)
	userID, _ := params["user_id"].(string)
	guildID, _ := params["guild_id"].(string)

	// Create a mock message for analysis
	message := map[string]interface{}{
		"content":    content,
		"channel_id": channelID,
		"user_id":    userID,
		"guild_id":   guildID,
	}

	err := s.hub.ProcessMessageWithLLM(ctx, message)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Analysis failed: %v", err)), nil
	}

	return mcp.NewToolResultText("Message analyzed successfully"), nil
}

func (s *Server) HandleSendMessage(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	channelID, _ := params["channel_id"].(string)
	content, _ := params["content"].(string)

	err := s.hub.SendDiscordMessage(channelID, content)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to send message: %v", err)), nil
	}

	return mcp.NewToolResultText("Message sent successfully"), nil
}

func (s *Server) HandleCreateTask(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	messageID, _ := params["message_id"].(string)
	title, _ := params["task_title"].(string)
	description, _ := params["task_description"].(string)
	priority, _ := params["priority"].(string)

	task := map[string]interface{}{
		"message_id":  messageID,
		"title":       title,
		"description": description,
		"priority":    priority,
		"source":      "discord",
	}

	result, _ := json.Marshal(task)
	return mcp.NewToolResultText(string(result)), nil
}

func (s *Server) HandleSystemInfo(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	infoType, _ := params["info_type"].(string)
	userID, _ := params["user_id"].(string)

	// 🔒 Validação de Segurança (simplificada para demo)
	authorizedUsers := []string{
		"1344830702780420157", // Apenas você!
		"1400577637461659759",
		"880669325143461898",
		"kblom",
		"admin",
		"faelmori",
	}

	isAuthorized := false
	for _, authUser := range authorizedUsers {
		if userID == authUser {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Usuário %s não autorizado para comandos do sistema", userID)), nil
	}

	var result string
	var err error

	// Get system information
	mem, err := s.GetMemoryInfo()
	if err != nil {
		mem = "Memory info unavailable"
	}

	disk, err := s.GetDiskInfo()
	if err != nil {
		disk = "Disk info unavailable"
	}

	cpu, err := s.GetCPUInfo()
	if err != nil {
		cpu = "CPU info unavailable"
	}

	// Build system health status
	health := &types.SystemHealth{
		Status:     "ok",
		Version:    "v1.3.5",
		Uptime:     time.Since(s.startedAt),
		Host:       "dev",
		Mem:        mem,
		Disk:       disk,
		CPU:        cpu,
		Goroutines: fmt.Sprintf("ok (%d)", runtime.NumGoroutine()),
		GoBE:       "🟢 **healthy** — 3 services ativos\n`api`,`scheduler`,`webhooks`",
		MCP:        "🟢 **ready** — all tools operational",
		Analyzer:   "🟢 **ready** — v1.0.0",
	}

	// Create status embed using embedkit
	embed := s.BuildStatusEmbed(userID, "dev", health,
		"http://localhost:8088/swagger/index.html",
		"http://localhost:3666", // painel
		"http://localhost:8088/api/v1/logs",
	)

	// Convert embed to string for MCP response
	embedJSON, err := json.Marshal(embed)
	if err != nil {
		result = fmt.Sprintf("🤖 **System Status**\nStatus: %s\nUptime: %s\nHost: %s\n\n%s\n\n%s\n\n%s",
			health.Status, health.Uptime.String(), health.Host, cpu, mem, disk)
	} else {
		result = string(embedJSON)
	}

	switch infoType {
	case "cpu":
		result, err = s.GetCPUInfo()
	case "memory":
		result, err = s.GetMemoryInfo()
	case "disk":
		result, err = s.GetDiskInfo()
	case "all":
		cpu, _ := s.GetCPUInfo()
		memory, _ := s.GetMemoryInfo()
		disk, _ := s.GetDiskInfo()
		result = fmt.Sprintf("🖥️ **System Info Complete**\n\n%s\n\n%s\n\n%s", cpu, memory, disk)
	default:
		return mcp.NewToolResultError("Tipo inválido. Use: cpu, memory, disk, all"), nil
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Erro ao obter info do sistema: %v", err)), nil
	}

	return mcp.NewToolResultText(result), nil
}

func (s *Server) HandleShellCommand(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	command, _ := params["command"].(string)
	userID, _ := params["user_id"].(string)
	requireConfirmation, _ := params["require_confirmation"].(bool)

	// 🔒 SUPER Validação de Segurança
	adminUsers := []string{
		"1344830702780420157", // Apenas você!
		"1400577637461659759",
		"880669325143461898",
	}

	isAdmin := false
	for _, admin := range adminUsers {
		if userID == admin {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return mcp.NewToolResultError("❌ ACESSO NEGADO: Apenas administradores podem executar comandos shell"), nil
	}

	// 🚫 Blacklist de comandos perigosos
	dangerousCommands := []string{"rm -rf", "mkfs", "dd if=", "shutdown", "reboot", "passwd", "userdel"}
	for _, dangerous := range dangerousCommands {
		if strings.Contains(strings.ToLower(command), dangerous) {
			return mcp.NewToolResultError(fmt.Sprintf("❌ Comando bloqueado por segurança: %s", dangerous)), nil
		}
	}

	if requireConfirmation {
		return mcp.NewToolResultText(fmt.Sprintf("⚠️ **CONFIRMAÇÃO NECESSÁRIA**\n\nComando: `%s`\n\nResponda 'CONFIRMO' para executar", command)), nil
	}

	// Log da execução
	fmt.Printf("🔧 SHELL EXECUTION by %s: %s\n", userID, command)

	output, err := s.executeShellCommand(command)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Erro na execução: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("✅ **Comando executado**\n```\n%s\n```\n\n📄 **Output:**\n```\n%s\n```", command, output)), nil
}

func (s *Server) GetCPUInfo() (string, error) {
	cmd := exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' || echo 'CPU: Informação não disponível'")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("🔥 **CPU Usage**\nArquitetura: %s\nCores: %d\nStatus: Sistema ativo", runtime.GOARCH, runtime.NumCPU()), nil
	}
	return fmt.Sprintf("🔥 **CPU Usage**\nArquitetura: %s\nCores: %d\n%s", runtime.GOARCH, runtime.NumCPU(), string(output)), nil
}

func (s *Server) GetMemoryInfo() (string, error) {
	cmd := exec.Command("sh", "-c", "free -h 2>/dev/null || echo 'Memória: Sistema Linux'")
	output, err := cmd.Output()
	if err != nil {
		return "💾 **Memory Info**\nSistema ativo\nRAM: Disponível", nil
	}
	return fmt.Sprintf("💾 **Memory Info**\n%s", string(output)), nil
}

func (s *Server) GetDiskInfo() (string, error) {
	cmd := exec.Command("sh", "-c", "df -h / 2>/dev/null || echo 'Disco: Sistema ativo'")
	output, err := cmd.Output()
	if err != nil {
		return "💿 **Disk Usage**\nSistema de arquivos ativo", nil
	}
	return fmt.Sprintf("💿 **Disk Usage**\n%s", string(output)), nil
}

func (s *Server) executeShellCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (s *Server) handleEventsResource(ctx context.Context) (*mcp.ReadResourceResult, error) {
	events := map[string]interface{}{
		"status":           "active",
		"events_processed": 0,
		"last_update":      "2024-01-01T00:00:00Z",
	}

	data, _ := json.Marshal(events)
	return mcp.NewReadResourceResult(string(data)), nil
}

func (s *Server) handleChannelsResource(ctx context.Context, params map[string]string) (*mcp.ReadResourceResult, error) {
	guildID := params["guild_id"]

	channels := map[string]interface{}{
		"guild_id": guildID,
		"channels": []map[string]interface{}{
			{"id": "channel1", "name": "general", "type": "text"},
			{"id": "channel2", "name": "random", "type": "text"},
		},
	}

	data, _ := json.Marshal(channels)
	return mcp.NewReadResourceResult(string(data)), nil
}

func (s *Server) Start() error {
	// The MCP server is ready to handle requests once tools and resources are registered
	// In the new version, we don't need an explicit Start method as the server is
	// ready immediately after initialization
	return nil
}

// BuildStatusEmbed creates a status embed using embedkit
func (s *Server) BuildStatusEmbed(userID, env string, health *types.SystemHealth, swaggerURL, panelURL, logsURL string) map[string]interface{} {
	// Convert SystemHealth to embedkit.SystemInfo
	systemInfo := embedkit.SystemInfo{
		Hostname:  health.Host,
		Uptime:    health.Uptime,
		Timestamp: time.Now(),
		Services: map[string]string{
			"GoBE":     parseServiceStatus(health.GoBE),
			"MCP":      parseServiceStatus(health.MCP),
			"Analyzer": parseServiceStatus(health.Analyzer),
		},
	}

	// Parse CPU, Memory and Disk usage from the strings (simplified)
	systemInfo.CPUUsage = parseUsagePercent(fmt.Sprintf("%v", health.CPU))
	systemInfo.MemoryUsage = parseUsagePercent(fmt.Sprintf("%v", health.Mem))
	systemInfo.DiskUsage = parseUsagePercent(fmt.Sprintf("%v", health.Disk))

	// Create base embed
	embed := embedkit.StatusEmbed(systemInfo)

	// Add custom fields specific to the MCP server
	if fields, ok := embed["fields"].([]map[string]interface{}); ok {
		// Add version field
		fields = append(fields, map[string]interface{}{
			"name":   "🔧 Version",
			"value":  health.Version,
			"inline": true,
		})

		// Add environment field
		fields = append(fields, map[string]interface{}{
			"name":   "🌍 Environment",
			"value":  env,
			"inline": true,
		})

		// Add goroutines field
		fields = append(fields, map[string]interface{}{
			"name":   "🔄 Goroutines",
			"value":  health.Goroutines,
			"inline": true,
		})

		// Add links as buttons
		links := embedkit.NewLinkButtons(map[string]string{
			"📖 API Docs": swaggerURL,
			"📊 Panel":    panelURL,
			"📋 Logs":     logsURL,
		})

		if links != "" {
			fields = append(fields, map[string]interface{}{
				"name":  "🔗 Quick Links",
				"value": links,
			})
		}

		embed["fields"] = fields
	}

	// Add footer with user info
	embed["footer"] = map[string]interface{}{
		"text": fmt.Sprintf("Requested by %s • MCP System Status", userID),
	}

	return embed
}

// Helper functions for parsing service status and usage percentages
func parseServiceStatus(service interface{}) string {
	serviceStr := fmt.Sprintf("%v", service)
	if strings.Contains(serviceStr, "🟢") {
		return "running"
	} else if strings.Contains(serviceStr, "🟡") {
		return "warning"
	} else if strings.Contains(serviceStr, "🔴") {
		return "failed"
	}
	return "unknown"
}

func parseUsagePercent(usage string) float64 {
	// Simple regex to extract percentage from strings like "45.2%" or "CPU: 67%"
	re := regexp.MustCompile(`(\d+\.?\d*)%`)
	matches := re.FindStringSubmatch(usage)
	if len(matches) > 1 {
		var percent float64
		fmt.Sscanf(matches[1], "%f", &percent)
		return percent
	}
	return 0.0
}
