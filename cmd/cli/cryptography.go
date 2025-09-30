package cli

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/kubex-ecosystem/gobe/internal/app/security/crypto"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	"github.com/spf13/cobra"
)

func CryptographyCommand() *cobra.Command {

	shortDesc := "Cryptography related commands"
	longDesc := "Cryptography related commands for encoding, decoding, hashing, encryption, and more"

	var cmd = &cobra.Command{
		Use:         "cryptography",
		Short:       shortDesc,
		Long:        longDesc,
		Aliases:     []string{"crypto", "crypt", "crp"},
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GOBE_HIDEBANNER") == "true")),
	}

	cmd.AddCommand(encodeDataCmd())
	cmd.AddCommand(decodeDataCmd())
	cmd.AddCommand(hashDataCmd())
	cmd.AddCommand(verifyHashCmd())
	cmd.AddCommand(generateKeyCmd())
	cmd.AddCommand(encryptDataCmd())
	cmd.AddCommand(decryptDataCmd())
	cmd.AddCommand(signDataCmd())
	cmd.AddCommand(verifySignatureCmd())

	return cmd
}

func encodeDataCmd() *cobra.Command {
	var inputFile, outputFile, encoding string
	var input string

	cmd := &cobra.Command{
		Use:   "encode",
		Short: "Encode data using a specified encoding scheme (base64, hex)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return encodeData(input, inputFile, outputFile, encoding)
		},
	}

	cmd.Flags().StringVarP(&input, "input", "i", "", "Input data to encode")
	cmd.Flags().StringVarP(&inputFile, "input-file", "f", "", "Input file to encode")
	cmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVarP(&encoding, "encoding", "e", "base64", "Encoding type: base64, hex")

	return cmd
}

func decodeDataCmd() *cobra.Command {
	var inputFile, outputFile, encoding string
	var input string

	cmd := &cobra.Command{
		Use:   "decode",
		Short: "Decode data using a specified decoding scheme (base64, hex)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return decodeData(input, inputFile, outputFile, encoding)
		},
	}

	cmd.Flags().StringVarP(&input, "input", "i", "", "Input data to decode")
	cmd.Flags().StringVarP(&inputFile, "input-file", "f", "", "Input file to decode")
	cmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVarP(&encoding, "encoding", "e", "base64", "Encoding type: base64, hex")

	return cmd
}

func hashDataCmd() *cobra.Command {
	var inputFile, algorithm string
	var input string

	cmd := &cobra.Command{
		Use:   "hash",
		Short: "Generate a hash of the input data (sha256, md5)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return hashData(input, inputFile, algorithm)
		},
	}

	cmd.Flags().StringVarP(&input, "input", "i", "", "Input data to hash")
	cmd.Flags().StringVarP(&inputFile, "input-file", "f", "", "Input file to hash")
	cmd.Flags().StringVarP(&algorithm, "algorithm", "a", "sha256", "Hash algorithm: sha256, md5")

	return cmd
}

func verifyHashCmd() *cobra.Command {
	var inputFile, hashValue, algorithm string
	var input string

	cmd := &cobra.Command{
		Use:   "verify-hash",
		Short: "Verify the hash of the input data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return verifyHash(input, inputFile, hashValue, algorithm)
		},
	}

	cmd.Flags().StringVarP(&input, "input", "i", "", "Input data to verify")
	cmd.Flags().StringVarP(&inputFile, "input-file", "f", "", "Input file to verify")
	cmd.Flags().StringVarP(&hashValue, "hash", "H", "", "Expected hash value")
	cmd.Flags().StringVarP(&algorithm, "algorithm", "a", "sha256", "Hash algorithm: sha256, md5")

	return cmd
}

func generateKeyCmd() *cobra.Command {
	var length int
	var outputFile string

	cmd := &cobra.Command{
		Use:   "generate-key",
		Short: "Generate a new cryptographic key",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateKey(length, outputFile)
		},
	}

	cmd.Flags().IntVarP(&length, "length", "l", 32, "Key length in bytes")
	cmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Output file (default: stdout)")

	return cmd
}

func encryptDataCmd() *cobra.Command {
	var inputFile, outputFile, keyFile string
	var input, key string

	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt data using ChaCha20-Poly1305 algorithm",
		RunE: func(cmd *cobra.Command, args []string) error {
			return encryptData(input, inputFile, outputFile, key, keyFile)
		},
	}

	cmd.Flags().StringVarP(&input, "input", "i", "", "Input data to encrypt")
	cmd.Flags().StringVarP(&inputFile, "input-file", "f", "", "Input file to encrypt")
	cmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVarP(&key, "key", "k", "", "Encryption key")
	cmd.Flags().StringVarP(&keyFile, "key-file", "K", "", "Key file")

	return cmd
}

func decryptDataCmd() *cobra.Command {
	var inputFile, outputFile, keyFile string
	var input, key string

	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt data using ChaCha20-Poly1305 algorithm",
		RunE: func(cmd *cobra.Command, args []string) error {
			return decryptData(input, inputFile, outputFile, key, keyFile)
		},
	}

	cmd.Flags().StringVarP(&input, "input", "i", "", "Input data to decrypt")
	cmd.Flags().StringVarP(&inputFile, "input-file", "f", "", "Input file to decrypt")
	cmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVarP(&key, "key", "k", "", "Decryption key")
	cmd.Flags().StringVarP(&keyFile, "key-file", "K", "", "Key file")

	return cmd
}

func signDataCmd() *cobra.Command {
	var inputFile, keyFile, algorithm string
	var input, key string

	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Sign data using a private key (SHA256 signature)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return signData(input, inputFile, key, keyFile, algorithm)
		},
	}

	cmd.Flags().StringVarP(&input, "input", "i", "", "Input data to sign")
	cmd.Flags().StringVarP(&inputFile, "input-file", "f", "", "Input file to sign")
	cmd.Flags().StringVarP(&key, "key", "k", "", "Private key for signing")
	cmd.Flags().StringVarP(&keyFile, "key-file", "K", "", "Private key file")
	cmd.Flags().StringVarP(&algorithm, "algorithm", "a", "sha256", "Signature algorithm")

	return cmd
}

func verifySignatureCmd() *cobra.Command {
	var inputFile, keyFile, signature, algorithm string
	var input, key string

	cmd := &cobra.Command{
		Use:   "verify-signature",
		Short: "Verify the signature of the input data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return verifySignature(input, inputFile, key, keyFile, signature, algorithm)
		},
	}

	cmd.Flags().StringVarP(&input, "input", "i", "", "Input data to verify")
	cmd.Flags().StringVarP(&inputFile, "input-file", "f", "", "Input file to verify")
	cmd.Flags().StringVarP(&key, "key", "k", "", "Public key for verification")
	cmd.Flags().StringVarP(&keyFile, "key-file", "K", "", "Public key file")
	cmd.Flags().StringVarP(&signature, "signature", "s", "", "Signature to verify")
	cmd.Flags().StringVarP(&algorithm, "algorithm", "a", "sha256", "Signature algorithm")

	return cmd
}

// Implementation functions

func encodeData(input, inputFile, outputFile, encoding string) error {
	cryptoService := crypto.NewCryptoService()

	var data []byte
	var err error

	// Get input data
	if inputFile != "" {
		data, err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	} else if input != "" {
		data = []byte(input)
	} else {
		// Read from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter data to encode: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		data = []byte(strings.TrimSpace(input))
	}

	var encoded string
	switch strings.ToLower(encoding) {
	case "base64":
		encoded = cryptoService.EncodeBase64(data)
	case "hex":
		encoded = hex.EncodeToString(data)
	default:
		return fmt.Errorf("unsupported encoding: %s", encoding)
	}

	// Output result
	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(encoded), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		gl.Log("success", fmt.Sprintf("Encoded data written to %s", outputFile))
	} else {
		fmt.Println(encoded)
	}

	return nil
}

func decodeData(input, inputFile, outputFile, encoding string) error {
	cryptoService := crypto.NewCryptoService()

	var encodedData string
	var err error

	// Get input data
	if inputFile != "" {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
		encodedData = string(data)
	} else if input != "" {
		encodedData = input
	} else {
		// Read from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter data to decode: ")
		encodedData, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		encodedData = strings.TrimSpace(encodedData)
	}

	var decoded []byte
	switch strings.ToLower(encoding) {
	case "base64":
		decoded, err = cryptoService.DecodeBase64(encodedData)
		if err != nil {
			return fmt.Errorf("failed to decode base64: %w", err)
		}
	case "hex":
		decoded, err = hex.DecodeString(encodedData)
		if err != nil {
			return fmt.Errorf("failed to decode hex: %w", err)
		}
	default:
		return fmt.Errorf("unsupported encoding: %s", encoding)
	}

	// Output result
	if outputFile != "" {
		err = os.WriteFile(outputFile, decoded, 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		gl.Log("success", fmt.Sprintf("Decoded data written to %s", outputFile))
	} else {
		fmt.Print(string(decoded))
	}

	return nil
}

func hashData(input, inputFile, algorithm string) error {
	var data []byte
	var err error

	// Get input data
	if inputFile != "" {
		data, err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	} else if input != "" {
		data = []byte(input)
	} else {
		// Read from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter data to hash: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		data = []byte(strings.TrimSpace(input))
	}

	var hash string
	switch strings.ToLower(algorithm) {
	case "sha256":
		h := sha256.Sum256(data)
		hash = hex.EncodeToString(h[:])
	case "md5":
		h := md5.Sum(data)
		hash = hex.EncodeToString(h[:])
	default:
		return fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	fmt.Printf("Algorithm: %s\n", algorithm)
	fmt.Printf("Hash: %s\n", hash)

	return nil
}

func verifyHash(input, inputFile, expectedHash, algorithm string) error {
	var data []byte
	var err error

	// Get input data
	if inputFile != "" {
		data, err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	} else if input != "" {
		data = []byte(input)
	} else {
		return fmt.Errorf("no input data provided")
	}

	var calculatedHash string
	switch strings.ToLower(algorithm) {
	case "sha256":
		h := sha256.Sum256(data)
		calculatedHash = hex.EncodeToString(h[:])
	case "md5":
		h := md5.Sum(data)
		calculatedHash = hex.EncodeToString(h[:])
	default:
		return fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	fmt.Printf("Algorithm: %s\n", algorithm)
	fmt.Printf("Expected: %s\n", expectedHash)
	fmt.Printf("Calculated: %s\n", calculatedHash)

	if strings.EqualFold(expectedHash, calculatedHash) {
		gl.Log("success", "✅ Hash verification PASSED")
	} else {
		gl.Log("error", "❌ Hash verification FAILED")
		return fmt.Errorf("hash mismatch")
	}

	return nil
}

func generateKey(length int, outputFile string) error {
	cryptoService := crypto.NewCryptoService()

	var key []byte
	var err error

	if length == 32 {
		// Use ChaCha20 key generation
		key, err = cryptoService.GenerateKey()
	} else {
		// Use custom length key generation
		key, err = cryptoService.GenerateKeyWithLength(length)
	}

	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	keyHex := hex.EncodeToString(key)

	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(keyHex), 0600) // Restrictive permissions for key file
		if err != nil {
			return fmt.Errorf("failed to write key file: %w", err)
		}
		gl.Log("success", fmt.Sprintf("Key generated and saved to %s", outputFile))
	} else {
		fmt.Printf("Generated key (%d bytes): %s\n", len(key), keyHex)
	}

	return nil
}

func encryptData(input, inputFile, outputFile, key, keyFile string) error {
	cryptoService := crypto.NewCryptoService()

	// Get data to encrypt
	var data []byte
	var err error

	if inputFile != "" {
		data, err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	} else if input != "" {
		data = []byte(input)
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter data to encrypt: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		data = []byte(strings.TrimSpace(input))
	}

	// Get encryption key
	var keyData []byte
	if keyFile != "" {
		keyHex, err := os.ReadFile(keyFile)
		if err != nil {
			return fmt.Errorf("failed to read key file: %w", err)
		}
		keyData, err = hex.DecodeString(strings.TrimSpace(string(keyHex)))
		if err != nil {
			return fmt.Errorf("failed to decode key from file: %w", err)
		}
	} else if key != "" {
		keyData, err = hex.DecodeString(key)
		if err != nil {
			return fmt.Errorf("failed to decode key: %w", err)
		}
	} else {
		return fmt.Errorf("no encryption key provided")
	}

	// Encrypt
	encrypted, nonce, err := cryptoService.Encrypt(data, keyData)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	result := fmt.Sprintf("Encrypted: %s\nNonce: %s\n", encrypted, nonce)

	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(result), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		gl.Log("success", fmt.Sprintf("Encrypted data written to %s", outputFile))
	} else {
		fmt.Print(result)
	}

	return nil
}

func decryptData(input, inputFile, outputFile, key, keyFile string) error {
	cryptoService := crypto.NewCryptoService()

	// Get encrypted data
	var encryptedData []byte
	var err error

	if inputFile != "" {
		encryptedData, err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	} else if input != "" {
		encryptedData = []byte(input)
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter encrypted data: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		encryptedData = []byte(strings.TrimSpace(input))
	}

	// Get decryption key
	var keyData []byte
	if keyFile != "" {
		keyHex, err := os.ReadFile(keyFile)
		if err != nil {
			return fmt.Errorf("failed to read key file: %w", err)
		}
		keyData, err = hex.DecodeString(strings.TrimSpace(string(keyHex)))
		if err != nil {
			return fmt.Errorf("failed to decode key from file: %w", err)
		}
	} else if key != "" {
		keyData, err = hex.DecodeString(key)
		if err != nil {
			return fmt.Errorf("failed to decode key: %w", err)
		}
	} else {
		return fmt.Errorf("no decryption key provided")
	}

	// Decrypt
	decrypted, nonce, err := cryptoService.Decrypt(encryptedData, keyData)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(decrypted), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		gl.Log("success", fmt.Sprintf("Decrypted data written to %s", outputFile))
		gl.Log("info", fmt.Sprintf("Nonce used: %s", nonce))
	} else {
		fmt.Printf("Decrypted: %s\n", decrypted)
		fmt.Printf("Nonce: %s\n", nonce)
	}

	return nil
}

func signData(input, inputFile, key, keyFile, algorithm string) error {
	// Simple signature using hash + key concatenation (for demonstration)
	var data []byte
	var err error

	if inputFile != "" {
		data, err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	} else if input != "" {
		data = []byte(input)
	} else {
		return fmt.Errorf("no input data provided")
	}

	var keyData []byte
	if keyFile != "" {
		keyData, err = os.ReadFile(keyFile)
		if err != nil {
			return fmt.Errorf("failed to read key file: %w", err)
		}
	} else if key != "" {
		keyData = []byte(key)
	} else {
		return fmt.Errorf("no signing key provided")
	}

	// Create signature: hash(data + key)
	combined := append(data, keyData...)
	var signature string

	switch strings.ToLower(algorithm) {
	case "sha256":
		h := sha256.Sum256(combined)
		signature = hex.EncodeToString(h[:])
	case "md5":
		h := md5.Sum(data)
		signature = hex.EncodeToString(h[:])
	default:
		return fmt.Errorf("unsupported signature algorithm: %s", algorithm)
	}

	fmt.Printf("Algorithm: %s\n", algorithm)
	fmt.Printf("Signature: %s\n", signature)

	return nil
}

func verifySignature(input, inputFile, key, keyFile, expectedSignature, algorithm string) error {
	var data []byte
	var err error

	if inputFile != "" {
		data, err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	} else if input != "" {
		data = []byte(input)
	} else {
		return fmt.Errorf("no input data provided")
	}

	var keyData []byte
	if keyFile != "" {
		keyData, err = os.ReadFile(keyFile)
		if err != nil {
			return fmt.Errorf("failed to read key file: %w", err)
		}
	} else if key != "" {
		keyData = []byte(key)
	} else {
		return fmt.Errorf("no verification key provided")
	}

	// Calculate signature: hash(data + key)
	combined := append(data, keyData...)
	var calculatedSignature string

	switch strings.ToLower(algorithm) {
	case "sha256":
		h := sha256.Sum256(combined)
		calculatedSignature = hex.EncodeToString(h[:])
	case "md5":
		h := md5.Sum(data)
		calculatedSignature = hex.EncodeToString(h[:])
	default:
		return fmt.Errorf("unsupported signature algorithm: %s", algorithm)
	}

	fmt.Printf("Algorithm: %s\n", algorithm)
	fmt.Printf("Expected: %s\n", expectedSignature)
	fmt.Printf("Calculated: %s\n", calculatedSignature)

	if strings.EqualFold(expectedSignature, calculatedSignature) {
		gl.Log("success", "✅ Signature verification PASSED")
	} else {
		gl.Log("error", "❌ Signature verification FAILED")
		return fmt.Errorf("signature mismatch")
	}

	return nil
}
