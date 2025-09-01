package cli

import (
	"fmt"
	"os"

	crt "github.com/rafa-mori/gobe/internal/app/security/certificates"
	crp "github.com/rafa-mori/gobe/internal/app/security/crypto"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
	"github.com/spf13/cobra"
)

func CertificatesCmdList() *cobra.Command {

	shortDesc := "Certificates commands"
	longDesc := "Certificates commands for GoBE or any other service"

	certificatesCmd := &cobra.Command{
		Use:         "certificates",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				gl.Log("error", fmt.Sprintf("Error displaying help: %v", err))
				return
			}
		},
	}
	cmdList := []*cobra.Command{
		generateCommand(),
		verifyCert(),
		generateRandomKey(),
	}
	certificatesCmd.AddCommand(cmdList...)
	return certificatesCmd
}

func generateCommand() *cobra.Command {
	var keyPath, certFilePath, certPass string
	var debug bool

	shortDesc := "Generate certificates for GoBE or any other service"
	longDesc := "Generate certificates for GoBE or any other service using the provided configuration file"

	var startCmd = &cobra.Command{
		Use:         "generate",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			crtS := crt.NewCertService(keyPath, certFilePath)
			_, _, err := crtS.GenerateCertificate(certFilePath, keyPath, []byte(certPass))
			if err != nil {
				gl.Log("fatal", fmt.Sprintf("Error generating certificate: %v", err))
			}
			gl.Log("success", "Certificate generated successfully")
		},
	}

	startCmd.Flags().StringVarP(&keyPath, "key-path", "k", "", "Path to the private key file")
	startCmd.Flags().StringVarP(&certFilePath, "cert-file-path", "c", "", "Path to the certificate file")
	startCmd.Flags().StringVarP(&certPass, "cert-pass", "p", "", "Password for the certificate")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	return startCmd
}

func verifyCert() *cobra.Command {
	var keyPath, certFilePath string
	var debug bool

	shortDesc := "Verify certificates for GoBE or any other service"
	longDesc := "Verify certificates for GoBE or any other service using the provided configuration file"

	var startCmd = &cobra.Command{
		Use:         "verify",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			crtS := crt.NewCertService(keyPath, certFilePath)
			err := crtS.VerifyCert()
			if err != nil {
				gl.Log("fatal", fmt.Sprintf("Error verifying certificate: %v", err))
			}
			gl.Log("success", "Certificate verified successfully")
		},
	}

	startCmd.Flags().StringVarP(&keyPath, "key-path", "k", "", "Path to the private key file")
	startCmd.Flags().StringVarP(&certFilePath, "cert-file-path", "c", "", "Path to the certificate file")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	return startCmd
}

func generateRandomKey() *cobra.Command {
	var keyPath string //, fileFormat string
	var length int
	var debug bool

	shortDesc := "Generate a random key for GoBE or any other service"
	longDesc := "Generate a random key for GoBE or any other service using the provided configuration file"

	var startCmd = &cobra.Command{
		Use:         "random-key",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			crtS := crp.NewCryptoService()
			var bts []byte
			var btsErr error
			if length > 0 {
				bts, btsErr = crtS.GenerateKeyWithLength(length)
			} else {
				bts, btsErr = crtS.GenerateKey()
			}
			if btsErr != nil {
				gl.Log("fatal", fmt.Sprintf("Error generating random key: %v", btsErr))
			}
			key := string(bts)
			if keyPath != "" {
				// File cannot exist, because this method will truncate the file
				if f, err := os.Stat(keyPath); f != nil && !os.IsNotExist(err) {
					gl.Log("error", fmt.Sprintf("File already exists: %s", keyPath))
					return
				}
				writeErr := os.WriteFile(keyPath, bts, 0644)
				if writeErr != nil {
					gl.Log("fatal", fmt.Sprintf("Error writing random key to file: %v", writeErr))
					return
				}
			}
			gl.Log("success", fmt.Sprintf("Random key generated successfully: %s", key))
		},
	}

	startCmd.Flags().StringVarP(&keyPath, "key-path", "k", "", "Path to the private key file")
	//startCmd.Flags().StringVarP(&fileFormat, "file-format", "f", "", "File format for the key (e.g., PEM, DER)")
	startCmd.Flags().IntVarP(&length, "length", "l", 16, "Length of the random key")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	return startCmd
}
