package analyzer

import (
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	"github.com/kubex-ecosystem/gobe/internal/services/analyzer"
)

type ScorecardResponse = analyzer.ScorecardResponse
type RepositoryInfo = analyzer.RepositoryInfo
type DORAMetrics = analyzer.DORAMetrics
type CHIMetrics = analyzer.CHIMetrics
type AIMetrics = analyzer.AIMetrics
type ConfidenceMetrics = analyzer.ConfidenceMetrics

type (
	// ErrorResponse padroniza respostas de erro nos endpoints de contato.
	ErrorResponse = t.ErrorResponse
	// MessageResponse padroniza mensagens simples de sucesso.
	MessageResponse = t.MessageResponse
)
