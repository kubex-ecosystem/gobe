# 🧠 Sistema de Triagem Inteligente - Discord MCP Hub

## 🎯 Objetivo

O sistema de triagem inteligente decide automaticamente:

- **SE** o bot deve responder a uma mensagem
- **COMO** o bot deve responder (tipo de resposta)
- **QUANDO** usar ferramentas MCP avançadas

## 🔍 Como Funciona

### Etapa 1: Filtros Básicos

- ❌ Mensagens muito curtas (< 2 caracteres)
- ❌ Apenas emojis ou caracteres especiais  
- ❌ Mensagens irrelevantes ou spam

### Etapa 2: Classificação Inteligente

| Tipo | Detecta | Exemplo | Resposta |
|------|---------|---------|----------|
| **command** | Comandos diretos | `!ping`, `!help` | Comando específico |
| **question** | Perguntas | "Como fazer...?", "Quando..." | Análise LLM + resposta |
| **task_request** | Solicitações | "Criar tarefa", "Lembrar de..." | Criação de tarefa |
| **analysis** | Análises | "Analise isso", "O que acha?" | Análise detalhada |
| **casual** | Conversa | "Oi bot", "Obrigado" | Resposta casual |

### Etapa 3: Processamento Específico

Cada tipo tem processamento especializado com fallbacks inteligentes.

## 🚀 Comandos para Testar

### ✅ Comandos Diretos (sempre respondem)

```plaintext
!ping
!help
!analyze Este é um texto para análise
!task Criar um relatório mensal
```

### ✅ Perguntas (detecta automaticamente)

```plaintext
Como funciona este sistema?
Quando devo usar esta ferramenta?
Por que o bot não respondeu?
Qual é a melhor forma de...?
```

### ✅ Solicitações de Tarefa (detecta automaticamente)

```plaintext
Preciso criar uma tarefa para amanhã
Quero fazer um lembrete sobre reunião
Adicionar no sistema: revisar código
```

### ✅ Pedidos de Análise (detecta automaticamente)

```plaintext
Analise este documento para mim
O que você pensa sobre esta ideia?
Avalie esta proposta
Review deste código
```

### ✅ Mensagens Casuais (detecta automaticamente)

```plaintext
Oi bot, tudo bem?
Obrigado pela ajuda!
Legal, bot muito útil
```

### ❌ Mensagens Ignoradas (não respondem)

```plaintext
kkk
😀😄😊
a
ok
hm
```

## 🔧 Configuração

O sistema funciona em **modo híbrido**:

- ✅ **Discord Real**: Bot conectado ao servidor real
- 🧠 **LLM Dev Mode**: Respostas simuladas para desenvolvimento
- 🔧 **MCP Ativo**: Ferramentas MCP funcionais

## 📊 Logs e Monitoramento

Cada mensagem gera logs detalhados:

```plaintext
🧠 Processando mensagem com LLM: [mensagem]
✅ Triagem aprovada - Tipo: [tipo]
❓ Processando pergunta: [conteúdo]
⏭️ Mensagem ignorada pela triagem: [motivo]
```

## 🎯 Teste Completo

Para testar todos os tipos, envie estas mensagens no Discord:

1. **Comando**: `!ping`
2. **Pergunta**: `Como funciona a triagem?`
3. **Tarefa**: `Criar uma tarefa de teste`
4. **Análise**: `Analise esta mensagem`
5. **Casual**: `Oi bot!`
6. **Ignorada**: `kkk`

## ✨ Próximos Passos

1. **LLM Real**: Conectar OpenAI API real
2. **MCP Avançado**: Ferramentas personalizadas
3. **Persistência**: Salvar tarefas em banco
4. **Web Interface**: Dashboard de monitoramento

## 🎉 Status

- ✅ Triagem inteligente funcionando
- ✅ Bot conectado ao Discord
- ✅ Classificação automática
- ✅ Fallbacks inteligentes
- ✅ Logs detalhados
- 🚧 LLM em modo dev (próximo passo)
