// Package registration handles user registration requests.
package registration

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/services/registration"
)

// RegistrationController handles the HTTP requests for user registration.
type RegistrationController struct {
	RegistrationService *registration.RegistrationService
}

// NewRegistrationController creates a new instance of the registration controller.
func NewRegistrationController(service *registration.RegistrationService) *RegistrationController {
	return &RegistrationController{RegistrationService: service}
}

// RegisterUserRequest defines the request body for user registration.
type RegisterUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// RegisterUser handles the user registration request.
func (ctrl *RegistrationController) RegisterUser(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ctrl.RegistrationService.InitiateRegistration(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate registration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration initiated. Please check your email to complete the process."})
}

// VerifyEmail handles the email verification request.
func (ctrl *RegistrationController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification token is required"})
		return
	}

	err := ctrl.RegistrationService.CompleteRegistration(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete registration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully. You can now log in."})
}
