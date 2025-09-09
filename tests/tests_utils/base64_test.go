// O sufixo _test no nome do pacote indica que é um pacote de teste externo.
package utils_test

import (
	"testing"
	utils "github.com/kubex-ecosystem/gobe/internal/utils"
)

func TestIsBase64ByteSlice(t *testing.T) {
	tests := []struct{
		name string
		in   []byte
		want bool
	}{
		{"Válido com padding", []byte("SGVsbG8="), true},
		{"Válido sem padding", []byte("Zm9v"), true},
		{"Inválido caractere '!'", []byte("SGVsbG8!"), false},
		{"Vazio é válido pela regex", []byte(""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsBase64ByteSlice(tt.in); got != tt.want {
				t.Errorf("IsBase64ByteSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBase64ByteSliceString(t *testing.T) {
	tests := []struct{
		name string
		in   string
		want bool
	}{
		{"Válido (foobar)", "Zm9vYmFy", true},
		{"Inválido com espaço", "Zm 9v", false},
		{"Vazio é válido", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsBase64ByteSliceString(tt.in); got != tt.want {
				t.Errorf("IsBase64ByteSliceString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBase64ByteSliceStringWithPadding(t *testing.T) {
	tests := []struct{
		name string
		in   string
		want bool
	}{
		{"Válido com padding (Hello)", "SGVsbG8=", true},
		{"Inválido caractere fora da base64", "SGV$sbG8=", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsBase64ByteSliceStringWithPadding(tt.in); got != tt.want {
				t.Errorf("IsBase64ByteSliceStringWithPadding() = %v, want %v", got, tt.want)
			}
		})
	}
}
