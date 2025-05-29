package certificates

import (
	"crypto/rsa"
)

type PrivateKey struct{ *rsa.PrivateKey }
type PublicKey struct{ *rsa.PublicKey }
