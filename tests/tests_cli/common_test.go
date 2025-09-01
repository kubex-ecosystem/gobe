// Pacote de teste externo para o pacote cli.
package cli_test

import (
	"os"
	"testing"
	cli "github.com/rafa-mori/gobe/cmd/cli"
)

func TestGetDescriptions_BannerAndDescription(t *testing.T) {
	orig := os.Args
	os.Args = []string{"gobe", "-h"}
	t.Cleanup(func(){ os.Args = orig })

	m := cli.GetDescriptions([]string{"descrição longa", "descrição curta"}, true)
	if m == nil {
		t.Fatalf("esperava map retornado, veio nil")
	}
	if m["description"] != "descrição longa" {
		t.Fatalf("esperava descrição longa com -h, veio %q", m["description"])
	}
	if m["banner"] == "" {
		t.Fatalf("banner não deveria ser vazio")
	}
}
