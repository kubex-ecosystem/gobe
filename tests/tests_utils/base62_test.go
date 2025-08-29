// Pacote de teste externo para utils.
package utils_test

import (
	"testing"
	utils "github.com/rafa-mori/gobe/internal/utils"
)

func TestIsBase62String(t *testing.T) {
	cases := []struct{
		name string
		in   string
		want bool
	}{
		{"Começa com letra e underscore permitido", "abc_123", true},
		{"Começa com underscore", "_abc123", true},
		{"Começa com dígito deve falhar", "1abc", false},
		{"Caracter inválido '-'", "ab-c", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := utils.IsBase62String(c.in); got != c.want {
				t.Errorf("IsBase62String(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

func TestIsBase62ByteSlice(t *testing.T) {
	if got := utils.IsBase62ByteSlice([]byte("abc_123")); !got {
		t.Errorf("IsBase62ByteSlice(abc_123) = %v, want true", got)
	}
	if got := utils.IsBase62ByteSlice([]byte("1abc")); got {
		t.Errorf("IsBase62ByteSlice(1abc) = %v, want false (regex permite, mas string não)", got)
	}
}
