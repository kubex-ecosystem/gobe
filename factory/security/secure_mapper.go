package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ut "github.com/rafa-mori/gdbase/utils"
	crp "github.com/rafa-mori/gobe/internal/app/security/crypto"
	krs "github.com/rafa-mori/gobe/internal/app/security/external"
	sci "github.com/rafa-mori/gobe/internal/app/security/interfaces"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
	ci "github.com/rafa-mori/gobe/internal/proto/interfaces"
	t "github.com/rafa-mori/gobe/internal/proto/types"
)

type ISecureMapper[T any] interface {
	Serialize(format string) ([]byte, error)
	Deserialize(encryptedData []byte, format string) (*T, error)
	WriteDataFile(format string) error
	ReadDataFile(format string) (*T, error)
	GetFilePath() string
	SetFilePath(filePath string)
	SetKey(name string, key []byte)
	LoadOrGenerateKey(name string) ([]byte, error)
}

type SecureMapper[T any] struct {
	*t.Reference
	object        ci.IProperty[T]
	cryptoService sci.ICryptoService
	keyring       sci.IKeyringService
	filePath      string
	key           []byte
}

func NewSecureMapper[T any](name string, mapperObject *T, key []byte, filePath string) *SecureMapper[T] {
	var err error
	cryptoService := crp.NewCryptoService()
	if key == nil {
		key, err = cryptoService.GenerateKey()
		if err != nil {
			gl.Log("fatal", fmt.Sprintf("Failed to generate key: %v", err))
		}
	}
	keyring := krs.NewKeyringService(name, strings.ToValidUTF8(string(key), ""))
	if err := keyring.StorePassword(string(key)); err != nil {
		gl.Log("fatal", fmt.Sprintf("Failed to store key: %v", err))
	}
	return &SecureMapper[T]{
		Reference:     t.NewReference(name).GetReference(),
		key:           key,
		keyring:       keyring,
		filePath:      filePath,
		cryptoService: cryptoService,
		object:        t.NewProperty[T](name, mapperObject, false, nil),
	}
}

func (s *SecureMapper[T]) Serialize(format string) (string, error) {
	value := s.object.GetValue()
	mapper := t.NewMapper[T](&value, s.filePath)
	data, err := mapper.Serialize(format)
	if err != nil {
		gl.Log("error", fmt.Sprintf("failed to serialize data: %v", err))
		return "", err
	}
	//encryptedData, encodedEncryptedData, err := s.cryptoService.Encrypt(data, s.key)
	encryptedData, _, err := s.cryptoService.Encrypt(data, s.key)
	if err != nil {
		gl.Log("error", fmt.Sprintf("failed to encrypt data: %v", err))
		return "", err
	}
	gl.Log("success", fmt.Sprintf("data encrypted successfully: %s", s.filePath))
	return encryptedData, nil
}

func (s *SecureMapper[T]) Deserialize(encryptedData []byte, format string) (*T, error) {
	var err error
	var decryptedData string
	if s.cryptoService.IsEncrypted(encryptedData) {
		//decryptedData, encodedDecryptedData, err = s.cryptoService.Decrypt(encryptedData, s.key)
		decryptedData, _, err = s.cryptoService.Decrypt(encryptedData, s.key)
		if err != nil {
			gl.Log("error", fmt.Sprintf("failed to decrypt data: %v", err))
			return nil, err
		}
	} else {
		gl.Log("debug", "data is not encrypted, skipping decryption")
		decryptedData = string(encryptedData)
	}
	if len(decryptedData) == 0 {
		gl.Log("error", "decrypted data is empty")
		return nil, fmt.Errorf("decrypted data is empty")
	}
	var data *T
	mapper := t.NewMapper[T](data, s.filePath)
	data, err = mapper.Deserialize([]byte(decryptedData), format)
	if err != nil {
		gl.Log("error", fmt.Sprintf("failed to deserialize data: %v", err))
		return nil, err
	}
	if data == nil {
		gl.Log("error", "deserialized data is nil")
		return nil, fmt.Errorf("deserialized data is nil")
	}
	return data, nil
}

func (s *SecureMapper[T]) WriteDataFile(format string) error {
	if s.filePath == "" {
		gl.Log("error", "file path is not initialized")
		return fmt.Errorf("file path is not initialized")
	}
	value := s.object.GetValue()

	t.NewMapper[T](&value, s.filePath).SerializeToFile(format)

	if _, statErr := os.Stat(s.filePath); os.IsNotExist(statErr) {
		gl.Log("error", fmt.Sprintf("failed to write data to file: %v", statErr))
		return fmt.Errorf("failed to write data to file: %v", statErr)
	}

	gl.Log("success", fmt.Sprintf("data written to file: %s", s.filePath))

	return nil
}

func (s *SecureMapper[T]) ReadDataFile(format string) (*T, error) {
	if s.filePath == "" {
		gl.Log("error", "file path is not initialized")
		return nil, fmt.Errorf("file path is not initialized")
	}
	if _, statErr := os.Stat(s.filePath); os.IsNotExist(statErr) {
		gl.Log("error", fmt.Sprintf("file does not exist: %v", statErr))
		return nil, fmt.Errorf("file does not exist: %v", statErr)
	}
	value := s.object.GetValue()
	if data, err := t.NewMapper[T](&value, s.filePath).DeserializeFromFile(format); err != nil {
		gl.Log("error", fmt.Sprintf("failed to read data from file: %v", err))
		return nil, fmt.Errorf("failed to read data from file: %v", err)
	} else {
		return data, nil
	}
}

func (s *SecureMapper[T]) GetFilePath() string {
	if s.filePath == "" {
		gl.Log("error", "file path is not initialized")
		return ""
	}
	return s.filePath
}

func (s *SecureMapper[T]) SetFilePath(filePath string) {
	if filePath == "" {
		gl.Log("error", "file path is empty")
		return
	}
	s.filePath = filePath
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			gl.Log("error", fmt.Sprintf("failed to create directory: %v", err))
		}
		if err := ut.EnsureFile(s.filePath, 0644, []string{}); err != nil {
			gl.Log("error", fmt.Sprintf("failed to create file: %v", err))
		}
	}
}

func (s *SecureMapper[T]) SetKey(name string, key []byte) {
	if key == nil {
		gl.Log("error", "key is nil")
		return
	}
	if name != s.Reference.Name {
		gl.Log("error", "keyring name does not match the mapper name")
		return
	}
	s.key = key
	err := s.keyring.StorePassword(string(key))
	if err != nil {
		gl.Log("error", fmt.Sprintf("failed to store key: %v", err))
	}
}

func (s *SecureMapper[T]) LoadOrGenerateKey(name string) ([]byte, error) {
	// Check if the keyring service is initialized
	if s.keyring == nil {
		gl.Log("error", "keyring service is not initialized")
		return nil, fmt.Errorf("keyring service is not initialized")
	}
	// Check if the name matches the mapper name
	if name != s.Reference.Name {
		gl.Log("error", "keyring name does not match the mapper name")
		return nil, fmt.Errorf("keyring name does not match the mapper name")
	}
	// Check if the keyring service has a stored key
	storedKey, err := s.keyring.RetrievePassword()
	if err != nil || storedKey == "" {
		newKey, genErr := s.cryptoService.GenerateKey()
		if genErr != nil {
			return nil, fmt.Errorf("erro ao gerar chave: %v", genErr)
		}
		s.key = newKey
		err = s.keyring.StorePassword(string(newKey))
		if err != nil {
			gl.Log("error", fmt.Sprintf("erro ao armazenar chave de criptografia: %v", err))
			return nil, fmt.Errorf("erro ao armazenar chave de criptografia: %v", err)
		}
	} else {
		s.key = []byte(storedKey)
	}
	return s.key, nil
}

func (s *SecureMapper[T]) GetValue() T { return s.object.GetValue() }

func (s *SecureMapper[T]) SetValue(value *T) { s.object.SetValue(value) }
