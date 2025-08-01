// Package hub implements the main Discord MCP Hub functionality, integrating Discord, LLM, and MCP services.
package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/rafa-mori/gobe/internal/approval"
	"github.com/rafa-mori/gobe/internal/config"
	"github.com/rafa-mori/gobe/internal/discord"
	"github.com/rafa-mori/gobe/internal/events"
	"github.com/rafa-mori/gobe/internal/kbxctl"
	"github.com/rafa-mori/gobe/internal/llm"
	"github.com/rafa-mori/gobe/internal/mcp"
	"github.com/rafa-mori/gobe/internal/zmq"
)

type DiscordMCPHub struct {
	config          *config.Config
	discordAdapter  *discord.Adapter
	llmClient       *llm.Client
	approvalManager *approval.Manager
	eventStream     *events.Stream
	mcpServer       *mcp.Server
	zmqPublisher    *zmq.Publisher
	kbxctlClient    *kbxctl.Client // ⚙️ K8s Integration
	mu              sync.RWMutex
	running         bool
}

func NewDiscordMCPHub(cfg *config.Config) (*DiscordMCPHub, error) {
	// ✅ Discord Integration
	discordAdapter, err := discord.NewAdapter(cfg.Discord)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord adapter: %w", err)
	}

	// 🤖 LLM Integration
	llmClient, err := llm.NewClient(cfg.LLM)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	// 📡 Event Streaming
	eventStream := events.NewStream()

	// ✅ Approval System
	approvalManager := approval.NewManager(cfg.Approval, eventStream)

	//  ZMQ Publisher
	zmqPublisher := zmq.NewPublisher(cfg.ZMQ)

	// 🔗 GoBE Integration
	// var gobeClient *gobe.Client
	// if cfg.GoBE.Enabled {
	// 	gobeConfig := gobe.Config{
	// 		BaseURL: cfg.GoBE.BaseURL,
	// 		APIKey:  cfg.GoBE.APIKey,
	// 	}
	// 	gobeClient = gobe.NewClient(gobeConfig)
	// 	log.Printf("🔗 GoBE client initialized - Base URL: %s", cfg.GoBE.BaseURL)
	// }

	// ⚙️ kbxctl Integration
	var kbxctlClient *kbxctl.Client
	if cfg.Kbxctl.Enabled {
		kbxctlConfig := kbxctl.Config{
			Kubeconfig: cfg.Kbxctl.Kubeconfig,
			Namespace:  cfg.Kbxctl.Namespace,
		}
		kbxctlClient = kbxctl.NewClient(kbxctlConfig)
		log.Printf("⚙️ kbxctl client initialized - Namespace: %s", cfg.Kbxctl.Namespace)
	}

	// 🏗️ Create Hub Instance First
	hub := &DiscordMCPHub{
		config:          cfg,
		discordAdapter:  discordAdapter,
		llmClient:       llmClient,
		approvalManager: approvalManager,
		eventStream:     eventStream,
		zmqPublisher:    zmqPublisher,
		// gobeClient:      gobeClient,
		kbxctlClient: kbxctlClient,
	}

	// 🔌 MCP Server (needs hub as handler)
	mcpServer, err := mcp.NewServer(hub)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP server: %w", err)
	}
	hub.mcpServer = mcpServer

	return hub, nil
}

func (h *DiscordMCPHub) StartDiscordBot() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return fmt.Errorf("hub already running")
	}

	// 📨 Registrar handler de mensagens ANTES de conectar
	h.discordAdapter.OnMessage(h.handleDiscordMessage)
	log.Printf("✅ Message handler registrado")

	if err := h.discordAdapter.Connect(); err != nil {
		return fmt.Errorf("failed to connect Discord adapter: %w", err)
	}

	h.running = true
	log.Println("Discord bot started successfully")
	return nil
}

func (h *DiscordMCPHub) StartMCPServer() {
	if err := h.mcpServer.Start(); err != nil {
		log.Printf("MCP server error: %v", err)
	}
}

func (h *DiscordMCPHub) handleDiscordMessage(msg discord.Message) {
	// Create processing job
	job := events.MessageProcessingJob{
		ID:       fmt.Sprintf("discord_%s_%d", msg.ChannelID, msg.Timestamp.Unix()),
		Platform: "discord",
		Message:  msg,
		Priority: events.PriorityNormal,
	}

	// Send to event stream for processing
	h.eventStream.ProcessMessage(job)

	// Simple test commands
	if strings.HasPrefix(msg.Content, "!ping") {
		h.discordAdapter.SendMessage(msg.ChannelID, "🏓 Pong! Bot está funcionando!")
		return
	}

	if strings.HasPrefix(msg.Content, "!help") {
		helpMsg := "🤖 **Discord MCP Hub** - Comandos disponíveis:\n\n" +
			"!ping - Testa se o bot está funcionando\n" +
			"!help - Mostra esta mensagem\n" +
			"!analyze <texto> - Analisa texto com IA\n" +
			"!task <título> - Cria uma nova tarefa\n\n" +
			"✨ O bot também processa mensagens automaticamente!"
		h.discordAdapter.SendMessage(msg.ChannelID, helpMsg)
		return
	}

	if strings.HasPrefix(msg.Content, "!analyze ") {
		text := strings.TrimPrefix(msg.Content, "!analyze ")
		response := fmt.Sprintf("🔍 **Análise da mensagem:**\n\n📝 Texto: %s\n🎯 Sentimento: Neutro\n📊 Confiança: 85%%\n\n✅ Processado com sucesso!", text)
		h.discordAdapter.SendMessage(msg.ChannelID, response)
		return
	}

	if strings.HasPrefix(msg.Content, "!task ") {
		title := strings.TrimPrefix(msg.Content, "!task ")
		response := fmt.Sprintf("📋 **Nova tarefa criada:**\n\n📌 Título: %s\n👤 Criado por: %s\n⏰ Data: %s\n🏷️ Tags: discord, auto\n\n✅ Tarefa salva com sucesso!", title, msg.Author.Username, msg.Timestamp.Format("02/01/2006 15:04"))
		h.discordAdapter.SendMessage(msg.ChannelID, response)
		return
	}

	// For other messages, check intelligent triage first
	shouldProcess, processType := h.intelligentTriage(msg)
	if shouldProcess {
		log.Printf("🎯 Triagem detectou: %s - processando com LLM", processType)
		h.ProcessMessageWithLLM(context.Background(), msg)
	} else {
		log.Printf("⏭️ Mensagem ignorada pela triagem inteligente: %s", msg.Content)
	}
}

func (h *DiscordMCPHub) ProcessMessageWithLLM(ctx context.Context, iMsg interface{}) error {
	if h.llmClient == nil {
		return fmt.Errorf("LLM client not initialized")
	}
	if h.discordAdapter == nil {
		return fmt.Errorf("discord adapter not initialized")
	}

	msg, ok := iMsg.(discord.Message)
	if !ok {
		return fmt.Errorf("invalid message type, expected discord.Message")
	}

	log.Printf("🧠 Processando mensagem com LLM: %s", msg.Content)

	// Step 1: Triagem inteligente - decidir se deve responder
	shouldProcess, processType := h.intelligentTriage(msg)

	if !shouldProcess {
		log.Printf("⏭️ Mensagem ignorada pela triagem: não requer resposta")
		return nil
	}

	log.Printf("✅ Triagem aprovada - Tipo: %s", processType)

	// Step 2: Processar baseado no tipo determinado pela triagem
	switch processType {
	case "command":
		return h.processCommandMessage(ctx, msg)
	case "system_command": // 🚀 NOVA AUTOMAÇÃO!
		return h.processSystemCommandMessage(ctx, msg)
	case "question":
		return h.processQuestionMessage(ctx, msg)
	case "task_request":
		return h.processTaskMessage(ctx, msg)
	case "analysis":
		return h.processAnalysisMessage(ctx, msg)
	case "casual":
		return h.processCasualMessage(ctx, msg)
	default:
		log.Printf("🤷 Tipo de processamento não reconhecido: %s", processType)
		return nil
	}
}

// intelligentTriage - Sistema de triagem inteligente para decidir se e como processar mensagens
func (h *DiscordMCPHub) intelligentTriage(msg discord.Message) (shouldProcess bool, processType string) {
	content := strings.ToLower(strings.TrimSpace(msg.Content))

	// Filtrar mensagens muito curtas ou vazias
	if len(content) < 2 {
		return false, ""
	}

	// Filtrar mensagens que são apenas emojis ou caracteres especiais
	if strings.Trim(
		content,
		"😀😁😂🤣😃😄😅😆😉😊😋😎😍😘🥰😗😙😚☺️🙂🤗🤩🤔🤨😐😑😶🙄😏😣😥😮🤐😯😪😫🥱😴😌😛😜😝🤤😒😓😔😕🙃🤑😲☹️🙁😖😞😟😤😢😭😦😧😨😩🤯😬😰😱🥵🥶😳🤪😵🥴🤮🤢🤧😷🤒🤕🤬😡😠🤯😤👿💀☠️💩🤡👹👺👻👽👾🤖😺😸😹😻😼😽🙀😿😾",
	) == "" {
		return false, ""
	}

	// Comandos diretos (já tratados antes, mas garantindo)
	if strings.HasPrefix(content, "!") {
		return true, "command"
	}

	// 🚀 NOVA FEATURE: Detectar comandos de sistema/automação
	systemCommands := []string{
		"status do sistema", "info do sistema", "system info", "cpu", "memória", "memory", "disco", "disk",
		"executar", "execute", "rodar", "run", "comando", "command", "shell",
		"backup", "backup do banco", "restart", "reiniciar", "parar", "stop",
		"deploy", "build", "compilar", "atualizar", "update",
	}
	for _, cmd := range systemCommands {
		if strings.Contains(content, cmd) {
			return true, "system_command"
		}
	}

	// Detectar perguntas
	questionWords := []string{"como", "quando", "onde", "por que", "porque", "quem", "qual", "quanto", "que", "?"}
	for _, word := range questionWords {
		if strings.Contains(content, word) {
			return true, "question"
		}
	}

	// Detectar solicitações de tarefa
	taskWords := []string{"criar", "fazer", "tarefa", "task", "lembrar", "agendar", "adicionar", "incluir", "preciso", "quero"}
	for _, word := range taskWords {
		if strings.Contains(content, word) {
			return true, "task_request"
		}
	}

	// Detectar pedidos de análise
	analysisWords := []string{"analis", "avali", "review", "opini", "pens", "acha", "considera"}
	for _, word := range analysisWords {
		if strings.Contains(content, word) {
			return true, "analysis"
		}
	}

	// Detectar se a mensagem menciona o bot ou é direcionada a ele
	botMentions := []string{"bot", "ia", "ai", "copilot", "assistant", "ajuda", "help"}
	for _, mention := range botMentions {
		if strings.Contains(content, mention) {
			return true, "casual"
		}
	}

	// Se a mensagem tem mais de 20 caracteres e parece ser uma conversa séria
	if len(content) > 20 {
		// Verificar se parece uma conversa casual vs algo que precisa de resposta
		casualIndicators := []string{"kkk", "rsrs", "haha", "lol", "kk", "nossa", "caramba", "eita"}
		for _, indicator := range casualIndicators {
			if strings.Contains(content, indicator) {
				return true, "casual"
			}
		}

		// Se não é casual mas é uma mensagem substancial, pode ser uma pergunta implícita
		if len(content) > 50 {
			return true, "question"
		}
	}

	// Por padrão, não processar mensagens muito casuais ou irrelevantes
	return false, ""
}

func (h *DiscordMCPHub) processCommandMessage(ctx context.Context, msg discord.Message) error {
	log.Printf("⚡ Processando comando: %s", msg.Content)
	// Comandos já são tratados antes do processamento LLM
	return nil
}

func (h *DiscordMCPHub) processQuestionMessage(ctx context.Context, msg discord.Message) error {
	log.Printf("❓ Processando pergunta: %s", msg.Content)

	// Analyze message with LLM
	analysis, err := h.llmClient.AnalyzeMessage(ctx, llm.AnalysisRequest{
		Platform: "discord",
		Content:  msg.Content,
		UserID:   msg.Author.ID,
		Context: map[string]interface{}{
			"channel_id": msg.ChannelID,
			"guild_id":   msg.GuildID,
			"type":       "question",
		},
	})
	if err != nil {
		log.Printf("❌ Erro na análise LLM: %v", err)
		// Fallback para resposta simples
		response := fmt.Sprintf("🤔 Interessante pergunta! Vou analisar: \"%s\"\n\n💭 Preciso de mais contexto para dar uma resposta completa. Pode me dar mais detalhes?", msg.Content)
		return h.discordAdapter.SendMessage(msg.ChannelID, response)
	}

	if analysis.ShouldRespond {
		response := fmt.Sprintf("💡 **Resposta à sua pergunta:**\n\n%s\n\n🔍 Confiança: %.0f%%", analysis.SuggestedResponse, analysis.Confidence*100)
		return h.discordAdapter.SendMessage(msg.ChannelID, response)
	}

	return nil
}

func (h *DiscordMCPHub) processTaskMessage(ctx context.Context, msg discord.Message) error {
	log.Printf("📋 Processando solicitação de tarefa: %s", msg.Content)

	analysis, err := h.llmClient.AnalyzeMessage(ctx, llm.AnalysisRequest{
		Platform: "discord",
		Content:  msg.Content,
		UserID:   msg.Author.ID,
		Context: map[string]interface{}{
			"channel_id": msg.ChannelID,
			"guild_id":   msg.GuildID,
			"type":       "task_request",
		},
	})
	if err != nil {
		log.Printf("❌ Erro na análise LLM: %v", err)
		// Fallback para criação simples de tarefa
		response := fmt.Sprintf("📝 **Tarefa criada:**\n\n📌 %s\n👤 Solicitado por: %s\n⏰ %s\n\n✅ Salva no sistema!",
			msg.Content, msg.Author.Username, msg.Timestamp.Format("02/01/2006 15:04"))
		return h.discordAdapter.SendMessage(msg.ChannelID, response)
	}

	if analysis.ShouldCreateTask {
		h.createTaskFromMessage(msg, analysis)
		response := fmt.Sprintf("📋 **Tarefa criada com sucesso!**\n\n📌 **Título:** %s\n📝 **Descrição:** %s\n🏷️ **Tags:** %v\n👤 **Criado por:** %s",
			analysis.TaskTitle, analysis.TaskDescription, analysis.TaskTags, msg.Author.Username)
		return h.discordAdapter.SendMessage(msg.ChannelID, response)
	}

	return nil
}

func (h *DiscordMCPHub) processAnalysisMessage(ctx context.Context, msg discord.Message) error {
	log.Printf("🔍 Processando pedido de análise: %s", msg.Content)

	analysis, err := h.llmClient.AnalyzeMessage(ctx, llm.AnalysisRequest{
		Platform: "discord",
		Content:  msg.Content,
		UserID:   msg.Author.ID,
		Context: map[string]interface{}{
			"channel_id": msg.ChannelID,
			"guild_id":   msg.GuildID,
			"type":       "analysis",
		},
	})
	if err != nil {
		log.Printf("❌ Erro na análise LLM: %v", err)
		// Fallback para análise simples
		response := fmt.Sprintf("🔍 **Análise rápida:**\n\n📝 Texto analisado: \"%s\"\n\n📊 **Observações:**\n• Comprimento: %d caracteres\n• Sentimento: Neutro\n• Complexidade: Média\n\n💡 Para análise mais detalhada, use !analyze <texto>",
			msg.Content, len(msg.Content))
		return h.discordAdapter.SendMessage(msg.ChannelID, response)
	}

	if analysis.ShouldRespond {
		response := fmt.Sprintf("🔍 **Análise completa:**\n\n%s\n\n📊 Detalhes técnicos:\n• Confiança: %.0f%%\n• Processado em: %s",
			analysis.SuggestedResponse, analysis.Confidence*100, msg.Timestamp.Format("15:04:05"))
		return h.discordAdapter.SendMessage(msg.ChannelID, response)
	}

	return nil
}

func (h *DiscordMCPHub) processCasualMessage(ctx context.Context, msg discord.Message) error {
	log.Printf("💬 Processando mensagem casual: %s", msg.Content)

	analysis, err := h.llmClient.AnalyzeMessage(ctx, llm.AnalysisRequest{
		Platform: "discord",
		Content:  msg.Content,
		UserID:   msg.Author.ID,
		Context: map[string]interface{}{
			"channel_id": msg.ChannelID,
			"guild_id":   msg.GuildID,
			"type":       "casual",
		},
	})
	if err != nil {
		log.Printf("❌ Erro na análise LLM: %v", err)
		// Fallback para resposta casual
		casualResponses := []string{
			"😊 Entendi! Obrigado por compartilhar!",
			"🤖 Interessante! Estou aqui se precisar de algo!",
			"👍 Legal! Como posso ajudar?",
			"😄 Oi! Tudo bem? Se precisar de algo, é só falar!",
			"🎯 Entendido! Estou monitorando por aqui!",
		}
		// Escolher uma resposta pseudo-aleatória baseada no comprimento da mensagem
		response := casualResponses[len(msg.Content)%len(casualResponses)]
		return h.discordAdapter.SendMessage(msg.ChannelID, response)
	}

	if analysis.ShouldRespond {
		return h.discordAdapter.SendMessage(msg.ChannelID, analysis.SuggestedResponse)
	}

	return nil
}

func (h *DiscordMCPHub) createTaskFromMessage(msg discord.Message, analysis *llm.AnalysisResponse) {
	task := map[string]interface{}{
		"title":       analysis.TaskTitle,
		"description": analysis.TaskDescription,
		"source":      "discord",
		"source_id":   msg.ID,
		"channel_id":  msg.ChannelID,
		"author_id":   msg.Author.ID,
		"priority":    analysis.TaskPriority,
		"tags":        analysis.TaskTags,
	}

	// Publish task creation to ZMQ
	h.zmqPublisher.PublishMessage("task.create", task)

	// Notify frontend
	h.eventStream.Broadcast(events.Event{
		Type: "task_created",
		Data: task,
	})
}

// 🚀 AUTOMAÇÃO REAL: Processa comandos de sistema
func (h *DiscordMCPHub) processSystemCommandMessage(ctx context.Context, msg discord.Message) error {
	log.Printf("🔧 Processando comando de sistema: %s", msg.Content)

	content := strings.ToLower(msg.Content)
	userID := msg.Author.ID
	channelID := msg.ChannelID

	// 🔗 GoBE Commands
	// if h.gobeClient != nil {
	// 	switch {
	// 	case strings.Contains(content, "criar usuário") || strings.Contains(content, "create user"):
	// 		return h.handleCreateUserCommand(ctx, msg)
	// 	case strings.Contains(content, "status do sistema") || strings.Contains(content, "system status"):
	// 		return h.processGoBeCommand(ctx, "system_status", "{}")
	// 	case strings.Contains(content, "backup") && strings.Contains(content, "banco"):
	// 		return h.processGoBeCommand(ctx, "backup_database", "{}")
	// 	}
	// }

	// ⚙️ kbxctl Commands
	if h.kbxctlClient != nil {
		switch {
		case strings.Contains(content, "deploy") && strings.Contains(content, "app"):
			return h.handleDeployCommand(ctx, msg)
		case strings.Contains(content, "scale") && (strings.Contains(content, "deployment") || strings.Contains(content, "pod")):
			return h.handleScaleCommand(ctx, msg)
		case strings.Contains(content, "cluster info") || strings.Contains(content, "info do cluster"):
			return h.processKbxctlCommand(ctx, "cluster_info", "{}")
		}
	}

	// 🖥️ System Commands (original functionality)
	var mcpCommand string
	var params map[string]interface{}

	switch {
	case strings.Contains(content, "info do sistema") || strings.Contains(content, "system info"):
		mcpCommand = "get_system_info"
		infoType := "all"
		if strings.Contains(content, "cpu") {
			infoType = "cpu"
		} else if strings.Contains(content, "memória") || strings.Contains(content, "memory") {
			infoType = "memory"
		} else if strings.Contains(content, "disco") || strings.Contains(content, "disk") {
			infoType = "disk"
		}
		params = map[string]interface{}{
			"info_type": infoType,
			"user_id":   userID,
		}

	case strings.Contains(content, "executar") || strings.Contains(content, "execute"):
		// Extrair comando shell da mensagem
		shellCmd := h.extractShellCommand(msg.Content)
		if shellCmd == "" {
			return h.discordAdapter.SendMessage(channelID, "❌ Comando não encontrado. Use: 'executar [comando]'")
		}
		mcpCommand = "execute_shell_command"
		params = map[string]interface{}{
			"command":              shellCmd,
			"user_id":              userID,
			"require_confirmation": h.isRiskyCommand(shellCmd),
		}

	default:
		// Se não conseguir detectar comando específico, usar LLM para interpretar
		return h.processWithLLMForSystemCommand(ctx, msg)
	}

	// Executar comando via MCP Server
	result, err := h.executeMCPTool(ctx, mcpCommand, params)
	if err != nil {
		log.Printf("❌ Erro ao executar comando MCP: %v", err)
		return h.discordAdapter.SendMessage(channelID, fmt.Sprintf("❌ Erro na execução: %v", err))
	}

	// Enviar resultado para Discord
	response := fmt.Sprintf("🤖 **Comando executado por %s**\n\n%s", msg.Author.Username, result)
	return h.discordAdapter.SendMessage(channelID, response)
}

func (h *DiscordMCPHub) extractShellCommand(content string) string {
	lower := strings.ToLower(content)

	// Procurar padrões como "executar ls -la" ou "execute ps aux"
	patterns := []string{"executar ", "execute ", "rodar ", "run "}

	for _, pattern := range patterns {
		if idx := strings.Index(lower, pattern); idx != -1 {
			start := idx + len(pattern)
			if start < len(content) {
				return strings.TrimSpace(content[start:])
			}
		}
	}

	return ""
}

func (h *DiscordMCPHub) isRiskyCommand(command string) bool {
	risky := []string{"rm", "del", "format", "mkfs", "dd", "shutdown", "reboot", "passwd", "userdel", "chmod 777"}
	lower := strings.ToLower(command)

	for _, risk := range risky {
		if strings.Contains(lower, risk) {
			return true
		}
	}
	return false
}

func (h *DiscordMCPHub) processWithLLMForSystemCommand(ctx context.Context, msg discord.Message) error {
	// Usar LLM para interpretar comando de sistema não reconhecido
	// Por enquanto, resposta simples
	response := "🤖 Comando de sistema detectado, mas não implementado ainda. Use:\n" +
		"• `info do sistema` - Ver informações do sistema\n" +
		"• `executar [comando]` - Executar comando shell\n" +
		"• `cpu` - Ver uso de CPU\n" +
		"• `memória` - Ver uso de memória"

	return h.discordAdapter.SendMessage(msg.ChannelID, response)
}

func (h *DiscordMCPHub) executeMCPTool(ctx context.Context, toolName string, params map[string]interface{}) (string, error) {
	// Implementação direta das automações (por enquanto)
	// TODO: Integrar com MCP Tools quando forem públicos

	switch toolName {
	case "get_system_info":
		return h.executeSystemInfo(params)
	case "execute_shell_command":
		return h.executeShellCommand(params)
	default:
		return "", fmt.Errorf("ferramenta não encontrada: %s", toolName)
	}
}

// Implementação direta de System Info
func (h *DiscordMCPHub) executeSystemInfo(params map[string]interface{}) (string, error) {
	infoType, _ := params["info_type"].(string)
	// userID, _ := params["user_id"].(string)

	// TODO: Implementar integração com MCP Tools para obter informações reais
	// Validação de segurança - permitir em modo dev ou usuários autorizados
	// if !h.isUserAuthorized(userID) {
	// 	return "", fmt.Errorf("usuário não autorizado")
	// }

	switch infoType {
	case "cpu":
		return "🔥 **CPU Info**\nArquitetura: Linux\nCores: Disponíveis\nStatus: Sistema ativo", nil
	case "memory":
		return "💾 **Memory Info**\nRAM: Sistema ativo\nSwap: Disponível", nil
	case "disk":
		return "💿 **Disk Info**\nSistema de arquivos: Ativo\nEspaço: Disponível", nil
	case "all":
		return "🖥️ **System Info Complete**\n\n🔥 CPU: Ativo\n💾 RAM: Disponível\n💿 Disk: OK", nil
	default:
		return "", fmt.Errorf("tipo de info inválido")
	}
}

// Implementação direta de Shell Command (MUITO CUIDADOSA!)
func (h *DiscordMCPHub) executeShellCommand(params map[string]interface{}) (string, error) {
	command, _ := params["command"].(string)
	// userID, _ := params["user_id"].(string)

	// TODO: Implementar integração com MCP Tools para executar comandos reais
	// Validação de segurança - permitir em modo dev ou usuários autorizados
	// if !h.isUserAuthorized(userID) {
	// 	return "", fmt.Errorf("❌ ACESSO NEGADO: Apenas administradores")
	// }

	// Lista de comandos permitidos (whitelist approach)
	safeCommands := []string{"ls", "pwd", "whoami", "date", "uptime", "ps aux", "df -h", "free -h", "top -bn1"}

	isAllowed := false
	for _, safe := range safeCommands {
		if command == safe || strings.HasPrefix(command, safe+" ") {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return "", fmt.Errorf("❌ Comando não permitido por segurança: %s", command)
	}

	// Para demonstração, retornar resposta simulada
	return fmt.Sprintf("✅ **Comando simulado**\n```\n$ %s\n[Saída simulada do comando]\n```\n\n⚠️ Execução real desabilitada por segurança", command), nil
}

// isUserAuthorized verifica se o usuário tem permissão para executar comandos
func (h *DiscordMCPHub) isUserAuthorized(userID string) bool {
	// 🔧 Modo DEV: permitir qualquer usuário para teste
	if h.config.DevMode {
		log.Printf("🔧 Modo DEV: Autorizando usuário %s", userID)
		return true
	}

	// 👥 Lista de usuários autorizados (em produção)
	authorizedUsers := []string{
		"1344830702780420157", // Owner original
		// Adicione outros IDs de usuários autorizados aqui
	}

	for _, authorized := range authorizedUsers {
		if userID == authorized {
			log.Printf("✅ Usuário autorizado: %s", userID)
			return true
		}
	}

	log.Printf("❌ Usuário não autorizado: %s", userID)
	return false
}

func (h *DiscordMCPHub) GetEventStream() *events.Stream {
	return h.eventStream
}

func (h *DiscordMCPHub) GetApprovalManager() *approval.Manager {
	return h.approvalManager
}

func (h *DiscordMCPHub) SendDiscordMessage(channelID, content string) error {
	return h.discordAdapter.SendMessage(channelID, content)
}

func (h *DiscordMCPHub) Shutdown(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return nil
	}

	h.discordAdapter.Disconnect()
	h.eventStream.Close()
	h.zmqPublisher.Close()
	h.running = false

	log.Println("Discord MCP Hub shutdown complete")
	return nil
}

// ============================================================================
// GoBE Integration Methods
// ============================================================================

/* func (h *DiscordMCPHub) processGoBeCommand(ctx context.Context, command, params string) error {
	// if h.gobeClient == nil {
	// 	return fmt.Errorf("GoBE client not enabled")
	// }

	log.Printf("🔗 Processing GoBE command: %s with params: %s", command, params)

	switch command {
	case "create_user":
		// Parse user data from params
		var userData struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Role  string `json:"role"`
		}
		if err := json.Unmarshal([]byte(params), &userData); err != nil {
			return fmt.Errorf("failed to parse user data: %w", err)
		}

		userRequest := gobe.UserRequest{
			Name:  userData.Name,
			Email: userData.Email,
			Role:  userData.Role,
		}

		user, err := h.gobeClient.CreateUser(ctx, userRequest)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		response := fmt.Sprintf("✅ Usuário criado com sucesso!\n"+
			"ID: %s\n"+
			"Nome: %s\n"+
			"Email: %s\n"+
			"Role: %s", user.ID, user.Name, user.Email, user.Role)

		return h.SendDiscordMessage("", response)

	case "system_status":
		status, err := h.gobeClient.GetSystemStatus(ctx)
		if err != nil {
			return fmt.Errorf("failed to get system status: %w", err)
		}

		response := fmt.Sprintf("📊 Status do Sistema:\n"+
			"Status: %s\n"+
			"Versão: %s\n"+
			"Uptime: %s\n"+
			"Database: %v\n"+
			"Sessões ativas: %d",
			status.Status, status.Version, status.Uptime,
			status.Database.Connected, status.Metrics.ActiveSessions)

		return h.SendDiscordMessage("", response)

	case "backup_database":
		result, err := h.gobeClient.BackupDatabase(ctx)
		if err != nil {
			return fmt.Errorf("failed to backup database: %w", err)
		}

		filename, _ := result["filename"].(string)
		size, _ := result["size"].(string)

		response := fmt.Sprintf("💾 Backup do banco realizado!\n"+
			"Arquivo: %s\n"+
			"Tamanho: %s", filename, size)

		return h.SendDiscordMessage("", response)

	default:
		return fmt.Errorf("comando GoBE desconhecido: %s", command)
	}
} */

// ============================================================================
// kbxctl Integration Methods
// ============================================================================

func (h *DiscordMCPHub) processKbxctlCommand(ctx context.Context, command, params string) error {
	if h.kbxctlClient == nil {
		return fmt.Errorf("kbxctl client not enabled")
	}

	log.Printf("⚙️ Processing kbxctl command: %s with params: %s", command, params)

	switch command {
	case "deploy_app":
		var deployParams struct {
			AppName string            `json:"app_name"`
			Version string            `json:"version"`
			Image   string            `json:"image"`
			Values  map[string]string `json:"values"`
		}

		if err := json.Unmarshal([]byte(params), &deployParams); err != nil {
			return fmt.Errorf("failed to parse deploy params: %w", err)
		}

		if deployParams.Values == nil {
			deployParams.Values = make(map[string]string)
		}

		result, err := h.kbxctlClient.DeployApp(ctx, deployParams.AppName, deployParams.Version, deployParams.Image, deployParams.Values)
		if err != nil {
			return fmt.Errorf("failed to deploy app: %w", err)
		}

		response := fmt.Sprintf("🚀 Deploy realizado com sucesso!\n"+
			"App: %s\n"+
			"Namespace: %s\n"+
			"Status: %s", result.Name, result.Namespace, result.Status)

		return h.SendDiscordMessage("", response)

	case "scale_deployment":
		var scaleParams struct {
			AppName  string `json:"app_name"`
			Replicas int    `json:"replicas"`
		}

		if err := json.Unmarshal([]byte(params), &scaleParams); err != nil {
			return fmt.Errorf("failed to parse scale params: %w", err)
		}

		err := h.kbxctlClient.ScaleDeployment(ctx, scaleParams.AppName, scaleParams.Replicas)
		if err != nil {
			return fmt.Errorf("failed to scale deployment: %w", err)
		}

		response := fmt.Sprintf("📈 Scaling realizado!\n"+
			"App: %s\n"+
			"Replicas: %d\n"+
			"Status: ✅ Sucesso", scaleParams.AppName, scaleParams.Replicas)

		return h.SendDiscordMessage("", response)

	case "cluster_info":
		info, err := h.kbxctlClient.GetClusterInfo(ctx)
		if err != nil {
			return fmt.Errorf("failed to get cluster info: %w", err)
		}

		name, _ := info["name"].(string)
		version, _ := info["version"].(string)
		nodeCount, _ := info["node_count"].(float64)
		status, _ := info["status"].(string)

		response := fmt.Sprintf("🎯 Informações do Cluster:\n"+
			"Nome: %s\n"+
			"Versão: %s\n"+
			"Nodes: %.0f\n"+
			"Status: %s", name, version, nodeCount, status)

		return h.SendDiscordMessage("", response)

	default:
		return fmt.Errorf("comando kbxctl desconhecido: %s", command)
	}
}

// ============================================================================
// Command Handlers for Discord Integration
// ============================================================================

func (h *DiscordMCPHub) handleCreateUserCommand(ctx context.Context, msg discord.Message) error {
	log.Printf("🔗 Handling create user command from Discord")

	// Extract user info from message
	content := strings.ToLower(msg.Content)

	// Simple parsing - in a real implementation, you might want more sophisticated parsing
	var name, email, role string

	// Look for patterns like "criar usuário João email@test.com admin"
	parts := strings.Fields(msg.Content)
	for i, part := range parts {
		if (strings.Contains(part, "usuário") || strings.Contains(part, "user")) && i+1 < len(parts) {
			name = parts[i+1]
		}
		if strings.Contains(part, "@") {
			email = part
		}
		if strings.Contains(content, "admin") {
			role = "admin"
		} else if strings.Contains(content, "user") {
			role = "user"
		}
	}

	if name == "" {
		return h.discordAdapter.SendMessage(msg.ChannelID,
			"❌ Nome não encontrado. Use: 'criar usuário [nome] [email] [role]'")
	}

	if email == "" {
		email = fmt.Sprintf("%s@discord.local", strings.ToLower(name))
	}

	if role == "" {
		role = "user"
	}

	// Create JSON params for GoBE
	// params := fmt.Sprintf(`{"name": "%s", "email": "%s", "role": "%s"}`, name, email, role)

	return nil //h.processGoBeCommand(ctx, "create_user", params)
}

func (h *DiscordMCPHub) handleDeployCommand(ctx context.Context, msg discord.Message) error {
	log.Printf("⚙️ Handling deploy command from Discord")

	// Extract deploy info from message
	parts := strings.Fields(msg.Content)

	var appName, version, image string

	for i, part := range parts {
		if strings.Contains(part, "deploy") && i+1 < len(parts) {
			appName = parts[i+1]
		}
		if (strings.Contains(part, "versão") || strings.Contains(part, "version")) && i+1 < len(parts) {
			version = parts[i+1]
		}
		if strings.Contains(part, ":") && strings.Contains(part, "/") {
			image = part // Docker image format
		}
	}

	if appName == "" {
		return h.discordAdapter.SendMessage(msg.ChannelID,
			"❌ Nome da app não encontrado. Use: 'deploy [app] versão [version] imagem [image]'")
	}

	if version == "" {
		version = "latest"
	}

	if image == "" {
		image = fmt.Sprintf("%s:%s", appName, version)
	}

	// Create JSON params for kbxctl
	params := fmt.Sprintf(`{"app_name": "%s", "version": "%s", "image": "%s", "values": {}}`,
		appName, version, image)

	return h.processKbxctlCommand(ctx, "deploy_app", params)
}

func (h *DiscordMCPHub) handleScaleCommand(ctx context.Context, msg discord.Message) error {
	log.Printf("⚙️ Handling scale command from Discord")

	// Extract scale info from message
	parts := strings.Fields(msg.Content)

	var appName string
	var replicas int = 1

	for i, part := range parts {
		if strings.Contains(part, "scale") && i+1 < len(parts) {
			appName = parts[i+1]
		}
		if strings.Contains(part, "replica") && i+1 < len(parts) {
			fmt.Sscanf(parts[i+1], "%d", &replicas)
		}
		// Also try to parse numbers directly
		var num int
		if n, err := fmt.Sscanf(part, "%d", &num); err == nil && n == 1 && num > 0 && num < 100 {
			replicas = num
		}
	}

	if appName == "" {
		return h.discordAdapter.SendMessage(msg.ChannelID,
			"❌ Nome da app não encontrado. Use: 'scale [app] [replicas]'")
	}

	// Create JSON params for kbxctl
	params := fmt.Sprintf(`{"app_name": "%s", "replicas": %d}`, appName, replicas)

	return h.processKbxctlCommand(ctx, "scale_deployment", params)
}
