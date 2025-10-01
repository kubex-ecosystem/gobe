package router

import (
	"context"
	"fmt"
	"time"

	"github.com/kubex-ecosystem/gobe/internal/app/executor"
	"github.com/kubex-ecosystem/gobe/internal/app/screening"
	"github.com/kubex-ecosystem/gobe/internal/app/session"
)

type DiscordMsg struct {
	GuildID   string
	ChannelID string
	UserID    string
	Text      string
}

type Responder interface {
	Reply(msg string)
	Update(msg string)
}

type Handler struct {
	Store  session.Store
	Run    executor.Runner
	Engine *screening.Engine
	TTL    time.Duration
}

func (h Handler) OnMessage(ctx context.Context, m DiscordMsg, ui Responder) {
	st, _ := h.Store.Load(ctx, m.GuildID, m.ChannelID, m.UserID)
	if st == nil {
		// cria sessão básica
		st = h.newSession(m)
		_ = h.Store.Save(ctx, st, h.sessionTTL())
	}

	engine := h.Engine
	if engine == nil {
		engine = screening.NewEngine(screening.Config{})
	}

	decision := engine.Analyze(m.Text, screening.Context{
		LastBotState:    st.LastBotState,
		LastIntent:      screening.Intent(st.LastUserIntent),
		LastMessageHash: st.LastMessageHash,
		LastMessageUnix: st.LastMessageUnix,
		Now:             time.Now(),
	})

	if decision.Duplicate {
		ui.Reply(fmt.Sprintf("🔁 Mensagem repetida detectada. Continuo no passo %s (%d%%).", st.NextStep, st.ProgressPct))
		return
	}

	var (
		output string
		err    error
	)

	switch decision.Action {
	case screening.ActionReplyStatus:
		ui.Reply(fmt.Sprintf("⏳ Sessão %s — %d%%. Próximo: %s.", st.ID, st.ProgressPct, st.NextStep))
		output, err = h.Run.Step(ctx, st)
	case screening.ActionContinue:
		output, err = h.Run.Step(ctx, st)
		if err == nil {
			ui.Reply(fmt.Sprintf("▶️ Continuando: %s", st.NextStep))
		}
	case screening.ActionExecute:
		output, err = h.Run.Step(ctx, st)
		if err == nil {
			ui.Reply(fmt.Sprintf("✅ Executado: %s", st.NextStep))
		}
	case screening.ActionClarify:
		ui.Reply("Posso seguir com o processamento atual ou prefere que eu detalhe o que já fiz? Use `continuar` ou `status`.")
	case screening.ActionResetSession:
		st = h.newSession(m)
		_ = h.Store.Save(ctx, st, h.sessionTTL())
		ui.Reply("🔄 Nova sessão criada. Me diz o que devemos fazer agora.")
	case screening.ActionAbortSession:
		st.LastBotState = "IDLE"
		st.NextStep = ""
		st.ProgressPct = 0
		ui.Reply("⏹️ Sessão atual pausada. Use `continuar` para retomar ou `nova` para recomeçar.")
	case screening.ActionAcknowledge:
		ui.Reply("👍 Valeu! Continuo por aqui.")
	case screening.ActionIgnore:
		if st.LastBotState == "WORKING" || st.LastBotState == "PENDING" {
			ui.Reply(fmt.Sprintf("⏳ Sessão %s — %d%%. Continuo no %s.", st.ID, st.ProgressPct, st.NextStep))
		}
	case screening.ActionPrompt, screening.ActionNone:
		ui.Reply("Não captei. Quer que eu continue a sessão atual ou abrir outra? (`continuar` / `nova`)")
	case screening.ActionDuplicate:
		// já tratado anteriormente, mas mantemos por segurança
		ui.Reply(fmt.Sprintf("🔁 Ainda estou no passo %s (%d%%).", st.NextStep, st.ProgressPct))
	default:
		ui.Reply("Beleza, estou acompanhando. Se quiser algo é só avisar!")
	}

	if err != nil {
		ui.Reply("⚠️ Tive um erro ao executar seu pedido, mas já retomei a sessão. Tenta novamente em instantes.")
		return
	}

	if output != "" {
		ui.Update(fmt.Sprintf("🧩 %s", summarizeOneLine(output)))
	}

	st.LastUserIntent = string(decision.Intent)
	st.LastMessageHash = decision.Fingerprint
	st.LastMessageUnix = decision.ObservedAt.Unix()
	_ = h.Store.Save(ctx, st, h.sessionTTL())
}

func (h Handler) newSession(m DiscordMsg) *session.State {
	return &session.State{
		ID:           fmt.Sprintf("%s:%s:%s:%d", m.GuildID, m.ChannelID, m.UserID, time.Now().Unix()),
		GuildID:      m.GuildID,
		ChannelID:    m.ChannelID,
		UserID:       m.UserID,
		LastBotState: "IDLE",
		NextStep:     "",
		ProgressPct:  0,
	}
}

func (h Handler) sessionTTL() time.Duration {
	if h.TTL > 0 {
		return h.TTL
	}
	return 24 * time.Hour
}

func summarizeOneLine(s string) string {
	const n = 240
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
