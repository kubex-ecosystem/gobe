<!-- ---
title: GoBE - Backend Modular & Seguro
version: 1.3.5
owner: kubex
audience: dev
languages: [en, pt-BR]
sources: [internal/module/info/manifest.json, https://github.com/kubex-ecosystem/gobe]
assumptions: []
--- -->

<!-- markdownlint-disable MD013 MD025 -->
# GoBE - Backend Modular & Seguro

![GoBE Banner](docs/assets/top_banner_lg_b.png)

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Licença: MIT](https://img.shields.io/badge/licen%C3%A7a-MIT-green.svg)](https://github.com/kubex-ecosystem/gobe/blob/main/LICENSE)
[![Automação](https://img.shields.io/badge/automação-zero%20config-blue)](#recursos)
[![Modular](https://img.shields.io/badge/modular-sim-yellow)](#recursos)
[![Segurança](https://img.shields.io/badge/segurança-alta-red)](#recursos)
[![MCP](https://img.shields.io/badge/MCP-habilitado-orange)](#suporte-mcp)
[![Provedores IA](https://img.shields.io/badge/Provedores%20IA-4-purple)](#provedores-ia)
[![Webhooks](https://img.shields.io/badge/webhooks-funcional-green)](#webhooks)
[![Contribuições Bem-vindas](https://img.shields.io/badge/contribuições-bem--vindas-brightgreen.svg)](https://github.com/kubex-ecosystem/gobe/blob/main/CONTRIBUTING.md)
[![Build](https://github.com/kubex-ecosystem/gobe/actions/workflows/kubex_go_release.yml/badge.svg)](https://github.com/kubex-ecosystem/gobe/actions/workflows/kubex_go_release.yml)

**Programe Rápido. Seja Dono de Tudo.** — Um backend modular, seguro e com configuração zero para aplicações Go modernas com integração completa de IA.

## Resumo

GoBE é um backend modular em Go que funciona **com zero configuração** e fornece APIs REST prontas para uso, provedores de IA, webhooks e MCP (Model Context Protocol). Um comando = servidor completo com autenticação, banco de dados, 4 provedores de IA e ferramentas de sistema integradas.

```bash
make build && ./gobe start  # Zero config, backend instantâneo com IA
curl http://localhost:3666/mcp/tools  # Ferramentas MCP prontas
curl http://localhost:3666/providers  # OpenAI, Anthropic, Gemini, Groq prontos
curl http://localhost:3666/v1/webhooks/health  # Sistema de webhook pronto
```

## **Índice**

1. [Sobre o Projeto](#sobre-o-projeto)
2. [Recursos](#recursos)
3. [Como Executar](#como-executar)
4. [Provedores IA](#provedores-ia)
5. [Suporte MCP](#suporte-mcp)
6. [Webhooks](#webhooks)
7. [Uso](#uso)
    - [CLI](#cli)
    - [Configuração](#configuração)
8. [Referência da API](#referência-da-api)
9. [Roteiro](#roteiro)
10. [Contribuindo](#contribuindo)
11. [Contato](#contato)

---

## **Sobre o Projeto**

GoBE é um backend modular construído com Go que incorpora o princípio Kubex: **Sem Aprisionamento. Sem Desculpas.** Ele entrega **segurança, automação e flexibilidade** em um binário único que roda em qualquer lugar — do seu laptop a clusters empresariais.

### **Alinhamento da Missão**

Seguindo a missão da Kubex de democratizar tecnologia modular, GoBE fornece:

- **DX Primeiro:** Um comando inicia tudo — servidor, banco de dados, autenticação, ferramentas MCP
- **Acessibilidade Total:** Funciona sem Kubernetes, Docker ou configuração complexa
- **Independência de Módulo:** Todo componente (CLI/HTTP/Jobs/Eventos) é um cidadão pleno

### **Status Atual - Pronto para Produção**

✅ **Zero-config:** Gera automaticamente certificados, senhas, armazenamento keyring
✅ **Provedores IA:** OpenAI, Anthropic Claude, Google Gemini, Groq com suporte a streaming
✅ **Protocolo MCP:** Model Context Protocol com registro dinâmico de ferramentas e comandos shell
✅ **Sistema Webhook:** Processamento funcional de webhook com integração AMQP e lógica de retry
✅ **Integração Discord:** Execução real de ferramentas MCP via bot Discord com respostas formatadas
✅ **Arquitetura Modular:** Interfaces limpas, exportáveis via `factory/`
✅ **Integração Banco:** PostgreSQL/SQLite via gerenciamento Docker `gdbase`
✅ **API REST:** Autenticação, usuários, produtos, clientes, jobs, webhooks, chat IA
✅ **Stack Segurança:** Certificados dinâmicos, JWT, keyring, rate limiting, comandos shell com whitelist
✅ **Interface CLI:** Gerenciamento completo via comandos Cobra
✅ **Multi-plataforma:** Linux, macOS, Windows (AMD64, ARM64)
✅ **Testes:** Testes unitários + testes de integração para endpoints MCP e provedores
✅ **CI/CD:** Builds e releases automatizados

## **Evolução do Projeto**

GoBE evoluiu de um servidor backend simples para uma plataforma abrangente integrada com IA. A **Versão 1.3.5** representa um marco significativo com a adição de:

### **Adições Importantes Recentes (v1.3.5)**

- **🤖 Ecossistema Completo de Provedores IA:** Integração completa com 4 grandes provedores de IA
- **🔗 Integração Real Discord MCP:** Transformação de bots Discord de placeholder para ferramentas IA funcionais
- **📬 Webhooks de Produção:** Sistema de webhook totalmente funcional com persistência e lógica de retry
- **🪝 Multi-webhooks no Discord:** Configure múltiplos destinos de webhook com compatibilidade automática para configs legadas
- **🛡️ Guardião de Assinaturas Discord:** Middleware opcional para verificar assinaturas Ed25519 mantendo bypass seguro por padrão
- **📡 Stream de webhooks Discord:** Hub publica eventos recebidos no serviço de webhooks e no event bus interno
- **⚡ Respostas IA com Streaming:** Chat IA em tempo real com Server-Sent Events
- **🔧 Ferramentas MCP Aprimoradas:** Execução de comando shell com whitelist de segurança

O sistema agora serve como uma **solução completa de backend IA**, permitindo que desenvolvedores construam aplicações com IA sem configuração de infraestrutura complexa. A arquitetura modular torna adequado para tudo, desde projetos pessoais até aplicações empresariais.

---

## **Recursos**

### **🤖 Integração IA**

✨ **4 Provedores IA Prontos**
- **OpenAI** (GPT-3.5, série GPT-4) com URLs base customizáveis
- **Anthropic Claude** (3.5 Sonnet, Opus, Haiku) com suporte a streaming
- **Google Gemini** (1.5 Pro, Flash) com cálculo correto de custos
- **Groq** (modelos Llama, Mixtral) com inferência ultra-rápida

🎯 **Gerenciamento Inteligente de Provedores**
- Troca dinâmica de provedores e verificações de disponibilidade
- Suporte a chave API externa por requisição
- Estimativa automática de custos e rastreamento de uso
- Respostas com streaming via Server-Sent Events

### **🔧 Integração de Sistema**

✨ **Protocolo MCP (Model Context Protocol)**
- Registro dinâmico de ferramentas com execução thread-safe
- Ferramentas de monitoramento de sistema integradas
- Execução de comandos shell com whitelist de segurança
- Integração bot Discord com funcionalidade real de ferramenta

📬 **Sistema de Webhook de Produção**
- Integração AMQP/RabbitMQ para processamento assíncrono
- Armazenamento persistente de webhook com lógica de retry
- Handlers especializados (GitHub, Discord, Stripe, etc.)
- API de gerenciamento RESTful com paginação

### **🏗️ Plataforma Principal**

✨ **Totalmente modular**
- Toda lógica segue interfaces bem definidas, garantindo encapsulamento
- Pode ser usado como servidor ou biblioteca/módulo
- Padrão Factory para todos os componentes principais

🔒 **Zero-config, mas customizável**
- Funciona sem configuração inicial, mas suporta customização via arquivos
- Gera automaticamente certificados, senhas e configurações seguras
- Suporte a override de variáveis de ambiente

🔗 **Integração direta com `gdbase`**
- Gerenciamento de banco via Docker
- Otimizações automáticas para persistência e performance
- Suporte multi-banco (PostgreSQL, SQLite)

🛡️ **Autenticação avançada**
- Certificados gerados dinamicamente
- Senhas aleatórias e keyring seguro
- Gerenciamento de token JWT com lógica de refresh

🌐 **API REST abrangente**
- Endpoints de chat IA com streaming
- Gerenciamento e monitoramento de webhooks
- Autenticação, gerenciamento de usuários, produtos, clientes
- Saúde do sistema e métricas

📋 **Monitoramento nível empresarial**
- Rotas protegidas, armazenamento seguro e monitoramento de requisições
- Métricas de sistema em tempo real via ferramentas MCP
- Verificações de saúde de conexão (DB, AMQP, webhooks)

🧑‍💻 **CLI poderoso**
- Comandos para iniciar, configurar e monitorar o servidor
- Startup zero-config com logging detalhado

---

## **Como Executar**

**Um Comando. Todo o Poder.**

### Início Rápido

```bash
# Clonar e compilar
git clone https://github.com/kubex-ecosystem/gobe.git
cd gobe
make build

# Iniciar tudo (zero config)
./gobe start

# Servidor pronto em http://localhost:3666
# Endpoints MCP: /mcp/tools, /mcp/exec
# Verificação de saúde: /health
```

### Requisitos

- **Go 1.25+** (para compilar do código-fonte)
- **Docker** (opcional, para recursos avançados de banco de dados)
- **Chaves API** (opcional, para provedores IA - podem ser definidas por requisição)
- **Não requer Kubernetes, nem configuração complexa**

### Opções de Build

```bash
make build          # Build de produção
make build-dev      # Build de desenvolvimento
make install        # Instalar binário + configuração do ambiente
make clean          # Limpar artefatos
make test           # Executar todos os testes
```

---

## **Provedores IA**

GoBE inclui **4 provedores IA prontos para produção** com suporte a streaming e rastreamento de custos.

### **Provedores Disponíveis**

| Provedor | Modelos | Recursos | Preços |
|----------|---------|----------|---------|
| **OpenAI** | GPT-3.5, GPT-4, GPT-4o | Streaming, URL base customizada | $0.002-$0.03 por 1K tokens |
| **Anthropic** | Claude 3.5 Sonnet, Opus, Haiku | Streaming, contexto longo | $0.25-$75 por 1M tokens |
| **Google Gemini** | Gemini 1.5 Pro, Flash | Streaming, multimodal | $0.075-$10.50 por 1M tokens |
| **Groq** | Llama 3.1, Mixtral | Inferência ultra-rápida | $0.05-$0.79 por 1M tokens |

### **Exemplos de Uso**

#### **Listar Provedores Disponíveis**
```bash
curl http://localhost:3666/providers
```

#### **Chat com Streaming**
```bash
curl -X POST http://localhost:3666/chat \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Olá!"}],
    "stream": true,
    "temperature": 0.7
  }'
```

#### **Troca de Provedor**
```bash
# Usar diferentes provedores para diferentes tarefas
curl -X POST http://localhost:3666/chat \
  -d '{"provider": "anthropic", "model": "claude-3-5-sonnet-20241022", ...}'

curl -X POST http://localhost:3666/chat \
  -d '{"provider": "groq", "model": "llama-3.1-70b-versatile", ...}'
```

#### **Chaves API Externas**
```bash
# Usar sua própria chave API para uma requisição específica
curl -X POST http://localhost:3666/chat \
  -H "X-External-API-Key: sua-chave-api-aqui" \
  -d '{"provider": "openai", ...}'
```

### **Configuração de Provedores**

Configure provedores via variáveis de ambiente ou arquivos de configuração:

```bash
# Variáveis de Ambiente
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export GEMINI_API_KEY="..."
export GROQ_API_KEY="gsk_..."

# URLs Base Customizadas (para endpoints compatíveis com OpenAI)
export OPENAI_BASE_URL="https://api.groq.com"
```

### **Rastreamento de Custos**

Cada resposta inclui estimativa de custo:

```json
{
  "content": "Texto da resposta...",
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 20,
    "total_tokens": 30,
    "latency_ms": 1500,
    "cost_usd": 0.0006,
    "provider": "openai",
    "model": "gpt-4"
  }
}
```

---

## **Suporte MCP**

GoBE implementa o **Model Context Protocol (MCP)** para integração perfeita de ferramentas IA.

### Endpoints Disponíveis

```bash
GET  /mcp/tools     # Listar ferramentas disponíveis
POST /mcp/exec      # Executar ferramentas
```

### Ferramentas Integradas

| Ferramenta | Descrição | Argumentos | Recursos |
|-----------|-----------|------------|----------|
| `system.status` | Status abrangente do sistema com métricas | `detailed: boolean` | Stats runtime, uso memória, saúde conexões |
| `shell.command` | Executar comandos shell seguros | `command: string, args: array` | Comandos com whitelist, timeout 10s, captura saída |

### Exemplos de Uso

#### **Listar Ferramentas Disponíveis**
```bash
curl http://localhost:3666/mcp/tools
```

#### **Status do Sistema (Básico)**
```bash
curl -X POST http://localhost:3666/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{"tool": "system.status", "args": {"detailed": false}}'
```

#### **Status do Sistema (Detalhado)**
```bash
curl -X POST http://localhost:3666/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{"tool": "system.status", "args": {"detailed": true}}'
```

#### **Executar Comandos Shell**
```bash
# Listar arquivos
curl -X POST http://localhost:3666/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{"tool": "shell.command", "args": {"command": "ls", "args": ["-la"]}}'

# Verificar info do sistema
curl -X POST http://localhost:3666/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{"tool": "shell.command", "args": {"command": "uname", "args": ["-a"]}}'

# Verificar uso do disco
curl -X POST http://localhost:3666/mcp/exec \
  -H "Content-Type: application/json" \
  -d '{"tool": "shell.command", "args": {"command": "df", "args": ["-h"]}}'
```

### **Recursos de Segurança**

- **Comandos com Whitelist:** Apenas comandos seguros são permitidos (`ls`, `pwd`, `date`, `uname`, etc.)
- **Proteção Timeout:** Todos os comandos fazem timeout após 10 segundos
- **Tratamento de Erro:** Captura e relatório adequado de erros
- **Autenticação Admin:** Comandos shell requerem privilégios de admin

### Arquitetura MCP

- **Registro Dinâmico:** Ferramentas registradas em runtime, sem necessidade de restart
- **Thread-Safe:** Execução concorrente de ferramenta com proteção RWMutex
- **Extensível:** Adicionar ferramentas customizadas via interface Registry
- **Integrado:** Usa controladores de gerenciamento existentes para dados do sistema

---

## **Webhooks**

GoBE fornece um **sistema de webhook pronto para produção** com persistência, lógica de retry e integração AMQP.

### **Recursos**

- ✅ **Armazenamento Persistente:** Armazenamento em memória com planos para persistência em banco
- ✅ **Integração AMQP:** Processamento async via RabbitMQ
- ✅ **Lógica Retry:** Retry automático de eventos webhook falidos
- ✅ **Handlers Especializados:** GitHub, Discord, Stripe, webhooks genéricos
- ✅ **API RESTful:** Operações CRUD completas com paginação
- ✅ **Stats Tempo Real:** Monitor processamento de webhook em tempo real

### **Endpoints Disponíveis**

| Método | Endpoint | Descrição |
|--------|----------|------------|
| `POST` | `/v1/webhooks` | Receber eventos webhook |
| `GET` | `/v1/webhooks/health` | Saúde do sistema webhook + stats |
| `GET` | `/v1/webhooks/events` | Listar eventos webhook (paginado) |
| `GET` | `/v1/webhooks/events/:id` | Obter evento webhook específico |
| `POST` | `/v1/webhooks/retry` | Retry todos eventos webhook falidos |

### **Exemplos de Uso**

#### **Receber Webhooks**
```bash
# Webhook genérico
curl -X POST http://localhost:3666/v1/webhooks \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Source: github" \
  -H "X-Event-Type: push" \
  -d '{"repository": "meu-repo", "commits": [...]}'
```

#### **Verificar Saúde do Sistema**
```bash
curl http://localhost:3666/v1/webhooks/health
```
**Resposta:**
```json
{
  "status": "ok",
  "timestamp": "2025-01-20T12:00:00Z",
  "stats": {
    "total_events": 150,
    "processed_events": 145,
    "failed_events": 3,
    "pending_events": 2,
    "uptime_seconds": 3600,
    "amqp_connected": true
  }
}
```

#### **Listar Eventos Webhook**
```bash
# Todos os eventos (paginado)
curl "http://localhost:3666/v1/webhooks/events?limit=10&offset=0"

# Filtrar por fonte
curl "http://localhost:3666/v1/webhooks/events?source=github&limit=20"
```

#### **Obter Detalhes do Evento**
```bash
curl http://localhost:3666/v1/webhooks/events/123e4567-e89b-12d3-a456-426614174000
```

#### **Retry Eventos Falidos**
```bash
curl -X POST http://localhost:3666/v1/webhooks/retry
```
**Resposta:**
```json
{
  "status": "success",
  "retried_count": 3,
  "timestamp": "2025-01-20T12:00:00Z",
  "message": "eventos falidos enfileirados para retry"
}
```

### **Tipos de Eventos Webhook**

O sistema fornece tratamento especializado para diferentes tipos de webhook:

| Fonte | Tipo Evento | Recursos Handler |
|-------|-------------|------------------|
| **GitHub** | `github.push` | Info repositório, detalhes commit, notificações AMQP |
| **Discord** | `discord.message` | Log canal, triggers resposta bot |
| **Stripe** | `stripe.payment` | Processamento pagamento, atualizações cobrança usuário |
| **Genérico** | `user.created` | Emails boas-vindas, criação perfil |

### **Exemplos de Integração**

#### **Integração GitHub**
```bash
# Configurar webhook GitHub para apontar para sua instância GoBE
# URL Webhook: https://seu-servidor.com/v1/webhooks
# Content-Type: application/json
# Eventos: push, pull_request, issues

curl -X POST https://seu-servidor.com/v1/webhooks \
  -H "X-GitHub-Event: push" \
  -H "X-Webhook-Source: github" \
  -d @github_push_payload.json
```

#### **Integração Bot Discord**
```bash
# Eventos Discord são processados automaticamente e podem disparar ferramentas MCP
curl -X POST http://localhost:3666/v1/webhooks \
  -H "X-Webhook-Source: discord" \
  -H "X-Event-Type: message" \
  -d '{"channel_id": "123", "content": "!system status", "author": {...}}'
```

### **Integração AMQP**

Webhooks são automaticamente publicados em filas RabbitMQ:

- **Exchange:** `gobe.events`
- **Routing Key:** `webhook.received`
- **Queue Bindings:** `gobe.system.events`, `gobe.mcp.tasks`

**Formato Mensagem AMQP:**
```json
{
  "id": "uuid",
  "source": "github",
  "event_type": "push",
  "payload": {...},
  "headers": {...},
  "timestamp": "2025-01-20T12:00:00Z",
  "processed": false,
  "status": "received"
}
```

---

## **Uso**

### CLI

Iniciar o servidor principal:

```sh
./gobe start -p 3666 -b "0.0.0.0"
```

Isso inicia o servidor, gera certificados, configura bancos de dados e começa a escutar requisições!

Ver todos os comandos disponíveis:

```sh
./gobe --help
```

**Comandos principais:**

| Comando   | Função                                         |
|-----------|------------------------------------------------|
| `start`   | Inicia o servidor                              |
| `stop`    | Para o servidor com segurança                  |
| `restart` | Reinicia todos os serviços                     |
| `status`  | Mostra o status do servidor e serviços ativos |
| `config`  | Gera arquivo de configuração inicial           |
| `logs`    | Mostra logs do servidor                        |

---

### Configuração

GoBE pode funcionar sem configuração inicial, mas suporta customização via arquivos YAML/JSON. Por padrão, tudo é gerado automaticamente no primeiro uso.

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

#### Integrações de Mensagem

Bots WhatsApp e Telegram podem ser configurados via arquivo `config/discord_config.json` na seção `integrations`:

```json
{
  "integrations": {
    "whatsapp": {
      "enabled": true,
      "access_token": "<token>",
      "verify_token": "<verify>",
      "phone_number_id": "<number>",
      "webhook_url": "https://seu.servidor/whatsapp/webhook"
    },
    "telegram": {
      "enabled": true,
      "bot_token": "<bot token>",
      "webhook_url": "https://seu.servidor/telegram/webhook",
      "allowed_updates": ["message", "callback_query"]
    }
  }
}
```

Após configurar o arquivo ou variáveis de ambiente, o servidor irá expor os seguintes endpoints:

- `POST /api/v1/whatsapp/send` e `/api/v1/whatsapp/webhook`
- `POST /api/v1/telegram/send` e `/api/v1/telegram/webhook`

Cada rota também fornece um endpoint `/ping` para verificações de saúde.

---

## **Referência da API**

### **Endpoints Principais**

| Categoria | Método | Endpoint | Descrição |
|-----------|--------|----------|-----------|
| **Saúde** | `GET` | `/health` | Verificação básica de saúde |
| **Saúde** | `GET` | `/healthz` | Verificação saúde estilo Kubernetes |
| **Saúde** | `GET` | `/status` | Status detalhado do sistema |
| **Saúde** | `GET` | `/api/v1/health` | Saúde API com métricas |

### **Endpoints Provedores IA**

| Método | Endpoint | Descrição | Streaming |
|--------|----------|-----------|-----------|
| `GET` | `/providers` | Listar todos provedores IA e disponibilidade | ❌ |
| `POST` | `/chat` | Chat com provedores IA | ✅ SSE |
| `POST` | `/v1/advise` | Obter conselhos/recomendações IA | ✅ SSE |

### **Endpoints MCP (Model Context Protocol)**

| Método | Endpoint | Descrição | Auth |
|--------|----------|-----------|------|
| `GET` | `/mcp/tools` | Listar ferramentas MCP disponíveis | Bearer |
| `POST` | `/mcp/exec` | Executar ferramenta MCP | Bearer |

### **Endpoints Webhook**

| Método | Endpoint | Descrição | Auth |
|--------|----------|-----------|------|
| `POST` | `/v1/webhooks` | Receber eventos webhook | Bearer |
| `GET` | `/v1/webhooks/health` | Saúde sistema webhook | Bearer |
| `GET` | `/v1/webhooks/events` | Listar eventos webhook (paginado) | Bearer |
| `GET` | `/v1/webhooks/events/:id` | Obter evento webhook específico | Bearer |
| `POST` | `/v1/webhooks/retry` | Retry eventos webhook falidos | Bearer |

### **Endpoints Monitoramento Sistema**

| Método | Endpoint | Descrição | Auth |
|--------|----------|-----------|------|
| `GET` | `/api/v1/mcp/system/info` | Informações do sistema | Bearer |
| `GET` | `/api/v1/mcp/system/cpu-info` | Métricas CPU | Bearer |
| `GET` | `/api/v1/mcp/system/memory-info` | Métricas memória | Bearer |
| `GET` | `/api/v1/mcp/system/disk-info` | Métricas disco | Bearer |

### **Endpoints Agendador**

| Método | Endpoint | Descrição | Auth |
|--------|----------|-----------|------|
| `GET` | `/health/scheduler/stats` | Estatísticas agendador | Bearer |
| `POST` | `/health/scheduler/force` | Forçar execução agendador | Bearer |

### **Endpoints Web UI**

| Método | Endpoint | Descrição | Auth |
|--------|----------|-----------|------|
| `GET` | `/` | Servir raiz web UI | Público |
| `GET` | `/app/*path` | Servir aplicação web UI | Público |

### **Autenticação**

GoBE usa autenticação **Bearer Token** para endpoints protegidos:

```bash
# Obter token (implementação depende da sua configuração auth)
TOKEN="seu-jwt-token-aqui"

# Usar token nas requisições
curl -H "Authorization: Bearer $TOKEN" http://localhost:3666/mcp/tools
```

### **Formatos de Resposta**

#### **Resposta Sucesso**
```json
{
  "status": "success",
  "data": { ... },
  "timestamp": "2025-01-20T12:00:00Z"
}
```

#### **Resposta Erro**
```json
{
  "status": "error",
  "message": "Descrição do erro",
  "code": 400,
  "timestamp": "2025-01-20T12:00:00Z"
}
```

#### **Resposta Streaming (SSE)**
```
data: {"content": "Olá", "done": false}

data: {"content": " mundo!", "done": false}

data: {"done": true, "usage": {"total_tokens": 10, "cost_usd": 0.0002}}
```

### **Rate Limiting**

- **Padrão:** 100 requisições por minuto por IP
- **Endpoints IA:** 30 requisições por minuto por chave API
- **Headers:** `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

### **Suporte CORS**

CORS está habilitado para integração web UI:

```javascript
// Exemplo frontend JavaScript
fetch('http://localhost:3666/providers')
  .then(response => response.json())
  .then(providers => console.log(providers));
```

---

## **Roteiro**

### ✅ **Concluído (v1.3.5)**

- [x] Modularização completa e interfaces plugáveis
- [x] Zero-config com geração automática de certificados
- [x] Integração com keyring do sistema
- [x] API REST para autenticação e gerenciamento
- [x] Autenticação via certificados e senhas seguras
- [x] CLI para gerenciamento e monitoramento
- [x] Integração com `gdbase` para gerenciamento banco via Docker
- [x] **Implementação Protocolo MCP com registro dinâmico ferramentas**
- [x] **Ferramentas integradas system.status e shell.command**
- [x] **Execução thread-safe ferramentas com testes abrangentes**
- [x] **4 Provedores IA: OpenAI, Anthropic, Gemini, Groq**
- [x] **Respostas IA com streaming e rastreamento custos**
- [x] **Sistema webhook produção com integração AMQP**
- [x] **Integração bot Discord com ferramentas MCP reais**
- [x] **Documentação API completa com exemplos**
- [x] Suporte multi-banco (PostgreSQL, SQLite)
- [x] Testes automatizados e CI/CD

### 🚧 **Em Progresso (v1.4.0)**

- [ ] Persistência banco para eventos webhook
- [ ] Biblioteca estendida ferramentas MCP (operações arquivo, ferramentas rede)
- [ ] Suporte WebSocket para comunicação MCP tempo real
- [ ] Integração Prometheus para monitoramento

### 📋 **Planejado (v1.5.0+)**

- [ ] Sistema plugin para registro ferramenta externa
- [ ] Políticas segurança avançadas e RBAC
- [ ] Integração Grafana para visualização métricas
- [ ] Suporte multi-tenant
- [ ] Orquestração workflow IA avançada
- [ ] Suporte middleware customizado

### 🎯 **Próximos Marcos**

1. **v1.4.0** - Persistência banco, WebSocket MCP, monitoramento
2. **v1.5.0** - Sistema plugin e segurança avançada
3. **v2.0.0** - Plataforma multi-tenant com orquestração workflow

---

## **Contribuindo**

Contribuições são bem-vindas! Sinta-se à vontade para abrir issues ou submeter pull requests. Veja o [Guia de Contribuição](docs/CONTRIBUTING.md) para mais detalhes.

---

## **Contato**

💌 **Desenvolvedor**:
[Rafael Mori](mailto:faelmori@gmail.com)
💼 [Me siga no GitHub](https://github.com/kubex-ecosystem)
Estou aberto a colaborações e novas ideias. Se achou o projeto interessante, entre em contato!

---

## **Riscos & Mitigações**

• **Zero-config pode ocultar configurações necessárias** → Logs verbosos + documentação override
• **Thread-safety registro MCP** → RWMutex implementado + testes concorrência
• **Dependência gdbase para DB** → Fallback SQLite sempre disponível
• **Auto-geração certificado** → Backup keyring + regeneração automática

---

## **Próximos Passos**

1. **Estender ferramentas MCP** - operações arquivo, diagnósticos rede, queries banco
2. **WebSocket MCP** - comunicação ferramenta tempo real para agentes IA
3. **Sistema plugin** - registro ferramenta externa via bibliotecas compartilhadas
4. **Monitoramento avançado** - métricas Prometheus + dashboards Grafana
5. **Hardening segurança** - RBAC, logs auditoria, aplicação política

---

## **Changelog**

### v1.3.5 (20/01/2025)

- ✅ **Ecossistema Provedores IA:** Integração completa OpenAI, Anthropic, Gemini, Groq
- ✅ **Respostas IA Streaming:** Server-Sent Events com rastreamento custos
- ✅ **Webhooks Produção:** Sistema webhook completo com AMQP e lógica retry
- ✅ **Ferramentas MCP Melhoradas:** Adicionado shell.command com whitelist segurança
- ✅ **Integração Discord:** Execução real ferramenta MCP via bot Discord
- ✅ **API Abrangente:** REST API completa com documentação detalhada

### v1.3.4 (23/12/2024)

- ✅ Implementação Protocolo MCP com registro dinâmico
- ✅ Ferramenta integrada system.status com métricas runtime
- ✅ Execução thread-safe ferramentas com testes abrangentes
- ✅ Documentação atualizada seguindo padrões Kubex
- ✅ Endpoints MCP zero-config (/mcp/tools, /mcp/exec)
