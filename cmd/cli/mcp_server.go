package cli

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	f "github.com/kubex-ecosystem/gobe/factory"
	l "github.com/kubex-ecosystem/logz"
	"github.com/spf13/cobra"
)

var (
	mcpServerCmd = &cobra.Command{
		Use:   "mcp-server",
		Short: "Start the MCP server",
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
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Interact with a Large Language Model (LLM) in a chat format",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Implement chat logic here
			return nil
		},
	}
	return cmd
}

func generateTextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate text using a specified LLM",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Implement text generation logic here
			return nil
		},
	}
	return cmd
}

func analyzeTextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze text using a specified LLM",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Implement text analysis logic here
			return nil
		},
	}
	return cmd
}

func summarizeTextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summarize",
		Short: "Summarize text using a specified LLM",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Implement text summarization logic here
			return nil
		},
	}
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
