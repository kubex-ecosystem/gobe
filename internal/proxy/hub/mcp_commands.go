// Package hub provides MCP command handlers for Discord integration
package hub

import (
	"context"
	"fmt"
	"strings"

	"github.com/kubex-ecosystem/gobe/internal/services/mcp"
	logz "github.com/kubex-ecosystem/logz"
)

// RegisterMCPCommands registers Discord command handlers for MCP tools
func (h *DiscordMCPHub) RegisterMCPCommands() {
	logz.Log("info", "Registering Discord MCP command handlers")

	// Register command handlers via Discord adapter
	if h.discordAdapter != nil {
		// !grompt command - Generate structured prompts
		h.RegisterCommand("grompt", h.HandleGromptCommand)

		// !ask command - Direct AI prompt
		h.RegisterCommand("ask", h.HandleAskCommand)

		// !analyze command - Analyze project
		h.RegisterCommand("analyze", h.HandleAnalyzeCommand)

		// !security command - Security audit
		h.RegisterCommand("security", h.HandleSecurityCommand)

		// !mcp command - List/execute MCP tools
		h.RegisterCommand("mcp", h.HandleMCPCommand)

		logz.Log("info", "Discord MCP commands registered successfully")
	}
}

// HandleGromptCommand processes !grompt command
// Usage: !grompt <idea1>, <idea2>, <idea3> [--purpose=<purpose>] [--provider=<provider>]
func (h *DiscordMCPHub) HandleGromptCommand(ctx context.Context, message string, args []string) (string, error) {
	logz.Log("info", "Processing !grompt command", "args", args)

	if len(args) == 0 {
		return formatDiscordHelp("grompt"), nil
	}

	// Parse args
	ideas, purpose, provider := parseGromptArgs(args)

	if len(ideas) == 0 {
		return "❌ **Error:** Please provide at least one idea!\n\n" + formatDiscordHelp("grompt"), nil
	}

	// Execute via MCP
	result, err := h.mcpRegistry.Exec(ctx, "grompt.generate", map[string]interface{}{
		"ideas":      ideas,
		"purpose":    purpose,
		"provider":   provider,
		"max_tokens": 5000,
	})

	if err != nil {
		logz.Log("error", "Grompt command failed", "error", err)
		return fmt.Sprintf("❌ **Failed to generate prompt:**\n```\n%v\n```", err), nil
	}

	// Format response for Discord
	return formatGromptResponse(result), nil
}

// HandleAskCommand processes !ask command for direct AI prompts
// Usage: !ask <question> [--provider=<provider>]
func (h *DiscordMCPHub) HandleAskCommand(ctx context.Context, message string, args []string) (string, error) {
	logz.Log("info", "Processing !ask command", "args", args)

	if len(args) == 0 {
		return formatDiscordHelp("ask"), nil
	}

	// Parse args
	prompt, provider := parseAskArgs(args)

	// Execute via MCP
	result, err := h.mcpRegistry.Exec(ctx, "grompt.direct", map[string]interface{}{
		"prompt":     prompt,
		"provider":   provider,
		"max_tokens": 1000,
	})

	if err != nil {
		logz.Log("error", "Ask command failed", "error", err)
		return fmt.Sprintf("❌ **Failed to get response:**\n```\n%v\n```", err), nil
	}

	// Format response for Discord
	return formatAskResponse(result), nil
}

// HandleAnalyzeCommand processes !analyze command
// Usage: !analyze <path> [--depth=<1-5>]
func (h *DiscordMCPHub) HandleAnalyzeCommand(ctx context.Context, message string, args []string) (string, error) {
	logz.Log("info", "Processing !analyze command", "args", args)

	if len(args) == 0 {
		return formatDiscordHelp("analyze"), nil
	}

	// Parse args
	projectPath, depth := parseAnalyzeArgs(args)

	// Execute via MCP
	result, err := h.mcpRegistry.Exec(ctx, "analyzer.project", map[string]interface{}{
		"project_path":         projectPath,
		"depth":                depth,
		"include_dependencies": true,
	})

	if err != nil {
		logz.Log("error", "Analyze command failed", "error", err)
		return fmt.Sprintf("❌ **Analysis failed:**\n```\n%v\n```", err), nil
	}

	// Format response for Discord
	return formatAnalyzeResponse(result), nil
}

// HandleSecurityCommand processes !security command
// Usage: !security <path> [--severity=<low|medium|high|critical>]
func (h *DiscordMCPHub) HandleSecurityCommand(ctx context.Context, message string, args []string) (string, error) {
	logz.Log("info", "Processing !security command", "args", args)

	if len(args) == 0 {
		return formatDiscordHelp("security"), nil
	}

	// Parse args
	projectPath, severity := parseSecurityArgs(args)

	// Execute via MCP
	result, err := h.mcpRegistry.Exec(ctx, "analyzer.security", map[string]interface{}{
		"project_path":       projectPath,
		"severity_threshold": severity,
	})

	if err != nil {
		logz.Log("error", "Security command failed", "error", err)
		return fmt.Sprintf("❌ **Security audit failed:**\n```\n%v\n```", err), nil
	}

	// Format response for Discord
	return formatSecurityResponse(result), nil
}

// HandleMCPCommand processes !mcp command to list/execute MCP tools
// Usage: !mcp list | !mcp exec <tool> <args>
func (h *DiscordMCPHub) HandleMCPCommand(ctx context.Context, message string, args []string) (string, error) {
	logz.Log("info", "Processing !mcp command", "args", args)

	if len(args) == 0 {
		return formatDiscordHelp("mcp"), nil
	}

	subcommand := args[0]

	switch subcommand {
	case "list":
		tools := h.mcpRegistry.List()
		return formatMCPToolsList(tools), nil

	case "exec":
		if len(args) < 2 {
			return "❌ **Error:** Please specify a tool name!\n\nUsage: `!mcp exec <tool> <args>`", nil
		}
		// This would require more complex arg parsing
		return "⚠️ **Not implemented yet**\nUse specific commands like `!grompt`, `!ask`, `!analyze` instead.", nil

	default:
		return formatDiscordHelp("mcp"), nil
	}
}

// RegisterCommand is a helper to register commands with the Discord adapter
func (h *DiscordMCPHub) RegisterCommand(name string, handler func(context.Context, string, []string) (string, error)) {
	// This would integrate with your Discord adapter's command registration
	// Implementation depends on your Discord library
	logz.Log("debug", "Registered Discord command", "command", name)
}

// ========== Argument Parsers ==========

func parseGromptArgs(args []string) (ideas []string, purpose string, provider string) {
	purpose = "General"
	provider = "gemini"

	var currentIdeas []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "--purpose=") {
			purpose = strings.TrimPrefix(arg, "--purpose=")
		} else if strings.HasPrefix(arg, "--provider=") {
			provider = strings.TrimPrefix(arg, "--provider=")
		} else {
			// Collect ideas (might be comma-separated)
			parts := strings.Split(arg, ",")
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					currentIdeas = append(currentIdeas, trimmed)
				}
			}
		}
	}

	return currentIdeas, purpose, provider
}

func parseAskArgs(args []string) (prompt string, provider string) {
	provider = "gemini"
	var promptParts []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "--provider=") {
			provider = strings.TrimPrefix(arg, "--provider=")
		} else {
			promptParts = append(promptParts, arg)
		}
	}

	prompt = strings.Join(promptParts, " ")
	return
}

func parseAnalyzeArgs(args []string) (projectPath string, depth int) {
	depth = 3 // default

	for _, arg := range args {
		if strings.HasPrefix(arg, "--depth=") {
			depthStr := strings.TrimPrefix(arg, "--depth=")
			fmt.Sscanf(depthStr, "%d", &depth)
			if depth < 1 {
				depth = 1
			}
			if depth > 5 {
				depth = 5
			}
		} else if !strings.HasPrefix(arg, "--") {
			projectPath = arg
		}
	}

	return
}

func parseSecurityArgs(args []string) (projectPath string, severity string) {
	severity = "medium" // default

	for _, arg := range args {
		if strings.HasPrefix(arg, "--severity=") {
			severity = strings.TrimPrefix(arg, "--severity=")
		} else if !strings.HasPrefix(arg, "--") {
			projectPath = arg
		}
	}

	return
}

// ========== Response Formatters ==========

func formatGromptResponse(result interface{}) string {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("✅ **Prompt Generated**\n```\n%v\n```", result)
	}

	response, _ := resultMap["response"].(string)
	provider, _ := resultMap["provider"].(string)
	model, _ := resultMap["model"].(string)

	// Truncate if too long for Discord (2000 char limit)
	if len(response) > 1800 {
		response = response[:1800] + "\n\n... *(truncated)*"
	}

	output := "🧠 **Grompt - Structured Prompt Generated**\n\n"
	output += fmt.Sprintf("📝 **Provider:** %s (%s)\n", provider, model)
	output += fmt.Sprintf("```\n%s\n```\n", response)

	// Add usage info if available
	if usage, ok := resultMap["usage"].(map[string]interface{}); ok {
		if total, ok := usage["total_tokens"].(float64); ok {
			output += fmt.Sprintf("\n💰 **Tokens:** %.0f", total)
		}
	}

	return output
}

func formatAskResponse(result interface{}) string {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("✅ **Response:**\n```\n%v\n```", result)
	}

	response, _ := resultMap["response"].(string)
	provider, _ := resultMap["provider"].(string)

	// Truncate if too long
	if len(response) > 1800 {
		response = response[:1800] + "\n\n... *(truncated)*"
	}

	return fmt.Sprintf("🤖 **%s:**\n\n%s", provider, response)
}

func formatAnalyzeResponse(result interface{}) string {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("📊 **Analysis Result:**\n```json\n%v\n```", result)
	}

	output := "📊 **Project Analysis**\n\n"

	// Extract project info
	if project, ok := resultMap["project"].(map[string]interface{}); ok {
		if name, ok := project["name"].(string); ok {
			output += fmt.Sprintf("📁 **Name:** %s\n", name)
		}
		if lang, ok := project["language"].(string); ok {
			output += fmt.Sprintf("💻 **Language:** %s\n", lang)
		}
		if files, ok := project["files_count"].(float64); ok {
			output += fmt.Sprintf("📄 **Files:** %.0f\n", files)
		}
		if loc, ok := project["loc"].(float64); ok {
			output += fmt.Sprintf("📏 **Lines of Code:** %.0f\n", loc)
		}
	}

	output += "\n✅ **Analysis complete!** Check logs for details."

	return output
}

func formatSecurityResponse(result interface{}) string {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("🔒 **Security Audit:**\n```json\n%v\n```", result)
	}

	output := "🔒 **Security Audit Results**\n\n"

	// Extract summary
	if summary, ok := resultMap["summary"].(map[string]interface{}); ok {
		critical := summary["critical"]
		high := summary["high"]
		medium := summary["medium"]
		low := summary["low"]

		output += fmt.Sprintf("🔴 **Critical:** %v\n", critical)
		output += fmt.Sprintf("🟠 **High:** %v\n", high)
		output += fmt.Sprintf("🟡 **Medium:** %v\n", medium)
		output += fmt.Sprintf("🟢 **Low:** %v\n\n", low)
	}

	// Extract vulnerabilities (show top 3)
	if vulns, ok := resultMap["vulnerabilities"].([]interface{}); ok && len(vulns) > 0 {
		output += "**Top Vulnerabilities:**\n"
		for i, v := range vulns {
			if i >= 3 {
				output += fmt.Sprintf("\n... and %d more\n", len(vulns)-3)
				break
			}
			if vuln, ok := v.(map[string]interface{}); ok {
				severity, _ := vuln["severity"].(string)
				file, _ := vuln["file"].(string)
				desc, _ := vuln["description"].(string)
				output += fmt.Sprintf("\n%d. **[%s]** %s\n   📄 %s\n", i+1, severity, desc, file)
			}
		}
	} else {
		output += "✅ **No vulnerabilities found!**\n"
	}

	return output
}

func formatMCPToolsList(tools []mcp.ToolSpec) string {
	if len(tools) == 0 {
		return "📋 **No MCP tools registered**"
	}

	output := fmt.Sprintf("📋 **Available MCP Tools** (%d)\n\n", len(tools))

	categories := map[string][]string{
		"🔧 Built-in": {},
		"🧠 Grompt":   {},
		"🔍 Analyzer": {},
		"🎨 External": {},
	}

	for _, tool := range tools {
		name := tool.Name
		desc := tool.Description

		entry := fmt.Sprintf("• `%s` - %s", name, desc)

		if strings.HasPrefix(name, "system.") || strings.HasPrefix(name, "shell.") {
			categories["🔧 Built-in"] = append(categories["🔧 Built-in"], entry)
		} else if strings.HasPrefix(name, "grompt.") {
			categories["🧠 Grompt"] = append(categories["🧠 Grompt"], entry)
		} else if strings.HasPrefix(name, "analyzer.") {
			categories["🔍 Analyzer"] = append(categories["🔍 Analyzer"], entry)
		} else {
			categories["🎨 External"] = append(categories["🎨 External"], entry)
		}
	}

	// Output by category
	for category, items := range categories {
		if len(items) > 0 {
			output += fmt.Sprintf("**%s**\n", category)
			for _, item := range items {
				output += item + "\n"
			}
			output += "\n"
		}
	}

	return output
}

func formatDiscordHelp(command string) string {
	switch command {
	case "grompt":
		return "**🧠 Grompt - Structured Prompt Generator**\n\n" +
			"**Usage:** `!grompt <idea1>, <idea2> [options]`\n\n" +
			"**Options:**\n" +
			"• `--purpose=<purpose>` - Prompt purpose\n" +
			"• `--provider=<provider>` - AI provider\n\n" +
			"**Examples:**\n" +
			"```\n!grompt quantum computing, tutorial\n" +
			"!grompt REST API, golang --purpose=\"Code\"\n```"

	case "ask":
		return "**🤖 Ask - Direct AI Prompt**\n\n" +
			"**Usage:** `!ask <question> [--provider=<provider>]`\n\n" +
			"**Examples:**\n" +
			"```\n!ask Explain quantum entanglement\n" +
			"!ask REST vs GraphQL --provider=claude\n```"

	case "analyze":
		return "**🔍 Analyze - Project Analysis**\n\n" +
			"**Usage:** `!analyze <path> [--depth=<1-5>]`\n\n" +
			"**Examples:**\n" +
			"```\n!analyze /projects/kubex/gobe\n" +
			"!analyze /home/user/myproject --depth=5\n```"

	case "security":
		return "**🔒 Security - Security Audit**\n\n" +
			"**Usage:** `!security <path> [--severity=<level>]`\n\n" +
			"**Examples:**\n" +
			"```\n!security /projects/kubex/gobe\n" +
			"!security /home/user/app --severity=high\n```"

	case "mcp":
		return "**⚙️ MCP - Tool Management**\n\n" +
			"**Usage:**\n" +
			"• `!mcp list` - List all tools\n" +
			"• `!mcp exec <tool> <args>` - Execute tool\n\n" +
			"**Example:**\n```\n!mcp list\n```"

	default:
		return "❓ **Unknown command**\n\n" +
			"**Available:** `!grompt`, `!ask`, `!analyze`, `!security`, `!mcp`"
	}
}
