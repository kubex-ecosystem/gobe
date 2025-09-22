// Package contacts provides the ContactController for handling contact form submissions.
package contacts

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
)

type SMTPConfig struct {
	Host string
	Port string
	User string
	Pass string
}

type ContactController struct {
	queue      chan ci.ContactForm
	properties map[string]any
	APIWrapper *t.APIWrapper[ci.ContactForm]
}

type (
	// ErrorResponse padroniza respostas de erro nos endpoints de contato.
	ErrorResponse = t.ErrorResponse
	// MessageResponse padroniza mensagens simples de sucesso.
	MessageResponse = t.MessageResponse
)

func NewContactController(properties map[string]any) *ContactController {
	return &ContactController{
		queue:      make(chan ci.ContactForm, 100),
		properties: properties,
		APIWrapper: t.NewAPIWrapper[ci.ContactForm](),
	}
}

// HandleContact processa o formulário e encaminha para o canal configurado.
//
// @Summary     Processar contato
// @Description Valida o token secreto e dispara o fluxo de envio de mensagem.
// @Tags        contact
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body t.ContactForm true "Dados do formulário de contato"
// @Success     200 {object} MessageResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     403 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/contact/handle [post]
func (c *ContactController) HandleContact(ctx *gin.Context) {
	var form t.ContactForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "error processing data"})
		gl.Log("debug", fmt.Sprintf("Error processing data: %v", err.Error()))
		return
	}

	envT := c.properties["env"].(*t.Property[ci.IEnvironment])
	env := envT.GetValue()
	secretToken := env.Getenv("SECRET_TOKEN")

	if form.Token != secretToken {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Status: "error", Message: "invalid token"})
		gl.Log("warn", fmt.Sprintf("Invalid token: %s", form.Token))
		return
	}

	if err := sendEmailWithRetry(c, form, 2); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "error sending email"})
		gl.Log("debug", fmt.Sprintf("Error sending email: %v", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, MessageResponse{Status: "ok", Message: "Message sent successfully!"})
	gl.Log("success", "Message sent successfully!")
}

// GetContact retorna o status do fluxo de contato validando o token informado.
//
// @Summary     Consultar contato
// @Description Executa a mesma validação e envio do fluxo principal, retornando o resultado.
// @Tags        contact
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body t.ContactForm true "Dados do formulário de contato"
// @Success     200 {object} MessageResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     403 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/contact [get]
func (c *ContactController) GetContact(ctx *gin.Context) {
	var form t.ContactForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "error processing data"})
		gl.Log("debug", fmt.Sprintf("Error processing data: %v", err.Error()))
		return
	}

	envT := c.properties["env"].(*t.Property[ci.IEnvironment])
	env := envT.GetValue()
	secretToken := env.Getenv("SECRET_TOKEN")

	if form.Token != secretToken {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Status: "error", Message: "invalid token"})
		gl.Log("warn", fmt.Sprintf("Invalid token: %s", form.Token))
		return
	}

	if err := sendEmailWithRetry(c, form, 2); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "error sending email"})
		gl.Log("debug", fmt.Sprintf("Error sending email: %v", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, MessageResponse{Status: "ok", Message: "Message sent successfully!"})
	gl.Log("success", "Message sent successfully!")
}

// PostContact cria um novo contato seguindo as mesmas validações do fluxo padrão.
//
// @Summary     Enviar contato
// @Description Cria uma nova entrada de contato e dispara notificações conforme configuração.
// @Tags        contact
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body t.ContactForm true "Dados do formulário de contato"
// @Success     200 {object} MessageResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     403 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/contact [post]
func (c *ContactController) PostContact(ctx *gin.Context) {
	var form t.ContactForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "error processing data"})
		gl.Log("debug", fmt.Sprintf("Error processing data: %v", err.Error()))
		return
	}

	envT := c.properties["env"].(*t.Property[ci.IEnvironment])
	env := envT.GetValue()
	secretToken := env.Getenv("SECRET_TOKEN")

	if form.Token != secretToken {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Status: "error", Message: "invalid token"})
		gl.Log("warn", fmt.Sprintf("Invalid token: %s", form.Token))
		return
	}

	if err := sendEmailWithRetry(c, form, 2); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "error sending email"})
		gl.Log("debug", fmt.Sprintf("Error sending email: %v", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, MessageResponse{Status: "ok", Message: "Message sent successfully!"})
	gl.Log("success", "Message sent successfully!")
}

func (c *ContactController) GetContactForm(ctx *gin.Context) {
	var form t.ContactForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error processing data"})
		gl.Log("debug", fmt.Sprintf("Error processing data: %v", err.Error()))
		return
	}

	envT := c.properties["env"].(*t.Property[ci.IEnvironment])
	env := envT.GetValue()
	secretToken := env.Getenv("SECRET_TOKEN")

	if form.Token != secretToken {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
		gl.Log("warn", fmt.Sprintf("Invalid token: %s", form.Token))
		return
	}

	if err := sendEmailWithRetry(c, form, 2); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending email"})
		gl.Log("debug", fmt.Sprintf("Error sending email: %v", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Message sent successfully!"})
	gl.Log("success", "Message sent successfully!")
}

func (c *ContactController) GetContactFormByID(ctx *gin.Context) {
	var form t.ContactForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error processing data"})
		gl.Log("debug", fmt.Sprintf("Error processing data: %v", err.Error()))
		return
	}

	envT := c.properties["env"].(*t.Property[ci.IEnvironment])
	env := envT.GetValue()
	secretToken := env.Getenv("SECRET_TOKEN")

	if form.Token != secretToken {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
		gl.Log("warn", fmt.Sprintf("Invalid token: %s", form.Token))
		return
	}

	if err := sendEmailWithRetry(c, form, 2); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending email"})
		gl.Log("debug", fmt.Sprintf("Error sending email: %v", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Message sent successfully!"})
	gl.Log("success", "Message sent successfully!")
}
