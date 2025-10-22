package testsintegration

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/kubex-ecosystem/gobe/internal/app/controllers/registration"
// 	reg_svc "github.com/kubex-ecosystem/gobe/internal/services/registration"
// )

// // MockRegistrationService is a mock for testing the controller.
// type MockRegistrationService struct {
// 	InitiateRegistrationFunc func(ctx context.Context, name, email, password string) error
// 	CompleteRegistrationFunc func(ctx context.Context, token string) error
// }

// func (m *MockRegistrationService) InitiateRegistration(ctx context.Context, name, email, password string) error {
// 	if m.InitiateRegistrationFunc != nil {
// 		return m.InitiateRegistrationFunc(ctx, name, email, password)
// 	}
// 	return nil
// }

// func (m *MockRegistrationService) CompleteRegistration(ctx context.Context, token string) error {
// 	if m.CompleteRegistrationFunc != nil {
// 		return m.CompleteRegistrationFunc(ctx, token)
// 	}
// 	return nil
// }

// func setupRouter(service reg_svc.RegistrationService) *gin.Engine {
// 	gin.SetMode(gin.TestMode)
// 	router := gin.New()
// 	ctrl := registration.NewRegistrationController(&service)
// 	api := router.Group("/api/v1")
// 	{
// 		api.POST("/register", ctrl.RegisterUser)
// 		api.GET("/verify-email", ctrl.VerifyEmail)
// 	}
// 	return router
// }

// func TestRegisterUserEndpoint(t *testing.T) {
// 	mockService := &MockRegistrationService{}
// 	router := setupRouter(mockService)

// 	t.Run("SuccessfulRegistration", func(t *testing.T) {
// 		mockService.InitiateRegistrationFunc = func(ctx context.Context, name, email, password string) error {
// 			return nil // Simulate success
// 		}

// 		payload := registration.RegisterUserRequest{Name: "Test", Email: "test@example.com", Password: "password123"}
// 		body, _ := json.Marshal(payload)

// 		req, _ := http.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(body))
// 		req.Header.Set("Content-Type", "application/json")

// 		w := httptest.NewRecorder()
// 		router.ServeHTTP(w, req)

// 		if w.Code != http.StatusOK {
// 			t.Errorf("Expected status OK, got %d", w.Code)
// 		}
// 	})

// 	t.Run("InvalidPayload", func(t *testing.T) {
// 		// Missing password
// 		payload := `{"name": "Test", "email": "test@example.com"}`

// 		req, _ := http.NewRequest("POST", "/api/v1/register", bytes.NewBufferString(payload))
// 		req.Header.Set("Content-Type", "application/json")

// 		w := httptest.NewRecorder()
// 		router.ServeHTTP(w, req)

// 		if w.Code != http.StatusBadRequest {
// 			t.Errorf("Expected status BadRequest, got %d", w.Code)
// 		}
// 	})
// }

// func TestVerifyEmailEndpoint(t *testing.T) {
// 	mockService := &MockRegistrationService{}
// 	router := setupRouter(mockService)

// 	t.Run("SuccessfulVerification", func(t *testing.T) {
// 		mockService.CompleteRegistrationFunc = func(ctx context.Context, token string) error {
// 			return nil // Simulate success
// 		}

// 		req, _ := http.NewRequest("GET", "/api/v1/verify-email?token=valid_token", nil)
// 		w := httptest.NewRecorder()
// 		router.ServeHTTP(w, req)

// 		if w.Code != http.StatusOK {
// 			t.Errorf("Expected status OK, got %d", w.Code)
// 		}
// 	})

// 	t.Run("MissingToken", func(t *testing.T) {
// 		req, _ := http.NewRequest("GET", "/api/v1/verify-email", nil)
// 		w := httptest.NewRecorder()
// 		router.ServeHTTP(w, req)

// 		if w.Code != http.StatusBadRequest {
// 			t.Errorf("Expected status BadRequest, got %d", w.Code)
// 		}
// 	})
// }
