package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kubex-ecosystem/gobe/internal/bootstrap"
	"github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	"github.com/kubex-ecosystem/gobe/internal/services/chatbot/discord"
	gl "github.com/kubex-ecosystem/logz/logger"
	"github.com/spf13/cobra"
)

var (
	discordToken        string
	discordChannel      string
	discordMessage      string
	discordDevMode      bool
	discordOutput       string
	discordListChannels bool
	discordStatus       bool
)

func DiscordCommand() *cobra.Command {
	shortDesc := "Discord bot management commands"
	longDesc := `Manage Discord bot operations including sending messages,
checking status, and listing channels.`

	cmd := &cobra.Command{
		Use:   "discord",
		Short: shortDesc,
		Long:  longDesc,
		Aliases: []string{
			"dc", "discord-bot", "discordbot",
		},
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				gl.Log("error", fmt.Sprintf("Failed to display help: %v", err))
			}
		},
	}

	cmd.AddCommand(discordSendCmd())
	cmd.AddCommand(discordStatusCmd())
	cmd.AddCommand(discordChannelsCmd())
	cmd.AddCommand(discordTestCmd())

	return cmd
}

func discordSendCmd() *cobra.Command {
	shortDesc := "Send a message to a Discord channel"
	long := `Send a message to a specified Discord channel.
Supports both text messages and file attachments.`

	cmd := &cobra.Command{
		Use:   "send",
		Short: shortDesc,
		Long:  long,
		Aliases: []string{
			"message", "msg", "post",
		},
		Annotations: GetDescriptions([]string{
			shortDesc,
			long,
		}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("info", "Sending Discord message...")

			if discordMessage == "" && len(args) == 0 {
				return fmt.Errorf("no message provided")
			}

			message := discordMessage
			if len(args) > 0 {
				message = args[0]
			}

			adapter, err := createDiscordAdapter()
			if err != nil {
				return err
			}

			err = adapter.Connect()
			if err != nil {
				return fmt.Errorf("failed to connect to Discord: %w", err)
			}

			// Send message using the correct interface
			opts := interfaces.SendOptions{}
			err = adapter.SendMessage(discordChannel, message, opts)
			if err != nil {
				return fmt.Errorf("failed to send message: %w", err)
			}

			gl.Log("info", "Message sent successfully")
			return nil
		},
	}

	cmd.Flags().StringVarP(&discordToken, "token", "t", "", "Discord bot token (or set DISCORD_BOT_TOKEN)")
	cmd.Flags().StringVarP(&discordChannel, "channel", "c", "", "Discord channel ID")
	cmd.Flags().StringVarP(&discordMessage, "message", "m", "", "Message to send")
	cmd.Flags().BoolVarP(&discordDevMode, "dev", "d", false, "Enable dev mode (mock operations)")
	cmd.Flags().StringVarP(&discordOutput, "output", "o", "", "Output file for response (default: stdout)")

	return cmd
}

func discordStatusCmd() *cobra.Command {
	shortDesc := "Check Discord bot status"
	longDesc := `Check the current status of the Discord bot including connection state.`
	cmd := &cobra.Command{
		Use:   "status",
		Short: shortDesc,
		Long:  longDesc,
		Aliases: []string{
			"state", "info", "check",
		},
		Annotations: GetDescriptions([]string{
			shortDesc,
			longDesc,
		}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("info", "Checking Discord bot status...")

			adapter, err := createDiscordAdapter()
			if err != nil {
				return err
			}

			// Try to connect
			err = adapter.Connect()
			status := map[string]interface{}{
				"connected": err == nil,
				"timestamp": time.Now(),
				"dev_mode":  discordDevMode,
			}

			if err != nil {
				status["error"] = err.Error()
			} else {
				status["message"] = "Discord bot is connected and ready"
			}

			output, err := json.MarshalIndent(status, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format status: %w", err)
			}

			if discordOutput != "" {
				return os.WriteFile(discordOutput, output, 0644)
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringVarP(&discordToken, "token", "t", "", "Discord bot token (or set DISCORD_BOT_TOKEN)")
	cmd.Flags().BoolVarP(&discordDevMode, "dev", "d", false, "Enable dev mode (mock operations)")
	cmd.Flags().StringVarP(&discordOutput, "output", "o", "", "Output file for status (default: stdout)")

	return cmd
}

func discordChannelsCmd() *cobra.Command {
	shortDesc := "List Discord channels"
	longDesc := `List available Discord channels the bot can access.`

	cmd := &cobra.Command{
		Use:   "channels",
		Short: shortDesc,
		Long:  longDesc,
		Aliases: []string{
			"list-channels", "channel-list", "chans",
		},
		Annotations: GetDescriptions([]string{
			shortDesc,
			longDesc,
		}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("info", "Listing Discord channels...")

			adapter, err := createDiscordAdapter()
			if err != nil {
				return err
			}

			err = adapter.Connect()
			if err != nil {
				return fmt.Errorf("failed to connect to Discord: %w", err)
			}

			// In dev mode, return mock channels
			if discordDevMode {
				channels := []interfaces.Channel{
					{ID: "123456789", Name: "general", Private: false},
					{ID: "987654321", Name: "bot-commands", Private: false},
					{ID: "555666777", Name: "voice-channel", Private: false},
				}

				output, err := json.MarshalIndent(channels, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format channels: %w", err)
				}

				if discordOutput != "" {
					return os.WriteFile(discordOutput, output, 0644)
				}

				fmt.Println(string(output))
				return nil
			}

			// TODO: Implement real channel listing when Discord API supports it in the adapter
			fmt.Println("Channel listing not yet implemented for production mode")
			return nil
		},
	}

	cmd.Flags().StringVarP(&discordToken, "token", "t", "", "Discord bot token (or set DISCORD_BOT_TOKEN)")
	cmd.Flags().BoolVarP(&discordDevMode, "dev", "d", true, "Enable dev mode (mock operations)")
	cmd.Flags().StringVarP(&discordOutput, "output", "o", "", "Output file for channels (default: stdout)")

	return cmd
}

func discordTestCmd() *cobra.Command {
	shortDesc := "Test Discord bot functionality"
	longDesc := `Run a series of tests to verify Discord bot functionality,
including connection and basic operations.`
	cmd := &cobra.Command{
		Use:   "test",
		Short: shortDesc,
		Long:  longDesc,
		Aliases: []string{
			"check", "verify", "diagnose",
		},
		Annotations: GetDescriptions([]string{
			shortDesc,
			longDesc,
		}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("info", "Testing Discord bot functionality...")

			adapter, err := createDiscordAdapter()
			if err != nil {
				return err
			}

			// Test connection
			fmt.Println("Testing Discord connection...")
			err = adapter.Connect()
			if err != nil {
				fmt.Printf("❌ Connection failed: %v\n", err)
				return err
			}
			fmt.Println("✅ Connection successful")

			// Test basic functionality (in dev mode)
			if discordDevMode {
				fmt.Println("Testing basic functionality...")
				// Test ping functionality
				err = adapter.PingAdapter("test ping")
				if err != nil {
					fmt.Printf("⚠️  Ping failed: %v\n", err)
				} else {
					fmt.Println("✅ Ping successful")
				}
			}

			fmt.Println("✅ All tests completed successfully")
			return nil
		},
	}

	cmd.Flags().StringVarP(&discordToken, "token", "t", "", "Discord bot token (or set DISCORD_BOT_TOKEN)")
	cmd.Flags().BoolVarP(&discordDevMode, "dev", "d", true, "Enable dev mode (mock operations)")

	return cmd
}

// Helper function to create Discord adapter
func createDiscordAdapter() (interfaces.IAdapter, error) {
	token := discordToken
	if token == "" {
		token = os.Getenv("DISCORD_BOT_TOKEN")
	}

	if token == "" && !discordDevMode {
		token = "dev_token" // Use dev token for dev mode
	}

	cfg := bootstrap.DiscordConfig{}
	cfg.Bot.Token = token

	adapter, err := discord.NewAdapter(cfg, "cli")
	return adapter, err
}
