package gateway

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LookAtniController holds placeholder endpoints for LookAtni automation hooks.
type LookAtniController struct{}

func NewLookAtniController() *LookAtniController { return &LookAtniController{} }

// Extract queues a LookAtni extraction job.
//
// @Summary     Extrair LookAtni
// @Description Enfileira uma extração de artefatos para processamento assíncrono.
// @Tags        gateway
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body map[string]interface{} true "Configuração da extração"
// @Success     202 {object} LookAtniActionResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Router      /api/v1/lookatni/extract [post]
func (lc *LookAtniController) Extract(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "invalid request body"})
		return
	}
	c.JSON(http.StatusAccepted, LookAtniActionResponse{
		Status:    "queued",
		Operation: "extract",
		Payload:   payload,
		Message:   "TODO: wire LookAtni extract pipeline",
		Timestamp: time.Now().UTC(),
	})
}

// Archive queues an archive operation for LookAtni artifacts.
//
// @Summary     Arquivar LookAtni
// @Description Agenda o arquivamento de artefatos processados pelo LookAtni.
// @Tags        gateway
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body map[string]interface{} true "Configuração do arquivamento"
// @Success     202 {object} LookAtniActionResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Router      /api/v1/lookatni/archive [post]
func (lc *LookAtniController) Archive(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "invalid request body"})
		return
	}
	c.JSON(http.StatusAccepted, LookAtniActionResponse{
		Status:    "queued",
		Operation: "archive",
		Payload:   payload,
		Message:   "TODO: connect LookAtni archive endpoint",
		Timestamp: time.Now().UTC(),
	})
}

// Download issues a temporary URL to fetch LookAtni artifacts.
//
// @Summary     Baixar ativo LookAtni
// @Description Retorna URL temporária para download do artefato processado.
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Identificador do recurso"
// @Success     200 {object} LookAtniDownloadResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/lookatni/download/{id} [get]
func (lc *LookAtniController) Download(c *gin.Context) {
	resourceID := strings.TrimSpace(c.Param("id"))
	if resourceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "missing resource id"})
		return
	}
	c.JSON(http.StatusOK, LookAtniDownloadResponse{
		DownloadURL: fmt.Sprintf("https://lookatni.local/%s", resourceID),
		ExpiresIn:   3600,
		Note:        "TODO: proxy real LookAtni artifact",
	})
}

// Projects lists available LookAtni projects.
//
// @Summary     Listar projetos LookAtni
// @Description Lista projetos cadastrados para automações LookAtni.
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} LookAtniProjectsResponse
// @Failure     401 {object} ErrorResponse
// @Router      /api/v1/lookatni/projects [get]
func (lc *LookAtniController) Projects(c *gin.Context) {
	c.JSON(http.StatusOK, LookAtniProjectsResponse{
		Projects: []map[string]interface{}{
			{
				"id":          "demo-project",
				"name":        "Demo Project",
				"description": "Placeholder LookAtni project",
			},
		},
		Version: "gateway-placeholder-1",
	})
}
