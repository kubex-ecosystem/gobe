// Pacote de teste externo para utils.
package utils_test

import (
	"testing"
	utils "github.com/rafa-mori/gobe/internal/utils"
)

func TestIsURLEncodeValidators(t *testing.T) {
	cases := []struct{
		name string
		inS  string
		want bool
	}{
		{"Permitidos básicos", "abc-XYZ_123.%", true},
		{"Caracter inválido '?'", "abc?123", false},
		{"Espaço é inválido", "abc 123", false},
	}
	for _, c := range cases {
		t.Run("String/"+c.name, func(t *testing.T) {
			if got := utils.IsURLEncodeString(c.inS); got != c.want {
				t.Errorf("IsURLEncodeString() = %v, want %v", got, c.want)
			}
		})
		t.Run("Bytes/"+c.name, func(t *testing.T) {
			if got := utils.IsURLEncodeByteSlice([]byte(c.inS)); got != c.want {
				t.Errorf("IsURLEncodeByteSlice() = %v, want %v", got, c.want)
			}
		})
	}
}
