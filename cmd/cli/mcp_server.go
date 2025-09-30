package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	f "github.com/kubex-ecosystem/gobe/factory"
	"github.com/kubex-ecosystem/gobe/internal/config"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	"github.com/kubex-ecosystem/gobe/internal/services/llm"
	l "github.com/kubex-ecosystem/logz"
	"github.com/spf13/cobra"
)

var (
	shortDesc    string = "Start the MCP server"
	longDesc     string = "Start the MCP server with GoBE"
	mcpServerCmd        = &cobra.Command{
		Use:         "mcp-server",
		Short:       shortDesc,
		Long:        longDesc,
		Aliases:     []string{"mcp", "mcpserver", "mcp_srv", "mcp-srv"},
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			startMCPServer()
		},
	}

	mcpServerPort           string
	mcpServerBind           string
	mcpServerLogFile        string
	mcpServerConfigFile     string
	mcpServerIsConfidential bool
	mcpServerDebug          bool
	mcpServerReleaseMode    bool

	// LLM command flags
	llmProvider    string
	llmModel       string
	llmMaxTokens   int
	llmTemperature float64
	llmInteractive bool
	llmOutput      string
	llmInput       string
)

func init() {
	mcpServerCmd.Flags().StringVarP(&mcpServerPort, "port", "p", "8080", "Port for the MCP server")
	mcpServerCmd.Flags().StringVarP(&mcpServerBind, "bind", "b", "0.0.0.0", "Bind address for the MCP server")
	mcpServerCmd.Flags().StringVarP(&mcpServerLogFile, "log-file", "l", "mcp_server.log", "Log file for the MCP server")
	mcpServerCmd.Flags().StringVarP(&mcpServerConfigFile, "config-file", "c", "mcp_server.yaml", "Config file for the MCP server")
	mcpServerCmd.Flags().BoolVarP(&mcpServerIsConfidential, "confidential", "C", false, "Enable confidential mode for the MCP server")
	mcpServerCmd.Flags().BoolVarP(&mcpServerDebug, "debug", "d", false, "Enable debug mode for the MCP server")
	mcpServerCmd.Flags().BoolVarP(&mcpServerReleaseMode, "release", "r", false, "Enable release mode for the MCP server")
}

func MCPServerCmd() *cobra.Command {
	mcpServerCmd.AddCommand(llmCmd())
	mcpServerCmd.AddCommand(chatCmd())
	mcpServerCmd.AddCommand(generateTextCmd())
	mcpServerCmd.AddCommand(analyzeTextCmd())
	mcpServerCmd.AddCommand(summarizeTextCmd())

	return mcpServerCmd
}

func startMCPServer() {
	logger := l.NewLogger("gobe_mcp_server")
	goBe, err := f.NewGoBE("MCPServer", mcpServerPort, mcpServerBind, mcpServerLogFile, mcpServerConfigFile, mcpServerIsConfidential, logger, mcpServerDebug, mcpServerReleaseMode)
	if err != nil {
		fmt.Printf("Error initializing GoBE: %s\n", err)
		os.Exit(1)
	}

	// dbService, err := f.GetDatabaseService(goBe)
	// if err != nil {
	// 	fmt.Printf("Error getting database service: %s\n", err)
	// 	os.Exit(1)
	// }

	// if dbService != nil {
	// 	err = dbService.Connect()
	// 	if err != nil {
	// 		fmt.Printf("Error connecting to database: %s\n", err)
	// 		os.Exit(1)
	// 	}
	// 	defer dbService.Disconnect()
	// }

	// err = goBe.Start()
	// if err != nil {
	// 	fmt.Printf("Error starting GoBE: %s\n", err)
	// 	os.Exit(1)
	// }

	if err := goBe.InitializeResources(); err != nil {
		fmt.Printf("Error initializing resources: %s\n", err)
		os.Exit(1)
	}

	// Start consuming messages from RabbitMQ in a separate goroutine
	go f.ConsumeMessages("mcp_queue")

	// Keep the main goroutine running
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println("MCP Server is running...")
	select {}
}
func llmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llm",
		Short: "Commands related to Large Language Models (LLMs)",
	}
	// Add subcommands related to LLMs here
	// like chat, generate, analyze, summarize, etc.
	cmd.AddCommand(chatCmd())
	cmd.AddCommand(generateTextCmd())
	cmd.AddCommand(analyzeTextCmd())
	cmd.AddCommand(summarizeTextCmd())
	return cmd
}

func chatCmd() *cobra.Command {
	shortDesc := "Interact with a Large Language Model (LLM) in a chat format"
	longDesc := `Start an interactive chat session with an LLM.
You can chat with OpenAI GPT, Google Gemini, or Groq models.`

	cmd := &cobra.Command{
		Use:         "chat",
		Short:       shortDesc,
		Long:        longDesc,
		Aliases:     []string{"conversation", "talk", "dialogue"},
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("info", "Starting LLM chat session...")

			client, err := createLLMClient()
			if err != nil {
				return err
			}

			if llmInteractive {
				return runInteractiveChat(client)
			}

			// Single message mode
			input, err := readInput()
			if err != nil {
				return err
			}

			response, err := generateLLMResponse(client, input, "chat")
			if err != nil {
				return err
			}

			return writeOutput(response)
		},
	}

	// Add flags specific to chat
	cmd.Flags().StringVarP(&llmProvider, "provider", "p", "gemini", "LLM provider (openai, gemini, groq)")
	cmd.Flags().StringVarP(&llmModel, "model", "m", "", "Model name (optional, uses provider default)")
	cmd.Flags().IntVar(&llmMaxTokens, "max-tokens", 2048, "Maximum tokens to generate")
	cmd.Flags().Float64Var(&llmTemperature, "temperature", 0.7, "Temperature for response generation")
	cmd.Flags().BoolVarP(&llmInteractive, "interactive", "i", false, "Start interactive chat session")
	cmd.Flags().StringVarP(&llmInput, "input", "f", "", "Input file (default: stdin)")
	cmd.Flags().StringVarP(&llmOutput, "output", "o", "", "Output file (default: stdout)")

	return cmd
}

func generateTextCmd() *cobra.Command {
	shortDesc := "Generate text using a specified LLM"
	longDesc := `Generate text based on a prompt using an LLM.
Supports various generation tasks like creative writing, code generation, etc.`

	cmd := &cobra.Command{
		Use:         "generate",
		Short:       shortDesc,
		Long:        longDesc,
		Aliases:     []string{"gen", "textgen", "create"},
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("info", "Starting text generation...")

			client, err := createLLMClient()
			if err != nil {
				return err
			}

			input, err := readInput()
			if err != nil {
				return err
			}

			if input == "" {
				return fmt.Errorf("no input provided for text generation")
			}

			response, err := generateLLMResponse(client, input, "generate")
			if err != nil {
				return err
			}

			return writeOutput(response)
		},
	}

	// Add flags specific to generation
	cmd.Flags().StringVarP(&llmProvider, "provider", "p", "gemini", "LLM provider (openai, gemini, groq)")
	cmd.Flags().StringVarP(&llmModel, "model", "m", "", "Model name (optional, uses provider default)")
	cmd.Flags().IntVar(&llmMaxTokens, "max-tokens", 2048, "Maximum tokens to generate")
	cmd.Flags().Float64Var(&llmTemperature, "temperature", 0.8, "Temperature for response generation (higher = more creative)")
	cmd.Flags().StringVarP(&llmInput, "input", "f", "", "Input file with prompt (default: stdin)")
	cmd.Flags().StringVarP(&llmOutput, "output", "o", "", "Output file (default: stdout)")

	return cmd
}

func analyzeTextCmd() *cobra.Command {
	shortDesc := "Analyze text using a specified LLM"
	longDesc := `Analyze text content using an LLM.
Provides insights, sentiment analysis, classification, and other analytical tasks.`
	cmd := &cobra.Command{
		Use:         "analyze",
		Short:       shortDesc,
		Long:        longDesc,
		Aliases:     []string{"analysis", "text-analysis", "nlp-analyze", "nlp"},
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("info", "Starting text analysis...")

			client, err := createLLMClient()
			if err != nil {
				return err
			}

			input, err := readInput()
			if err != nil {
				return err
			}

			if input == "" {
				return fmt.Errorf("no input provided for text analysis")
			}

			// Use the existing AnalyzeMessage function for proper analysis
			req := llm.AnalysisRequest{
				Platform: "cli",
				Content:  input,
				UserID:   "cli-user",
				Context:  map[string]interface{}{"task": "analyze"},
			}

			ctx := context.Background()
			result, err := client.AnalyzeMessage(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to analyze text: %w", err)
			}

			// Format analysis result as JSON
			output, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format analysis result: %w", err)
			}

			return writeOutput(string(output))
		},
	}

	// Add flags specific to analysis
	cmd.Flags().StringVarP(&llmProvider, "provider", "p", "gemini", "LLM provider (openai, gemini, groq)")
	cmd.Flags().StringVarP(&llmModel, "model", "m", "", "Model name (optional, uses provider default)")
	cmd.Flags().IntVar(&llmMaxTokens, "max-tokens", 1024, "Maximum tokens for analysis")
	cmd.Flags().Float64Var(&llmTemperature, "temperature", 0.3, "Temperature for analysis (lower = more focused)")
	cmd.Flags().StringVarP(&llmInput, "input", "f", "", "Input file with text to analyze (default: stdin)")
	cmd.Flags().StringVarP(&llmOutput, "output", "o", "", "Output file for analysis result (default: stdout)")

	return cmd
}

func summarizeTextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summarize",
		Short: "Summarize text using a specified LLM",
		Long: `Summarize long text content using an LLM.
Provides concise summaries, key points extraction, and content condensation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("info", "Starting text summarization...")

			client, err := createLLMClient()
			if err != nil {
				return err
			}

			input, err := readInput()
			if err != nil {
				return err
			}

			if input == "" {
				return fmt.Errorf("no input provided for text summarization")
			}

			response, err := generateLLMResponse(client, input, "summarize")
			if err != nil {
				return err
			}

			return writeOutput(response)
		},
	}

	// Add flags specific to summarization
	cmd.Flags().StringVarP(&llmProvider, "provider", "p", "gemini", "LLM provider (openai, gemini, groq)")
	cmd.Flags().StringVarP(&llmModel, "model", "m", "", "Model name (optional, uses provider default)")
	cmd.Flags().IntVar(&llmMaxTokens, "max-tokens", 1024, "Maximum tokens for summary")
	cmd.Flags().Float64Var(&llmTemperature, "temperature", 0.5, "Temperature for summarization (balanced)")
	cmd.Flags().StringVarP(&llmInput, "input", "f", "", "Input file with text to summarize (default: stdin)")
	cmd.Flags().StringVarP(&llmOutput, "output", "o", "", "Output file for summary (default: stdout)")

	return cmd
}

func retry(attempts int, sleep time.Duration, fn func() error) error {
	for i := 0; i < attempts; i++ {
		err := fn()
		if err != nil {
			if i == attempts-1 {
				return err
			}
			time.Sleep(sleep)
		}
	}
	return nil
}

func forever() {
	select {}
}

// Helper functions for LLM commands
func createLLMClient() (*llm.Client, error) {
	cfg := config.LLMConfig{
		Provider:    llmProvider,
		Model:       llmModel,
		MaxTokens:   llmMaxTokens,
		Temperature: llmTemperature,
	}

	client, err := llm.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	return client, nil
}

func readInput() (string, error) {
	if llmInput != "" {
		// Read from file
		content, err := os.ReadFile(llmInput)
		if err != nil {
			return "", fmt.Errorf("failed to read input file: %w", err)
		}
		return string(content), nil
	}

	// Read from stdin
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string

	fmt.Print("Enter your text (Ctrl+D to finish):\n> ")
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read from stdin: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}

func writeOutput(content string) error {
	if llmOutput != "" {
		return os.WriteFile(llmOutput, []byte(content), 0644)
	}

	fmt.Println(content)
	return nil
}

func runInteractiveChat(client *llm.Client) error {
	fmt.Println("=== Interactive LLM Chat ===")
	fmt.Println("Type 'exit' or 'quit' to end the session")
	fmt.Println("Type 'clear' to clear the conversation history")
	fmt.Println("=====================================")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		if input == "clear" {
			fmt.Println("Conversation history cleared.")
			continue
		}

		response, err := generateLLMResponse(client, input, "chat")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\nAssistant: %s\n", response)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	return nil
}

func generateLLMResponse(client *llm.Client, input, taskType string) (string, error) {
	// Create appropriate prompt based on task type
	var prompt string
	switch taskType {
	case "chat":
		prompt = input
	case "generate":
		prompt = fmt.Sprintf("Generate content based on the following prompt:\n\n%s", input)
	case "summarize":
		prompt = fmt.Sprintf("Please provide a concise summary of the following text:\n\n%s", input)
	default:
		prompt = input
	}

	// Use AnalysisRequest structure to leverage existing LLM client
	req := llm.AnalysisRequest{
		Platform: "cli",
		Content:  prompt,
		UserID:   "cli-user",
		Context: map[string]interface{}{
			"task": taskType,
			"cli":  true,
		},
	}

	ctx := context.Background()
	result, err := client.AnalyzeMessage(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get LLM response: %w", err)
	}

	// For generation and chat tasks, return the suggested response
	// For analysis tasks, this is handled separately in analyzeTextCmd
	if result.SuggestedResponse != "" {
		return result.SuggestedResponse, nil
	}

	// Fallback: return a basic response
	return fmt.Sprintf("Task completed. Confidence: %.2f", result.Confidence), nil
}
