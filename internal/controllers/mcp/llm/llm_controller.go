// Package llm provides the controller for managing LLM (Large Language Model) operations.
package llm

import (
	"net/http"

	models "github.com/rafa-mori/gdbase/factory/models/mcp"
	svc "github.com/rafa-mori/gobe/internal/services"
	gl "github.com/rafa-mori/gobe/logger"

	"github.com/gin-gonic/gin"
	"github.com/rafa-mori/gobe/internal/types"
	"gorm.io/gorm"
)

type LLMController struct {
	llmService svc.LLMService
	ApiWrapper *types.APIWrapper[svc.LLMModel]
}

func NewLLMController(db *gorm.DB) *LLMController {
	return &LLMController{
		llmService: svc.NewLLMService(models.NewLLMRepo(db)),
		ApiWrapper: types.NewApiWrapper[svc.LLMModel](),
	}
}

func (lc *LLMController) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/mcp/llm")
	{
		api.GET("/", lc.GetAllLLMModels)
		api.GET("/:id", lc.GetLLMModelByID)
		api.POST("/", lc.CreateLLMModel)
		api.PUT("/:id", lc.UpdateLLMModel)
		api.DELETE("/:id", lc.DeleteLLMModel)
		api.GET("/provider/:provider", lc.GetLLMModelsByProvider)
		api.GET("/provider/:provider/model/:model", lc.GetLLMModelByProviderAndModel)
		api.GET("/enabled", lc.GetEnabledLLMModels)
	}
}

func (lc *LLMController) GetAllLLMModels(c *gin.Context) {
	llmModels, err := lc.llmService.ListLLMModels()
	if err != nil {
		gl.Log("error", "Failed to get all LLM svc", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get LLM svc"})
		return
	}
	c.JSON(http.StatusOK, llmModels)
}

func (lc *LLMController) GetLLMModelByID(c *gin.Context) {
	id := c.Param("id")
	model, err := lc.llmService.GetLLMModelByID(id)
	if err != nil {
		gl.Log("error", "Failed to get LLM model by ID", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "LLM model not found"})
		return
	}
	c.JSON(http.StatusOK, model)
}

func (lc *LLMController) CreateLLMModel(c *gin.Context) {
	var modelRequest svc.LLMModel

	if err := c.ShouldBindJSON(&modelRequest); err != nil {
		gl.Log("error", "Failed to bind LLM model request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	createdModel, err := lc.llmService.CreateLLMModel(modelRequest)
	if err != nil {
		gl.Log("error", "Failed to create LLM model", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create LLM model"})
		return
	}
	c.JSON(http.StatusCreated, createdModel)
}

func (lc *LLMController) UpdateLLMModel(c *gin.Context) {
	id := c.Param("id")
	var modelRequest svc.LLMModel
	if err := c.ShouldBindJSON(&modelRequest); err != nil {
		gl.Log("error", "Failed to bind LLM model update request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	modelRequest.SetID(id)
	updatedModel, err := lc.llmService.UpdateLLMModel(modelRequest)
	if err != nil {
		gl.Log("error", "Failed to update LLM model", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update LLM model"})
		return
	}
	c.JSON(http.StatusOK, updatedModel)
}

func (lc *LLMController) DeleteLLMModel(c *gin.Context) {
	id := c.Param("id")
	err := lc.llmService.DeleteLLMModel(id)
	if err != nil {
		gl.Log("error", "Failed to delete LLM model", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete LLM model"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "LLM model deleted successfully"})
}

func (lc *LLMController) GetLLMModelsByProvider(c *gin.Context) {
	provider := c.Param("provider")
	llmModels, err := lc.llmService.GetLLMModelByProvider(provider)
	if err != nil {
		gl.Log("error", "Failed to get LLM svc by provider", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get LLM svc"})
		return
	}
	c.JSON(http.StatusOK, llmModels)
}

func (lc *LLMController) GetLLMModelByProviderAndModel(c *gin.Context) {
	provider := c.Param("provider")
	modelName := c.Param("model")
	model, err := lc.llmService.GetLLMModelByProviderAndModel(provider, modelName)
	if err != nil {
		gl.Log("error", "Failed to get LLM model by provider and model", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "LLM model not found"})
		return
	}
	c.JSON(http.StatusOK, model)
}

func (lc *LLMController) GetEnabledLLMModels(c *gin.Context) {
	llmModels, err := lc.llmService.GetEnabledLLMModels()
	if err != nil {
		gl.Log("error", "Failed to get enabled LLM svc", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get enabled LLM svc"})
		return
	}
	c.JSON(http.StatusOK, llmModels)
}
