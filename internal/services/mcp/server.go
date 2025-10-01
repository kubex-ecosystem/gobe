// Package mcp provides the implementation of the MCP server for Discord integration.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/kubex-ecosystem/gobe/internal/app/security/execsafe"
	"github.com/kubex-ecosystem/gobe/internal/commons/embedkit/components"
	"github.com/kubex-ecosystem/gobe/internal/commons/embedkit/helpers"
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

var systemInfoAllowList = map[string]struct{}{
	"1344830702780420157": {}, // Apenas você!
	"1400577637461659759": {},
	"880669325143461898":  {},
	"kblom":               {},
	"admin":               {},
	"faelmori":            {},
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
		startedAt: time.Now(),
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
	rawInfoType, _ := params["info_type"].(string)
	infoType := strings.TrimSpace(strings.ToLower(rawInfoType))
	if infoType == "" {
		infoType = "all"
	}

	userID, _ := params["user_id"].(string)

	if _, ok := systemInfoAllowList[userID]; !ok {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Usuário %s não autorizado para comandos do sistema", userID)), nil
	}

	cpuInfo, cpuErr := s.GetCPUInfo()
	memoryInfo, memoryErr := s.GetMemoryInfo()
	diskInfo, diskErr := s.GetDiskInfo()

	metricWarnings := map[string][]string{
		"cpu":    {},
		"memory": {},
		"disk":   {},
	}
	overallWarnings := make([]string, 0, 3)

	if cpuErr != nil {
		msg := fmt.Sprintf("CPU metrics fallback (%v)", cpuErr)
		metricWarnings["cpu"] = append(metricWarnings["cpu"], msg)
		overallWarnings = append(overallWarnings, msg)
	}
	if memoryErr != nil {
		msg := fmt.Sprintf("Memory metrics fallback (%v)", memoryErr)
		metricWarnings["memory"] = append(metricWarnings["memory"], msg)
		overallWarnings = append(overallWarnings, msg)
	}
	if diskErr != nil {
		msg := fmt.Sprintf("Disk metrics fallback (%v)", diskErr)
		metricWarnings["disk"] = append(metricWarnings["disk"], msg)
		overallWarnings = append(overallWarnings, msg)
	}

	status := "ok"
	switch {
	case len(overallWarnings) >= 2:
		status = "critical"
	case len(overallWarnings) == 1:
		status = "warning"
	}

	env := s.resolveEnvironment()
	envLabel := strings.ToUpper(env)
	host := resolveHostname()

	mcpStatus := "🟢 **ready** — all tools operational"
	analyzerStatus := "🟢 **ready** — v1.0.0"
	if len(overallWarnings) > 0 {
		mcpStatus = "🟡 **degraded** — métricas parciais disponíveis"
		analyzerStatus = "🟡 **standby** — aguardando métricas estáveis"
	}

	health := &types.SystemHealth{
		Status:     status,
		Version:    "v1.3.5",
		Uptime:     time.Since(s.startedAt),
		Host:       host,
		Mem:        memoryInfo,
		Disk:       diskInfo,
		CPU:        cpuInfo,
		Goroutines: fmt.Sprintf("ok (%d)", runtime.NumGoroutine()),
		GoBE:       fmt.Sprintf("🟢 **healthy** — %s com 3 servicos ativos\n`api`,`scheduler`,`webhooks`", envLabel),
		MCP:        mcpStatus,
		Analyzer:   analyzerStatus,
	}

	embed := s.BuildStatusEmbed(userID, env, health,
		"http://localhost:8088/swagger/index.html",
		"http://localhost:3666",
		"http://localhost:8088/api/v1/logs",
		overallWarnings...,
	)

	switch infoType {
	case "cpu":
		response := cpuInfo
		if warningText := formatBulletList(metricWarnings["cpu"]); warningText != "" {
			response = fmt.Sprintf("%s\n\n⚠️ Alertas:\n%s", response, warningText)
		}
		return mcp.NewToolResultText(response), nil
	case "memory":
		response := memoryInfo
		if warningText := formatBulletList(metricWarnings["memory"]); warningText != "" {
			response = fmt.Sprintf("%s\n\n⚠️ Alertas:\n%s", response, warningText)
		}
		return mcp.NewToolResultText(response), nil
	case "disk":
		response := diskInfo
		if warningText := formatBulletList(metricWarnings["disk"]); warningText != "" {
			response = fmt.Sprintf("%s\n\n⚠️ Alertas:\n%s", response, warningText)
		}
		return mcp.NewToolResultText(response), nil
	case "all", "status", "overview", "embed":
		payload, err := embedToJSON(embed)
		if err != nil {
			fallback := s.buildSystemSummary(env, health)
			if warningText := formatBulletList(overallWarnings); warningText != "" {
				fallback = fmt.Sprintf("%s\n\n⚠️ Alertas:\n%s", fallback, warningText)
			}
			return mcp.NewToolResultText(fallback), nil
		}
		return mcp.NewToolResultText(payload), nil
	default:
		return mcp.NewToolResultError("Tipo inválido. Use: cpu, memory, disk, all, status, overview, embed"), nil
	}
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
		return fmt.Sprintf("🔥 **CPU Usage**\nArquitetura: %s\nCores: %d\nStatus: Sistema ativo", runtime.GOARCH, runtime.NumCPU()), fmt.Errorf("collecting cpu info: %w", err)
	}
	data := strings.TrimSpace(string(output))
	return fmt.Sprintf("🔥 **CPU Usage**\nArquitetura: %s\nCores: %d\n%s", runtime.GOARCH, runtime.NumCPU(), data), nil
}

func (s *Server) GetMemoryInfo() (string, error) {
	cmd := exec.Command("sh", "-c", "free -h 2>/dev/null || echo 'Memória: Sistema Linux'")
	output, err := cmd.Output()
	if err != nil {
		return "💾 **Memory Info**\nSistema ativo\nRAM: Disponível", fmt.Errorf("collecting memory info: %w", err)
	}
	return fmt.Sprintf("💾 **Memory Info**\n%s", strings.TrimSpace(string(output))), nil
}

func (s *Server) GetDiskInfo() (string, error) {
	cmd := exec.Command("sh", "-c", "df -h / 2>/dev/null || echo 'Disco: Sistema ativo'")
	output, err := cmd.Output()
	if err != nil {
		return "💿 **Disk Usage**\nSistema de arquivos ativo", fmt.Errorf("collecting disk info: %w", err)
	}
	return fmt.Sprintf("💿 **Disk Usage**\n%s", strings.TrimSpace(string(output))), nil
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
func (s *Server) BuildStatusEmbed(userID, env string, health *types.SystemHealth, swaggerURL, panelURL, logsURL string, warnings ...string) map[string]interface{} {
	systemInfo := components.SystemInfo{
		Hostname:  health.Host,
		Uptime:    health.Uptime,
		Timestamp: time.Now(),
		Services: map[string]string{
			"GoBE":     parseServiceStatus(health.GoBE),
			"MCP":      parseServiceStatus(health.MCP),
			"Analyzer": parseServiceStatus(health.Analyzer),
		},
	}
	systemInfo.CPUUsage = parseUsagePercent(fmt.Sprintf("%v", health.CPU))
	systemInfo.MemoryUsage = parseUsagePercent(fmt.Sprintf("%v", health.Mem))
	systemInfo.DiskUsage = parseUsagePercent(fmt.Sprintf("%v", health.Disk))

	builder := components.NewStatusEmbedBuilder(systemInfo)
	builder.WithDescription(fmt.Sprintf(
		"Env `%s` • Host `%s`\nStatus: **%s** • Versão `%s`",
		strings.ToUpper(env),
		health.Host,
		formatStatusLabel(health.Status),
		health.Version,
	))
	builder.WithColor(helpers.StatusColor(health.Status, 0))
	builder.WithFooter(fmt.Sprintf("Requested by %s • MCP System Status", userID))
	builder.WithTimestamp(time.Now())

	builder.AddInlineField("🌍 Environment", strings.ToUpper(env))
	builder.AddInlineField("🔖 Version", health.Version)
	builder.AddInlineField("🔄 Goroutines", health.Goroutines)

	serviceDetails := make([]string, 0, 3)
	for _, entry := range []struct {
		label string
		value any
	}{
		{label: "GoBE", value: health.GoBE},
		{label: "MCP", value: health.MCP},
		{label: "Analyzer", value: health.Analyzer},
	} {
		if detail := formatServiceDetail(entry.label, entry.value); detail != "" {
			serviceDetails = append(serviceDetails, detail)
		}
	}
	if len(serviceDetails) > 0 {
		builder.AddField("🛰️ Modules", strings.Join(serviceDetails, "\n\n"), false)
	}

	links := components.NewLinkButtons(map[string]string{
		"📖 API Docs": swaggerURL,
		"📊 Panel":    panelURL,
		"📋 Logs":     logsURL,
	})
	if links != "" {
		builder.AddField("🔗 Quick Links", links, false)
	}

	if warningsText := formatBulletList(warnings); warningsText != "" {
		builder.AddField("⚠️ Alertas", warningsText, false)
	}

	return builder.Build()
}

func (s *Server) resolveEnvironment() string {
	if env := strings.TrimSpace(os.Getenv("APP_ENV")); env != "" {
		return env
	}
	if env := strings.TrimSpace(os.Getenv("ENVIRONMENT")); env != "" {
		return env
	}
	if env := strings.TrimSpace(os.Getenv("GO_ENV")); env != "" {
		return env
	}
	return "dev"
}

func resolveHostname() string {
	if host, err := os.Hostname(); err == nil {
		if trimmed := strings.TrimSpace(host); trimmed != "" {
			return trimmed
		}
	}
	return "dev"
}

func embedToJSON(embed map[string]interface{}) (string, error) {
	data, err := json.Marshal(embed)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func formatBulletList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		builder.WriteString("• ")
		builder.WriteString(trimmed)
		builder.WriteByte('\n')
	}
	return strings.TrimSpace(builder.String())
}

func formatServiceDetail(label string, value any) string {
	if value == nil {
		return ""
	}
	var body string
	switch v := value.(type) {
	case string:
		body = strings.TrimSpace(v)
	case fmt.Stringer:
		body = strings.TrimSpace(v.String())
	case []string:
		body = formatBulletList(v)
	default:
		encoded, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			body = strings.TrimSpace(fmt.Sprintf("%v", v))
		} else {
			body = fmt.Sprintf("```json\n%s\n```", string(encoded))
		}
	}

	if body == "" {
		return ""
	}
	return fmt.Sprintf("**%s**\n%s", label, body)
}

func (s *Server) buildSystemSummary(env string, health *types.SystemHealth) string {
	var builder strings.Builder
	builder.WriteString("🤖 **System Status**\n")
	builder.WriteString(fmt.Sprintf("Env: `%s`\n", strings.ToUpper(env)))
	builder.WriteString(fmt.Sprintf("Host: `%s`\n", health.Host))
	builder.WriteString(fmt.Sprintf("Status: %s\n", formatStatusLabel(health.Status)))
	builder.WriteString(fmt.Sprintf("Uptime: %s\n\n", health.Uptime.Truncate(time.Second)))
	builder.WriteString(strings.TrimSpace(fmt.Sprintf("%v", health.CPU)))
	builder.WriteString("\n\n")
	builder.WriteString(strings.TrimSpace(fmt.Sprintf("%v", health.Mem)))
	builder.WriteString("\n\n")
	builder.WriteString(strings.TrimSpace(fmt.Sprintf("%v", health.Disk)))
	return strings.TrimSpace(builder.String())
}

func formatStatusLabel(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "UNKNOWN"
	}
	return strings.ToUpper(status)
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
