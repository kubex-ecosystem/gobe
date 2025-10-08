# 🔗 MCP Integration Guide - Kubex Ecosystem

## Overview

GoBE now integrates **Grompt** and **Analyzer** as MCP (Model Context Protocol) tools, creating a unified hub for all Kubex ecosystem functionality.

```
┌────────────────────────────────────────────────────────────┐
│                    GoBE MCP Hub (v1.3.5)                   │
├────────────────────────────────────────────────────────────┤
│  Built-in Tools:                                           │
│  • system.status     - System health and metrics           │
│  • shell.command     - Safe shell execution                │
├────────────────────────────────────────────────────────────┤
│  External Kubex Ecosystem Tools:                           │
│  • grompt.generate   - Structured prompt engineering       │
│  • grompt.direct     - Direct AI prompts                   │
│  • analyzer.project  - Project structure analysis          │
│  • analyzer.security - Security vulnerability detection    │
└────────────────────────────────────────────────────────────┘
                    │                 │
          ┌─────────┴─────┐   ┌──────┴────────┐
          │  Grompt:8080  │   │  Analyzer:8081│
          └───────────────┘   └───────────────┘
```

---

## 🚀 Quick Start

### 1. Start Required Services

```bash
# Terminal 1: Start Grompt
cd /projects/kubex/grompt
./grompt start -p 8080

# Terminal 2: Start Analyzer (if available)
cd /projects/kubex/analyzer
./analyzer start -p 8081

# Terminal 3: Start GoBE with MCP
cd /projects/kubex/gobe
./gobe start -p 3666
```

### 2. Configure External Tools (Optional)

```bash
# Override default URLs via environment variables
export GROMPT_URL=http://localhost:8080
export ANALYZER_URL=http://localhost:8081

./gobe start -p 3666
```

---

## 📋 Available MCP Tools

### Built-in Tools

#### `system.status`
Get comprehensive system status including health, version, and runtime metrics.

```bash
curl -X POST http://localhost:3666/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "system.status",
    "args": {
      "detailed": true
    }
  }'
```

**Response:**
```json
{
  "status": "ok",
  "version": "v1.3.5",
  "uptime_seconds": 3600,
  "health": {
    "status": "healthy",
    "checks": {...}
  },
  "connections": {
    "database": {...},
    "rabbitmq": {...},
    "webhooks": {...}
  }
}
```

---

#### `shell.command`
Execute safe shell commands with whitelist validation.

```bash
curl -X POST http://localhost:3666/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "shell.command",
    "args": {
      "command": "ls",
      "args": ["-la", "/tmp"]
    }
  }'
```

**Whitelisted commands:**
`ls`, `pwd`, `whoami`, `date`, `uptime`, `ps`, `df`, `free`, `uname`, `echo`, `cat`, `head`, `tail`, `grep`, `wc`, `sort`, `uniq`

---

### External Kubex Ecosystem Tools

#### `grompt.generate`
Generate professional, structured prompts from raw ideas using Grompt engine.

```bash
curl -X POST http://localhost:3666/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "grompt.generate",
    "args": {
      "ideas": [
        "quantum computing",
        "beginner-friendly explanation",
        "visual analogies"
      ],
      "purpose": "Educational Content",
      "provider": "gemini",
      "api_key": "AIzaYOUR_KEY_HERE"
    }
  }'
```

**Parameters:**
- `ideas` (array, required): Raw ideas/concepts
- `purpose` (string, required): Prompt purpose
- `provider` (string): AI provider (`openai`, `claude`, `gemini`, `deepseek`, `chatgpt`)
- `model` (string): Specific model name
- `api_key` (string): **BYOK** - External API key
- `max_tokens` (int): Maximum tokens (default: 5000)

**Response:**
```json
{
  "response": "Professional structured prompt...",
  "provider": "gemini",
  "model": "gemini-2.0-flash-exp",
  "usage": {
    "prompt_tokens": 120,
    "completion_tokens": 850,
    "total_tokens": 970
  }
}
```

---

#### `grompt.direct`
Send a direct prompt to AI provider without prompt engineering.

```bash
curl -X POST http://localhost:3666/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "grompt.direct",
    "args": {
      "prompt": "Explain quantum entanglement in 3 sentences",
      "provider": "openai",
      "model": "gpt-4o-mini",
      "api_key": "sk-proj-YOUR_KEY",
      "max_tokens": 500
    }
  }'
```

**Parameters:**
- `prompt` (string, required): Direct prompt text
- `provider` (string): AI provider
- `model` (string): Specific model
- `api_key` (string): **BYOK** - External API key
- `max_tokens` (int): Maximum tokens (default: 1000)

---

#### `analyzer.project`
Deep analysis of project structure, dependencies, and code quality.

```bash
curl -X POST http://localhost:3666/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "analyzer.project",
    "args": {
      "project_path": "/projects/kubex/gobe",
      "depth": 3,
      "include_dependencies": true
    }
  }'
```

**Parameters:**
- `project_path` (string, required): Path to project directory
- `depth` (int): Analysis depth 1-5 (default: 3)
- `include_dependencies` (bool): Include dependency graph (default: true)

**Response:**
```json
{
  "project": {
    "name": "gobe",
    "language": "go",
    "files_count": 145,
    "loc": 12453
  },
  "structure": {...},
  "dependencies": {...},
  "metrics": {...}
}
```

---

#### `analyzer.security`
Perform security audit and detect potential vulnerabilities.

```bash
curl -X POST http://localhost:3666/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "analyzer.security",
    "args": {
      "project_path": "/projects/kubex/gobe",
      "severity_threshold": "medium"
    }
  }'
```

**Parameters:**
- `project_path` (string, required): Path to project
- `severity_threshold` (string): Minimum severity (`low`, `medium`, `high`, `critical`)

**Response:**
```json
{
  "vulnerabilities": [
    {
      "severity": "high",
      "file": "internal/security/auth.go",
      "line": 45,
      "description": "Potential SQL injection",
      "recommendation": "Use parameterized queries"
    }
  ],
  "summary": {
    "critical": 0,
    "high": 1,
    "medium": 3,
    "low": 5
  }
}
```

---

## 🔑 BYOK (Bring Your Own Key) Support

All Grompt tools support external API keys via the `api_key` parameter:

```json
{
  "tool": "grompt.generate",
  "args": {
    "ideas": ["AI safety", "ethics"],
    "purpose": "Research",
    "provider": "claude",
    "api_key": "sk-ant-YOUR_CLAUDE_KEY"  // ← BYOK
  }
}
```

**Benefits:**
- ✅ No server-side API key storage required
- ✅ Each request can use different keys
- ✅ Perfect for multi-tenant scenarios
- ✅ Supports all AI providers (OpenAI, Claude, Gemini, DeepSeek, ChatGPT)

---

## 🧪 Testing MCP Integration

### Test Script
```bash
#!/bin/bash
# test_mcp_integration.sh

GOBE_URL="http://localhost:3666"
GROMPT_APIKEY="YOUR_API_KEY_HERE"

echo "=== Testing Built-in Tools ==="

# Test system.status
echo "1. Testing system.status..."
curl -s -X POST $GOBE_URL/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "system.status",
    "args": {"detailed": false}
  }' | jq .

# Test shell.command
echo "2. Testing shell.command..."
curl -s -X POST $GOBE_URL/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "shell.command",
    "args": {"command": "uname", "args": ["-a"]}
  }' | jq .

echo ""
echo "=== Testing External Kubex Tools ==="

# Test grompt.direct with BYOK
echo "3. Testing grompt.direct (BYOK)..."
curl -s -X POST $GOBE_URL/mcp/exec \
  -H "Content-Type: application/json" \
  -d "{
    \"tool\": \"grompt.direct\",
    \"args\": {
      \"prompt\": \"Hello from MCP integration test!\",
      \"provider\": \"gemini\",
      \"api_key\": \"$GROMPT_APIKEY\",
      \"max_tokens\": 100
    }
  }" | jq .

# List all available tools
echo ""
echo "4. Listing all MCP tools..."
curl -s -X GET $GOBE_URL/mcp/tools | jq .

echo ""
echo "=== Test Complete ==="
```

**Run tests:**
```bash
chmod +x test_mcp_integration.sh
./test_mcp_integration.sh
```

---

## 🌐 Discord Integration

Use MCP tools via Discord bot commands:

```
# System status
!mcp system.status

# Generate prompt with Grompt
!grompt quantum computing, beginner explanation

# Direct AI prompt
!ask Explain quantum computing

# Analyze project
!analyze /projects/kubex/gobe
```

**Discord Bot Integration** (`internal/proxy/hub/hub.go`):
```go
func (h *DiscordMCPHub) HandleGromptCommand(args []string) {
    result, err := h.mcpRegistry.Exec(context.Background(), "grompt.generate", map[string]interface{}{
        "ideas":    args,
        "purpose":  "Code Generation",
        "provider": "gemini",
    })

    // Format and send to Discord
    h.SendFormattedResponse(result)
}
```

---

## ⚙️ Configuration

### Environment Variables

```bash
# GoBE Server
export PORT=3666
export DEBUG=true

# External Tool URLs
export GROMPT_URL=http://localhost:8080
export ANALYZER_URL=http://localhost:8081

# AI Provider Keys (optional, can use BYOK)
export OPENAI_API_KEY=sk-...
export ANTHROPIC_API_KEY=sk-ant-...
export GEMINI_API_KEY=AIza...
```

### Config File (`~/.kubex/gobe/config.yaml`)
```yaml
server:
  port: 3666
  debug: true

mcp:
  external_tools:
    grompt:
      url: http://localhost:8080
      timeout: 30s
    analyzer:
      url: http://localhost:8081
      timeout: 60s

ai_providers:
  default: gemini
  openai:
    api_key: ${OPENAI_API_KEY}
  gemini:
    api_key: ${GEMINI_API_KEY}
```

---

## 🔒 Security Considerations

### MCP Tool Authentication

- **Built-in tools**: `system.status` (none), `shell.command` (admin)
- **External tools**: `grompt.*` (none), `analyzer.project` (none), `analyzer.security` (admin)

### Shell Command Whitelist

Only safe, read-only commands are allowed. Dangerous commands like `rm`, `mv`, `chmod`, `curl` are blocked.

### BYOK Security

- API keys sent via MCP are only used for that specific request
- Keys are never stored or logged by GoBE
- Use HTTPS in production to encrypt keys in transit

---

## 🐛 Troubleshooting

### "Tool not found: grompt.generate"

**Cause:** Grompt service not running or URL misconfigured

**Solution:**
```bash
# Check if Grompt is running
curl http://localhost:8080/api/health

# Set correct URL
export GROMPT_URL=http://localhost:8080
```

### "Connection refused"

**Cause:** External service not accessible

**Solution:**
```bash
# Verify services are running
ps aux | grep grompt
ps aux | grep analyzer

# Check network connectivity
nc -zv localhost 8080
nc -zv localhost 8081
```

### "API Key not configured"

**Cause:** No API key provided and no server-side config

**Solution:**
Use BYOK by including `api_key` in request:
```json
{
  "tool": "grompt.generate",
  "args": {
    "api_key": "YOUR_KEY_HERE",
    ...
  }
}
```

---

## 📊 Architecture

### MCP Registry Pattern

```go
// Register tools
registry := mcp.NewRegistry()
mcp.RegisterBuiltinTools(registry)
mcp.RegisterExternalTools(registry, config)

// Execute tool
result, err := registry.Exec(ctx, "grompt.generate", args)
```

### Tool Handler Signature

```go
type ToolHandler func(context.Context, map[string]interface{}) (interface{}, error)
```

### External API Integration

```go
// Grompt API call with BYOK
func callGromptAPI(ctx context.Context, config, endpoint, payload, apiKey) {
    req.Header.Set("X-API-Key", apiKey)  // BYOK support
    // ... HTTP request logic
}
```

---

## 🚀 Roadmap

### Future MCP Tools

- ✅ Grompt prompt engineering (v1.3.5)
- ✅ Analyzer code analysis (v1.3.5)
- ⏳ GemX image generation
- ⏳ CI/CD pipeline integration
- ⏳ Database query tool
- ⏳ API testing tool

---

## 📚 Related Documentation

- [Grompt BYOK Guide](/projects/kubex/grompt/docs/BYOK_GUIDE.md)
- [GoBE CLAUDE.md](/projects/kubex/gobe/CLAUDE.md)
- [MCP Protocol Spec](https://modelcontextprotocol.io/)

---

**Version:** 1.3.5
**Last Updated:** 2025-01-20
**License:** MIT
