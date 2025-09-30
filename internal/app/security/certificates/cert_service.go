package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	crp "github.com/kubex-ecosystem/gobe/internal/app/security/crypto"
	krs "github.com/kubex-ecosystem/gobe/internal/app/security/external"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
	cm "github.com/kubex-ecosystem/gobe/internal/commons"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	"golang.org/x/crypto/chacha20poly1305"
)

// CertService provides methods for managing certificates and private keys.
// It supports generating, encrypting, decrypting, and verifying certificates.
type CertService struct {
	keyPath  string             // Path to the private key file.
	certPath string             // Path to the certificate file.
	security *crp.CryptoService // Service for cryptographic operations.
}

// GenerateCertificate generates a self-signed certificate and encrypts the private key.
// Parameters:
// - certPath: Path to save the certificate.
// - keyPath: Path to save the private key.
// - password: Password used to encrypt the private key.
// Returns: The encrypted private key, the certificate bytes, and an error if any.
func (c *CertService) GenerateCertificate(certPath, keyPath string, password []byte) ([]byte, []byte, error) {
	priv, generateKeyErr := rsa.GenerateKey(rand.Reader, 2048)
	if generateKeyErr != nil {
		gl.Log("error", fmt.Sprintf("error generating private key: %v", generateKeyErr))
		return nil, nil, fmt.Errorf("error generating private key: %v", generateKeyErr)
	}

	sn, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	template := x509.Certificate{
		SerialNumber: sn,
		Subject:      pkix.Name{CommonName: "Kubex Self-Signed"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:         true,
	}

	certDER, certDERErr := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if certDERErr != nil {
		gl.Log("error", fmt.Sprintf("error creating certificate: %v", certDERErr))
		return nil, nil, fmt.Errorf("error creating certificate: %v", certDERErr)
	}

	var pwd string
	var pwdErr error
	if len(password) == 0 {
		pwd, pwdErr = GetOrGenPasswordKeyringPass("jwt_secret")

		if pwdErr != nil {
			gl.Log("error", fmt.Sprintf("error retrieving password: %v", pwdErr))
			return nil, nil, fmt.Errorf("error retrieving password: %w", pwdErr)
		}
		password = []byte(pwd)
	} else {
		pwd = string(password)
	}

	isEncoded := c.security.IsBase64String(pwd)
	var decodedPassword []byte
	var err error
	if isEncoded {
		decodedPassword, err = c.security.DecodeIfEncoded(password)
		if err != nil {
			gl.Log("error", fmt.Sprintf("error decoding password: %v", err))
			return nil, nil, fmt.Errorf("error decoding password: %w", err)
		}
	} else {
		decodedPassword = []byte(pwd)
	}
	pkcs1PrivBytes := x509.MarshalPKCS1PrivateKey(priv)

	block, err := chacha20poly1305.NewX(decodedPassword)
	if err != nil {
		gl.Log("error", fmt.Sprintf("error creating cipher: %v, %d", err, len(decodedPassword)))
		return nil, nil, fmt.Errorf("error creating cipher: %w", err)
	}

	nonce := make([]byte, block.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		gl.Log("error", fmt.Sprintf("error generating nonce: %v", err))
		return nil, nil, fmt.Errorf("error generating nonce: %w", err)
	}

	ciphertext := block.Seal(nonce, nonce, pkcs1PrivBytes, nil)
	if err := os.MkdirAll(filepath.Dir(keyPath), 0755); err != nil {
		gl.Log("error", fmt.Sprintf("error creating directory for key file: %v", err))
		return nil, nil, fmt.Errorf("error creating directory for key file: %w", err)
	}

	certFile, err := os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		gl.Log("error", fmt.Sprintf("error opening certificate file: %v", err))
		return nil, nil, fmt.Errorf("error opening certificate file: %w", err)
	}
	defer func(certFile *os.File) {
		_ = certFile.Close()
	}(certFile)

	pemBlock := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	if err := pem.Encode(certFile, &pemBlock); err != nil {
		gl.Log("error", fmt.Sprintf("error encoding certificate: %v", err))
		return nil, nil, fmt.Errorf("error encoding certificate: %w", err)
	}

	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		gl.Log("error", fmt.Sprintf("error opening key file: %v", err))
		return nil, nil, fmt.Errorf("error opening key file: %w", err)
	}
	defer func(keyFile *os.File) {
		_ = keyFile.Close()
	}(keyFile)

	pemBlock = pem.Block{Type: "RSA PRIVATE KEY", Bytes: ciphertext}
	if err := pem.Encode(keyFile, &pemBlock); err != nil {
		gl.Log("error", fmt.Sprintf("error encoding private key: %v", err))
		return nil, nil, fmt.Errorf("error encoding private key: %w", err)
	}

	return ciphertext, certDER, nil
}

// GenSelfCert generates a self-signed certificate and stores it in the configured paths.
// Returns: The encrypted private key, the certificate bytes, and an error if any.
func (c *CertService) GenSelfCert() ([]byte, []byte, error) {

	// HERE WE ARE USING THE KEYRING TO STORE THE PASSWORD
	// FOR THE CERTIFICATE AND PRIVATE KEY!!! THE NAME GIVEN
	// TO THE SECRET IS "jwt_secret" AND IT WILL BE USED TO
	// ENCRYPT THE PRIVATE KEY AND STORE IT IN THE KEYRING
	key, keyErr := GetOrGenPasswordKeyringPass("jwt_secret")

	if keyErr != nil {
		gl.Log("error", fmt.Sprintf("error retrieving password: %v", keyErr))
		return nil, nil, fmt.Errorf("error retrieving password: %w", keyErr)
	}
	return c.GenerateCertificate(c.certPath, c.keyPath, []byte(key))
}

// DecryptPrivateKey decrypts an encrypted private key using the provided password.
// Parameters:
// - ciphertext: The encrypted private key.
// - password: The password used for decryption.
// Returns: The decrypted private key and an error if any.
func (c *CertService) DecryptPrivateKey(privKeyBytes []byte, password []byte) (*rsa.PrivateKey, error) {
	if c == nil {
		gl.Log("fatal", "CertService is nil, trying to create a new one")
	}
	if password == nil {
		strPassword, passwordErr := GetOrGenPasswordKeyringPass("jwt_secret")
		if passwordErr != nil {
			gl.Log("error", fmt.Sprintf("error retrieving password: %v", passwordErr))
			return nil, fmt.Errorf("error retrieving password: %w", passwordErr)
		}
		password = []byte(strPassword)
	}

	privKeyDecrypted, _, err := crp.NewCryptoServiceType().Decrypt(privKeyBytes, password)
	if err != nil {
		return nil, fmt.Errorf("erro ao descriptografar chave privada: %w", err)
	}
	if len(privKeyDecrypted) == 0 {
		return nil, fmt.Errorf("erro ao descriptografar chave privada: %w", err)
	}

	return x509.ParsePKCS1PrivateKey([]byte(privKeyDecrypted))
}

// GetCertAndKeyFromFile reads the certificate and private key from their respective files.
// Returns: The certificate bytes, the private key bytes, and an error if any.
func (c *CertService) GetCertAndKeyFromFile() ([]byte, []byte, error) {
	if c == nil {
		gl.Log("warn", "CertService is nil, trying to create a new one")
		c = new(CertService)
	}
	if c.keyPath == "" {
		c.keyPath = os.ExpandEnv(cm.DefaultGoBEKeyPath)
	}
	if c.certPath == "" {
		c.certPath = os.ExpandEnv(cm.DefaultGoBECertPath)
	}
	certBytes, err := os.ReadFile(os.ExpandEnv(c.certPath))
	if err != nil {
		return nil, nil, fmt.Errorf("error reading certificate file: %w", err)
	}

	keyBytes, err := os.ReadFile(os.ExpandEnv(c.keyPath))
	if err != nil {
		return nil, nil, fmt.Errorf("error reading key file: %w", err)
	}

	return certBytes, keyBytes, nil
}

// VerifyCert verifies the validity of the certificate stored in the configured path.
// Returns: An error if the certificate is invalid or cannot be read.
func (c *CertService) VerifyCert() error {
	if c == nil {
		gl.Log("warn", "CertService is nil, trying to create a new one")
		c = new(CertService)
	}
	if c.keyPath == "" {
		c.keyPath = os.ExpandEnv(cm.DefaultGoBEKeyPath)
	}
	if c.certPath == "" {
		c.certPath = os.ExpandEnv(cm.DefaultGoBECertPath)
	}
	certFile, err := os.Open(c.certPath)
	if err != nil {
		return fmt.Errorf("error opening certificate file: %w", err)
	}
	defer func(certFile *os.File) {
		_ = certFile.Close()
	}(certFile)

	certBytes, err := os.ReadFile(c.certPath)
	if err != nil {
		return fmt.Errorf("error reading certificate file: %w", err)
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		return fmt.Errorf("error decoding certificate")
	}

	_, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("error parsing certificate: %w", err)
	}

	return nil
}

// GetPublicKey retrieves the public key from the certificate file.
// Returns: The public key and an error if any.
func (c *CertService) GetPublicKey() (*rsa.PublicKey, error) {
	if c == nil {
		gl.Log("warn", "CertService is nil, trying to create a new one")
		c = new(CertService)
	}
	if c.keyPath == "" {
		c.keyPath = os.ExpandEnv(cm.DefaultGoBEKeyPath)
	}
	if c.certPath == "" {
		c.certPath = os.ExpandEnv(cm.DefaultGoBECertPath)
	}
	certBytes, err := os.ReadFile(os.ExpandEnv(c.certPath))
	if err != nil {
		gl.Log("error", fmt.Sprintf("error reading certificate file: %v", err))
		return nil, fmt.Errorf("error reading certificate file: %w", err)
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		gl.Log("error", "error decoding certificate")
		return nil, fmt.Errorf("error decoding certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		gl.Log("error", fmt.Sprintf("error parsing certificate: %v", err))
		return nil, fmt.Errorf("error parsing certificate: %w", err)
	}

	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		gl.Log("error", "error asserting public key type")
		return nil, fmt.Errorf("error asserting public key type")
	}

	return pubKey, nil
}

// GetPrivateKey retrieves the private key from the key file.
// Returns: The private key and an error if any.
func (c *CertService) GetPrivateKey() (*rsa.PrivateKey, error) {
	var err error
	if c.keyPath == "" {
		c.keyPath = os.ExpandEnv(cm.DefaultGoBEKeyPath)
	}
	keyBytes, err := os.ReadFile(c.keyPath)
	if err != nil {
		return nil, fmt.Errorf("error reading certificate file: %w", err)
	}
	privateKeyBlock, _ := pem.Decode(keyBytes)
	if privateKeyBlock == nil {
		return nil, fmt.Errorf("error decoding private key")
	}
	// pwd, err := GetOrGenPasswordKeyringPass(cm.KeyringService)
	pwd, err := GetOrGenPasswordKeyringPass("jwt_secret")
	if err != nil {
		return nil, fmt.Errorf("error retrieving password: %w", err)
	}

	isEncoded := c.security.IsBase64String(pwd)
	var decodedPassword []byte
	if isEncoded {
		decodedPassword, err = c.security.DecodeBase64(pwd)
		if err != nil {
			return nil, fmt.Errorf("error decoding password: %w", err)
		}
	} else {
		decodedPassword = []byte(pwd)
	}

	copyKey := make([]byte, len(privateKeyBlock.Bytes))
	copy(copyKey, privateKeyBlock.Bytes)
	privKeyDecrypted, privKeyDecryptedDecoded, err := c.security.Decrypt(copyKey, decodedPassword)
	if err != nil {
		gl.Log("error", fmt.Sprintf("error decrypting private key: %v", err))
		return nil, fmt.Errorf("erro ao descriptografar chave privada: %w", err)
	}
	if len(privKeyDecrypted) == 0 {
		gl.Log("error", "error decrypting private key: empty result")
		return nil, fmt.Errorf("erro ao descriptografar chave privada: %w", err)
	}

	isEncoded = c.security.IsBase64String(string(privKeyDecrypted))
	if isEncoded {
		gl.Log("debug", "private key is encoded, decoding it")
		privKeyDecryptedBytes, err := c.security.DecodeBase64(string(privKeyDecrypted))
		if err != nil {
			return nil, fmt.Errorf("error decoding private key: %w", err)
		}
		privKeyDecrypted = string(privKeyDecryptedBytes)
		if privKeyDecrypted != string(privKeyDecryptedDecoded) {
			gl.Log("error", "decoded private key is not equal to decrypted private key")
			return nil, fmt.Errorf("decoded private key is not equal to decrypted private key")
		}
	}
	isEncoded = c.security.IsBase64String(string(privKeyDecrypted))
	var privKeyDecryptedDecodedAgain []byte
	if isEncoded {
		privKeyDecryptedDecodedAgain, err = c.security.DecodeBase64(string(privKeyDecrypted))
		if err != nil {
			return nil, fmt.Errorf("error decoding private key: %w", err)
		}
	} else {
		privKeyDecryptedDecodedAgain = []byte(string(privKeyDecrypted))
	}

	privKey, err := x509.ParsePKCS1PrivateKey(privKeyDecryptedDecodedAgain)
	if err != nil {
		return nil, fmt.Errorf("error decrypting private key: %w", err)
	}
	return privKey, nil
}

// newCertService creates a new instance of CertService with the provided paths.
// Parameters:
// - keyPath: Path to the private key file.
// - certPath: Path to the certificate file.
// Returns: A pointer to a CertService instance.
func newCertService(keyPath, certPath string) *CertService {
	if keyPath == "" {
		keyPath = os.ExpandEnv(cm.DefaultGoBEKeyPath)
	}
	if certPath == "" {
		certPath = os.ExpandEnv(cm.DefaultGoBECertPath)
	}
	crtService := &CertService{
		keyPath:  os.ExpandEnv(keyPath),
		certPath: os.ExpandEnv(certPath),
	}
	return crtService
}

// NewCertService creates a new CertService and returns it as an interface.
// Parameters:
// - keyPath: Path to the private key file.
// - certPath: Path to the certificate file.
// Returns: An implementation of sci.ICertService.
func NewCertService(keyPath, certPath string) sci.ICertService {
	return newCertService(keyPath, certPath)
}

// NewCertServiceType creates a new CertService and returns it as a concrete type.
// Parameters:
// - keyPath: Path to the private key file.
// - certPath: Path to the certificate file.
// Returns: A pointer to a CertService instance.
func NewCertServiceType(keyPath, certPath string) *CertService {
	return newCertService(keyPath, certPath)
}

// GetOrGenPasswordKeyringPass retrieves the password from the keyring or generates a new one if it doesn't exist
// It uses the keyring service name to store and retrieve the password
// These methods aren't exposed to the outside world, only accessible through the package main logic
func GetOrGenPasswordKeyringPass(name string) (string, error) {
	cryptoService := crp.NewCryptoServiceType()

	// Try to retrieve the password from the keyring
	krPass, krPassErr := krs.NewKeyringService(cm.KeyringService, fmt.Sprintf("gobe-%s", name)).RetrievePassword()
	if krPassErr != nil {
		if errors.Is(krPassErr, os.ErrNotExist) {
			// If the error is "keyring: item not found", generate a new key
			gl.Log("debug", fmt.Sprintf("Key not found, generating new key for %s", name))
			krPassKey, krPassKeyErr := cryptoService.GenerateKey()
			if krPassKeyErr != nil {
				gl.Log("error", fmt.Sprintf("Error generating key: %v", krPassKeyErr))
				return "", krPassKeyErr
			}

			// Store the password in the keyring and return the encoded password
			// Passing a string, we avoid the pointless conversion
			// to []byte and then back to string
			// This is a better practice for performance and readability
			encodedPass, storeErr := storeKeyringPassword(name, string(krPassKey))
			if storeErr != nil {
				gl.Log("error", fmt.Sprintf("Error storing key: %v", storeErr))
				return "", storeErr
			}

			return encodedPass, nil
		} else {
			gl.Log("error", fmt.Sprintf("Error retrieving key: %v", krPassErr))
			return "", krPassErr
		}
	}

	isEncoded := cryptoService.IsBase64String(krPass)

	if !isEncoded {
		gl.Log("debug", fmt.Sprintf("Keyring password is not encoded, encoding it for %s", name))
		return cryptoService.EncodeBase64([]byte(krPass)), nil
	}

	return krPass, nil
}

// storeKeyringPassword stores the password in the keyring
// It will check if data is encoded, if so, will decode, store and then
// encode again or encode for the first time, returning always a portable data for
// the caller/logic outside this package be able to use it better and safer
// This method is not exposed to the outside world, only accessible through the package main logic
func storeKeyringPassword(name string, pass string) (string, error) {
	cryptoService := crp.NewCryptoServiceType()
	// Will decode if encoded, but only if the password is not empty, not nil and not ENCODED

	var outputPass string

	isEncoded := cryptoService.IsBase64String(pass)
	if isEncoded {
		var decodeErr error
		var decodedPassByte []byte
		// Will decode if encoded, but only if the password is not empty, not nil and not ENCODED
		decodedPassByte, decodeErr = cryptoService.DecodeBase64(pass)
		if decodeErr != nil {
			gl.Log("error", fmt.Sprintf("Error decoding password: %v", decodeErr))
			return "", decodeErr
		}
		outputPass = string(decodedPassByte)
	} else {
		outputPass = pass
	}

	// Check if the decoded password is empty
	if len(outputPass) == 0 {
		gl.Log("error", "Decoded password is empty")
		return "", errors.New("decoded password is empty")
	}

	// Store the password in the keyring DECODED to avoid storing the encoded password
	// locally are much better for security keep binary static and encoded to handle with transport
	// integration and other utilities
	storeErr := krs.NewKeyringService(cm.KeyringService, fmt.Sprintf("gobe-%s", name)).StorePassword(outputPass)
	if storeErr != nil {
		gl.Log("error", fmt.Sprintf("Error storing key: %v", storeErr))
		return "", storeErr
	}

	isEncoded = cryptoService.IsBase64String(outputPass)
	if !isEncoded {
		outputPass = cryptoService.EncodeBase64([]byte(outputPass))
	}

	// Retrieve the password ENCODED to provide a portable data for the caller/logic outside this package
	return outputPass, nil
}
