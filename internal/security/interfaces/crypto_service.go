package interfaces

type ICryptoService interface {
	Encrypt([]byte, []byte) (string, string, error)
	Decrypt([]byte, []byte) (string, string, error)

	GenerateKey() ([]byte, error)
	GenerateKeyWithLength(int) ([]byte, error)

	EncodeIfDecoded([]byte) (string, error)
	DecodeIfEncoded([]byte) ([]byte, error)
	EncodeBase64([]byte) string
	DecodeBase64(string) ([]byte, error)

	IsBase64String(string) bool
	IsKeyValid([]byte) bool
	IsEncrypted([]byte) bool
}
