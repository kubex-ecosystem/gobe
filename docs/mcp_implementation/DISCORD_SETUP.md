# 🤖 Discord MCP Hub - Configuração Completa

## 🎯 Problema Identificado

O erro `invalid_scope` que você está vendo acontece porque o Discord tem regras específicas para OAuth2. Aqui está a solução completa:

## 📋 URLs Importantes

### 1. 🔗 URL de Convite do Bot (Use esta!)
```
https://discord.com/api/oauth2/authorize?client_id=1344830702780420157&permissions=274877908992&scope=bot%20applications.commands
```

### 2. 🎯 URL de Redirect OAuth2 (Configurar no Discord Developer Portal)
```
https://4a1308bc7fab.ngrok.app/api/v1/oauth2/authorize
```

### 3. 🪝 Webhook URL (Já configurado)
```
https://discord.com/api/webhooks/1381317940649132162/KZro3msMCG1h_jl_eW-EGXPIldUpbRf8R0DC04bpFRcSOC4ZeW1HzMAGDvNdiO1jVcKj
```

## ⚙️ Configuração no Discord Developer Portal

### Passo 1: OAuth2 Settings
1. Vá para https://discord.com/developers/applications/1344830702780420157/oauth2/general
2. Em **Redirects**, adicione:
   ```
   https://4a1308bc7fab.ngrok.app/api/v1/oauth2/authorize
   ```

### Passo 2: Bot Permissions
1. Vá para https://discord.com/developers/applications/1344830702780420157/bot
2. Certifique-se que estas permissões estão habilitadas:
   - ✅ Send Messages
   - ✅ Read Messages/View Channels
   - ✅ Read Message History
   - ✅ Use Slash Commands
   - ✅ Embed Links
   - ✅ Attach Files

### Passo 3: Interactions Endpoint URL
1. Vá para https://discord.com/developers/applications/1344830702780420157/general
2. Em **Interactions Endpoint URL**, adicione:
   ```
   https://4a1308bc7fab.ngrok.app/api/v1/discord/interactions
   ```

## 🚀 Endpoints Implementados

Nosso servidor agora tem estes endpoints:

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/api/v1/oauth2/authorize` | GET | OAuth2 authorization |
| `/api/v1/oauth2/token` | POST | Token exchange |
| `/api/v1/discord/webhook/:id/:token` | POST | Webhook receiver |
| `/api/v1/discord/interactions` | POST | Slash commands |
| `/api/v1/health` | GET | Health check |

## 🔧 Como Testar

### 1. Iniciar o Servidor
```bash
cd /srv/apps/LIFE/KUBEX/booster
./scripts/test-discord.sh
```

### 2. Convidar o Bot
Use a URL de convite acima para adicionar o bot ao seu servidor

### 3. Testar Comandos
No Discord, digite:
- `!ping` - Teste básico
- `!help` - Lista de comandos
- `!analyze Esta é uma mensagem teste` - Análise com IA
- `!task Criar uma tarefa de exemplo` - Criação de tarefa

### 4. Verificar Logs
O servidor mostrará logs detalhados de todas as interações

## 🐛 Resolução de Problemas

### Error: invalid_scope
- ✅ **Resolvido**: URLs corretas implementadas
- ✅ **Causa**: Discord OAuth2 requer endpoints específicos
- ✅ **Solução**: Usar URL de convite direta para bots

### Bot não responde
1. Verifique se o bot está online: logs devem mostrar "Discord bot logged in"
2. Verifique permissões no servidor
3. Teste com `!ping` primeiro

### Webhook não funciona
1. Verifique se a URL do webhook está correta
2. Verifique se o endpoint `/api/v1/discord/webhook/:id/:token` está funcionando
3. Teste com POST request manual

## 🎯 Próximos Passos

1. **Adicione o redirect URI no Discord Developer Portal**
2. **Use a URL de convite para adicionar o bot**
3. **Teste os comandos no Discord**
4. **Configure slash commands (opcional)**

## 📚 Documentação Útil

- [Discord OAuth2](https://discord.com/developers/docs/topics/oauth2)
- [Discord Bot Permissions](https://discord.com/developers/docs/topics/permissions)
- [Discord Webhooks](https://discord.com/developers/docs/resources/webhook)
- [Discord Interactions](https://discord.com/developers/docs/interactions/receiving-and-responding)

## 🤝 Integração MCP

Nosso sistema agora está totalmente integrado:

```
Discord Bot → HTTP Server → Hub → MCP Server → Tools
     ↓              ↓         ↓        ↓         ↓
  Comandos    OAuth2/Webhook  Logic   Protocol  Actions
```

O MCP está funcionando através dos comandos `!analyze` e `!task` que usam as ferramentas MCP implementadas.
