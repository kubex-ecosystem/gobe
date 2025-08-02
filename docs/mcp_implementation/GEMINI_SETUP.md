# 🧠 Configuração Gemini - Discord MCP Hub

## 🎯 Como Configurar Gemini

### Passo 1: Obter API Key do Gemini

1. Vá para: <https://aistudio.google.com/app/apikey>
2. Faça login com sua conta Google
3. Clique em "Create API Key"
4. Copie a API key (começa com "AI...")

### Passo 2: Configurar no Projeto

Edite o arquivo `.env` e substitua:

```bash
# De:
GEMINI_API_KEY=SUA_GEMINI_API_KEY_AQUI

# Para:
GEMINI_API_KEY=AIzaSy...sua_api_key_aqui
```

### Passo 3: Testar

```bash
./scripts/test-gemini.sh
```

## 🔧 Provedores Suportados

| Provedor | Variável | Exemplo API Key | Status |
|----------|----------|-----------------|--------|
| **Gemini** | `GEMINI_API_KEY` | `AIzaSy...` | ✅ Recomendado |
| **OpenAI** | `OPENAI_API_KEY` | `sk-...` | ✅ Suportado |
| **Dev Mode** | `dev_api_key` | `dev_api_key` | ✅ Para testes |

## 🚀 Teste Rápido (Sem API Key)

Se você quiser testar primeiro sem configurar Gemini:

```bash
# Usar modo dev temporariamente
export GEMINI_API_KEY="dev_api_key"
./scripts/test-gemini.sh
```

O sistema funcionará com mocks inteligentes!

## 🎯 Comandos para Testar com Gemini Real

Uma vez configurado com sua API key, teste no Discord:

### ✅ Perguntas Inteligentes

```plaintext
Como funciona inteligência artificial?
Qual a diferença entre Gemini e GPT?
Explique o que é machine learning
```

### ✅ Análises Complexas

```plaintext
Analise esta frase: "A tecnologia está mudando o mundo"
O que você pensa sobre inteligência artificial?
Avalie os prós e contras do trabalho remoto
```

### ✅ Criação de Tarefas

```plaintext
Criar uma tarefa para estudar Python
Preciso lembrar de fazer backup dos dados
Adicionar no cronograma: reunião com equipe
```

### ✅ Conversas Casuais

```plaintext
Oi Gemini, como você está?
Obrigado pela ajuda!
Que legal este sistema!
```

## 📊 Logs e Monitoramento

Com Gemini ativo, você verá logs como:

```plaintext
🧠 LLM usando Gemini: AIzaSy...
🧠 Processando mensagem com LLM: [sua mensagem]
✅ Triagem aprovada - Tipo: question
❓ Processando pergunta: [conteúdo]
```

## 🔄 Fallback Inteligente

Se o Gemini falhar ou não responder, o sistema:

1. ✅ Usa resposta mock inteligente
2. ✅ Continua funcionando normalmente
3. ✅ Loga o erro para debug

## 💡 Vantagens do Gemini

- 🚀 **Rápido**: Respostas em ~1-2 segundos
- 🧠 **Inteligente**: Compreensão contextual excelente
- 💰 **Econômico**: Mais barato que OpenAI
- 🔒 **Confiável**: Infraestrutura Google
- 🌍 **Multilíngue**: Suporte nativo ao português

## 🔧 Configuração Avançada

### Ajustar Temperatura

No arquivo `config/discord_config.json`:

```json
{
  "llm": {
    "provider": "gemini",
    "model": "gemini-pro",
    "temperature": 0.7,  // 0.0 = mais conservador, 1.0 = mais criativo
    "max_tokens": 1000
  }
}
```

### Modelos Disponíveis

- `gemini-pro`: Modelo padrão (recomendado)
- `gemini-pro-vision`: Para análise de imagens (futuro)

## 🎉 Status

- ✅ Gemini integrado
- ✅ Fallback inteligente
- ✅ Parsing JSON robusto
- ✅ Cache implementado
- ✅ Logs detalhados
- ✅ Auto-detecção de provedor

## 🚀 Próximos Passos

1. **Configurar sua API key**
2. **Testar com `./scripts/test-gemini.sh`**
3. **Enviar mensagens no Discord**
4. **Ver a mágica acontecer!** ✨
