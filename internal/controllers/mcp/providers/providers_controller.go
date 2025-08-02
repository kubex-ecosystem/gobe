// Package providers provides the controller for managing user providers.
package providers

import (
	"net/http"

	models "github.com/rafa-mori/gdbase/factory/models/mcp"
	t "github.com/rafa-mori/gdbase/types"
	svc "github.com/rafa-mori/gobe/internal/services"
	gl "github.com/rafa-mori/gobe/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProvidersController struct {
	providersService svc.ProvidersService
}

func NewProvidersController(db *gorm.DB) *ProvidersController {
	return &ProvidersController{
		providersService: svc.NewProvidersService(models.NewProvidersRepo(db)),
	}
}

// GetAllProviders retrieves all providers
func (pc *ProvidersController) GetAllProviders(c *gin.Context) {
	providers, err := pc.providersService.ListProviders()
	if err != nil {
		gl.Log("error", "Failed to get providers", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get providers"})
		return
	}
	c.JSON(http.StatusOK, providers)
}

// GetProviderByID retrieves a provider by ID
func (pc *ProvidersController) GetProviderByID(c *gin.Context) {
	id := c.Param("id")
	provider, err := pc.providersService.GetProviderByID(id)
	if err != nil {
		gl.Log("error", "Failed to get provider by ID", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}
	c.JSON(http.StatusOK, provider)
}

// CreateProvider creates a new provider
func (pc *ProvidersController) CreateProvider(c *gin.Context) {
	var providerRequest struct {
		Provider   string  `json:"provider" binding:"required"`
		OrgOrGroup string  `json:"org_or_group" binding:"required"`
		Config     t.JsonB `json:"config,omitempty"`
	}

	if err := c.ShouldBindJSON(&providerRequest); err != nil {
		gl.Log("error", "Failed to bind provider request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create a new provider model
	newProvider := models.NewProvidersModel(
		providerRequest.Provider,
		providerRequest.OrgOrGroup,
		providerRequest.Config,
	)

	createdProvider, err := pc.providersService.CreateProvider(newProvider)
	if err != nil {
		gl.Log("error", "Failed to create provider", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create provider"})
		return
	}
	c.JSON(http.StatusCreated, createdProvider)
}

// UpdateProvider updates an existing provider
func (pc *ProvidersController) UpdateProvider(c *gin.Context) {
	id := c.Param("id")
	var providerRequest struct {
		Provider   string  `json:"provider"`
		OrgOrGroup string  `json:"org_or_group"`
		Config     t.JsonB `json:"config"`
	}

	if err := c.ShouldBindJSON(&providerRequest); err != nil {
		gl.Log("error", "Failed to bind provider update request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get existing provider
	existingProvider, err := pc.providersService.GetProviderByID(id)
	if err != nil {
		gl.Log("error", "Failed to get provider for update", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// Update fields
	if providerRequest.Provider != "" {
		existingProvider.SetProvider(providerRequest.Provider)
	}
	if providerRequest.OrgOrGroup != "" {
		existingProvider.SetOrgOrGroup(providerRequest.OrgOrGroup)
	}
	if providerRequest.Config != nil {
		existingProvider.SetConfig(providerRequest.Config)
	}

	updatedProvider, err := pc.providersService.UpdateProvider(existingProvider)
	if err != nil {
		gl.Log("error", "Failed to update provider", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider"})
		return
	}
	c.JSON(http.StatusOK, updatedProvider)
}

// DeleteProvider deletes a provider by ID
func (pc *ProvidersController) DeleteProvider(c *gin.Context) {
	id := c.Param("id")

	if err := pc.providersService.DeleteProvider(id); err != nil {
		gl.Log("error", "Failed to delete provider", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete provider"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Provider deleted successfully"})
}

// GetProvidersByProvider retrieves providers by provider name
func (pc *ProvidersController) GetProvidersByProvider(c *gin.Context) {
	provider := c.Param("provider")
	providers, err := pc.providersService.GetProviderByName(provider)
	if err != nil {
		gl.Log("error", "Failed to get providers by provider", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get providers by provider"})
		return
	}
	c.JSON(http.StatusOK, providers)
}

// GetProvidersByOrgOrGroup retrieves providers by organization or group
func (pc *ProvidersController) GetProvidersByOrgOrGroup(c *gin.Context) {
	orgOrGroup := c.Param("org_or_group")
	providers, err := pc.providersService.GetProviderByOrgOrGroup(orgOrGroup)
	if err != nil {
		gl.Log("error", "Failed to get providers by org or group", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get providers by org or group"})
		return
	}
	c.JSON(http.StatusOK, providers)
}

// GetActiveProviders retrieves all active providers
func (pc *ProvidersController) GetActiveProviders(c *gin.Context) {
	// Como não existe um método específico para ativos, vamos retornar todos
	providers, err := pc.providersService.ListProviders()
	if err != nil {
		gl.Log("error", "Failed to get active providers", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active providers"})
		return
	}
	c.JSON(http.StatusOK, providers)
}

// UpsertProviderByNameAndOrg creates or updates a provider by name and org_or_group
func (pc *ProvidersController) UpsertProviderByNameAndOrg(c *gin.Context) {
	var providerRequest struct {
		Provider   string  `json:"provider" binding:"required"`
		OrgOrGroup string  `json:"org_or_group" binding:"required"`
		Config     t.JsonB `json:"config"`
	}

	if err := c.ShouldBindJSON(&providerRequest); err != nil {
		gl.Log("error", "Failed to bind provider upsert request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Use UpsertProviderByNameAndOrg com os parâmetros corretos
	result, err := pc.providersService.UpsertProviderByNameAndOrg(
		providerRequest.Provider,
		providerRequest.OrgOrGroup,
		providerRequest.Config,
		"admin", // userID temporário
	)
	if err != nil {
		gl.Log("error", "Failed to upsert provider", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upsert provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider": result,
	})
}
