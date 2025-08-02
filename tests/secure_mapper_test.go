package tests

import (
	fsc "github.com/rafa-mori/gobe/factory/security"
	//at "github.com/rafa-mori/gobe/internal/types"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCryptoService struct {
	mock.Mock
}

func (m *MockCryptoService) GenerateKeyWithLength(length int) ([]byte, error) {
	args := m.Called(length)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCryptoService) IsEncrypted(data []byte) bool {
	args := m.Called(data)
	return args.Bool(0)
}

func (m *MockCryptoService) Encrypt(data []byte, key []byte) ([]byte, error) {
	args := m.Called(data, key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCryptoService) Decrypt(encryptedData []byte, key []byte) ([]byte, error) {
	args := m.Called(encryptedData, key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCryptoService) GenerateKey() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

// MockKeyringService para simular armazenamento de chaves
type MockKeyringService struct {
	mock.Mock
}

func (m *MockKeyringService) StorePassword(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

func (m *MockKeyringService) RetrievePassword() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// Teste do SecureMapper usando mocks

func TestNewSecureMapperWithValidInputsInitializesSuccessfully(t *testing.T) {
	name := "test-mapper"
	mapperObject := "test-value"
	key := []byte("test-key")
	filePath := "test-path"

	secureMapper := fsc.NewSecureMapper(name, &mapperObject, key, filePath)

	loadedKey, err := secureMapper.LoadOrGenerateKey(name)
	if err != nil {
		t.Fatalf("Failed to load or generate key: %v", err)
	}

	require.NotNil(t, secureMapper)
	require.Equal(t, name, secureMapper.Reference.Name)
	require.Equal(t, key, loadedKey)
	require.Equal(t, filePath, secureMapper.GetFilePath())
	require.NotNil(t, secureMapper.GetValue())

	// This is not possible to test because these properties are not exported
	// or accessible from outside the struct
	// require.NotNil(t, secureMapper.cryptoService)
	// require.NotNil(t, secureMapper.keyring)

}

func TestNewSecureMapperWithNilKeyGeneratesKey(t *testing.T) {
	name := "test-mapper"
	mapperObject := "test-value"
	filePath := "test-path"

	secureMapper := fsc.NewSecureMapper(name, &mapperObject, nil, filePath)

	loadedKey, err := secureMapper.LoadOrGenerateKey(name)
	if err != nil {
		t.Fatalf("Failed to load or generate key: %v", err)
	}

	require.NotNil(t, secureMapper)
	require.NotNil(t, loadedKey)
	require.NotEmpty(t, loadedKey)
}

//func TestNewSecureMapperWithMocks(t *testing.T) {
//	// Criando mocks
//	mockCrypto := new(MockCryptoService)
//	mockKeyring := new(MockKeyringService)
//
//	// Definindo comportamento esperado dos mocks
//	mockCrypto.On("GenerateKey").Return([]byte("mocked-key"), nil)
//	mockKeyring.On("StorePassword", "mocked-key").Return(nil)
//	mockKeyring.On("RetrievePassword").Return("mocked-key", nil)
//
//	// Inicialização com os mocks
//	name := "test-mapper"
//	mapperObject := "test-value"
//	filePath := "test-path"
//
//	secureMapper := &SecureMapper[string]{
//		Reference:     at.NewReference(name).GetReference(),
//		key:           []byte("mocked-key"),
//		keyring:       mockKeyring,
//		filePath:      filePath,
//		cryptoService: mockCrypto,
//		object:        at.NewProperty[string](name, &mapperObject, false, nil),
//	}
//
//	loadedKey, err := secureMapper.LoadOrGenerateKey(name)
//	require.NoError(t, err)
//	require.Equal(t, []byte("mocked-key"), loadedKey)
//	require.Equal(t, name, secureMapper.Reference.Name)
//
//	mockKeyring.AssertExpectations(t)
//	mockCrypto.AssertExpectations(t)
//}
