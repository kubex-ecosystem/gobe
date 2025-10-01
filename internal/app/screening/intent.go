// Package screening provides utilities for intent detection in user messages.
package screening

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type Intent string

const (
	IntentQuestion  Intent = "QUESTION"
	IntentStatus    Intent = "STATUS_CHECK"
	IntentContinue  Intent = "CONTINUE"
	IntentAck       Intent = "ACK"
	IntentClarify   Intent = "CLARIFY"
	IntentCommand   Intent = "COMMAND"
	IntentSmalltalk Intent = "SMALLTALK"
	IntentReset     Intent = "RESET"
	IntentStop      Intent = "STOP"
	IntentUnknown   Intent = "UNKNOWN"
)

type Detected struct {
	Intent     Intent
	Confidence float32
	Reasons    []string
}

var rxQuestion = regexp.MustCompile(`[?？！]`)

// PT-BR / coloquial (sem "?")
var statusLex = []string{
	"e agora", "novidade", "novidades", "segue", "seguindo", "andamento",
	"conseguiu", "rolou", "rolando", "alguma coisa", "como ficou", "e ai", "e aí",
	"status", "e oq", "update", "sigo aguardando",
}
var continueLex = []string{"bora", "segue", "continua", "manda", "imenda", "vai", "vamo", "partiu"}
var ackLex = []string{"ok", "show", "blz", "beleza", "fechou", "top", "massa", "perfeito"}
var clarifyLex = []string{"nao entendi", "não entendi", "explica melhor", "como assim"}
var commandLex = []string{"gera", "cria", "analisa", "resume", "executa", "roda", "faz", "monta"}
var resetLex = []string{"nova", "novo", "reinicia", "reiniciar", "reset", "recomeca", "recomeça", "começar do zero"}
var stopLex = []string{"para", "pare", "cancel", "cancelar", "stop", "interrompe", "interromper"}

type Context struct {
	LastBotState    string // IDLE|WORKING|PENDING|DONE
	LastIntent      Intent
	LastMessageHash string
	LastMessageUnix int64
	Now             time.Time `json:"-"`
}

func norm(s string) string { return strings.TrimSpace(strings.ToLower(s)) }
func tokens(s string) int  { return len(strings.Fields(s)) }

func containsAny(hay string, list []string) bool {
	for _, w := range list {
		if strings.Contains(hay, w) {
			return true
		}
	}
	return false
}

func DetectIntent(msg string, ctx Context) Detected {
	m := norm(msg)
	tc := tokens(m)

	// vazio/risos/curtíssimo
	if tc == 0 || m == "kk" || m == "kkk" || m == "kkkk" || m == "rs" {
		return Detected{Intent: IntentSmalltalk, Confidence: 0.6, Reasons: []string{"tiny/emoji"}}
	}

	if utf8.RuneCountInString(m) <= 2 && !containsAny(m, ackLex) {
		return Detected{Intent: IntentSmalltalk, Confidence: 0.6, Reasons: []string{"tiny"}}
	}

	// pergunta explícita
	if rxQuestion.MatchString(msg) {
		return Detected{Intent: IntentQuestion, Confidence: 0.95, Reasons: []string{"punct:?"}}
	}

	// STATUS_CHECK por contexto + léxico curto (sem "?")
	if (ctx.LastBotState == "WORKING" || ctx.LastBotState == "PENDING") && tc <= 6 && containsAny(m, statusLex) {
		return Detected{Intent: IntentStatus, Confidence: 0.9, Reasons: []string{"context:working", "lex:status"}}
	}

	// CONTINUE/COMMAND/ACK/CLARIFY
	switch {
	case containsAny(m, continueLex):
		return Detected{Intent: IntentContinue, Confidence: 0.8, Reasons: []string{"lex:continue"}}
	case containsAny(m, stopLex):
		return Detected{Intent: IntentStop, Confidence: 0.85, Reasons: []string{"lex:stop"}}
	case containsAny(m, resetLex):
		return Detected{Intent: IntentReset, Confidence: 0.85, Reasons: []string{"lex:reset"}}
	case containsAny(m, commandLex):
		return Detected{Intent: IntentCommand, Confidence: 0.8, Reasons: []string{"lex:command"}}
	case containsAny(m, ackLex):
		return Detected{Intent: IntentAck, Confidence: 0.8, Reasons: []string{"lex:ack"}}
	case containsAny(m, clarifyLex):
		return Detected{Intent: IntentClarify, Confidence: 0.8, Reasons: []string{"lex:clarify"}}
	}

	// fallback heurístico: curto + léxico de status
	if tc <= 6 && containsAny(m, statusLex) {
		return Detected{Intent: IntentStatus, Confidence: 0.7, Reasons: []string{"short+status"}}
	}

	return Detected{Intent: IntentUnknown, Confidence: 0.3, Reasons: []string{"default"}}
}
