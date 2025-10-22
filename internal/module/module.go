// Package module provides internal types and functions for the GoBE application.
package module

import (
	cc "github.com/kubex-ecosystem/gobe/cmd/cli"
	vs "github.com/kubex-ecosystem/gobe/internal/module/version"
	gl "github.com/kubex-ecosystem/logz/logger"
	"github.com/spf13/cobra"

	"os"
	"strings"
)

type GoBE struct {
	parentCmdName string
	printBanner   bool
}

func (m *GoBE) Alias() string {
	return ""
}
func (m *GoBE) ShortDescription() string {
	return "GoBE is a fully-featured backend, MCP server, CLI tool, and much more."
}
func (m *GoBE) LongDescription() string {
	return `GoBE is a fully-featured modular, open source backend solution designed to streamline the development
and deployment of applications. It serves as a robust backend framework, a Model Context Protocol (MCP), and a
versatile command-line interface (CLI) tool, among other functionalities and features. One Command. All the Power.
`
}
func (m *GoBE) Usage() string {
	return "gobe [command] [args]"
}
func (m *GoBE) Examples() []string {
	return []string{
		"gobe service start",
		"gobe mcp-server chat",
		"gobe cert generate",
	}
}
func (m *GoBE) Active() bool {
	return true
}
func (m *GoBE) Module() string {
	return "gobe"
}
func (m *GoBE) Execute() error {
	return m.Command().Execute()
}
func (m *GoBE) Command() *cobra.Command {
	gl.Log("debug", "Starting GoBE CLI...")

	var rtCmd = &cobra.Command{
		Use:     m.Module(),
		Aliases: []string{m.Alias()},
		Example: m.concatenateExamples(),
		Version: vs.GetVersion(),
		Annotations: cc.GetDescriptions([]string{
			m.LongDescription(),
			m.ShortDescription(),
		}, m.printBanner),
	}

	rtCmd.AddCommand(cc.CertificatesCmdList())
	rtCmd.AddCommand(cc.ServiceCmd())
	rtCmd.AddCommand(vs.CliCommand())
	rtCmd.AddCommand(cc.MCPServerCmd())
	rtCmd.AddCommand(cc.CryptographyCommand())
	rtCmd.AddCommand(cc.DiscordCommand())
	rtCmd.AddCommand(cc.WebhookCommand())
	rtCmd.AddCommand(cc.DatabaseCommand())
	rtCmd.AddCommand(cc.ConfigCommand())

	// Set usage definitions for the command and its subcommands
	setUsageDefinition(rtCmd)
	for _, c := range rtCmd.Commands() {
		setUsageDefinition(c)
		if !strings.Contains(strings.Join(os.Args, " "), c.Use) {
			if c.Short == "" {
				c.Short = c.Annotations["description"]
			}
		}
	}

	return rtCmd
}
func (m *GoBE) SetParentCmdName(rtCmd string) {
	m.parentCmdName = rtCmd
}
func (m *GoBE) concatenateExamples() string {
	examples := ""
	rtCmd := m.parentCmdName
	if rtCmd != "" {
		rtCmd = rtCmd + " "
	}
	for _, example := range m.Examples() {
		examples += rtCmd + example + "\n  "
	}
	return examples
}
