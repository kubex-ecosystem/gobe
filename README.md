<!-- ---
title: GoBE - Modular & Secure Back-end
version: 1.3.4
owner: kubex
audience: dev
languages: [en, pt-BR]
sources: [internal/module/info/manifest.json, https://github.com/kubex-ecosystem/gobe]
assumptions: []
--- -->

<!-- markdownlint-disable MD013 MD025 -->
# GoBE - Modular & Secure Back-end

![GoBE Banner](docs/assets/top_banner_lg_b.png)

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](https://github.com/kubex-ecosystem/gobe/blob/main/LICENSE)
[![Automation](https://img.shields.io/badge/automation-zero%20config-blue)](#features)
[![Modular](https://img.shields.io/badge/modular-yes-yellow)](#features)
[![Security](https://img.shields.io/badge/security-high-red)](#features)
[![MCP](https://img.shields.io/badge/MCP-enabled-orange)](#mcp-support)
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg)](https://github.com/kubex-ecosystem/gobe/blob/main/CONTRIBUTING.md)
[![Build](https://github.com/kubex-ecosystem/gobe/actions/workflows/kubex_go_release.yml/badge.svg)](https://github.com/kubex-ecosystem/gobe/actions/workflows/kubex_go_release.yml)

**Code Fast. Own Everything.** — A modular, secure, and zero-config backend for modern Go applications.

## TL;DR

GoBE is a modular Go backend that runs **with zero configuration** and provides ready-to-use REST APIs + MCP (Model Context Protocol). One command = a complete server with authentication, database, and integrated system tools.

```bash
make build && ./gobe start  # Zero config, instant backend
curl http://localhost:3666/mcp/tools  # MCP tools ready
```

## **Table of Contents**

1. [About the Project](#about-the-project)
2. [Features](#features)
3. [How to Run](#how-to-run)
4. [MCP Support](#mcp-support)
5. [Usage](#usage)
    - [CLI](#cli)
    - [Configuration](#configuration)
6. [Roadmap](#roadmap)
7. [Contributing](#contributing)
8. [Contact](#contact)

---

## **About the Project**

GoBE is a modular backend built with Go that embodies the Kubex principle: **No Lock-in. No Excuses.** It delivers **security, automation, and flexibility** in a single binary that runs anywhere — from your laptop to enterprise clusters.

### **Mission Alignment**

Following Kubex's mission to democratize modular technology, GoBE provides:

- **DX First:** One command starts everything — server, database, authentication, MCP tools
- **Total Accessibility:** Runs without Kubernetes, Docker, or complex setup
- **Module Independence:** Every component (CLI/HTTP/Jobs/Events) is a full citizen

### **Current Status - Production Ready**

✅ **Zero-config:** Auto-generates certificates, passwords, keyring storage
✅ **MCP Protocol:** Model Context Protocol with dynamic tool registry
✅ **Modular Architecture:** Clean interfaces, exportable via `factory/`
✅ **Database Integration:** PostgreSQL/SQLite via `gdbase` Docker management
✅ **REST API:** Authentication, users, products, clients, jobs, webhooks
✅ **Security Stack:** Dynamic certificates, JWT, keyring, rate limiting
✅ **CLI Interface:** Complete management via Cobra commands
✅ **Multi-platform:** Linux, macOS, Windows (AMD64, ARM64)
✅ **Testing:** Unit tests + integration tests for MCP endpoints
✅ **CI/CD:** Automated builds and releases

## **Project Evolution**

The project has undergone significant evolution since its inception. Initially focused on basic functionalities, it has now expanded to include a comprehensive feature set that enhances security, modularity, and ease of use.

The current version of GoBE represents continuous improvements and refinements, with strong emphasis on security and automation. The system is designed to be developer-friendly, allowing teams to focus on building applications without worrying about backend complexities.

The modular architecture enables seamless integration with other systems, making GoBE a versatile choice for modern Go applications. The project is actively maintained with ongoing efforts to enhance capabilities and meet evolving developer needs.

Documentation and CI/CD remain key focus areas for upcoming updates.

---

## **Features**

✨ **Fully modular**

- All logic follows well-defined interfaces, ensuring encapsulation.
- Can be used as a server or as a library/module.

🔒 **Zero-config, but customizable**

- Runs without initial configuration, but supports customization via files.
- Automatically generates certificates, passwords, and secure settings.

🔗 **Direct integration with `gdbase`**

- Database management via Docker.
- Automatic optimizations for persistence and performance.

🛡️ **Advanced authentication**

- Dynamically generated certificates.
- Random passwords and secure keyring.

🌐 **Robust REST API**

- Endpoints for authentication, user management, products, clients, and cronjobs.

📋 **Log and security management**

- Protected routes, secure storage, and request monitoring.

🧑‍💻 **Powerful CLI**

- Commands to start, configure, and monitor the server.

---

## **How to Run**

**One Command. All the Power.**

### Quick Start

```bash
# Clone and build
git clone https://github.com/kubex-ecosystem/gobe.git
cd gobe
make build

# Start everything (zero config)
./gobe start

# Server ready at http://localhost:3666
# MCP endpoints: /mcp/tools, /mcp/exec
# Health check: /health
```

### Requirements

- **Go 1.24+** (for building from source)
- **Docker** (optional, for advanced database features)
- **No Kubernetes, no complex setup required**

### Build Options

```bash
make build          # Production build
make build-dev      # Development build
make install        # Install binary + environment setup
make clean          # Clean artifacts
make test           # Run all tests
```

---

## **MCP Support**

GoBE implements the **Model Context Protocol (MCP)** for seamless AI tool integration.

### Available Endpoints

```bash
GET  /mcp/tools     # List available tools
POST /mcp/exec      # Execute tools
```

### Built-in Tools

| Tool | Description | Args |
|------|-------------|------|
| `system.status` | Comprehensive system status | `detailed: boolean` |

### Example Usage

```bash
# List available tools
curl http://localhost:3666/mcp/tools

# Execute system status
curl -X POST http://localhost:3666/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{"tool": "system.status", "args": {"detailed": true}}'
```

### MCP Architecture

- **Dynamic Registry:** Tools registered at runtime, no restart needed
- **Thread-Safe:** Concurrent tool execution with RWMutex protection
- **Extensible:** Add custom tools via the Registry interface
- **Integrated:** Uses existing manage controllers for system data

---

## **Usage**

### CLI

Start the main server:

```sh
./gobe start -p 3666 -b "0.0.0.0"
```

This starts the server, generates certificates, sets up databases, and begins listening for requests!

See all available commands:

```sh
./gobe --help
```

**Main commands:**

| Command   | Function                                         |
|-----------|--------------------------------------------------|
| `start`   | Starts the server                                |
| `stop`    | Safely stops the server                          |
| `restart` | Restarts all services                            |
| `status`  | Shows the status of the server and active services|
| `config`  | Generates an initial configuration file          |
| `logs`    | Displays server logs                             |

---

### Configuration

GoBE can run without any initial configuration, but supports customization via YAML/JSON files. By default, everything is generated automatically on first use.

Example configuration:

```yaml
port: 3666
bindAddress: 0.0.0.0
database:
  type: postgres
  host: localhost
  port: 5432
  user: gobe
  password: secure
```

#### Messaging Integrations

WhatsApp and Telegram bots can be configured via the `config/discord_config.json` file under the `integrations` section:

```json
{
  "integrations": {
    "whatsapp": {
      "enabled": true,
      "access_token": "<token>",
      "verify_token": "<verify>",
      "phone_number_id": "<number>",
      "webhook_url": "https://your.server/whatsapp/webhook"
    },
    "telegram": {
      "enabled": true,
      "bot_token": "<bot token>",
      "webhook_url": "https://your.server/telegram/webhook",
      "allowed_updates": ["message", "callback_query"]
    }
  }
}
```

After setting up the file or environment variables, the server will expose the following endpoints:

- `POST /api/v1/whatsapp/send` and `/api/v1/whatsapp/webhook`
- `POST /api/v1/telegram/send` and `/api/v1/telegram/webhook`

Each route also provides a `/ping` endpoint for health checks.

---

## **Roadmap**

### ✅ **Completed (v1.3.3)**

- [x] Full modularization and pluggable interfaces
- [x] Zero-config with automatic certificate generation
- [x] Integration with system keyring
- [x] REST API for authentication and management
- [x] Authentication via certificates and secure passwords
- [x] CLI for management and monitoring
- [x] Integration with `gdbase` for database management via Docker
- [x] **MCP Protocol implementation with dynamic tool registry**
- [x] **Built-in system.status tool with runtime metrics**
- [x] **Thread-safe tool execution with comprehensive testing**
- [x] Multi-database support (PostgreSQL, SQLite)
- [x] Complete documentation following Kubex standards
- [x] Automated tests and CI/CD

### 🚧 **In Progress**

- [ ] Extended MCP tool library (file operations, network tools)
- [ ] Prometheus integration for monitoring
- [ ] Grafana integration for metrics visualization

### 📋 **Planned**

- [ ] Support for custom middlewares
- [ ] WebSocket support for real-time MCP communication
- [ ] Plugin system for external tool registration
- [ ] Advanced security policies and RBAC

### 🎯 **Next Milestones**

1. **v1.4.0** - Extended MCP tools and WebSocket support
2. **v1.5.0** - Monitoring stack integration (Prometheus/Grafana)
3. **v2.0.0** - Plugin architecture and advanced security

---

## **Contributing**

Contributions are welcome! Feel free to open issues or submit pull requests. See the [Contribution Guide](docs/CONTRIBUTING.md) for more details.

---

## **Contact**

💌 **Developer**:
[Rafael Mori](mailto:faelmori@gmail.com)
💼 [Follow me on GitHub](https://github.com/kubex-ecosystem)
I'm open to collaborations and new ideas. If you found the project interesting, get in touch!

---

## **Risks & Mitigations**

• **Zero-config may hide necessary configurations** → Verbose logs + override documentation
• **MCP registry thread-safety** → RWMutex implemented + concurrency tests
• **Dependency on gdbase for DB** → Fallback SQLite always available
• **Certificate auto-generation** → Keyring backup + automatic regeneration

---

## **Next Steps**

1. **Extend MCP tools** - file operations, network diagnostics, database queries
2. **WebSocket MCP** - real-time tool communication for AI agents
3. **Plugin system** - external tool registration via shared libraries
4. **Advanced monitoring** - Prometheus metrics + Grafana dashboards
5. **Security hardening** - RBAC, audit logs, policy enforcement

---

## **Changelog**

### v1.3.4 (2025-12-23)

- ✅ MCP Protocol implementation with dynamic registry
- ✅ Built-in system.status tool with runtime metrics
- ✅ Thread-safe tool execution with comprehensive tests
- ✅ Documentation updated following Kubex standards
- ✅ Zero-config MCP endpoints (/mcp/tools, /mcp/exec)
