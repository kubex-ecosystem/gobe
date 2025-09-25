package manage

import t "github.com/kubex-ecosystem/gobe/internal/contracts/types"

type (
	// ErrorResponse padroniza respostas de erro para os endpoints de gestão.
	ErrorResponse = t.ErrorResponse
)

// HealthResponse representa a resposta básica de healthcheck.
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// PingResponse representa a resposta do endpoint de ping.
type PingResponse struct {
	Message string `json:"message"`
}

// VersionResponse contém informações de versão.
type VersionResponse struct {
	Version string `json:"version"`
}

// ConfigResponse encapsula dados de configuração resumidos.
type ConfigResponse struct {
	Config map[string]any `json:"config"`
}

// ActionResponse descreve o resultado de ações de start/stop.
type ActionResponse struct {
	Message string `json:"message"`
}
