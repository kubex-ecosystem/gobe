// Package certificates provides an interface for managing SSL/TLS certificates.
package certificates

type ICertManager interface {
	GenerateCertificate(certPath, keyPath string, password []byte) ([]byte, []byte, error)
	VerifyCert() error
	GetCertAndKeyFromFile() ([]byte, []byte, error)
}
