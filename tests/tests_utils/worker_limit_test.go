// Pacote de teste externo para utils.
package utils_test

import (
	"errors"
	"testing"
	utils "github.com/kubex-ecosystem/gobe/internal/utils"
)

func TestValidateWorkerLimit(t *testing.T) {
	t.Run("Aceita inteiro não-negativo", func(t *testing.T) {
		if err := utils.ValidateWorkerLimit(5); err != nil {
			t.Fatalf("esperava nil, obteve erro: %v", err)
		}
	})
	
	t.Run("Rejeita negativo", func(t *testing.T) {
		if err := utils.ValidateWorkerLimit(-1); err == nil {
			t.Fatalf("esperava erro para negativo, obteve nil")
		}
	})

	t.Run("Rejeita tipo inválido", func(t *testing.T) {
		err := utils.ValidateWorkerLimit("cinco")
		if err == nil {
			t.Fatalf("esperava erro de tipo inválido, obteve nil")
		}
		if !errors.Is(err, err) { // apenas validar que veio um erro
			// no-op: mantemos a verificação para evitar lints
		}
	})
}
