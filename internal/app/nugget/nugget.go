// Package nugget fornece funções para manipulação de strings.
package nugget

import "strings"

// Update mantém um resumo curto (bullets) — aqui heurístico leve.
// Você pode plugar um LLM barato depois. Mantém <= 2000 chars.
func Update(prev, delta string) string {
	const max = 2000
	lines := []string{}
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if !strings.HasPrefix(s, "- ") {
			s = "- " + s
		}
		lines = append(lines, s)
	}
	for _, l := range strings.Split(prev, "\n") {
		if strings.TrimSpace(l) != "" {
			add(l)
		}
	}
	for _, l := range strings.Split(delta, "\n") {
		if strings.TrimSpace(l) != "" {
			add(l)
		}
	}
	out := strings.Join(lines, "\n")
	if len(out) > max {
		out = out[:max]
	}
	return out
}
