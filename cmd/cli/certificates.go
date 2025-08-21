package cli

import (
	"fmt"
	"os"

	gl "github.com/rafa-mori/gobe/internal/module/logger"
	crt "github.com/rafa-mori/gobe/internal/security/certificates"
	crp "github.com/rafa-mori/gobe/internal/security/crypto"
	"github.com/spf13/cobra"
)

func CertificatesCmdList() *cobra.Command {
	certificatesCmd := &cobra.Command{
		Use:   "certificates",
		Short: "Certificates commands",
		Long:  "Certificates commands for GoBE or any other service",
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

	short := "Generate certificates for GoBE or any other service"
	long := "Generate certificates for GoBE or any other service using the provided configuration file"

	var startCmd = &cobra.Command{
		Use:         "generate",
		Short:       short,
		Long:        long,
		Annotations: GetDescriptions([]string{short, long}, false),
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

	short := "Verify certificates for GoBE or any other service"
	long := "Verify certificates for GoBE or any other service using the provided configuration file"

	var startCmd = &cobra.Command{
		Use:         "verify",
		Short:       short,
		Long:        long,
		Annotations: GetDescriptions([]string{short, long}, false),
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

	short := "Generate a random key for GoBE or any other service"
	long := "Generate a random key for GoBE or any other service using the provided configuration file"

	var startCmd = &cobra.Command{
		Use:         "random-key",
		Short:       short,
		Long:        long,
		Annotations: GetDescriptions([]string{short, long}, false),
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
