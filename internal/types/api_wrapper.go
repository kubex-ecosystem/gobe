// Package types contém definições de tipos e estruturas para o wrapper de API
package types

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIResponse encapsulando respostas
type APIResponse struct {
	Status string                 `json:"status"`
	Hash   string                 `json:"hash,omitempty"`
	Msg    string                 `json:"msg,omitempty"`
	Filter map[string]interface{} `json:"filter,omitempty"`
	Data   interface{}            `json:"data,omitempty"`
}

func NewAPIResponse() *APIResponse {
	return &APIResponse{
		Status: "success",
		Hash:   "",
		Msg:    "",
		Filter: make(map[string]interface{}),
		Data:   nil,
	}
}

// APIRequest (futuro espaço para lógica extra)
type APIRequest struct{}

func NewAPIRequest() *APIRequest {
	return &APIRequest{}
}

// API Wrapper para gerenciar requisições e respostas de maneira padronizada
type APIWrapper[T any] struct{}

func NewApiWrapper[T any]() *APIWrapper[T] {
	return &APIWrapper[T]{}
}

// Gerencia requisições de forma genérica
func (w *APIWrapper[T]) HandleRequest(c *gin.Context, method string, endpoint string, payload interface{}) {
	switch method {
	case "GET":
		c.JSON(http.StatusOK, gin.H{"message": "GET request handled", "endpoint": endpoint})
	case "POST":
		c.JSON(http.StatusCreated, gin.H{"message": "POST request handled", "endpoint": endpoint, "payload": payload})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unsupported method", "method": method})
	}
}

// Middleware para interceptar e padronizar respostas
func (w *APIWrapper[T]) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Writer.Status() >= 400 {
			errMsg := c.Errors.ByType(gin.ErrorTypePrivate).String()
			if errMsg == "" {
				errMsg = "Erro desconhecido"
			}
			w.JSONResponseWithError(c, fmt.Errorf(errMsg))
		}
	}
}

// Enviar resposta padronizada
func (w *APIWrapper[T]) JSONResponse(c *gin.Context, status string, msg, hash string, data interface{}, filter map[string]interface{}, httpStatus int) {
	r := NewAPIResponse()
	r.Status = status
	r.Msg = msg
	r.Hash = hash
	r.Data = data
	r.Filter = filter

	c.JSON(httpStatus, r)
}

// JSONResponseWithError sends a JSON response to the client with an error message.
func (w *APIWrapper[T]) JSONResponseWithError(c *gin.Context, err error) {
	r := NewAPIResponse()
	r.Status = "error"
	r.Msg = err.Error()
	r.Hash = ""
	r.Data = nil
	r.Filter = make(map[string]interface{})

	c.JSON(http.StatusBadRequest, r)
}

// JSONResponseWithSuccess sends a JSON response to the client with a success message.
func (w *APIWrapper[T]) JSONResponseWithSuccess(c *gin.Context, msgKey, hash string, data interface{}) {
	//msg := translateMessage(msgKey) // Função fictícia para traduzir mensagens
	r := NewAPIResponse()
	r.Status = "success"
	r.Msg = msgKey
	r.Hash = hash
	r.Data = data
	r.Filter = make(map[string]interface{})

	c.JSON(http.StatusOK, r)
}

func (w *APIWrapper[T]) GetContext(c *gin.Context) (context.Context, error) {
	userId := c.GetHeader("X-User-ID")
	if userId == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	uuserID, err := uuid.Parse(userId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %s", err)
	}
	ctx := context.WithValue(c.Request.Context(), "userID", uuserID)

	cronID := c.Param("id")
	if cronID != "" {
		if cronID == "" {
			return nil, fmt.Errorf("cron job ID is required")
		}
		cronUUID, err := uuid.Parse(cronID)
		if err != nil {
			return nil, fmt.Errorf("invalid cron job ID: %s", err)
		}
		ctx = context.WithValue(ctx, "cronID", cronUUID)
	}
	return ctx, nil
}
