package main

import (
	cc "github.com/rafa-mori/gobe/cmd/cli"
	gl "github.com/rafa-mori/gobe/logger"
	vs "github.com/rafa-mori/gobe/version"
	"github.com/spf13/cobra"

	"os"
	"strings"
)

type GoBE struct {
	parentCmdName string
	printBanner   bool
}

func (m *GoBE) Alias() string { return "" }
func (m *GoBE) ShortDescription() string {
	return "GoBE is a minimalistic backend service with Go."
}
func (m *GoBE) LongDescription() string {
	return `GoBE: A minimalistic backend service with Go.`
}
func (m *GoBE) Usage() string {
	return "gobe [command] [args]"
}
func (m *GoBE) Examples() []string {
	return []string{"gobe start -p ':8080' -b '0.0.0.0' -n 'MyService' -d"}
}
func (m *GoBE) Active() bool {
	return true
}
func (m *GoBE) Module() string {
	return "gobe"
}
func (m *GoBE) Execute() error { return m.Command().Execute() }
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
	rtCmd.AddCommand(cc.ServiceCmdList()...)
	rtCmd.AddCommand(vs.CliCommand())

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

func RegX() *GoBE {
	var printBannerV = os.Getenv("GOBEMIN_PRINT_BANNER")
	if printBannerV == "" {
		printBannerV = "true"
	}

	return &GoBE{
		printBanner: strings.ToLower(printBannerV) == "true",
	}
}
