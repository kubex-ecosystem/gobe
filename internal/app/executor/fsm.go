// Package executor fornece a implementação de um executor de tarefas baseado em máquina de estados finitos (FSM).
package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/kubex-ecosystem/gobe/internal/app/nugget"
	"github.com/kubex-ecosystem/gobe/internal/app/session"
)

type LLM interface {
	// Abstração do cliente (Files API + GenerateContent), já existente no seu projeto
	AnalyzeStep(ctx context.Context, step string, payload any) (string, error)
}

type Runner struct {
	Store session.Store
	LLM   LLM
}

func (r Runner) Step(ctx context.Context, s *session.State) (string, error) {
	// FSM simples: MAP:* → avança; REDUCE → finaliza; DONE → noop
	switch {
	case s.NextStep == "":
		s.NextStep = "MAP:chunk_000"
	case s.NextStep == "DONE":
		return "done", nil
	}

	out, err := r.LLM.AnalyzeStep(ctx, s.NextStep, nil)
	if err != nil {
		s.LastBotState = "PENDING"
		_ = r.Store.Save(ctx, s, 24*time.Hour)
		return "", err
	}

	// atualiza progresso/nugget
	s.ContextNugget = nugget.Update(s.ContextNugget, out)
	switch {
	case s.NextStep == "REDUCE":
		s.NextStep = "DONE"
		s.ProgressPct = 100
		s.LastBotState = "DONE"
	default:
		// exemplo de incremento: MAP:chunk_007 → MAP:chunk_008
		var idx int
		if _, err := fmt.Sscanf(s.NextStep, "MAP:chunk_%03d", &idx); err == nil {
			idx++
			s.NextStep = fmt.Sprintf("MAP:chunk_%03d", idx)
			if s.ProgressPct < 95 {
				s.ProgressPct += 3
			}
			s.LastBotState = "WORKING"
		} else {
			// quando terminar os MAPs, muda pra REDUCE
			s.NextStep = "REDUCE"
			s.LastBotState = "WORKING"
			if s.ProgressPct < 98 {
				s.ProgressPct = 98
			}
		}
	}

	_ = r.Store.Save(ctx, s, 24*time.Hour)
	return out, nil
}
