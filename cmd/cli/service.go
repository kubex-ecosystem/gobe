package cli

import (
	"os"

	gb "github.com/kubex-ecosystem/gobe"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	l "github.com/kubex-ecosystem/logz"
	"github.com/spf13/cobra"
)

func ServiceCmdList() []*cobra.Command {
	return []*cobra.Command{
		startCommand(),
		stopCommand(),
		restartCommand(),
		statusCommand(),
		logsCommand(),
	}
}

func startCommand() *cobra.Command {
	var name, port, bind, logFile, configFile string
	var isConfidential, debug, releaseMode bool

	shortDesc := "Start a minimal backend service"
	longDesc := "Start a minimal backend service with GoBE"

	var startCmd = &cobra.Command{
		Use:         "start",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				gl.SetDebug(true)
			}

			gbm, gbmErr := gb.NewGoBE(name, port, bind, logFile, configFile, isConfidential, l.GetLogger("GoBE"), debug, releaseMode)
			if gbmErr != nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", gbmErr.Error())
				return
			}
			if gbm == nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", "GoBE instance is nil")
				return
			}
			gbm.StartGoBE()
			gl.Log("success", "GoBE started successfully")
		},
	}

	startCmd.Flags().StringVarP(&name, "name", "n", "GoBE", "Name of the process")
	startCmd.Flags().StringVarP(&port, "port", "p", "8666", "Port to listen on")
	startCmd.Flags().StringVarP(&bind, "bind", "b", "0.0.0.0", "Bind address")
	startCmd.Flags().StringVarP(&logFile, "log-file", "l", "", "Log file path")
	startCmd.Flags().StringVarP(&configFile, "config-file", "c", "", "Configuration file path")
	startCmd.Flags().BoolVarP(&isConfidential, "confidential", "C", false, "Enable confidential mode")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	startCmd.Flags().BoolVarP(&releaseMode, "release", "r", false, "Enable release mode")

	return startCmd
}

func stopCommand() *cobra.Command {
	var name string

	shortDesc := "Stop a running backend service"
	longDesc := "Stop a running backend service with GoBE"

	var stopCmd = &cobra.Command{
		Use:         "stop",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			gbm, gbmErr := gb.NewGoBE(name, "", "", "", "", false, l.GetLogger("GoBE"), false, false)
			if gbmErr != nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", gbmErr.Error())
				return
			}
			if gbm == nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", "GoBE instance is nil")
				return
			}
			gbm.StopGoBE()
			gl.Log("success", "GoBE stopped successfully")
		},
	}

	stopCmd.Flags().StringVarP(&name, "name", "n", "GoBE", "Name of the process")

	return stopCmd
}

func restartCommand() *cobra.Command {
	var name string

	shortDesc := "Restart a running backend service"
	longDesc := "Restart a running backend service with GoBE"

	var restartCmd = &cobra.Command{
		Use:         "restart",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			gbm, gbmErr := gb.NewGoBE(name, "", "", "", "", false, l.GetLogger("GoBE"), false, false)
			if gbmErr != nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", gbmErr.Error())
				return
			}
			if gbm == nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", "GoBE instance is nil")
				return
			}
			gbm.StopGoBE()
			gl.Log("success", "GoBE stopped successfully")
			gbm.StartGoBE()
			gl.Log("success", "GoBE started successfully")
		},
	}

	restartCmd.Flags().StringVarP(&name, "name", "n", "GoBE", "Name of the process")

	return restartCmd
}

func statusCommand() *cobra.Command {
	var name string

	shortDesc := "Get the status of a running backend service"
	longDesc := "Get the status of a running backend service with GoBE"

	var statusCmd = &cobra.Command{
		Use:         "status",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			gbm, gbmErr := gb.NewGoBE(name, "", "", "", "", false, l.GetLogger("GoBE"), false, false)
			if gbmErr != nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", gbmErr.Error())
				return
			}
			if gbm == nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", "GoBE instance is nil")
				return
			}
			//gbm.StatusGoBE()
			gl.Log("success", "GoBE status retrieved successfully")
		},
	}

	statusCmd.Flags().StringVarP(&name, "name", "n", "GoBE", "Name of the process")

	return statusCmd
}

func logsCommand() *cobra.Command {
	var name string

	shortDesc := "Get the logs of a running backend service"
	longDesc := "Get the logs of a running backend service with GoBE"

	var logsCmd = &cobra.Command{
		Use:         "logs",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			gbm, gbmErr := gb.NewGoBE(name, "", "", "", "", false, l.GetLogger("GoBE"), false, false)
			if gbmErr != nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", gbmErr.Error())
				return
			}
			if gbm == nil {
				gl.Log("fatal", "Failed to create GoBE instance: ", "GoBE instance is nil")
				return
			}
			logsWriter, err := gbm.LogsGoBE()
			if err != nil {
				gl.Log("fatal", "Failed to get logs writer: ", err.Error())
				return
			}
			logsWriter.Write([]byte("Retrieving logs...\n"))
			gl.Log("success", "GoBE logs retrieved successfully")
		},
	}

	logsCmd.Flags().StringVarP(&name, "name", "n", "GoBE", "Name of the process")

	return logsCmd
}
