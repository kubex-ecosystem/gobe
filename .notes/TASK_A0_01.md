# Especificação Técnica: Implementação do Gateway HTTP no GoBE

Você é um engenheiro sênior de Go responsável por implementar no **GoBE** um **Gateway HTTP** com feature-parity (ou superior) ao gateway do **Grompt/Analyzer**, mas **seguindo o padrão de rotas do GoBE** (Gin + `proto.NewRoute` + `ar.IRouter`) e **internalizando** as rotas principais.

## Contexto e Restrição

* O servidor do GoBE usa **Gin** e registra rotas via o wrapper interno `proto.NewRoute` retornando `ar.IRoute`, com injeção de `dbService`, `middlewares` e `secureProperties`. Examine as rotas já existentes para **MCP Providers**, **System** e **Tasks** como referência de **estilo e wiring** (serviços, construtores, mapas de rotas, etc.). \[Referências: MCP Providers/Tasks/System]  &#x20;
* O gateway do **Grompt** já implementa handlers e endpoints HTTP/SSE que você deve **espelhar/melhorar** no GoBE, adaptando para o padrão de rotas do GoBE. Veja `transport/http.go` (Web UI embarcada, /chat SSE, /providers, /advise, scorecard/health, LookAtni e webhooks), `transport/http_sse.go` (variante SSE minimalista) e `transport/sse_coalescer.go` (coalescador SSE). \[Referências: Grompt gateway + SSE]  &#x20;

> **Não altere o framework** (Gin) nem o **estilo declarativo de rotas** do GoBE. Toda rota nova deve ser exposta como `map[string]ar.IRoute` via `proto.NewRoute(...)`, seguindo a mesma ideia dos arquivos `mcp_*`.&#x20;

---

## Objetivo

1. **Criar um “gateway” interno do GoBE** com rotas equivalentes às do Grompt/Analyzer, **mantendo**:

   * **SSE** para chat/completion com **coalescimento de chunks** (UX suave).
   * Suporte a **BYOK** via header `x-external-api-key` e **metadados** (`x-tenant-id`, `x-user-id`).
   * Rotas de **providers**, **advise**, **scorecard/metrics/health**, **LookAtni**, **webhooks**, e **status/health**.
   * **Web UI** embarcada opcional sob `/app/` e `/` (quando disponível).
2. **Padronizar segurança** via `secureProperties` (em produção, `secure: true`) e manter **validações**/sanitização configuráveis.
3. Integrar com os **serviços existentes** (`dbService`, middlewares) e o **registry/provider** já usado pelos módulos MCP/Tasks/Providers.

---

## Arquitetura e Pastas (sugestão)

Crie os seguintes pacotes, seguindo o padrão do GoBE:

```
internal/
  app/
    controllers/
      gateway/
        chat.go          // SSE chat controller (usa registry de providers)
        providers.go     // lista providers e status
        advise.go        // POST /v1/advise (SSE ou chunked)
        scorecard.go     // /api/v1/scorecard, /api/v1/scorecard/advice, /api/v1/metrics/ai
        health.go        // /api/v1/health + /healthz + /status
        lookatni.go      // endpoints LookAtni: extract/archive/download/projects
        webhooks.go      // /v1/webhooks e /v1/webhooks/health
        scheduler.go     // /health/scheduler/stats e /health/scheduler/force
        webui.go         // mounting de /app/ e /
    router/
      gateway/
        routes.go        // NewGatewayRoutes(*ar.IRouter) map[string]ar.IRoute
    transport/
      sse/
        coalescer.go     // portar SSE Coalescer do Grompt (melhorado)
    // reutilize types/middlewares dos módulos existentes
```

**Motivo**: manter simetria com `mcp_*` e permitir registrar **um único map de rotas** do gateway, como já é feito nos MCP *routes*.  &#x20;

---

## Especificação de Rotas (GoBE Gateway)

> Use `proto.NewRoute(method, path, contentType, handlerFunc, middlewaresMap, dbService, secureProps, nil)` ao declarar cada rota no `routes.go`. Siga o padrão de `NewMCP*Routes` para montar e retornar `map[string]ar.IRoute`.&#x20;

### Básicas / Estado

* `GET /healthz` → **ping** simples (200/JSON). (espelhar Grompt)&#x20;
* `GET /status` → status detalhado + métricas de middlewares se houver. (espelhar Grompt)&#x20;

### Chat SSE (Gateway LLM)

* `POST /chat` → SSE com **coalescing** (porta do Grompt). Respeite:

  * Headers BYOK: `x-external-api-key` (injetar em `req.Meta["external_api_key"]` ou `Headers`).
  * Headers de multi-tenant: `x-tenant-id`, `x-user-id` (propagar para provider).
  * `Temp` default 0.7; `Stream = true` forçado.
  * **SSE**: `Content-Type: text/event-stream`, flush por chunk coalescido, evento final com `done:true` e `usage` (quando houver).
  * Em caso de erro, emita JSON SSE `{error: "...", done:true}`.
    Replique/una o comportamento de `transport/http.go` e `transport/http_sse.go`. &#x20;
* **Coalescer**: implemente `transport/sse/coalescer.go` baseado em `SSECoalescer` do Grompt (timeout \~75ms, flush por pontuação, limite de buffer). Deve ser **thread-safe** o suficiente para uso no handler.&#x20;

### Providers

* `GET /providers` → lista providers com enrich de saúde/uptime/latência quando disponível; compatível com saída do Grompt.&#x20;

### Advise (AI-powered advice)

* `POST /v1/advise` → streaming SSE (ou resposta normal JSON) com as mesmas regras de headers/limites do Grompt (BF1 modo etc., se aplicável).&#x20;
* (Opcional) manter rota enxuta `POST /advise` para compatibilidade (se necessário).&#x20;

### Repository Intelligence (placeholder com rota viva)

* `GET  /api/v1/scorecard`
* `POST /api/v1/scorecard/advice`
* `GET  /api/v1/metrics/ai`
* `GET  /api/v1/health`
  Replicar placeholders do Grompt com headers `X-Schema-Version` e `X-Server-Version`. Manter TODOs para integrar engine real.&#x20;

### LookAtni (integração navegação de código)

* `POST   /api/v1/lookatni/extract`
* `POST   /api/v1/lookatni/archive`
* `GET    /api/v1/lookatni/download/`
* `GET    /api/v1/lookatni/projects`
* `GET    /api/v1/lookatni/projects/`
  Espelhar handlers do Grompt (crie controllers finos que chamem a lib/engine interna que você expuser).&#x20;

### Webhooks (meta-recursivo)

* `POST /v1/webhooks`
* `GET  /v1/webhooks/health`
  Crie controlador e registre. Placeholder funcional é aceitável inicialmente.&#x20;

### Health Scheduler (AI Provider Health)

* `GET  /health/scheduler/stats`
* `POST /health/scheduler/force`
  Estruture um `scheduler` dentro do controller com mesma semântica das rotas do Grompt.&#x20;

### Web UI (opcional)

* Montar **/app/** e **/** para o frontend embarcado quando disponível (não bloquear API). Trate erros de init com log e continue.&#x20;

---

## Padrões de Segurança e Middlewares

* Use `secureProperties["secure"]=true` para produção (nos exemplos MCP alguns estão `false` temporários; **corrija** para o gateway). Ative `validateAndSanitize` quando a validação existir no projeto. &#x20;
* **CORS** e **Access-Control-Allow-Origin:**\* no SSE (como no Grompt).&#x20;
* Respeite `dbService` quando aplicável; injete onde for necessário manter simetria com design MCP (vide `NewMCP*Routes`).&#x20;

---

## Integração com Providers/Registry

* No handler de chat, resolva provider via registry do GoBE (equivalente ao Grompt). Propague:

  * `x-external-api-key` (**BYOK**)
  * `x-tenant-id`, `x-user-id`
  * `req.Meta`, `req.Headers`
* Force `Stream=true` para SSE e use o coalescer para reduzir “micro-chunks”. &#x20;

---

## Registro de Rotas (estilo GoBE)

Implemente `internal/app/router/gateway/routes.go`:

```go
package gateway

import (
  "net/http"
  "github.com/gin-gonic/gin"
  proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
  ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
  gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
  // importe seus controllers em internal/app/controllers/gateway/*
)

type GatewayRoutes struct{ ar.IRouter }

func NewGatewayRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
  if rtr == nil { gl.Log("error","Router is nil for GatewayRoutes"); return nil }
  rtl := *rtr

  dbService := rtl.GetDatabaseService()
  middlewaresMap := make(map[string]gin.HandlerFunc)

  secure := map[string]bool{
    "secure": true, "validateAndSanitize": false, "validateAndSanitizeBody": false,
  }

  routes := map[string]ar.IRoute{}

  // Exemplo:
  // routes["Healthz"] = proto.NewRoute(http.MethodGet, "/healthz", "application/json", gatewayController.Healthz, middlewaresMap, dbService, secure, nil)
  // routes["ChatSSE"] = proto.NewRoute(http.MethodPost, "/chat", "text/event-stream", gatewayController.ChatSSE, middlewaresMap, dbService, secure, nil)
  // ... e assim por diante, para todas as rotas especificadas.

  return routes
}
```

Siga o padrão de `NewMCPProvidersRoutes`, `NewMCPTasksRoutes`, `NewMCPSystemRoutes` para consistência.  &#x20;

---

## SSE Coalescer

Portar `SSECoalescer` do Grompt para `internal/app/transport/sse/coalescer.go`, mantendo a lógica de **bufferTimeout \~75ms**, **maxBufferSize \~100**, flush em pontuação/quebras de linha e flush final no `Close()`. Ajuste para evitar data races.&#x20;

---

## Critérios de Aceite

* **Build** e **tests** passando.
* `curl -N -X POST http://localhost:PORT/chat` com um payload válido produz SSE **coalescido** e encerra com `{"done":true,"usage":...}`.
* `GET /providers` retorna lista com health (quando houver) e mantém compatibilidade com o formato do Grompt.&#x20;
* Rotas `advise`, `scorecard`, `metrics/ai`, `health` operam (placeholders onde ainda não há engine).&#x20;
* Rotas LookAtni e webhooks expostas, mesmo que inicialmente com lógicas mocked/placeholder.&#x20;
* Registro **único** do Gateway via `NewGatewayRoutes` retornando `map[string]ar.IRoute`, pronto para ser plugado no **registrador central** do GoBE (exatamente como os `mcp_*`).&#x20;

---

## Dicas de Implementação

* **Não** reescreva o sistema de roteamento; **apenas** adicione o módulo Gateway no mesmo **padrão** dos MCPs.
* Reaproveite a semântica do `http_sse.go` (headers e payloads) e do `http.go` (forçar stream, BYOK, erros via SSE) ao compor os handlers no GoBE. &#x20;
* Onde faltar engine real (scorecard, health RI), mantenha **placeholders** com headers de versão como no Grompt.&#x20;

---

## Exemplos de cURLs ( smoke tests )

```bash
# Health
curl -i http://localhost:PORT/healthz

# Providers
curl -s http://localhost:PORT/providers | jq

# Chat SSE (BYOK)
curl -N -H "Content-Type: application/json" \
     -H "x-external-api-key: $KEY" \
     -H "x-tenant-id: demo" \
     -H "x-user-id: rafa" \
     -X POST http://localhost:PORT/chat \
     -d '{"provider":"groq","model":"mixtral","messages":[{"role":"user","content":"hi"}],"stream":true}'
```

---

## Entregáveis

1. Pacotes criados, com **controllers**, **rotas** e **coalescer SSE** funcionais.
2. Arquivo `internal/app/router/gateway/routes.go` expondo **todas** as rotas em `map[string]ar.IRoute`.
3. Integração com `dbService`/middlewares idêntica ao padrão `mcp_*`.
4. Comentários `TODO` claros nos pontos ainda “placeholder”.

> Referências para estilo/API:
>
> * **MCP Providers/Tasks/System (GoBE)**: estilo de rotas, `secureProperties`, injeção de serviços.  &#x20;
> * **Grompt Gateway**: endpoints equivalentes, SSE, BYOK e Web UI/LookAtni/Webhooks/Scheduler.  &#x20;
