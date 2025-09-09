package external

import (
	"errors"
	"fmt"
	"os"

	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	"github.com/zalando/go-keyring"
)

type KeyringService struct {
	keyringService ci.IProperty[string]
	keyringName    ci.IProperty[string]
}

func newKeyringService(service, name string) *KeyringService {
	return &KeyringService{
		keyringService: t.NewProperty("keyringService", &service, false, nil),
		keyringName:    t.NewProperty("keyringName", &name, false, nil),
	}
}
func NewKeyringService(service, name string) sci.IKeyringService {
	return newKeyringService(service, name)
}
func NewKeyringServiceType(service, name string) *KeyringService {
	return newKeyringService(service, name)
}

func (k *KeyringService) StorePassword(password string) error {
	if password == "" {
		gl.Log("error", "key cannot be empty")
		return fmt.Errorf("key cannot be empty")
	}
	if err := keyring.Set(k.keyringService.GetValue(), k.keyringName.GetValue(), password); err != nil {
		return fmt.Errorf("error storing key: %v", err)
	}
	gl.Log("debug", fmt.Sprintf("key stored successfully: %s", k.keyringName.GetValue()))
	return nil
}
func (k *KeyringService) RetrievePassword() (string, error) {
	if password, err := keyring.Get(k.keyringService.GetValue(), k.keyringName.GetValue()); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", os.ErrNotExist
		}
		gl.Log("debug", fmt.Sprintf("error retrieving key: %v", err))
		return "", fmt.Errorf("error retrieving key: %v", err)
	} else {
		return password, nil
	}
}
