package interfaces

import "crypto/rsa"

type ICertManager interface {
	GenerateCertificate(certPath, keyPath string, password []byte) ([]byte, []byte, error)
	VerifyCert() error
	GetCertAndKeyFromFile() ([]byte, []byte, error)
}

type ICertService interface {
	GenerateCertificate(certPath, keyPath string, password []byte) ([]byte, []byte, error)
	GenSelfCert() ([]byte, []byte, error)
	DecryptPrivateKey(ciphertext []byte, password []byte) (*rsa.PrivateKey, error)
	VerifyCert() error
	GetCertAndKeyFromFile() ([]byte, []byte, error)
	GetPublicKey() (*rsa.PublicKey, error)
	GetPrivateKey() (*rsa.PrivateKey, error)
}
