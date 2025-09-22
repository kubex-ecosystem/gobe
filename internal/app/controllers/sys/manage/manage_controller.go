package manage

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ServerController agrupa endpoints de gestão da instância.
type ServerController struct{}

func NewServerController() *ServerController { return &ServerController{} }

// respondError devolve erro padronizado.
// Health verifica o estado básico da aplicação.
//
// @Summary     Healthcheck
// @Description Retorna status "healthy" para monitoramento. [Em desenvolvimento]
// @Tags        system beta
// @Produce     json
// @Success     200 {object} HealthResponse
// @Router      /health [get]
// @Router      /health [post]
func (sc *ServerController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Status: "ok", Message: "healthy"})
}

// Ping responde com "pong" para verificar conectividade.
//
// @Summary     Ping
// @Description Retorna mensagem de ping/pong para verificação de conectividade. [Em desenvolvimento]
// @Tags        system beta
// @Produce     json
// @Success     200 {object} PingResponse
// @Router      /ping [get]
// @Router      /ping [post]
func (sc *ServerController) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, PingResponse{Message: "pong"})
}

// Version retorna a versão atual da aplicação.
//
// @Summary     Versão
// @Description Informa a versão corrente da aplicação. [Em desenvolvimento]
// @Tags        system beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} VersionResponse
// @Failure     401 {object} ErrorResponse
// @Router      /version [get]
func (sc *ServerController) Version(c *gin.Context) {
	c.JSON(http.StatusOK, VersionResponse{Version: "v1.0.0"})
}

// Config retorna informações resumidas de configuração.
//
// @Summary     Configuração
// @Description Retorna configuração básica da aplicação. [Em desenvolvimento]
// @Tags        system beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} ConfigResponse
// @Failure     401 {object} ErrorResponse
// @Router      /api/v1/config [get]
func (sc *ServerController) Config(c *gin.Context) {
	c.JSON(http.StatusOK, ConfigResponse{Config: map[string]any{"config": "config"}})
}

// Start simula a inicialização do serviço.
//
// @Summary     Iniciar serviço
// @Description Dispara processo de inicialização do serviço GoBE. [Em desenvolvimento]
// @Tags        system beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} ActionResponse
// @Failure     401 {object} ErrorResponse
// @Router      /api/v1/start [post]
func (sc *ServerController) Start(c *gin.Context) {
	c.JSON(http.StatusOK, ActionResponse{Message: "gobe started successfully"})
}

// Stop simula a parada do serviço.
//
// @Summary     Parar serviço
// @Description Finaliza o serviço GoBE. [Em desenvolvimento]
// @Tags        system beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} ActionResponse
// @Failure     401 {object} ErrorResponse
// @Router      /api/v1/stop [post]
func (sc *ServerController) Stop(c *gin.Context) {
	c.JSON(http.StatusOK, ActionResponse{Message: "gobe stopped successfully"})
}
