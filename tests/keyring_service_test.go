package tests

import (
	"testing"

	krs "github.com/rafa-mori/gobe/internal/security/external"
	"github.com/stretchr/testify/require"
)

var (
	keyringService = ""
	keyringName    = ""
)

func TestStorePasswordStoresPasswordSuccessfully(t *testing.T) {
	certService := krs.NewKeyringService(keyringName, keyringService)
	err := certService.StorePassword("test-password")
	require.NoError(t, err)
}

func TestRetrievePasswordReturnsStoredPassword(t *testing.T) {
	certService := krs.NewKeyringService(keyringName, keyringService)
	_ = certService.StorePassword("test-password")
	password, err := certService.RetrievePassword()
	require.NoError(t, err)
	require.Equal(t, "test-password", password)
}
