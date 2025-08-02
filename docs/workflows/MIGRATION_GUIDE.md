# 🎯 Guia de Migração: Como Aplicar Este Workflow em Outros Projetos

Este é um template universal que pode ser adaptado para qualquer projeto Go. Siga estes passos para implementar um workflow de release de classe mundial.

## 🚀 Passo a Passo de Implementação

### 1. **Copie o Workflow Base**

```bash
# Copie o arquivo para seu projeto
cp .github/workflows/release.yml seu-projeto/.github/workflows/
```

### 2. **Adaptações Necessárias**

#### **a) Nome do Binário**

```yaml
# Linha ~67: Altere o padrão do nome do binário
BIN_NAME=$(ls -1t bin/SEU-PROJETO-${{ matrix.platform }}-${{ matrix.goarch }}* | head -n1)

# Exemplos:
# bin/myapp-linux-amd64
# bin/api-server-windows-amd64
# bin/cli-tool-darwin-amd64
```

#### **b) Comando de Build**

```yaml
# Linha ~65: Adapte para seu Makefile
make build ${{ matrix.platform }} all

# Se você usa outros comandos:
# go build -o bin/myapp-${{ matrix.platform }}-${{ matrix.goarch }}
# ./scripts/build.sh ${{ matrix.platform }} ${{ matrix.goarch }}
```

#### **c) Dependências do Sistema**

```yaml
# Linha ~85: Adapte suas dependências
sudo apt-get install -y upx zip tar curl gzip
# Adicione outras dependências específicas do seu projeto
# sudo apt-get install -y libssl-dev libpq-dev
```

### 3. **Configurações por Tipo de Projeto**

#### **🌐 API/Web Server**

```yaml
# Adicione health check antes do release
- name: 🔍 Health Check
  run: |
    # Build temporário para teste
    make build linux test
    timeout 10s ./bin/myapi-linux-amd64 &
    sleep 5
    curl -f http://localhost:8080/health || exit 1
```

#### **🛠️ CLI Tool**

```yaml
# Adicione teste de execução
- name: 🧪 CLI Test
  run: |
    make build linux all
    ./bin/mycli-linux-amd64 --version
    ./bin/mycli-linux-amd64 --help
```

#### **📚 Library**

```yaml
# Foque nos testes
- name: 🧪 Extended Tests
  run: |
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
```

### 4. **Personalizações Avançadas**

#### **🐳 Suporte a Docker**

```yaml
docker:
  name: 🐳 Docker Build
  runs-on: ubuntu-latest
  needs: [setup, build]
  steps:
    - name: 🏗️ Build Docker Image
      run: |
        docker build -t ${{ github.repository }}:${{ needs.setup.outputs.version }} .
        docker tag ${{ github.repository }}:${{ needs.setup.outputs.version }} ${{ github.repository }}:latest
    
    - name: 📤 Push to Registry
      run: |
        echo ${{ secrets.GITHUB_TOKEN }} | docker login ghcr.io -u ${{ github.actor }} --password-stdin
        docker push ghcr.io/${{ github.repository }}:${{ needs.setup.outputs.version }}
```

#### **📱 Notificações**

```yaml
notify:
  name: 📢 Notifications
  runs-on: ubuntu-latest
  needs: [release]
  if: always()
  steps:
    - name: 📱 Slack Notification
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ needs.release.result }}
        text: "🚀 Release ${{ needs.setup.outputs.version }} completed!"
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

### 5. **Estrutura de Projeto Recomendada**

```plaintext
seu-projeto/
├── .github/
│   └── workflows/
│       ├── release.yml          # Workflow de release
│       ├── ci.yml              # CI contínuo
│       └── README.md           # Documentação
├── bin/                        # Binários gerados
├── cmd/                        # Entrypoints
├── internal/                   # Código interno
├── pkg/                        # Código público
├── scripts/                    # Scripts de build
├── Makefile                    # Build automation
├── go.mod
├── go.sum
└── README.md
```

### 6. **Makefile Template**

```makefile
# Variáveis
APP_NAME := $(shell basename $(CURDIR))
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Flags de build
LDFLAGS := -X main.Version=$(VERSION) \
           -X main.BuildTime=$(BUILD_TIME) \
           -X main.Commit=$(COMMIT) \
           -s -w

# Targets principais
.PHONY: build build-all clean test

build:
 @echo "Building $(APP_NAME)..."
 @mkdir -p bin
 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-$(GOOS)-$(GOARCH) ./cmd

build-all: build-linux build-windows build-darwin

build-linux:
 @$(MAKE) build GOOS=linux GOARCH=amd64

build-windows:
 @$(MAKE) build GOOS=windows GOARCH=amd64

build-darwin:
 @$(MAKE) build GOOS=darwin GOARCH=amd64
```

### 7. **Secrets e Configurações**

#### **GitHub Secrets (se necessário)**

```yaml
# Para projetos que precisam de secrets adicionais
SLACK_WEBHOOK_URL         # Notificações Slack
DOCKER_REGISTRY_TOKEN     # Push para registries
CODE_SIGNING_CERT         # Assinatura de código
```

#### **Permissions**

```yaml
permissions:
  contents: write         # Criar releases
  packages: write         # Push containers
  security-events: write  # SARIF upload
  id-token: write         # OIDC tokens
```

### 8. **Debugging e Troubleshooting**

#### **Debug Mode**

```yaml
# Adicione para debug detalhado
- name: 🐛 Debug Information
  run: |
    echo "Go Version: $(go version)"
    echo "Git Status: $(git status --porcelain)"
    echo "Environment Variables:"
    env | grep -E "(GO|GIT|GITHUB)" | sort
```

#### **Logs Detalhados**

```yaml
# Para troubleshooting de builds
- name: 📊 Build Diagnostics
  if: failure()
  run: |
    echo "Build failed. Diagnostics:"
    ls -la bin/ || echo "No bin directory"
    go env
    make --version
```

### 9. **Exemplos de Uso**

#### **Para Microserviços**

```yaml
# Multi-service build
strategy:
  matrix:
    service: [api, worker, scheduler]
    platform: [linux, windows, darwin]
    
steps:
  - name: Build ${{ matrix.service }}
    run: make build-${{ matrix.service }} ${{ matrix.platform }}
```

#### **Para Monorepos**

```yaml
# Conditional builds baseado em mudanças
- name: Detect Changes
  id: changes
  run: |
    if git diff --name-only HEAD^..HEAD | grep -q "^api/"; then
      echo "api=true" >> $GITHUB_OUTPUT
    fi
    if git diff --name-only HEAD^..HEAD | grep -q "^worker/"; then
      echo "worker=true" >> $GITHUB_OUTPUT
    fi
```

## ✅ Checklist Final

- [ ] ✏️ **Nomes de binários** adaptados
- [ ] 🏗️ **Comandos de build** configurados  
- [ ] 📦 **Dependências** atualizadas
- [ ] 🧪 **Testes** específicos adicionados
- [ ] 📝 **Release notes** personalizadas
- [ ] 🔐 **Secrets** configurados (se necessário)
- [ ] 📱 **Notificações** configuradas (opcional)
- [ ] 🧹 **Cleanup** adaptado

## 🎉 Resultado Final

Com essas adaptações, você terá:

✨ **Builds paralelos** para múltiplas plataformas
⚡ **Cache inteligente** para builds rápidos
🔒 **Segurança integrada** com scans automáticos  
📦 **Releases profissionais** com assets organizados
📊 **Documentação automática** e instruções de uso
🧹 **Cleanup automático** para otimização de resources

---

**🚀 Agora seu projeto tem um workflow de release de classe mundial!**
