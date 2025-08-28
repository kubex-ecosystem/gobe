# GDBase Tunnel Controller

O **GDBase Tunnel Controller** fornece uma interface HTTP para gerenciar túneis Cloudflare, permitindo exposição segura de serviços Docker via internet.

## 🚀 Funcionalidades

- **Quick Tunnel**: Túneis HTTP efêmeros com URL dinâmica
- **Named Tunnel**: Túneis HTTP+TCP persistentes com domínio fixo
- **Estado Centralizado**: Gerenciamento thread-safe do estado dos túneis
- **Integração Docker**: Controle nativo de containers Cloudflared

## 📋 API Endpoints

### GET `/api/v1/mcp/db/tunnel/status`

Retorna o status atual do túnel.

**Resposta:**

```json
{
  "mode": "quick|named",
  "public": "https://xyz.trycloudflare.com",
  "running": true,
  "network": "gdbase_net",
  "target": "pgadmin:80"
}
```

### POST `/api/v1/mcp/db/tunnel/up`

Cria um novo túnel.

**Quick Tunnel:**

```json
{
  "mode": "quick",
  "network": "gdbase_net",
  "target": "pgadmin",
  "port": 80,
  "timeout": "30s"
}
```

**Named Tunnel:**

```json
{
  "mode": "named",
  "network": "gdbase_net",
  "token": "your-cloudflare-tunnel-token"
}
```

### POST `/api/v1/mcp/db/tunnel/down`

Para o túnel ativo.

**Resposta:** `204 No Content`

## 🛠️ Exemplos de Uso

### Expor PgAdmin via Quick Tunnel

```bash
curl -X POST http://localhost:8080/api/v1/mcp/db/tunnel/up \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "quick",
    "target": "pgadmin",
    "port": 80,
    "network": "gdbase_net"
  }'
```

### Usar Named Tunnel

```bash
curl -X POST http://localhost:8080/api/v1/mcp/db/tunnel/up \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "named",
    "network": "gdbase_net",
    "token": "eyJhIjoiN..."
  }'
```

### Verificar Status

```bash
curl http://localhost:8080/api/v1/mcp/db/tunnel/status
```

### Parar Túnel

```bash
curl -X POST http://localhost:8080/api/v1/mcp/db/tunnel/down
```

## ⚙️ Configuração

### Requisitos

- Docker client configurado
- Rede Docker existente (padrão: `gdbase_net`)
- Cloudflare Tunnel Token (para modo named)

### Variáveis de Ambiente

```bash
DOCKER_HOST=unix:///var/run/docker.sock  # Docker daemon socket
```

## 🔧 Detalhes Técnicos

### Modos de Túnel

#### Quick Tunnel

- **Uso**: Desenvolvimento e testes
- **URL**: Dinâmica (`*.trycloudflare.com`)
- **Duração**: Temporária
- **Configuração**: Apenas target e porta

#### Named Tunnel

- **Uso**: Produção
- **URL**: Fixa (seu domínio)
- **Duração**: Persistente
- **Configuração**: Token do dashboard Cloudflare

### Thread Safety

O controller usa `sync.RWMutex` para garantir operações thread-safe:

- **Read locks** para consultas de status
- **Write locks** para operações de criação/destruição

### Error Handling

Respostas padronizadas seguem o formato:

```json
{
  "error": "Error Type",
  "message": "Detailed error message"
}
```

## 🧪 Testes

Execute o script de teste:

```bash
./scripts/test_tunnel.sh
```

## 🏗️ Arquitetura

```
GDBaseController
├── Docker Client ──> Cloudflared Container
├── Tunnel State ──> Thread-safe status
└── Bridge Layer ──> gdbasez.TunnelHandle
```

### Fluxo de Operação

1. **Criação**: Valida parâmetros → Cria container → Atualiza estado
2. **Status**: Lê estado thread-safe
3. **Destruição**: Para container → Reset estado

## 🚨 Limitações

- **Um túnel por vez**: Sistema single-tunnel
- **Dependência Docker**: Requer Docker daemon ativo
- **Rede específica**: Containers devem estar na mesma rede Docker

## 📚 Referências

- [Cloudflare Tunnel Docs](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
- [Docker API Reference](https://docs.docker.com/engine/api/)
- [GoBE Architecture Guide](docs/architecture.md)
