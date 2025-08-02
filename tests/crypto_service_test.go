package tests

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"

	crp "github.com/rafa-mori/gobe/internal/security/crypto"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/chacha20poly1305"
)

func TestEncryptWithValidDataReturnsEncryptedData(t *testing.T) {
	service := crp.NewCryptoServiceType()
	key, _ := service.GenerateKey()
	data := []byte("test data")

	encrypted, _, err := service.Encrypt(data, key)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)
}

func TestEncryptWithEmptyDataReturnsError(t *testing.T) {
	service := crp.NewCryptoServiceType()
	key, _ := service.GenerateKey()

	encrypted, _, err := service.Encrypt([]byte{}, key)
	require.Error(t, err)
	require.Nil(t, encrypted)
}

func TestDecryptWithValidEncryptedDataReturnsOriginalData(t *testing.T) {
	service := crp.NewCryptoServiceType()
	key, _ := service.GenerateKey()
	data := []byte("test data")

	encrypted, _, _ := service.Encrypt(data, key)
	decrypted, _, err := service.Decrypt([]byte(encrypted), key)
	require.NoError(t, err)
	require.Equal(t, data, decrypted)
}

func TestDecryptWithInvalidKeyReturnsError(t *testing.T) {
	service := crp.NewCryptoServiceType()
	key, _ := service.GenerateKey()
	invalidKey, _ := service.GenerateKey()
	data := []byte("test data")

	encrypted, _, _ := service.Encrypt(data, key)
	decrypted, _, err := service.Decrypt([]byte(encrypted), invalidKey)
	require.Error(t, err)
	require.Nil(t, decrypted)
}

func TestDecryptWithEmptyDataReturnsError(t *testing.T) {
	service := crp.NewCryptoServiceType()
	key, _ := service.GenerateKey()

	decrypted, _, err := service.Decrypt([]byte{}, key)
	require.Error(t, err)
	require.Nil(t, decrypted)
}

func TestGenerateKeyReturnsValidKey(t *testing.T) {
	service := crp.NewCryptoServiceType()

	key, err := service.GenerateKey()
	require.NoError(t, err)
	require.Len(t, key, chacha20poly1305.KeySize)
}

func TestGenerateKeyWithLengthReturnsKeyOfSpecifiedLength(t *testing.T) {
	service := crp.NewCryptoServiceType()
	length := 32

	key, err := service.GenerateKeyWithLength(length)
	require.NoError(t, err)
	require.Len(t, key, length)
}

func TestIsEncryptedWithEncryptedDataReturnsTrue(t *testing.T) {
	service := crp.NewCryptoServiceType()
	key, _ := service.GenerateKey()
	data := []byte("test data")

	encrypted, _, _ := service.Encrypt(data, key)
	// Simulate the encrypted data
	isEncrypted := service.IsEncrypted([]byte(encrypted))
	require.True(t, isEncrypted)
}

func TestIsEncryptedWithUnencryptedDataReturnsFalse(t *testing.T) {
	service := crp.NewCryptoServiceType()
	data := []byte("test data")

	isEncrypted := service.IsEncrypted(data)
	require.False(t, isEncrypted)
}

func TestDecodeIfEncodedWithBase64EncodedDataReturnsDecodedData(t *testing.T) {
	data := []byte("dGVzdCBkYXRh") // "test data" in Base64
	decoded, err := crp.NewCryptoService().DecodeIfEncoded(data)
	require.NoError(t, err)
	require.Equal(t, "test data", string(decoded))
}

func TestDecodeIfEncodedWithNonBase64DataReturnsOriginalData(t *testing.T) {
	data := []byte("test data")
	decoded, err := crp.NewCryptoService().DecodeIfEncoded(data)
	require.NoError(t, err)
	require.Equal(t, data, decoded)
}

func TestEncodeIfDecodedWithNonBase64DataReturnsEncodedData(t *testing.T) {
	data := []byte("test data")
	encoded, err := crp.NewCryptoService().EncodeIfDecoded(data)
	require.NoError(t, err)
	require.Equal(t, base64.URLEncoding.EncodeToString(data), string(encoded))
}

func TestEncodeIfDecodedWithBase64DataReturnsOriginalData(t *testing.T) {
	data := []byte("dGVzdCBkYXRh") // "test data" in Base64
	encoded, err := crp.NewCryptoService().EncodeIfDecoded(data)
	require.NoError(t, err)
	require.Equal(t, data, encoded)
}

func TestEncryptDecryptLargeData(t *testing.T) {
	service := crp.NewCryptoServiceType()
	key, _ := service.GenerateKey()
	largeData := make([]byte, 100)
	_, err := rand.Read(largeData)
	if err != nil {
		t.Fatalf("failed to read random data: %v", err)
		return
	}

	encodedLargeData, err := service.EncodeIfDecoded(largeData)
	require.NoError(t, err)
	require.NotEmpty(t, encodedLargeData)

	encrypted, encryptedEncoded, err := service.Encrypt(largeData, key)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)
	require.NotEmpty(t, encryptedEncoded)

	decrypted, decryptedEncoded, err := service.Decrypt([]byte(encrypted), key)
	require.NoError(t, err)
	require.NotEmpty(t, decrypted)
	require.NotEmpty(t, decryptedEncoded)
	require.Equal(t, largeData, decrypted)
}

func TestMultiLayerEncryption(t *testing.T) {
	service := crp.NewCryptoServiceType()
	key, _ := service.GenerateKey()
	data := []byte("test data")

	encryptedOnce, _, err := service.Encrypt(data, key)
	require.NoError(t, err)

	encryptedTwice, _, err := service.Encrypt([]byte(encryptedOnce), key)
	require.NoError(t, err)

	decryptedOnce, _, err := service.Decrypt([]byte(encryptedTwice), key)
	require.NoError(t, err)

	decryptedTwice, _, err := service.Decrypt([]byte(decryptedOnce), key)
	require.NoError(t, err)
	require.Equal(t, data, decryptedTwice)
}

func TestPreventDoubleEncryption(t *testing.T) {
	service := crp.NewCryptoServiceType()
	key, _ := service.GenerateKey()
	data := []byte("test data")

	encrypted, _, errA := service.Encrypt(data, key)
	doubleEncrypted, _, errB := service.Encrypt([]byte(encrypted), key)

	require.NoError(t, errA)
	require.NoError(t, errB)
	require.Equal(t, encrypted, doubleEncrypted)
}

func BenchmarkEncryption(b *testing.B) {
	service := crp.NewCryptoServiceType()
	key, _ := service.GenerateKey()
	data := make([]byte, 100000) // 100KB de dados
	_, err := rand.Read(data)
	if err != nil {
		b.Fatalf("failed to read random data: %v", err)
		return
	}

	for i := 0; i < b.N; i++ {
		_, _, _ = service.Encrypt(data, key)
	}
}

// TestEncodeBase62 verifica se a função EncodeBase62 está funcionando corretamente
func TestEncodeBase62(t *testing.T) {
	data := []byte("TesteUnitario")
	encoded := crp.EncodeBase64(data)

	if len(encoded) == 0 {
		t.Errorf("Falha ao codificar Base62, resultado vazio")
	}
}

// TestDecodeBase62 verifica se a função DecodeBase62 retorna os bytes corretos
func TestDecodeBase62(t *testing.T) {
	data := []byte("TesteUnitario")
	encoded := crp.EncodeBase64(data)
	decoded, err := crp.DecodeBase64(encoded)

	require.NoError(t, err)
	require.NotEmpty(t, decoded)
	require.Equal(t, data, decoded)
	require.Equal(t, strings.TrimSpace(string(data)), strings.TrimSpace(string(decoded)))
}

// TestEncryptDecrypt verifica se o fluxo de criptografia e descriptografia funciona corretamente
func TestEncryptDecrypt(t *testing.T) {
	key, err := crp.NewCryptoService().GenerateKey()
	if err != nil {
		t.Fatalf("Erro ao gerar chave: %v", err)
	}

	originalData := []byte("Dados Sigilosos")
	encryptedData, encryptedDataEncoded, err := crp.NewCryptoService().Encrypt(originalData, key)
	if err != nil {
		t.Fatalf("Erro na criptografia: %v", err)
	}
	require.NotEmpty(t, encryptedData)
	require.NotEmpty(t, encryptedDataEncoded)

	decryptedData, decryptedDataEncoded, err := crp.NewCryptoService().Decrypt([]byte(encryptedData), key)
	if err != nil {
		t.Fatalf("Erro na descriptografia: %v", err)
	}
	require.NotEmpty(t, decryptedData)
	require.NotEmpty(t, decryptedDataEncoded)

	if !bytes.Equal(originalData, []byte(decryptedData)) {
		t.Errorf("Dados não correspondem após descriptografia. Esperado: %s, Obtido: %s", originalData, decryptedData)
	}
}

// TestEncryptDecrypt verifica se o fluxo de criptografia e descriptografia funciona corretamente
func TestEncryptDecryptSimpleStringA(t *testing.T) {
	key, err := crp.NewCryptoService().GenerateKey()
	if err != nil {
		t.Fatalf("Erro ao gerar chave: %v", err)
	}

	originalData := []byte("Dados Sigilosos Para")
	originalDataEncoded, err := crp.NewCryptoService().EncodeIfDecoded(originalData)
	if err != nil {
		t.Fatalf("Erro ao codificar Base62: %v", err)
	}
	require.NotEmpty(t, originalDataEncoded)

	encryptedData, encryptedDataEncoded, err := crp.NewCryptoService().Encrypt(originalData, key)
	if err != nil {
		t.Fatalf("Erro na criptografia: %v", err)
	}
	require.NotEmpty(t, encryptedData)
	require.NotEmpty(t, encryptedDataEncoded)

	decryptedData, decryptedDataEncoded, err := crp.NewCryptoService().Decrypt([]byte(encryptedData), key)
	if err != nil {
		t.Fatalf("Erro na descriptografia: %v", err)
	}
	require.NotEmpty(t, decryptedData)
	require.NotEmpty(t, decryptedDataEncoded)

	if !bytes.Equal(originalData, []byte(decryptedData)) {
		t.Errorf("Dados não correspondem após descriptografia. Esperado: %s, Obtido: %s", originalData, decryptedData)
	}
	if !bytes.Equal([]byte(originalDataEncoded), []byte(decryptedDataEncoded)) {
		t.Errorf("Dados codificados não correspondem após descriptografia. Esperado: %s, Obtido: %s", originalDataEncoded, decryptedDataEncoded)
	}
}

// TestEncryptDecrypt verifica se o fluxo de criptografia e descriptografia funciona corretamente
func TestEncryptDecryptComplexStringA(t *testing.T) {
	key, err := crp.NewCryptoService().GenerateKey()
	if err != nil {
		t.Fatalf("Erro ao gerar chave: %v", err)
	}

	originalData := []byte("Dados Sigilosos Para. Teste de criptografia com dados maiores e mais complexos! Será que funciona? Vamos testar com mais dados e ver se tudo está certo.")
	encryptedData, _, err := crp.NewCryptoService().Encrypt(originalData, key)
	if err != nil {
		t.Fatalf("Erro na criptografia: %v", err)
	}

	decryptedData, _, err := crp.NewCryptoService().Decrypt([]byte(encryptedData), key)
	if err != nil {
		t.Fatalf("Erro na descriptografia: %v", err)
	}

	if !bytes.Equal(originalData, []byte(decryptedData)) {
		t.Errorf("Dados não correspondem após descriptografia. Esperado: %s, Obtido: %s", originalData, decryptedData)
	}
}

// TestGenerateKey verifica se a chave gerada tem o tamanho correto
func TestGenerateKey(t *testing.T) {
	key, err := crp.NewCryptoService().GenerateKey()
	if err != nil {
		t.Fatalf("Erro ao gerar chave: %v", err)
	}

	if len(key) != chacha20poly1305.KeySize {
		t.Errorf("Tamanho da chave incorreto. Esperado: %d, Obtido: %d", chacha20poly1305.KeySize, len(key))
	}
}
