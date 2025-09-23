---
title: GoBE - Modular & Secure Back-end
version: 1.3.4
owner: kubex
audience: dev
languages: [pt-BR, en]
sources: [internal/module/info/manifest.json, https://github.com/kubex-ecosystem/gobe]
assumptions: []
---
<!-- markdownlint-disable MD013 MD025 -->
# GoBE - Modular & Secure Back-end

![GoBE Banner](/docs/assets/top_banner_lg_b.png)

[![Build Status](https://img.shields.io/github/actions/workflow/status/kubex-ecosystem/gobe/release.yml?branch=main)](https://github.com/kubex-ecosystem/gobe/actions)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](https://github.com/kubex-ecosystem/gobe/blob/main/LICENSE)
[![Automation](https://img.shields.io/badge/automation-zero%20config-blue)](#features)
[![Modular](https://img.shields.io/badge/modular-yes-yellow)](#features)
[![Security](https://img.shields.io/badge/security-high-red)](#features)
[![MCP](https://img.shields.io/badge/MCP-enabled-orange)](#suporte-mcp)
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg)](https://github.com/kubex-ecosystem/gobe/blob/main/CONTRIBUTING.md)

**Code Fast. Own Everything.** — Backend modular, seguro e zero-config para aplicações Go modernas.

## TL;DR

GoBE é um backend Go modular que roda **sem configuração** e oferece APIs REST + MCP (Model Context Protocol) prontas para uso. Um comando = servidor completo com autenticação, banco de dados, e ferramentas de sistema integradas.

```bash
make build && ./gobe start  # Zero config, backend instantâneo
curl http://localhost:3666/mcp/tools  # Ferramentas MCP prontas
```

---

## **Índice**

1. [Sobre o Projeto](#sobre-o-projeto)
2. [Funcionalidades](#features)
3. [Como Executar](#como-executar)
4. [Suporte MCP](#suporte-mcp)
5. [Uso](#usage
    - [CLI](#cli)
    - [Configuração](#configuration)
6. [Roadmap](#roadmap)
7. [Contribuindo](#contributing)
8. [Contato](#contact)

---

## **Sobre o Projeto**

GoBE é um backend modular construído em Go que incorpora o princípio Kubex: **No Lock-in. No Excuses.** Oferece **segurança, automação e flexibilidade** em um único binário que roda em qualquer lugar — do seu laptop a clusters enterprise.

### **Alinhamento com a Missão**

Seguindo a missão Kubex de democratizar tecnologia modular, GoBE oferece:

- **DX Primeiro:** Um comando inicia tudo — servidor, banco, autenticação, ferramentas MCP
- **Acessibilidade Total:** Roda sem Kubernetes, Docker ou setup complexo
- **Independência Modular:** Cada componente (CLI/HTTP/Jobs/Events) é cidadão pleno

### **Status Atual - Pronto para Produção**

✅ **Zero-config:** Auto-gera certificados, senhas, armazenamento keyring
✅ **Protocolo MCP:** Model Context Protocol com registry dinâmico de ferramentas
✅ **Arquitetura Modular:** Interfaces limpas, exportável via `factory/`
✅ **Integração Banco:** PostgreSQL/SQLite via gerenciamento Docker `gdbase`
✅ **API REST:** Autenticação, usuários, produtos, clientes, jobs, webhooks
✅ **Stack Segurança:** Certificados dinâmicos, JWT, keyring, rate limiting
✅ **Interface CLI:** Gerenciamento completo via comandos Cobra
✅ **Multi-plataforma:** Linux, macOS, Windows (AMD64, ARM64)
✅ **Testes:** Testes unitários + integração para endpoints MCP
✅ **CI/CD:** Builds e releases automatizados

---

## **Features**

✨ **Totalmente modular**

- Todas as lógicas seguem interfaces bem definidas, garantindo encapsulamento.
- Pode ser usado como servidor ou como biblioteca/módulo.

🔒 **Zero-config, mas personalizável**

- Roda sem configuração inicial, mas aceita customização via arquivos.
- Gera certificados, senhas e configurações seguras automaticamente.

🔗 **Integração direta com `gdbase`**

- Gerenciamento de bancos de dados via Docker.
- Otimizações automáticas para persistência e performance.

🛡️ **Autenticação avançada**

- Certificados gerados dinamicamente.
- Senhas aleatórias e keyring seguro.

🌐 **API REST robusta**

- Endpoints para autenticação, gerenciamento de usuários, produtos, clientes e cronjobs.

📋 **Gerenciamento de logs e segurança**

- Rotas protegidas, armazenamento seguro e monitoramento de requisições.

🧑‍💻 **CLI poderosa**

- Comandos para iniciar, configurar e monitorar o servidor.

---

## **Como Executar**

**One Command. All the Power.**

### Início Rápido

```bash
# Clone e compile
git clone https://github.com/kubex-ecosystem/gobe.git
cd gobe
make build

# Inicia tudo (zero config)
./gobe start

# Servidor pronto em http://localhost:3666
# Endpoints MCP: /mcp/tools, /mcp/exec
# Health check: /health
```

### Requisitos

- **Go 1.24+** (para compilar do código fonte)
- **Docker** (opcional, para recursos avançados de banco)
- **Não precisa Kubernetes nem setup complexo**

### Opções de Build

```bash
make build          # Build produção
make build-dev      # Build desenvolvimento
make install        # Instala binário + setup ambiente
make clean          # Limpa artefatos
make test           # Executa todos os testes
```

---

## **Suporte MCP**

GoBE implementa o **Model Context Protocol (MCP)** para integração perfeita com ferramentas de IA.

### Endpoints Disponíveis

```bash
GET  /mcp/tools     # Lista ferramentas disponíveis
POST /mcp/exec      # Executa ferramentas
```

### Ferramentas Built-in

| Ferramenta | Descrição | Args |
|------------|-----------|------|
| `system.status` | Status abrangente do sistema | `detailed: boolean` |

### Exemplo de Uso

```bash
# Lista ferramentas disponíveis
curl http://localhost:3666/mcp/tools

# Executa status do sistema
curl -X POST http://localhost:3666/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{"tool": "system.status", "args": {"detailed": true}}'
```

### Arquitetura MCP

- **Registry Dinâmico:** Ferramentas registradas em runtime, sem restart
- **Thread-Safe:** Execução concorrente com proteção RWMutex
- **Extensível:** Adicione ferramentas customizadas via interface Registry
- **Integrado:** Usa controllers manage existentes para dados do sistema

---

## **Usage**

### CLI

Inicie o servidor principal:

```sh
./gobe start -p 3666 -b "0.0.0.0"
```

Isso inicializa o servidor, gera certificados, configura bancos de dados e começa a escutar requisições!

Veja todos os comandos disponíveis:

```sh
./gobe --help
```

**Principais comandos:**

| Comando   | Função                                             |
|-----------|----------------------------------------------------|
| `start`   | Inicializa o servidor                              |
| `stop`    | Encerra o servidor de forma segura                 |
| `restart` | Reinicia todos os serviços                         |
| `status`  | Exibe o status do servidor e dos serviços ativos   |
| `config`  | Gera um arquivo de configuração inicial            |
| `logs`    | Exibe os logs do servidor                          |

---

### Configuration

O GoBE pode rodar sem configuração inicial, mas aceita customização via arquivos YAML/JSON. Por padrão, tudo é gerado automaticamente no primeiro uso.

Exemplo de configuração:

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

---

## **Roadmap**

- [x] Modularização total e interfaces plugáveis
- [x] Zero-config com geração automática de certificados
- [x] Integração com keyring do sistema
- [x] API REST para autenticação e gerenciamento
- [x] Autenticação via certificados e senhas seguras
- [x] CLI para gerenciamento e monitoramento
- [x] Integração com `gdbase` para gerenciamento de bancos via Docker
- [–] Suporte a múltiplos bancos de dados (Parcial concluído)
- [&nbsp;&nbsp;] Integração com Prometheus para monitoramento
- [&nbsp;&nbsp;] Suporte a middlewares personalizados
- [&nbsp;&nbsp;] Integração com Grafana para visualização de métricas
- [–] Documentação completa e exemplos de uso (Parcial concluído)
- [–] Testes automatizados e CI/CD (Parcial concluído)

---

## **Contributing**

Contribuições são bem-vindas! Sinta-se à vontade para abrir issues ou enviar pull requests. Veja o [Guia de Contribuição](docs/CONTRIBUTING.md) para mais detalhes.

---

## **Contact**

💌 **Developer**:
[Rafael Mori](mailto:faelmori@gmail.com)
💼 [Follow me on GitHub](https://github.com/kubex-ecosystem)
Estou aberto a colaborações e novas ideias. Se achou o projeto interessante, entre em contato!

---

## **Riscos & Mitigações**

• **Zero-config pode mascarar configurações necessárias** → Logs verbosos + documentação de override
• **MCP registry thread-safety** → RWMutex implementado + testes de concorrência
• **Dependência do gdbase para DB** → Fallback SQLite sempre disponível
• **Auto-geração de certificados** → Backup keyring + regeneração automática

---

## **Próximos Passos**

1. **Estender ferramentas MCP** - operações de arquivo, diagnósticos de rede, queries de banco
2. **WebSocket MCP** - comunicação em tempo real para agentes IA
3. **Sistema de plugins** - registro de ferramentas externas via bibliotecas compartilhadas
4. **Monitoramento avançado** - métricas Prometheus + dashboards Grafana
5. **Hardening de segurança** - RBAC, logs de auditoria, enforcement de políticas

---

## **Changelog**

### v1.3.4 (2025-09-23)

- ✅ Implementação do Protocolo MCP com registry dinâmico
- ✅ Ferramenta built-in system.status com métricas de runtime
- ✅ Execução thread-safe de ferramentas com testes abrangentes
- ✅ Documentação atualizada seguindo padrões Kubex
- ✅ Endpoints MCP zero-config (/mcp/tools, /mcp/exec)
