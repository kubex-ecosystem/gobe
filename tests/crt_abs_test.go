package tests

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"os"
	"testing"

	crt "github.com/rafa-mori/gobe/internal/security/certificates"
	krs "github.com/rafa-mori/gobe/internal/security/external"
	gl "github.com/rafa-mori/gobe/logger"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/chacha20poly1305"
)

type MockPasswordStore struct {
	StoredPassword string
}

func (m *MockPasswordStore) StorePassword(password string) {
	m.StoredPassword = password
}

func (m *MockPasswordStore) RetrievePassword() (string, error) {
	if m.StoredPassword == "" {
		return "", fmt.Errorf("senha não encontrada")
	}
	return m.StoredPassword, nil
}

type MockCertStorage struct {
	CertBuffer bytes.Buffer
	KeyBuffer  bytes.Buffer
}

func (m *MockCertStorage) StoreCertAndKey(cert []byte, key []byte) {
	m.CertBuffer.Write(cert)
	m.KeyBuffer.Write(key)
}

func (m *MockCertStorage) GetCertAndKey() ([]byte, []byte, error) {
	if m.CertBuffer.Len() == 0 || m.KeyBuffer.Len() == 0 {
		return nil, nil, fmt.Errorf("certificado ou chave não encontrados")
	}
	return m.CertBuffer.Bytes(), m.KeyBuffer.Bytes(), nil
}

func TestGenerateCertificateCreatesValidCertificateAndKey(t *testing.T) {
	certService := crt.NewCertServiceType("test.key", "test.crt")
	defer func() {
		err := os.Remove("test.key")
		if err != nil {
			gl.Log("error", fmt.Sprintf("Failed to remove test.key: %v", err))
		}
	}()
	defer func() {
		err := os.Remove("test.crt")
		if err != nil {
			gl.Log("error", fmt.Sprintf("Failed to remove test.crt: %v", err))
		}
	}()

	password := []byte("test-password")
	_, _, err := certService.GenerateCertificate("test.crt", "test.key", password)
	require.NoError(t, err)

	_, err = os.Stat("test.crt")
	require.NoError(t, err)

	_, err = os.Stat("test.key")
	require.NoError(t, err)
}

func TestGenSelfCertGeneratesCertificateAndStoresPassword(t *testing.T) {
	certService := crt.NewCertServiceType("test.key", "test.crt")
	defer func() {
		err := os.Remove("test.key")
		if err != nil {
			gl.Log("error", fmt.Sprintf("Failed to remove test.key: %v", err))
		}
	}()
	defer func() {
		err := os.Remove("test.crt")
		if err != nil {
			gl.Log("error", fmt.Sprintf("Failed to remove test.crt: %v", err))
		}
	}()

	_, _, err := certService.GenSelfCert()
	require.NoError(t, err)

	krServiceName := ""
	krKeyName := ""
	krService := krs.NewKeyringService(krKeyName, krServiceName)

	password, err := krService.RetrievePassword()
	require.NoError(t, err)
	require.NotEmpty(t, password)
}

func TestDecryptPrivateKeyWithValidDataDecryptsSuccessfully(t *testing.T) {
	certService := crt.NewCertServiceType("", "")
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	block, _ := chacha20poly1305.New([]byte("test-password"))
	nonce := make([]byte, block.NonceSize())
	_, _ = rand.Read(nonce)
	ciphertext := block.Seal(nonce, nonce, privateKeyBytes, nil)

	krServiceName := ""
	krKeyName := ""
	krService := krs.NewKeyringService(krKeyName, krServiceName)
	_ = krService.StorePassword("test-password")
	decryptedKey, err := certService.DecryptPrivateKey(ciphertext, []byte("test-password"))
	require.NoError(t, err)
	require.Equal(t, privateKey, decryptedKey)
}

func TestDecryptPrivateKeyWithInvalidPasswordReturnsError(t *testing.T) {
	certService := crt.NewCertServiceType("", "")
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	block, _ := chacha20poly1305.New([]byte("test-password"))
	nonce := make([]byte, block.NonceSize())
	_, _ = rand.Read(nonce)
	ciphertext := block.Seal(nonce, nonce, privateKeyBytes, nil)

	krServiceName := ""
	krKeyName := ""
	krService := krs.NewKeyringService(krKeyName, krServiceName)

	_ = krService.StorePassword("wrong-password")
	_, err := certService.DecryptPrivateKey(ciphertext, []byte("test-password"))
	require.Error(t, err)
}

func TestVerifyCertWithValidCertificateReturnsNoError(t *testing.T) {
	certService := crt.NewCertServiceType("test.key", "test.crt")
	defer func() {
		err := os.Remove("test.key")
		if err != nil {
			gl.Log("error", fmt.Sprintf("Failed to remove test.key: %v", err))
		}
	}()
	defer func() {
		err := os.Remove("test.crt")
		if err != nil {
			gl.Log("error", fmt.Sprintf("Failed to remove test.crt: %v", err))
		}
	}()

	_, _, _ = certService.GenSelfCert()
	err := certService.VerifyCert()
	require.NoError(t, err)
}

func TestVerifyCertWithInvalidCertificateReturnsError(t *testing.T) {
	certService := crt.NewCertServiceType("invalid.key", "invalid.crt")
	_, err := os.Create("invalid.crt")
	require.NoError(t, err)
	defer func() {
		err := os.Remove("invalid.crt")
		if err != nil {
			gl.Log("error", fmt.Sprintf("Failed to remove invalid.crt: %v", err))
		}
	}()

	err = certService.VerifyCert()
	require.Error(t, err)
}

func TestGetCertAndKeyFromFileReturnsCorrectData(t *testing.T) {
	certService := crt.NewCertServiceType("test.key", "test.crt")
	defer func() {
		err := os.Remove("test.key")
		if err != nil {
			gl.Log("error", fmt.Sprintf("Failed to remove test.key: %v", err))
		}
	}()
	defer func() {
		err := os.Remove("test.crt")
		if err != nil {
			gl.Log("error", fmt.Sprintf("Failed to remove test.crt: %v", err))
		}
	}()

	_, _, _ = certService.GenSelfCert()
	certBytes, keyBytes, err := certService.GetCertAndKeyFromFile()
	require.NoError(t, err)
	require.NotEmpty(t, certBytes)
	require.NotEmpty(t, keyBytes)
}

func TestGetCertAndKeyFromFileWithMissingFilesReturnsError(t *testing.T) {
	certService := crt.NewCertServiceType("missing.key", "missing.crt")
	_, _, err := certService.GetCertAndKeyFromFile()
	require.Error(t, err)
}
