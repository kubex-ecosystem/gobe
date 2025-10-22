// Package registration provides the routes for user registration.
package registration

import (
	"github.com/kubex-ecosystem/gobe/internal/app/controllers/registration"
	"github.com/kubex-ecosystem/gobe/internal/services/email"
	reg_svc "github.com/kubex-ecosystem/gobe/internal/services/registration"

	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	"github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
)

// NewRegistrationRoutes creates and returns the routes for user registration.
func NewRegistrationRoutes(rtr *ci.IRouter) map[string]ci.IRoute {
	// This is a placeholder. In a real application, these services would be
	// initialized properly, likely through dependency injection.
	// The bridge would be initialized in the main application setup.
	bridge := &gdbasez.Bridge{}

	// The email service would be initialized with the path to the config.
	emailService, _ := email.NewEmailService(".notes/email.yml")

	// The registration service depends on the bridge and email service.
	// The verification URL would come from a config file.
	verificationURL := "http://localhost:8080/api/v1/verify-email?token="
	regService := reg_svc.NewRegistrationService(bridge, emailService, verificationURL)

	// The controller uses the registration service.
	regController := registration.NewRegistrationController(regService)

	return map[string]ci.IRoute{
		"registerUser": proto.NewRoute("POST", "/api/v1/register", "application/json", regController.RegisterUser, nil, nil, nil, nil),
		"verifyEmail":  proto.NewRoute("GET", "/api/v1/verify-email", "application/json", regController.VerifyEmail, nil, nil, nil, nil),
	}
}
