#!/bin/bash

# === CONFIGURAÇÕES ===
BASE_URL="http://localhost:8080/api/v1"
LOOP_COUNT=20
LOG_FILE="mcp_benchmark_$(date +%Y%m%d_%H%M%S).log"

# === ROTAS A SEREM TESTADAS ===
ROUTES=(
  "GET /mcp/system/info"
  "GET /mcp/system/cpu-info"
  "GET /mcp/system/memory-info"
  "GET /mcp/system/disk-info"
  "POST /mcp/system/send-message"
  "GET /mcp/llm/enabled"
  "POST /mcp/system/shell-command"
)

# === FUNÇÃO PARA REQUISIÇÃO COM CURL ===
make_request() {
  local method=$1
  local route=$2
  local url="${BASE_URL}${route}"
  local start end duration

  start=$(date +%s%3N)
  
  if [[ "$method" == "POST" ]]; then
    response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$url" \
      -H "Content-Type: application/json" \
      -d '{"message":"Hello World","command":"uptime"}')
  else
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url")
  fi

  end=$(date +%s%3N)
  duration=$((end - start))

  timestamp=$(date '+%Y-%m-%d %H:%M:%S')
  log_line="[$timestamp] $method $route - Status: $response - ${duration}ms"
  echo "$log_line" | tee -a "$LOG_FILE"
}

# === LOOP PRINCIPAL ===
echo "🚀 Iniciando benchmark com $LOOP_COUNT iterações..."
echo "Log: $LOG_FILE"
echo "========================================"

for ((i=1; i<=LOOP_COUNT; i++)); do
  echo "▶️ Iteração $i/$LOOP_COUNT" | tee -a "$LOG_FILE"
  for entry in "${ROUTES[@]}"; do
    method=$(echo "$entry" | awk '{print $1}')
    route=$(echo "$entry" | cut -d' ' -f2-)
    make_request "$method" "$route"
  done
  echo "---"
  sleep 0.5
done

echo "✅ Benchmark finalizado!"
