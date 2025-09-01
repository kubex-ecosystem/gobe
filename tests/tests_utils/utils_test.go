// O sufixo _test no nome do pacote indica que é um pacote de teste externo.
package utils_test

import (
	// Importamos o pacote 'utils' com um alias para evitar ambiguidade.
	"testing"

	utils "github.com/rafa-mori/gobe/internal/utils"
)

// Teste corrigido para a função IsBase64String.
func TestIsBase64String(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "String Base64 válida (HelloWorld)",
			input: "SGVsbG9Xb3JsZA==", // "HelloWorld" em Base64, sem espaços.
			want:  true,
		},
		{
			name:  "String Base64 válida sem padding (foo)",
			input: "Zm9v",
			want:  true,
		},
		{
			name:  "String Base64 válida sem padding (foobar)",
			input: "Zm9vYmFy",
			want:  true,
		},
		{
			name:  "String inválida com caracteres especiais",
			input: "SGVsbG9Xb3JsZA!==", // Contém '!' que é inválido.
			want:  false,
		},
		{
			name:  "String inválida com espaços",
			input: "invalid string",
			want:  false,
		},
		{
			name:  "String vazia",
			input: "",
			want:  true, // A regex atual corretamente considera uma string vazia como válida.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Chamamos a função do pacote importado.
			got := utils.IsBase64String(tt.input)
			if got != tt.want {
				t.Errorf("IsBase64String() got = %v, want %v", got, tt.want)
			}
		})
	}
}
