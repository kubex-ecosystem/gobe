#!/bin/bash

# Script de teste para o controller de túnel GDBase
BASE_URL="http://localhost:8080"

echo "=== GDBase Tunnel Controller Test ==="
echo ""

# Função para fazer requisições
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3

    echo "🔄 $method $endpoint"
    if [ -n "$data" ]; then
        echo "📦 Data: $data"
        curl -s -X "$method" "$BASE_URL$endpoint" \
             -H "Content-Type: application/json" \
             -d "$data" | jq '.' 2>/dev/null || echo "Response received"
    else
        curl -s -X "$method" "$BASE_URL$endpoint" \
             -H "Content-Type: application/json" | jq '.' 2>/dev/null || echo "Response received"
    fi
    echo ""
}

# 1. Verificar status inicial
echo "1️⃣  Verificando status inicial do túnel:"
make_request "GET" "/api/v1/mcp/db/tunnel/status"

# 2. Criar túnel quick
echo "2️⃣  Criando túnel quick para PgAdmin:"
make_request "POST" "/api/v1/mcp/db/tunnel/up" '{
    "mode": "quick",
    "network": "gdbase_net",
    "target": "pgadmin",
    "port": 80,
    "timeout": "30s"
}'

# 3. Verificar status após criação
echo "3️⃣  Verificando status após criação:"
make_request "GET" "/api/v1/mcp/db/tunnel/status"

# 4. Tentar criar outro túnel (deve dar conflito)
echo "4️⃣  Tentando criar outro túnel (deve dar conflito):"
make_request "POST" "/api/v1/mcp/db/tunnel/up" '{
    "mode": "quick",
    "target": "postgres",
    "port": 5432
}'

# 5. Parar o túnel
echo "5️⃣  Parando o túnel:"
make_request "POST" "/api/v1/mcp/db/tunnel/down"

# 6. Verificar status final
echo "6️⃣  Verificando status final:"
make_request "GET" "/api/v1/mcp/db/tunnel/status"

# 7. Exemplo de túnel named
echo "7️⃣  Exemplo de túnel named (precisa de token válido):"
make_request "POST" "/api/v1/mcp/db/tunnel/up" '{
    "mode": "named",
    "network": "gdbase_net",
    "token": "your-cloudflare-tunnel-token-here"
}'

echo "✅ Teste concluído!"
echo ""
echo "📝 Endpoints disponíveis:"
echo "   GET  /api/v1/mcp/db/tunnel/status   - Status do túnel"
echo "   POST /api/v1/mcp/db/tunnel/up       - Criar túnel"
echo "   POST /api/v1/mcp/db/tunnel/down     - Parar túnel"
echo ""
echo "🔧 Modos suportados:"
echo "   quick - Túnel HTTP efêmero (URL dinâmica)"
echo "   named - Túnel HTTP+TCP fixo (requer token Cloudflare)"
