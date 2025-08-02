// Package preferences provides the controller for managing user preferences operations.
package preferences

import (
	"net/http"

	models "github.com/rafa-mori/gdbase/factory/models/mcp"
	t "github.com/rafa-mori/gdbase/types"
	svc "github.com/rafa-mori/gobe/internal/services"
	gl "github.com/rafa-mori/gobe/logger"

	"github.com/gin-gonic/gin"
	"github.com/rafa-mori/gobe/internal/types"
	"gorm.io/gorm"
)

type PreferencesController struct {
	preferencesService svc.PreferencesService
	APIWrapper         *types.APIWrapper[svc.PreferencesModel]
}

func NewPreferencesController(db *gorm.DB) *PreferencesController {
	return &PreferencesController{
		preferencesService: svc.NewPreferencesService(models.NewPreferencesRepo(db)),
		APIWrapper:         types.NewApiWrapper[svc.PreferencesModel](),
	}
}

func (pc *PreferencesController) GetAllPreferences(c *gin.Context) {
	preferences, err := pc.preferencesService.ListPreferences()
	if err != nil {
		gl.Log("error", "Failed to get all preferences", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get preferences"})
		return
	}
	c.JSON(http.StatusOK, preferences)
}

func (pc *PreferencesController) GetPreferencesByID(c *gin.Context) {
	id := c.Param("id")
	preferences, err := pc.preferencesService.GetPreferencesByID(id)
	if err != nil {
		gl.Log("error", "Failed to get preferences by ID", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Preferences not found"})
		return
	}
	c.JSON(http.StatusOK, preferences)
}

func (pc *PreferencesController) CreatePreferences(c *gin.Context) {
	var preferencesRequest svc.PreferencesModel

	if err := c.ShouldBindJSON(&preferencesRequest); err != nil {
		gl.Log("error", "Failed to bind preferences request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	createdPreferences, err := pc.preferencesService.CreatePreferences(preferencesRequest)
	if err != nil {
		gl.Log("error", "Failed to create preferences", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create preferences"})
		return
	}
	c.JSON(http.StatusCreated, createdPreferences)
}

func (pc *PreferencesController) UpdatePreferences(c *gin.Context) {
	id := c.Param("id")
	var preferencesRequest svc.PreferencesModel
	if err := c.ShouldBindJSON(&preferencesRequest); err != nil {
		gl.Log("error", "Failed to bind preferences update request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	preferencesRequest.SetID(id)
	updatedPreferences, err := pc.preferencesService.UpdatePreferences(preferencesRequest)
	if err != nil {
		gl.Log("error", "Failed to update preferences", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
		return
	}
	c.JSON(http.StatusOK, updatedPreferences)
}

func (pc *PreferencesController) DeletePreferences(c *gin.Context) {
	id := c.Param("id")
	err := pc.preferencesService.DeletePreferences(id)
	if err != nil {
		gl.Log("error", "Failed to delete preferences", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete preferences"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Preferences deleted successfully"})
}

func (pc *PreferencesController) GetPreferencesByScope(c *gin.Context) {
	scope := c.Param("scope")
	preferences, err := pc.preferencesService.GetPreferencesByScope(scope)
	if err != nil {
		gl.Log("error", "Failed to get preferences by scope", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Preferences not found"})
		return
	}
	c.JSON(http.StatusOK, preferences)
}

func (pc *PreferencesController) GetPreferencesByUserID(c *gin.Context) {
	userID := c.Param("userID")
	preferences, err := pc.preferencesService.GetPreferencesByUserID(userID)
	if err != nil {
		gl.Log("error", "Failed to get preferences by user ID", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get preferences"})
		return
	}
	c.JSON(http.StatusOK, preferences)
}

func (pc *PreferencesController) UpsertPreferencesByScope(c *gin.Context) {
	scope := c.Param("scope")

	var requestBody struct {
		Config t.JsonB `json:"config" binding:"required"`
		UserID string  `json:"user_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		gl.Log("error", "Failed to bind upsert preferences request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	preferences, err := pc.preferencesService.UpsertPreferencesByScope(scope, requestBody.Config, requestBody.UserID)
	if err != nil {
		gl.Log("error", "Failed to upsert preferences by scope", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upsert preferences"})
		return
	}
	c.JSON(http.StatusOK, preferences)
}
