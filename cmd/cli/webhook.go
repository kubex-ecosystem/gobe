package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	"github.com/spf13/cobra"
)

var (
	webhookURL      string
	webhookPort     string
	webhookEventID  string
	webhookOutput   string
	webhookFormat   string
	webhookPage     int
	webhookLimit    int
)

func WebhookCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Webhook management and monitoring commands",
		Long: `Manage webhook operations including listing events,
checking health, and retrying failed events.`,
	}

	cmd.AddCommand(webhookListCmd())
	cmd.AddCommand(webhookHealthCmd())
	cmd.AddCommand(webhookEventCmd())
	cmd.AddCommand(webhookRetryCmd())

	return cmd
}

func webhookListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhook events",
		Long:  `List webhook events with pagination support.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("info", "Listing webhook events...")

			baseURL := getWebhookBaseURL()
			url := fmt.Sprintf("%s/v1/webhooks/events?page=%d&limit=%d", baseURL, webhookPage, webhookLimit)

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get(url)
			if err != nil {
				return fmt.Errorf("failed to fetch webhook events: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("webhook API returned status %d", resp.StatusCode)
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}

			output, err := formatOutput(result, webhookFormat)
			if err != nil {
				return err
			}

			if webhookOutput != "" {
				return os.WriteFile(webhookOutput, []byte(output), 0644)
			}

			fmt.Println(output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&webhookPort, "port", "p", "8080", "Server port")
	cmd.Flags().StringVarP(&webhookFormat, "format", "f", "json", "Output format (json, table)")
	cmd.Flags().IntVar(&webhookPage, "page", 1, "Page number for pagination")
	cmd.Flags().IntVar(&webhookLimit, "limit", 10, "Number of events per page")
	cmd.Flags().StringVarP(&webhookOutput, "output", "o", "", "Output file (default: stdout)")

	return cmd
}

func webhookHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check webhook system health",
		Long:  `Check the health status of the webhook system including statistics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("info", "Checking webhook system health...")

			baseURL := getWebhookBaseURL()
			url := fmt.Sprintf("%s/v1/webhooks/health", baseURL)

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(url)
			if err != nil {
				return fmt.Errorf("failed to check webhook health: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("webhook health check returned status %d", resp.StatusCode)
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}

			output, err := formatOutput(result, webhookFormat)
			if err != nil {
				return err
			}

			if webhookOutput != "" {
				return os.WriteFile(webhookOutput, []byte(output), 0644)
			}

			fmt.Println(output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&webhookPort, "port", "p", "8080", "Server port")
	cmd.Flags().StringVarP(&webhookFormat, "format", "f", "json", "Output format (json, table)")
	cmd.Flags().StringVarP(&webhookOutput, "output", "o", "", "Output file (default: stdout)")

	return cmd
}

func webhookEventCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event [event-id]",
		Short: "Get details of a specific webhook event",
		Long:  `Get detailed information about a specific webhook event by ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := args[0]
			gl.Log("info", fmt.Sprintf("Getting webhook event details: %s", eventID))

			baseURL := getWebhookBaseURL()
			url := fmt.Sprintf("%s/v1/webhooks/events/%s", baseURL, eventID)

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(url)
			if err != nil {
				return fmt.Errorf("failed to fetch webhook event: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				return fmt.Errorf("webhook event not found: %s", eventID)
			}

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("webhook API returned status %d", resp.StatusCode)
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}

			output, err := formatOutput(result, webhookFormat)
			if err != nil {
				return err
			}

			if webhookOutput != "" {
				return os.WriteFile(webhookOutput, []byte(output), 0644)
			}

			fmt.Println(output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&webhookPort, "port", "p", "8080", "Server port")
	cmd.Flags().StringVarP(&webhookFormat, "format", "f", "json", "Output format (json, table)")
	cmd.Flags().StringVarP(&webhookOutput, "output", "o", "", "Output file (default: stdout)")

	return cmd
}

func webhookRetryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retry [event-id]",
		Short: "Retry a failed webhook event",
		Long:  `Retry processing of a failed webhook event by ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := args[0]
			gl.Log("info", fmt.Sprintf("Retrying webhook event: %s", eventID))

			baseURL := getWebhookBaseURL()
			url := fmt.Sprintf("%s/v1/webhooks/retry", baseURL)

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Post(url, "application/json", nil)
			if err != nil {
				return fmt.Errorf("failed to retry webhook event: %w", err)
			}
			defer resp.Body.Close()

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}

			if resp.StatusCode != http.StatusOK {
				if errMsg, ok := result["error"].(string); ok {
					return fmt.Errorf("webhook retry failed: %s", errMsg)
				}
				return fmt.Errorf("webhook retry failed with status %d", resp.StatusCode)
			}

			output, err := formatOutput(result, webhookFormat)
			if err != nil {
				return err
			}

			if webhookOutput != "" {
				return os.WriteFile(webhookOutput, []byte(output), 0644)
			}

			fmt.Println(output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&webhookPort, "port", "p", "8080", "Server port")
	cmd.Flags().StringVarP(&webhookFormat, "format", "f", "json", "Output format (json, table)")
	cmd.Flags().StringVarP(&webhookOutput, "output", "o", "", "Output file (default: stdout)")

	return cmd
}

// Helper functions
func getWebhookBaseURL() string {
	if webhookURL != "" {
		return webhookURL
	}
	return fmt.Sprintf("http://localhost:%s", webhookPort)
}

func formatOutput(data interface{}, format string) (string, error) {
	switch format {
	case "json":
		output, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to format as JSON: %w", err)
		}
		return string(output), nil
	case "table":
		// Simple table format for webhook data
		if dataMap, ok := data.(map[string]interface{}); ok {
			output := "Key\t\tValue\n"
			output += "---\t\t-----\n"
			for key, value := range dataMap {
				output += fmt.Sprintf("%s\t\t%v\n", key, value)
			}
			return output, nil
		}
		return fmt.Sprintf("%+v", data), nil
	default:
		return fmt.Sprintf("%+v", data), nil
	}
}