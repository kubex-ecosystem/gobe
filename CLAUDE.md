# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**Updated:** 2025-01-20 - Includes complete AI provider ecosystem, functional webhook system, enhanced MCP tools with shell commands, and **Kubex Ecosystem Integration** (Grompt + Analyzer via MCP).

## Common Development Commands

### Build and Installation

```bash
# Build the application
make build

# Build for development
make build-dev

# Install the binary and configure environment
make install

# Clean build artifacts
make clean

# Generate and serve documentation
make serve-docs        # Start docs server
make build-docs        # Build documentation static files
```

### Running the Application

```bash
# Start the server (basic)
./gobe start

# Start with custom port and bind address
./gobe start -p 3666 -b "0.0.0.0"

# View all available commands
./gobe --help

# Other useful commands
./gobe stop            # Stop the server
./gobe restart         # Restart all services
./gobe status          # Show server status
./gobe config          # Generate initial config file
./gobe logs            # Display server logs
```

### Testing

```bash
# Run all tests
make test

# Run specific test packages
go test ./internal/controllers/mcp/...
go test ./tests/tests_utils/...

# Test with coverage
go test -cover ./...

# Compile check for specific modules
go build -v ./internal/controllers/mcp/
go build -v ./internal/routes/
```

### Linting and Code Quality

```bash
# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Run golangci-lint if available
golangci-lint run
```

## Architecture Overview

**GoBE** is a modular, secure HTTP backend server built with Go following clean architecture principles and the Kubex ecosystem standards. **Latest version (v1.3.5) includes complete AI provider ecosystem, functional webhook system, and enhanced MCP Protocol implementation.**

### Core Technologies

- **Framework**: Gin HTTP router with comprehensive middleware
- **Database**: GORM with PostgreSQL/SQLite via gdbase factory system
- **Architecture**: Clean Architecture with Repository/Service patterns
- **AI Providers**: OpenAI, Anthropic Claude, Google Gemini, Groq with streaming support
- **MCP Protocol**: Dynamic tool registry with thread-safe execution and shell commands
- **Webhooks**: Production-ready system with AMQP integration and retry logic
- **Security**: Certificate-based authentication, keyring storage, CORS, command whitelisting
- **CLI**: Cobra-based command interface with modular design
- **Build**: Make-based with cross-platform support

### Key Architectural Patterns

#### 1. Modular Interface System (`internal/module/`)

All modules follow the Kubex universal interface pattern defined in `internal/module/module.go`:

- Common methods: `Alias()`, `ShortDescription()`, `LongDescription()`, `Usage()`, `Examples()`, `Active()`, `Module()`, `Execute()`, `Command()`
- Wrapper pattern in `internal/module/wrpr.go` with `RegX()` function for global access
- CLI integration via Cobra commands

#### 2. MCP (Model Context Protocol) Architecture

Follows strict tripé pattern: **Model → Repository → Service → Controller**

**Factory Pattern** (`gdbase/factory/models/mcp/`):

```go
func NewEntityService(entityRepo EntityRepo) EntityService
func NewEntityRepo(db *gorm.DB) EntityRepo
func NewEntityModel(fields...) EntityModel
```

**Repository Layer** (`gdbase/internal/models/mcp/*/repo.go`):

- Interface-based with standard CRUD operations
- Always validate nil models and UUIDs
- Call `Validate()` and `Sanitize()` before database operations

**Service Layer** (`gdbase/internal/models/mcp/*/service.go`):

- Business logic validation
- Error wrapping with context

**Controller Layer** (`gobe/internal/controllers/mcp/*/controller.go`):

- HTTP handlers using Gin
- Standard error response format with `gin.H{"error": err.Error()}`

#### 3. Router System (`internal/app/router/`)

Interface-based architecture with centralized route management:

```go
func NewEntityRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
    // Always validate router is not nil
    // Extract database service from router interface
    // Create controller with GORM database connection
    // Use descriptive route keys for debugging
}
```

#### 4. AI Provider System (`internal/services/gateway/providers/`)

**New in v1.3.5** - Complete AI provider ecosystem with streaming support:

**Factory Pattern** (`factory.go`):

```go
func New(cfg Config) (gateway.Provider, error) {
    switch cfg.Type {
    case "openai":
        return newOpenAIProvider(cfg)
    case "anthropic":
        return newAnthropicProvider(cfg)
    case "gemini":
        return newGeminiProvider(cfg)
    case "groq":
        return newGroqProvider(cfg)
    }
}
```

**Provider Implementation**:

- OpenAI: GPT-3.5/4 with custom base URLs
- Anthropic: Claude 3.5 Sonnet, Opus, Haiku with proper message conversion
- Google Gemini: 1.5 Pro/Flash with multimodal support
- Groq: Llama/Mixtral with ultra-fast inference

**Features**:

- Streaming responses via Server-Sent Events
- Automatic cost calculation per provider
- External API key support per request
- Thread-safe provider switching

#### 5. MCP (Model Context Protocol) System

**Enhanced in v1.3.5** - Complete MCP implementation with shell command support:

**Registry Architecture** (`internal/services/mcp/registry.go`):

- Thread-safe tool registration and execution
- Dynamic tool discovery without server restart
- Interface-based design for extensibility

**Built-in Tools** (`internal/services/mcp/register_builtin.go`):

- `system.status` - Comprehensive system metrics with connection health
- `shell.command` - Secure shell command execution with whitelist
- Integration with existing manage controllers
- Runtime registration in MetricsController

**Security Features**:

- Command whitelist: `ls`, `pwd`, `date`, `uname`, `df`, `free`, etc.
- 10-second timeout on all commands
- Proper error handling and output capture
- Admin-level authentication required

**API Endpoints** (via `internal/app/controllers/mcp/system/`):

- `GET /mcp/tools` - List available tools
- `POST /mcp/exec` - Execute tools with arguments
- Uses APIWrapper for consistent response format

**External Kubex Ecosystem Integration** (`internal/services/mcp/register_external.go`):

**NEW in v1.3.5** - GoBE now acts as a central MCP hub for the entire Kubex ecosystem:

**Grompt Tools** (Prompt Engineering):
- `grompt.generate` - Generate structured prompts from raw ideas with BYOK support
- `grompt.direct` - Direct AI prompts without prompt engineering

**Analyzer Tools** (Code Analysis):
- `analyzer.project` - Deep project structure and dependency analysis
- `analyzer.security` - Security audit and vulnerability detection

**Configuration**:
```bash
export GROMPT_URL=http://localhost:8080   # Default Grompt endpoint
export ANALYZER_URL=http://localhost:8081 # Default Analyzer endpoint
```

**Example Usage**:
```bash
# Generate prompt with BYOK
curl -X POST http://localhost:3666/mcp/exec \
  -d '{
    "tool": "grompt.generate",
    "args": {
      "ideas": ["quantum computing", "beginner tutorial"],
      "purpose": "Educational Content",
      "provider": "gemini",
      "api_key": "AIza..."
    }
  }'

# Analyze project security
curl -X POST http://localhost:3666/mcp/exec \
  -d '{
    "tool": "analyzer.security",
    "args": {
      "project_path": "/projects/kubex/gobe",
      "severity_threshold": "medium"
    }
  }'
```

See [MCP Integration Guide](docs/MCP_INTEGRATION.md) for complete documentation.

#### 6. Webhook System (`internal/services/webhooks/`)

**New in v1.3.5** - Production-ready webhook processing:

**Service Architecture** (`webhook_service.go`):

```go
type WebhookService struct {
    amqp      *messagery.AMQP
    events    []WebhookEvent
    mu        sync.RWMutex
    ctx       context.Context
    // Background processing with worker
}
```

**Features**:

- In-memory event storage with UUID tracking
- AMQP/RabbitMQ integration for async processing
- Background worker for event processing
- Specialized handlers for different webhook types
- Retry logic for failed events

**Event Types**:

- GitHub push events with repository notifications
- Discord message events for bot integration
- Stripe payment events for billing
- Generic user events

**API Endpoints** (`internal/app/controllers/gateway/webhooks.go`):

- `POST /v1/webhooks` - Receive webhook events
- `GET /v1/webhooks/health` - System health with statistics
- `GET /v1/webhooks/events` - List events with pagination
- `GET /v1/webhooks/events/:id` - Get specific event details
- `POST /v1/webhooks/retry` - Retry failed events

#### 7. Discord Integration System (`internal/proxy/hub/`)

**Enhanced in v1.3.5** - Real MCP tool integration:

**Hub Architecture** (`hub.go`):

```go
type DiscordMCPHub struct {
    mcpRegistry mcp.Registry // Real MCP registry integration
    // Discord bot connection and command handling
}
```

**Features**:

- Real MCP tool execution via Discord commands
- Discord-friendly formatting of technical data
- Command mapping: Discord commands → MCP tool names
- Rich formatting with emojis and structured output
- Error handling with user-friendly messages

**Command Examples**:

- `!system status` → `system.status` MCP tool
- `!shell ls` → `shell.command` with `ls` command
- Auto-formatted responses for system metrics

#### 8. Security System

- Zero-config security with automatic certificate generation
- Keyring-based secret storage using system keyring
- Comprehensive CORS headers and security middleware
- JWT-based authentication with certificates
- Command whitelisting for shell execution
- Input validation and sanitization across all endpoints

### Project Structure

```plaintext
gobe/
├── cmd/                          # CLI entry points
│   ├── main.go                   # Main CLI entrypoint (uses module.RegX().Command().Execute())
│   ├── cli/                      # CLI command implementations
│   └── swagger/                  # Swagger documentation server
├── internal/
│   ├── module/                   # Kubex module interface implementation
│   │   ├── module.go            # Universal module interface
│   │   ├── wrpr.go              # RegX() wrapper function
│   │   ├── info/                # Application info and banners
│   │   └── logger/              # Global logger (import as gl)
│   ├── app/
│   │   ├── router/              # HTTP router implementation
│   │   ├── controllers/         # HTTP handlers
│   │   ├── middlewares/         # HTTP middleware
│   │   └── security/            # Security services
│   ├── contracts/               # Interface definitions
│   │   ├── interfaces/          # Core interfaces
│   │   └── types/               # Type definitions
│   ├── services/                # Business services
│   │   ├── mcp/                 # MCP Protocol implementation
│   │   ├── gateway/             # AI provider gateway and registry
│   │   │   └── providers/       # AI provider implementations (OpenAI, Anthropic, Gemini, Groq)
│   │   └── webhooks/           # Webhook processing service
│   ├── bridges/                 # External service bridges (gdbase)
│   ├── config/                  # Configuration management
│   ├── observers/               # Event and approval managers
│   ├── proxy/                   # Proxy functionality (Discord integration)
│   ├── sockets/                 # Socket connections (AMQP/RabbitMQ)
│   ├── utils/                   # Utility functions
│   └── commons/                 # Common shared components
├── tests/                        # Test files (table-driven tests)
├── factory/                      # Dependency injection factories
├── support/                      # Build scripts and utilities
├── docs/                         # Documentation assets
├── config/                       # Configuration files
└── web/                          # Web assets and static files
```

### Key Dependencies and Modules

#### Internal Modules

- **gdbase**: Database layer (separate module) - handles database management via Docker
- **logz**: Logging wrapper (import as `l` or `gl` for global logger)

#### External Dependencies

- **Gin**: HTTP router and middleware
- **GORM**: ORM with PostgreSQL/SQLite support
- **Cobra**: CLI framework
- **Viper**: Configuration management
- **Discord/Telegram/WhatsApp**: Chatbot integrations
- **MCP Protocol**: Model Context Protocol support
- **AI Providers**: OpenAI, Anthropic, Google Gemini, Groq APIs
- **RabbitMQ**: Message queue for webhooks and notifications via AMQP

### Configuration and Environment

#### Zero-Config Startup

- Automatically generates certificates and passwords
- Creates default configuration files
- Sets up database connections via gdbase
- Stores secrets in system keyring

#### Environment Variables

- Follow `GOBE_*` prefix pattern
- `DEBUG=true` enables debug logging
- `GIN_MODE=release` for production mode

**AI Provider Configuration**:

- `OPENAI_API_KEY` - OpenAI API key
- `ANTHROPIC_API_KEY` - Anthropic Claude API key
- `GEMINI_API_KEY` - Google Gemini API key
- `GROQ_API_KEY` - Groq API key
- `OPENAI_BASE_URL` - Custom OpenAI-compatible base URL

#### Configuration Files

- Default config: `~/.kubex/gobe/config.yaml`
- Certificate storage: `~/.kubex/gobe/cert.json`
- Manifest: `internal/module/info/manifest.json`
- Messaging integrations: `config/discord_config.json` (includes WhatsApp/Telegram settings)

### Development Guidelines

#### Adding New AI Providers

1. Create provider implementation in `internal/services/gateway/providers/`
2. Follow existing provider pattern (OpenAI, Anthropic, Gemini, Groq)
3. Implement required interface methods: `Name()`, `Available()`, `Chat()`, `Notify()`
4. Add streaming support with Server-Sent Events
5. Include cost calculation logic
6. Update factory in `factory.go`
7. Add configuration support via environment variables

#### Adding New MCP Tools

1. Add tool specification in `internal/services/mcp/register_builtin.go`
2. Implement handler function with proper error handling
3. Follow security guidelines (whitelist, timeout, validation)
4. Add tool to registry during initialization
5. Update tests in `tests/tests_mcp/`

#### Adding New Webhook Handlers

1. Add event type to `internal/services/webhooks/webhook_service.go`
2. Implement specialized handler function
3. Add AMQP integration if needed
4. Update controller routes in `internal/app/controllers/gateway/webhooks.go`
5. Add proper response types

#### Adding New MCP Controllers

1. Create factory functions in `gdbase/factory/models/mcp/`
2. Implement model in `gdbase/internal/models/mcp/entity/`
3. Implement repository with interface validation
4. Implement service with business logic
5. Create controller in `gobe/internal/controllers/mcp/`
6. Add routes in `gobe/internal/routes/`
7. Register routes in router system

#### Code Style Standards

- Follow Kubex AGENTS.md universal standards (see AGENTS.md)
- Use interfaces for dependency injection
- Always validate nil pointers before dereferencing
- Use descriptive error messages with context
- Import global logger as `gl "github.com/kubex-ecosystem/logz/logger"`
- Follow table-driven testing patterns
- Use `context.Context` for cancellation and timeouts
- All modules follow universal interface pattern with methods: `Alias()`, `ShortDescription()`, `LongDescription()`, `Usage()`, `Examples()`, `Active()`, `Module()`, `Execute()`, `Command()`
- Wrapper pattern via `RegX()` function for global module access

#### Error Handling

- Repository level: `fmt.Errorf("Repository: operation failed: %w", err)`
- Service level: `fmt.Errorf("validation error: %w", err)`
- Controller level: `ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})`

#### Security Considerations

- Never commit secrets or API keys
- Use keyring for sensitive data storage
- Validate all inputs at controller level
- Always sanitize data before database operations
- Follow secure coding practices for JWT handling
- **AI Provider Security**: Support external API keys per request via headers
- **Shell Command Security**: Strict whitelist of allowed commands with timeout
- **Webhook Security**: Validate headers and implement AMQP for async processing
- **MCP Security**: Admin-level authentication for sensitive tools

### Testing Patterns

- Use table-driven tests with `testing` package
- External test packages with `_test` suffix
- Cover error paths and edge cases
- Mock dependencies via interfaces
- Test files located in `tests/` directory

### Build System

- Make-based build system that reads from `manifest.json`
- Cross-platform support (Linux, macOS, Windows)
- Automated GitHub Actions CI/CD
- Support for multiple architectures (AMD64, ARM64)

This project prioritizes security, modularity, and clean interfaces following the Kubex ecosystem standards.

## Recent Major Changes (v1.3.5)

### AI Provider Implementation

- **OpenAI Provider**: Complete with custom base URL support and GPT-4 integration
- **Anthropic Provider**: Claude 3.5 Sonnet with proper message format conversion
- **Google Gemini Provider**: 1.5 Pro/Flash with cost calculation and streaming
- **Groq Provider**: Llama/Mixtral with ultra-fast inference

### Enhanced MCP System

- **Shell Command Tool**: Secure execution with whitelist (`ls`, `pwd`, `date`, `uname`, etc.)
- **System Status Tool**: Enhanced with connection health checks (DB, AMQP, webhooks)
- **Discord Integration**: Real tool execution via Discord bot with formatted responses

### Production Webhook System

- **Event Processing**: In-memory storage with background worker processing
- **AMQP Integration**: Async processing via RabbitMQ with proper queue bindings
- **Retry Logic**: Automatic retry of failed webhook events
- **REST API**: Complete CRUD operations with pagination support

### Security Enhancements

- **Command Whitelisting**: Only safe shell commands allowed
- **External API Keys**: Support for per-request API key override
- **Input Validation**: Comprehensive validation across all endpoints
- **Timeout Protection**: All operations have appropriate timeouts

## Key Files Modified/Created

**New Files**:

- `internal/services/gateway/providers/anthropic.go`
- `internal/services/gateway/providers/gemini.go`
- `internal/services/webhooks/webhook_service.go`

**Modified Files**:

- `internal/services/gateway/providers/factory.go` - Added Anthropic and Gemini
- `internal/proxy/hub/hub.go` - Real MCP integration with Discord formatting
- `internal/services/mcp/register_builtin.go` - Added shell.command tool
- `internal/app/controllers/gateway/webhooks.go` - Complete webhook API
- `internal/app/router/gateway/routes.go` - Webhook service integration

This ensures GoBE v1.3.5 provides a complete, production-ready backend with AI integration, functional webhooks, and enhanced security.
