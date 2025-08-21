// Package mcp provides the implementation of the MCP server for Discord integration.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/rafa-mori/gobe/internal/events"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	mcpServer *server.MCPServer
	hub       MCPHandler
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
	// TODO: Fix resource handlers for new mcp-go version
	// Discord Events Resource
	eventsResource := mcp.NewResource(
		"discord://events", "Discord Events Stream",
		mcp.WithResourceDescription("Real-time Discord events and processing status"),
		mcp.WithMIMEType("application/json"),
	)
	_ = eventsResource // Temporary to avoid unused variable error

	// Discord Channels Template
	channelTemplate := mcp.NewResourceTemplate(
		"discord://channels/{guild_id}",
		"Discord Channels",
		mcp.WithTemplateDescription("List of Discord channels in a guild"),
		mcp.WithTemplateMIMEType("application/json"),
	)
	_ = channelTemplate // Temporary to avoid unused variable error
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
	// TODO: Fix Start method for new mcp-go version
	// For now, just return nil to allow compilation
	return nil
}
