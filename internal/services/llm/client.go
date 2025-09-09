// Package llm provides a client for interacting with LLM APIs (OpenAI, Gemini) to analyze Discord messages.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kubex-ecosystem/gobe/internal/config"

	"github.com/patrickmn/go-cache"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/genai"
)

type Client struct {
	openai     *openai.Client
	config     config.LLMConfig
	cache      *cache.Cache
	devMode    bool
	provider   string // "openai", "gemini", "dev"
	httpClient *http.Client
}

type AnalysisRequest struct {
	Platform string                 `json:"platform"`
	Content  string                 `json:"content"`
	UserID   string                 `json:"user_id"`
	Context  map[string]interface{} `json:"context"`
}

type AnalysisResponse struct {
	ShouldRespond     bool     `json:"should_respond"`
	SuggestedResponse string   `json:"suggested_response"`
	Confidence        float64  `json:"confidence"`
	ShouldCreateTask  bool     `json:"should_create_task"`
	TaskTitle         string   `json:"task_title"`
	TaskDescription   string   `json:"task_description"`
	TaskPriority      string   `json:"task_priority"`
	TaskTags          []string `json:"task_tags"`
	RequiresApproval  bool     `json:"requires_approval"`
	Sentiment         string   `json:"sentiment"`
	Category          string   `json:"category"`
}

func NewClient(config config.LLMConfig) (*Client, error) {
	devMode := config.APIKey == "dev_api_key" || config.APIKey == ""

	log.Printf("🔍 LLM Config Debug:")
	if len(config.APIKey) > 10 {
		log.Printf("   APIKey: %s...", config.APIKey[:10])
	} else {
		log.Printf("   APIKey: '%s' (len=%d)", config.APIKey, len(config.APIKey))
	}
	log.Printf("   Provider: %s", config.Provider)
	log.Printf("   DevMode: %v", devMode)

	// Determine provider
	provider := "dev"
	if !devMode {
		switch config.Provider {
		case "openai":
			provider = "openai"
		case "gemini":
			provider = "gemini"
		default:
			// Auto-detect based on API key format
			if strings.HasPrefix(config.APIKey, "sk-") {
				provider = "openai"
			} else if strings.HasPrefix(config.APIKey, "AI") {
				provider = "gemini"
			} else {
				return nil, fmt.Errorf("unknown LLM provider: %s (API key format not recognized)", config.Provider)
			}
		}
	}

	log.Printf("   Final Provider: %s", provider)

	var openaiClient *openai.Client
	httpClient := &http.Client{Timeout: 30 * time.Second}

	if provider == "openai" {
		openaiClient = openai.NewClient(config.APIKey)
	}

	// Cache for 5 minutes
	cache := cache.New(5*time.Minute, 10*time.Minute)

	return &Client{
		openai:     openaiClient,
		config:     config,
		cache:      cache,
		devMode:    devMode,
		provider:   provider,
		httpClient: httpClient,
	}, nil
}

func (c *Client) AnalyzeMessage(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s_%s_%s_%s", c.provider, req.Platform, req.UserID, req.Content)
	if cached, found := c.cache.Get(cacheKey); found {
		return cached.(*AnalysisResponse), nil
	}

	var response *AnalysisResponse
	var err error

	switch c.provider {
	case "dev":
		response = c.mockAnalysis(req)
	case "openai":
		response, err = c.analyzeWithOpenAI(ctx, req)
	case "gemini":
		response, err = c.analyzeWithGemini(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", c.provider)
	}

	if err != nil {
		return nil, err
	}

	// Cache the result
	c.cache.Set(cacheKey, response, cache.DefaultExpiration)
	return response, nil
}

func (c *Client) mockAnalysis(req AnalysisRequest) *AnalysisResponse {
	content := strings.ToLower(req.Content)

	// Enhanced mock based on message type from context
	msgType, _ := req.Context["type"].(string)

	response := &AnalysisResponse{
		ShouldRespond:     true,
		SuggestedResponse: c.generateMockResponse(req.Content, msgType),
		Confidence:        0.85,
		ShouldCreateTask:  strings.Contains(content, "task") || strings.Contains(content, "tarefa") || strings.Contains(content, "criar") || strings.Contains(content, "lembrar"),
		TaskTitle:         "Mock Task",
		TaskDescription:   "Mock task generated from message",
		TaskPriority:      "medium",
		TaskTags:          []string{"mock", "dev"},
		RequiresApproval:  false, // Dev mode doesn't require approval
		Sentiment:         "neutral",
		Category:          "general",
	}

	// Adjust based on message type
	switch msgType {
	case "question":
		response.Category = "question"
		response.SuggestedResponse = c.generateQuestionResponse(req.Content)
	case "task_request":
		response.Category = "request"
		response.ShouldCreateTask = true
		response.TaskTitle = c.extractTaskTitle(req.Content)
		response.SuggestedResponse = c.generateTaskResponse(req.Content)
	case "analysis":
		response.Category = "analysis"
		response.SuggestedResponse = c.generateAnalysisResponse(req.Content)
	case "casual":
		response.Category = "casual"
		response.SuggestedResponse = c.generateCasualResponse(req.Content)
	}

	return response
}

func (c *Client) generateMockResponse(content, msgType string) string {
	switch msgType {
	case "question":
		return c.generateQuestionResponse(content)
	case "task_request":
		return c.generateTaskResponse(content)
	case "analysis":
		return c.generateAnalysisResponse(content)
	case "casual":
		return c.generateCasualResponse(content)
	default:
		return fmt.Sprintf("📝 Entendi sua mensagem: \"%s\"\n\n🤖 Estou processando e posso ajudar com mais informações se precisar!", content)
	}
}

func (c *Client) generateQuestionResponse(content string) string {
	return fmt.Sprintf("🤔 **Sua pergunta:** %s\n\n💡 **Resposta:** Esta é uma resposta inteligente gerada pelo sistema. Baseado no contexto da sua pergunta, posso fornecer informações relevantes e sugestões práticas.\n\n❓ Precisa de mais detalhes sobre algum aspecto específico?", content)
}

func (c *Client) generateTaskResponse(content string) string {
	title := c.extractTaskTitle(content)
	return fmt.Sprintf("📋 **Tarefa criada com sucesso!**\n\n📌 **Título:** %s\n📝 **Descrição:** %s\n⏰ **Criada em:** %s\n🏷️ **Tags:** task, discord, auto\n\n✅ A tarefa foi salva no sistema e você receberá notificações sobre o progresso!", title, content, time.Now().Format("02/01/2006 15:04"))
}

func (c *Client) generateAnalysisResponse(content string) string {
	return fmt.Sprintf("🔍 **Análise completa do texto:**\n\n📝 **Conteúdo analisado:** %s\n\n📊 **Métricas:**\n• Comprimento: %d caracteres\n• Sentimento: Neutro\n• Complexidade: Média\n• Relevância: Alta\n\n💡 **Insights:** O texto apresenta características interessantes e pode beneficiar de análise mais aprofundada dependendo do contexto de uso.", content, len(content))
}

func (c *Client) generateCasualResponse(content string) string {
	responses := []string{
		"😊 Oi! Legal falar com você! Como posso ajudar?",
		"👋 Olá! Tudo bem? Estou aqui se precisar de alguma coisa!",
		"🤖 Oi! Sou o assistente inteligente. Em que posso ser útil?",
		"😄 Hey! Obrigado por conversar comigo! Posso ajudar com algo?",
		"👍 Entendi! Estou aqui para ajudar sempre que precisar!",
	}
	// Escolher resposta pseudo-aleatória baseada no comprimento da mensagem
	return responses[len(content)%len(responses)]
}

func (c *Client) extractTaskTitle(content string) string {
	// Remove palavras comuns de início
	title := strings.TrimSpace(content)
	title = strings.TrimPrefix(title, "criar ")
	title = strings.TrimPrefix(title, "preciso ")
	title = strings.TrimPrefix(title, "quero ")
	title = strings.TrimPrefix(title, "adicionar ")
	title = strings.TrimPrefix(title, "task ")
	title = strings.TrimPrefix(title, "tarefa ")

	// Limitar tamanho
	if len(title) > 50 {
		title = title[:50] + "..."
	}

	if title == "" {
		title = "Nova tarefa"
	}

	return title
}

func (c *Client) analyzeWithOpenAI(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	prompt := c.buildAnalysisPrompt(req)

	resp, err := c.openai.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: c.getSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: float32(c.config.Temperature),
		MaxTokens:   c.config.MaxTokens,
	})

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	return c.parseAnalysisResponse(resp.Choices[0].Message.Content), nil
}

// Gemini API structures
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

func (c *Client) analyzeWithGemini(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	prompt := c.buildAnalysisPrompt(req)
	systemPrompt := c.getSystemPrompt()

	// Combine system prompt and user prompt for Gemini
	fullPrompt := fmt.Sprintf("%s\n\n%s", systemPrompt, prompt)

	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		//"gemini-1.5-flash",
		genai.Text(fullPrompt),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content with Gemini: %w", err)
	}

	return c.parseAnalysisResponse(result.Text()), nil
}

func (c *Client) getSystemPrompt() string {
	return `Você é um assistente inteligente que analisa mensagens do Discord para determinar ações apropriadas.

Suas responsabilidades:
1. Analisar o conteúdo da mensagem
2. Determinar se uma resposta é necessária
3. Sugerir uma resposta apropriada se necessário
4. Identificar se uma tarefa deve ser criada baseada na mensagem
5. Avaliar o sentimento e categoria da mensagem

Responda sempre em JSON com a seguinte estrutura:
{
  "should_respond": boolean,
  "suggested_response": "string",
  "confidence": 0.0-1.0,
  "should_create_task": boolean,
  "task_title": "string",
  "task_description": "string", 
  "task_priority": "low|medium|high|urgent",
  "task_tags": ["tag1", "tag2"],
  "requires_approval": boolean,
  "sentiment": "positive|negative|neutral",
  "category": "question|request|complaint|information|other"
}`
}

func (c *Client) buildAnalysisPrompt(req AnalysisRequest) string {
	return fmt.Sprintf(`
Analise esta mensagem do Discord:

Plataforma: %s
Usuário: %s
Conteúdo: "%s"
Contexto: %v

Determine:
1. Se devemos responder automaticamente
2. Qual seria uma resposta apropriada
3. Se devemos criar uma tarefa baseada nesta mensagem
4. Nível de confiança na análise
5. Se requer aprovação humana antes de agir

Responda apenas com JSON válido.
`, req.Platform, req.UserID, req.Content, req.Context)
}

type Pagination struct {
	PageToken   string `json:"page_token" gorm:"column:page_token" binding:"required"`
	PageSize    uint64 `json:"page_size" gorm:"column:page_size" binding:"required"`
	PageCount   int    `json:"page_count" gorm:"column:page_count" binding:"required"`
	CurrentPage int    `json:"current_page" gorm:"column:current_page" binding:"required"`
	TotalSize   uint64 `json:"total_size" gorm:"column:total_size" binding:"required"`
}

func (c *Client) parseAnalysisResponse(content string) *AnalysisResponse {
	// Try to parse as JSON first
	var jsonResp struct {
		ShouldRespond     bool        `json:"should_respond" binding:"required"`
		SuggestedResponse string      `json:"suggested_response" binding:"required"`
		Confidence        float64     `json:"confidence" binding:"required"`
		ShouldCreateTask  bool        `json:"should_create_task" binding:"required"`
		TaskTitle         string      `json:"task_title,omitempty" binding:"required"`
		TaskDescription   string      `json:"task_description,omitempty" binding:"required"`
		TaskPriority      string      `json:"task_priority,omitempty" binding:"required"`
		TaskTags          []string    `json:"task_tags,omitempty" binding:"required"`
		RequiresApproval  bool        `json:"requires_approval" binding:"required"`
		Sentiment         string      `json:"sentiment,omitempty" binding:"required"`
		Category          string      `json:"category" binding:"required"`
		Pagination        *Pagination `json:"pagination,omitempty" binding:"omitempty"`
	}

	// Extract JSON from markdown code blocks if present
	jsonContent := content
	if start := strings.Index(content, "```json"); start != -1 {
		start += 7 // len("```json")
		if end := strings.Index(content[start:], "```"); end != -1 {
			jsonContent = strings.TrimSpace(content[start : start+end])
		}
	} else if start := strings.Index(content, "{"); start != -1 {
		if end := strings.LastIndex(content, "}"); end != -1 && end > start {
			jsonContent = content[start : end+1]
		}
	}

	// Try to parse JSON
	if err := json.Unmarshal([]byte(jsonContent), &jsonResp); err == nil {
		return &AnalysisResponse{
			ShouldRespond:     jsonResp.ShouldRespond,
			SuggestedResponse: jsonResp.SuggestedResponse,
			Confidence:        jsonResp.Confidence,
			ShouldCreateTask:  jsonResp.ShouldCreateTask,
			TaskTitle:         jsonResp.TaskTitle,
			TaskDescription:   jsonResp.TaskDescription,
			TaskPriority:      jsonResp.TaskPriority,
			TaskTags:          jsonResp.TaskTags,
			RequiresApproval:  jsonResp.RequiresApproval,
			Sentiment:         jsonResp.Sentiment,
			Category:          jsonResp.Category,
		}
	}

	// Fallback to simple parsing if JSON parsing fails
	analysis := &AnalysisResponse{
		ShouldRespond:    true,  // Default to responding
		Confidence:       0.8,   // Default confidence
		RequiresApproval: false, // Default to not requiring approval
		Sentiment:        "neutral",
		Category:         "other",
	}

	// Simple text-based extraction as fallback
	lowerContent := strings.ToLower(content)

	// Extract suggested response
	if start := strings.Index(lowerContent, "suggested_response"); start != -1 {
		remaining := content[start:]
		if colonIdx := strings.Index(remaining, ":"); colonIdx != -1 {
			afterColon := remaining[colonIdx+1:]
			if quoteStart := strings.Index(afterColon, "\""); quoteStart != -1 {
				quoteStart += 1
				if quoteEnd := strings.Index(afterColon[quoteStart:], "\""); quoteEnd != -1 {
					analysis.SuggestedResponse = afterColon[quoteStart : quoteStart+quoteEnd]
				}
			}
		}
	}

	// If no suggested response found, use the content as response
	if analysis.SuggestedResponse == "" {
		analysis.SuggestedResponse = content
	}

	// Check if should create task
	if strings.Contains(lowerContent, "should_create_task") && strings.Contains(lowerContent, "true") {
		analysis.ShouldCreateTask = true
		analysis.TaskTitle = "Tarefa criada automaticamente"
		analysis.TaskDescription = "Baseada na análise da mensagem"
		analysis.TaskPriority = "medium"
		analysis.TaskTags = []string{"auto-created", "llm"}
	}

	return analysis
}
