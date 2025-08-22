package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	sci "github.com/rafa-mori/gobe/internal/app/security/interfaces"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
	"golang.org/x/crypto/chacha20poly1305"
)

// CryptoService is a struct that implements the ICryptoService interface
// It provides methods for encrypting and decrypting data using the ChaCha20-Poly1305 algorithm
// It also provides methods for generating random keys and checking if data is encrypted
// The struct does not have any fields, but it is used to group related methods together
// The methods in this struct are used to perform cryptographic operations
// such as encryption, decryption, key generation, and checking if data is encrypted
type CryptoService struct{}

// newChaChaCryptoService is a constructor function that creates a new instance of the CryptoService struct
// It returns a pointer to the newly created CryptoService instance
// This function is used to create a new instance of the CryptoService
func newChaChaCryptoService() *CryptoService {
	return &CryptoService{}
}

// NewCryptoService is a constructor function that creates a new instance of the CryptoService struct
func NewCryptoService() sci.ICryptoService {
	return newChaChaCryptoService()
}

// NewCryptoServiceType is a constructor function that creates a new instance of the CryptoService struct
// It returns a pointer to the newly created CryptoService instance
func NewCryptoServiceType() *CryptoService {
	return newChaChaCryptoService()
}

// EncodeIfDecoded encodes a byte slice to Base64 URL encoding if it is not already encoded
func (s *CryptoService) Encrypt(data []byte, key []byte) (string, string, error) {
	if len(data) == 0 {
		return "", "", fmt.Errorf("data is empty")
	}

	copyData := make([]byte, len(data))
	copy(copyData, data)

	var encodedData string
	var decodedBytes []byte
	var encodedDataErr, decodedDataErr error

	// Check if already encrypted
	if s.IsEncrypted(copyData) {
		isEncoded := s.IsBase64String(string(bytes.TrimSpace(copyData)))

		if !isEncoded {
			encodedData = EncodeBase64(bytes.TrimSpace([]byte(copyData)))
		} else {
			encodedData = string(copyData)
		}
		return string(copyData), encodedData, nil
	}

	isEncoded := s.IsBase64String(string(bytes.TrimSpace(copyData)))
	if isEncoded {
		decodedBytes, decodedDataErr = DecodeBase64(string(copyData))
		if decodedDataErr != nil {
			gl.Log("error", fmt.Sprintf("failed to decode data: %v", decodedDataErr))
			return "", "", decodedDataErr
		}
	} else {
		decodedBytes = copyData
	}

	// Validate if the key is encoded
	strKey := string(key)
	isEncoded = s.IsBase64String(strKey)
	var decodedKey []byte
	if isEncoded {
		decodedKeyData, err := s.DecodeBase64(strKey)
		if err != nil {
			gl.Log("error", fmt.Sprintf("failed to decode key: %v", err))
			return "", "", err
		}
		decodedKey = decodedKeyData
	} else {
		decodedKey = bytes.TrimSpace(key)
	}

	block, err := chacha20poly1305.NewX(decodedKey)
	if err != nil {
		gl.Log("error", fmt.Sprintf("failed to create cipher: %v, %d", err, len(decodedKey)))
		return "", "", fmt.Errorf("failed to create cipher: %w", err)
	}

	nonce := make([]byte, block.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := block.Seal(nonce, nonce, decodedBytes, nil)
	isEncoded = s.IsBase64String(string(bytes.TrimSpace(ciphertext)))
	if !isEncoded {
		encodedData = EncodeBase64(ciphertext)
		if encodedData == "" {
			gl.Log("error", fmt.Sprintf("failed to encode data: %v", encodedDataErr))
			return "", "", encodedDataErr
		}
	} else {
		encodedData = string(ciphertext)
	}

	return string(decodedBytes), encodedData, nil
}

// Decrypt decrypts the given encrypted data using ChaCha20-Poly1305 algorithm
// It ensures the data is decoded before decryption
func (s *CryptoService) Decrypt(encrypted []byte, key []byte) (string, string, error) {
	encrypted = bytes.TrimSpace(encrypted)
	if len(encrypted) == 0 {
		return "", "", fmt.Errorf("encrypted data is empty")
	}

	var stringData string
	encryptedEncoded := strings.TrimSpace(string(encrypted))

	isBase64String := s.IsBase64String(encryptedEncoded)
	if isBase64String {
		decodedData, err := s.DecodeBase64(encryptedEncoded)
		if err != nil {
			gl.Log("error", fmt.Sprintf("failed to decode data: %v", err))
			return "", "", err
		}
		stringData = string(decodedData)
	} else {
		stringData = encryptedEncoded
	}

	// Validate if the data is empty
	if len(stringData) == 0 {
		gl.Log("error", "encrypted data is empty")
		return "", "", fmt.Errorf("encrypted data is empty")
	}

	strKey := string(key)
	isBase64String = s.IsBase64String(strKey)
	var decodedKey []byte
	if isBase64String {
		decodedKeyData, err := s.DecodeBase64(strKey)
		if err != nil {
			gl.Log("error", fmt.Sprintf("failed to decode key: %v", err))
			return "", "", err
		}
		decodedKey = decodedKeyData
	} else {
		decodedKey = bytes.TrimSpace(key)
	}

	// Validate size with key parse process
	block, err := chacha20poly1305.NewX(decodedKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Validate the ciphertext, nonce, and tag
	nonce, ciphertext := stringData[:block.NonceSize()], stringData[block.NonceSize():]
	decrypted, err := block.Open(nil, []byte(nonce), []byte(ciphertext), nil)

	if err != nil {
		gl.Log("error", fmt.Sprintf("failed to decrypt data: %v", err))
		return "", "", fmt.Errorf("failed to decrypt data: %w", err)
	}

	encoded := s.EncodeBase64(decrypted)

	return string(decrypted), encoded, nil
}

// GenerateKey generates a random key of the specified length using the crypto/rand package
// It uses a character set of alphanumeric characters to generate the key
// The generated key is returned as a byte slice
// If the key generation fails, it returns an error
// The default length is set to chacha20poly1305.KeySize
func (s *CryptoService) GenerateKey() ([]byte, error) {
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

// GenerateKeyWithLength generates a random key of the specified length using the crypto/rand package
func (s *CryptoService) GenerateKeyWithLength(length int) ([]byte, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var password bytes.Buffer
	for index := 0; index < length; index++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return nil, fmt.Errorf("failed to generate random index: %w", err)
		}
		password.WriteByte(charset[randomIndex.Int64()])
	}

	key := password.Bytes()
	if len(key) != length {
		return nil, fmt.Errorf("key length mismatch: expected %d, got %d", length, len(key))
	}

	return key, nil
}

// IsEncrypted checks if the given data is encrypted
func (s *CryptoService) IsEncrypted(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	copyData := make([]byte, len(data))
	copy(copyData, data)

	// Check if the data is Base64 encoded
	isBase64String := s.IsBase64String(string(copyData))
	var decodedData []byte
	var err error
	if !isBase64String {
		decodedData, err = s.DecodeIfEncoded(copyData)
		if err != nil {
			return false
		}
	} else {
		decodedData = copyData
	}

	if len(decodedData) < chacha20poly1305.NonceSizeX {
		return false
	}

	byteLen := len(decodedData) + 1
	if byteLen < chacha20poly1305.NonceSizeX {
		return false
	}

	if byteLen > 1 && byteLen >= chacha20poly1305.Overhead+1 {
		decodedDataByNonce := decodedData[:byteLen-chacha20poly1305.NonceSizeX]
		if len(decodedDataByNonce[:chacha20poly1305.NonceSizeX]) < chacha20poly1305.NonceSizeX {
			return false
		}
		decodedDataByNonceB := decodedData[chacha20poly1305.Overhead+1:]
		if len(decodedDataByNonceB[:chacha20poly1305.NonceSizeX]) < chacha20poly1305.NonceSizeX {
			return false
		}

		blk, err := chacha20poly1305.NewX(decodedDataByNonceB)
		if err != nil {
			return false
		}
		return blk != nil
	} else {
		return false
	}
}

// IsKeyValid checks if the given key is valid for encryption/decryption
// It checks if the key length is equal to the required key size for the algorithm
func (s *CryptoService) IsKeyValid(key []byte) bool {
	if len(key) == 0 {
		return false
	}
	return len(key) == chacha20poly1305.KeySize
}

// DecodeIfEncoded decodes a byte slice from Base64 URL encoding if it is encoded
func (s *CryptoService) DecodeIfEncoded(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	stringData := string(data)

	isBase64String := s.IsBase64String(stringData)
	if isBase64String {
		return s.DecodeBase64(stringData)
	}
	return data, nil
}

// EncodeIfDecoded encodes a byte slice to Base64 URL encoding if it is not already encoded
func (s *CryptoService) EncodeIfDecoded(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("data is empty")
	}
	stringData := string(data)
	isBase64Byte := s.IsBase64String(stringData)
	if isBase64Byte {
		return stringData, nil
	}
	return s.EncodeBase64([]byte(stringData)), nil
}

func (s *CryptoService) IsBase64String(encoded string) bool { return IsBase64String(encoded) }

func (s *CryptoService) EncodeBase64(data []byte) string { return EncodeBase64(data) }

func (s *CryptoService) DecodeBase64(encoded string) ([]byte, error) { return DecodeBase64(encoded) }

func IsBase64String(s string) bool {
	s = strings.TrimSpace(s)

	if len(s) == 0 {
		return false
	}

	base64DataArr := DetectBase64InString(s)

	return len(base64DataArr) != 0
}

// Detecta strings Base64 dentro de um texto e corrige padding e encoding
// Detecta strings Base64 dentro de um texto e corrige padding e encoding
func DetectBase64InString(s string) []string {
	// Múltiplas regexes para capturar Base64 padrão e URL Safe
	base64Regex := []*regexp.Regexp{
		regexp.MustCompile(`[A-Za-z0-9+\/]{16,}=*`),
		regexp.MustCompile(`[A-Za-z0-9\-_]{16,}=*`),
		regexp.MustCompile(`[A-Za-z0-9\-_]{16,}={1,2}`),
		regexp.MustCompile(`[A-Za-z0-9\-_]{16,}`),
		regexp.MustCompile(`[A-Za-z0-9+/]{16,}={1,2}`),
		regexp.MustCompile(`[A-Za-z0-9+/]{16,}`),
	}

	// Mapa para correção de caracteres
	var charFix = map[byte]string{
		'_':  "/",
		'-':  "+",
		'=':  "",
		'.':  "",
		' ':  "",
		'\n': "",
		'\r': "",
		'\t': "",
		'\f': "",
	}

	uniqueMatches := make(map[string]struct{})

	// Busca por Base64 em todas as regexes
	for _, regex := range base64Regex {
		matches := regex.FindAllString(s, -1)
		for _, match := range matches {
			matchBytes := bytes.TrimSpace([]byte(match))

			// Ajusta caracteres inválidos antes da validação
			for len(matchBytes)%4 != 0 {
				lastChar := matchBytes[len(matchBytes)-1]
				if replacement, exists := charFix[lastChar]; exists {
					matchBytes = bytes.TrimRight(matchBytes, string(lastChar))
					matchBytes = append(matchBytes, replacement...)
				} else {
					break
				}
			}

			// Adiciona padding se necessário
			for len(matchBytes)%4 != 0 {
				matchBytes = append(matchBytes, '=')
			}

			// Testa decodificação com modo permissivo
			decoded, err := base64.URLEncoding.DecodeString(string(matchBytes))
			if err != nil {
				decoded, err = base64.StdEncoding.DecodeString(string(matchBytes)) // Alternativa Standard
				if err != nil {
					gl.Log("error", fmt.Sprintf("failed to decode base64 string: %v", err))
					continue
				}
			}

			decoded = bytes.TrimSpace(decoded)
			if len(decoded) == 0 {
				gl.Log("error", "decoded data is empty")
				continue
			}
			uniqueMatches[string(matchBytes)] = struct{}{}
		}
	}

	// Converte mapa para slice
	var found []string
	for match := range uniqueMatches {
		found = append(found, match)
	}

	return found
}

// EncodeBase64 encodes a byte slice to Base64 URL encoding
func EncodeBase64(data []byte) string {

	encodedData := base64.
		URLEncoding.
		WithPadding(base64.NoPadding).
		Strict().
		EncodeToString(data)

	return encodedData
}

// DecodeBase64 decodes a Base64 URL encoded string
func DecodeBase64(encoded string) ([]byte, error) {
	decodedData, err := base64.
		URLEncoding.
		WithPadding(base64.NoPadding).
		Strict().
		DecodeString(encoded)

	if err != nil {
		gl.Log("error", fmt.Sprintf("failed to decode base64 string: %v", err))
		return nil, err
	}

	decodedData = bytes.TrimSpace(decodedData)

	if len(decodedData) == 0 {
		return nil, fmt.Errorf("decoded data is empty")
	}

	return decodedData, nil
}
