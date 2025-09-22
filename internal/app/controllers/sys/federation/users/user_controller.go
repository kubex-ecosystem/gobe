// Package users provides the UserController for managing user-related operations.
package users

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	user "github.com/kubex-ecosystem/gdbase/factory/models"
	sau "github.com/kubex-ecosystem/gobe/internal/app/security/authentication"
	crt "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
	cm "github.com/kubex-ecosystem/gobe/internal/commons"

	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/contracts/types"
	"gorm.io/gorm"
)

type UserController struct {
	userService    user.UserService
	APIWrapper     *types.APIWrapper[user.UserModel]
	APIAuthWrapper *types.APIWrapper[user.AuthRequestDTO]
}

func respondUserError(c *gin.Context, status int, message string) {
	c.JSON(status, ErrorResponse{Status: "error", Message: message})
}

func summaryFromUser(u user.UserModel) (UserSummary, bool) {
	if u == nil {
		return UserSummary{}, false
	}
	return UserSummary{
		ID:       u.GetID(),
		Username: u.GetUsername(),
		Email:    u.GetEmail(),
		Name:     u.GetName(),
		Role:     u.GetRoleID(),
		Active:   u.GetActive(),
	}, true
}

func summariesFromUsers(users []user.UserModel) []UserSummary {
	if len(users) == 0 {
		return []UserSummary{}
	}
	result := make([]UserSummary, 0, len(users))
	for _, u := range users {
		if summary, ok := summaryFromUser(u); ok {
			result = append(result, summary)
		}
	}
	return result
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{
		userService:    user.NewUserService(user.NewUserRepo(db)),
		APIWrapper:     types.NewAPIWrapper[user.UserModel](),
		APIAuthWrapper: types.NewAPIWrapper[user.AuthRequestDTO](),
	}
}

func (uc *UserController) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/users")
	{
		api.GET("/", uc.GetAllUsers)
		api.GET("/:id", uc.GetUserByID)
		api.POST("/", uc.CreateUser)
		api.POST("/:id", uc.UpdateUser)
		api.DELETE("/:id", uc.DeleteUser)
		api.POST("/sign-in", uc.AuthenticateUser)
		api.POST("/refresh-token", uc.RefreshToken)
		api.POST("/logout", uc.Logout)
		api.GET("/email/:email", uc.GetUserByEmail)
		api.GET("/username/:username", uc.GetUserByUsername)
	}
}

// GetAllUsers lista usuários ativos.
//
// @Summary     Listar usuários
// @Description Retorna a lista de usuários ativos do sistema. [Em desenvolvimento]
// @Tags        users beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} UserListResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users [get]
func (uc *UserController) GetAllUsers(c *gin.Context) {
	if _, err := uc.APIWrapper.GetContext(c); err != nil {
		respondUserError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	users, err := uc.userService.ListUsers()
	if err != nil {
		respondUserError(c, http.StatusInternalServerError, "failed to list users")
		return
	}
	ucList := make([]user.UserModel, 0, len(users))
	for _, u := range users {
		ucList = append(ucList, u)
	}
	c.JSON(http.StatusOK, UserListResponse{Users: summariesFromUsers(ucList)})
}

// GetUserByID recupera um usuário pelo identificador.
//
// @Summary     Obter usuário
// @Description Recupera um usuário específico pelo ID informado. [Em desenvolvimento]
// @Tags        users beta
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do usuário"
// @Success     200 {object} UserResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users/{id} [get]
func (uc *UserController) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
		respondUserError(c, http.StatusBadRequest, "id is required")
		return
	}
	if _, err := uc.APIWrapper.GetContext(c); err != nil {
		respondUserError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	usr, err := uc.userService.GetUserByID(id)
	if err != nil || usr == nil {
		respondUserError(c, http.StatusNotFound, "user not found")
		return
	}
	if summary, ok := summaryFromUser(usr); ok {
		c.JSON(http.StatusOK, UserResponse{User: summary})
		return
	}
	respondUserError(c, http.StatusInternalServerError, "failed to serialize user")
}

// CreateUser registra um novo usuário (sign-up).
//
// @Summary     Registrar usuário
// @Description Cria um novo usuário no sistema. [Em desenvolvimento]
// @Tags        users beta
// @Accept      json
// @Produce     json
// @Param       payload body CreateUserRequest true "Dados do usuário"
// @Success     201 {object} UserResponse
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/sign-up [post]
func (uc *UserController) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondUserError(c, http.StatusBadRequest, "invalid request payload")
		return
	}
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.Email) == "" {
		respondUserError(c, http.StatusBadRequest, "username, email and password are required")
		return
	}
	newUser := user.NewUserModel(req.Username, req.Name, req.Email)
	if err := newUser.SetPassword(req.Password); err != nil {
		respondUserError(c, http.StatusBadRequest, "failed to hash password")
		return
	}
	if strings.TrimSpace(req.RoleID) != "" {
		newUser.SetRoleID(req.RoleID)
	}
	created, err := uc.userService.CreateUser(newUser)
	if err != nil {
		respondUserError(c, http.StatusInternalServerError, "failed to create user")
		return
	}
	if summary, ok := summaryFromUser(created); ok {
		c.JSON(http.StatusCreated, UserResponse{User: summary})
		return
	}
	respondUserError(c, http.StatusInternalServerError, "failed to serialize user")
}


// AuthenticateUser valida credenciais e retorna tokens.
//
// @Summary     Autenticar usuário
// @Description Valida credenciais e retorna par de tokens. [Em desenvolvimento]
// @Tags        users beta
// @Accept      json
// @Produce     json
// @Param       payload body AuthRequest true "Credenciais"
// @Success     200 {object} AuthResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/sign-in [post]
func (uc *UserController) AuthenticateUser(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondUserError(c, http.StatusBadRequest, "invalid request payload")
		return
	}
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		respondUserError(c, http.StatusBadRequest, "username and password are required")
		return
	}
	usr, err := uc.userService.GetUserByUsername(req.Username)
	if err != nil || usr == nil {
		respondUserError(c, http.StatusUnauthorized, "invalid username or password")
		return
	}
	if !usr.CheckPasswordHash(req.Password) {
		respondUserError(c, http.StatusUnauthorized, "invalid username or password")
		return
	}
	tokenClient := sau.NewTokenClient(
		crt.NewCertService(
			os.ExpandEnv(cm.DefaultGoBEKeyPath),
			os.ExpandEnv(cm.DefaultGoBECertPath),
		),
		uc.userService.GetContextDBService(),
	)
	tokenService, idExpirationSecs, refreshExpirationSecs, err := tokenClient.LoadTokenCfg()
	if err != nil {
		respondUserError(c, http.StatusInternalServerError, err.Error())
		return
	}
	prevTokenID := strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")
	if prevTokenID != "" {
		if _, err := tokenService.ValidateIDToken(prevTokenID); err != nil {
			prevTokenID = ""
		}
	}
	tokenPair, err := tokenService.NewPairFromUser(c, usr, prevTokenID)
	if err != nil || tokenPair == nil {
		respondUserError(c, http.StatusInternalServerError, "failed to generate tokens")
		return
	}
	if idExpirationSecs <= 0 || refreshExpirationSecs <= 0 {
		respondUserError(c, http.StatusInternalServerError, "invalid token expiration time")
		return
	}
	c.Set("refresh_token", tokenPair.RefreshToken.ID)
	c.Set("user_id", usr.GetID())
	c.Header("Authorization", "Bearer "+tokenPair.RefreshToken.ID)
	c.Header("X-ID-Token", tokenPair.IDToken.SS)
	c.Header("X-Refresh-Token", tokenPair.RefreshToken.SS)
	c.Header("X-User-ID", usr.GetID())
	c.Header("X-User-Role", usr.GetRoleID())
	if summary, ok := summaryFromUser(usr); ok {
		c.JSON(http.StatusOK, AuthResponse{
			TokenType:        "Bearer",
			AccessToken:      tokenPair.IDToken.SS,
			RefreshToken:     tokenPair.RefreshToken.SS,
			ExpiresIn:        idExpirationSecs,
			RefreshExpiresIn: refreshExpirationSecs,
			User:             summary,
		})
		return
	}
	respondUserError(c, http.StatusInternalServerError, "failed to serialize user")
}

// @Summary Refresh Token
// @Description Refreshes the user's authentication token.
// @Accept json
// @Produce json
// @Success 200 {object} types.APIResponse[string]
// @Failure 400 {object} types.APIResponse[string]
// @Failure 401 {object} types.APIResponse[string]
// @Failure 500 {object} types.APIResponse[string]
// @Router /users/refresh-token [post]
func (uc *UserController) RefreshToken(c *gin.Context) {
	prevTokenID := strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")
	if prevTokenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing refresh token"})
		return
	}
	refreshTk := c.GetHeader("X-Refresh-Token")
	if refreshTk == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing refresh token"})
		return
	}
	tokenString := c.GetHeader("X-ID-Token")
	if tokenString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing ID token"})
		return
	}

	tokenClient := sau.NewTokenClient(crt.NewCertService("", ""), uc.userService.GetContextDBService())
	tokenService, idExpirationSecs, refreshExpirationSecs, err := tokenClient.LoadTokenCfg()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, err := tokenService.ValidateIDToken(tokenString)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error validating ID token: %v", err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid ID token"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID token not found"})
		return
	}
	// Generate a new token pair
	token, err := tokenService.NewPairFromUser(c, user, refreshTk)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if token == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set the refresh token in the response header
	c.Header("Authorization", "Bearer "+token.RefreshToken.ID)
	// Set the ID token in the response header
	c.Header("X-ID-Token", token.IDToken.SS)
	// Set the refresh token in the response header
	c.Header("X-Refresh-Token", token.RefreshToken.SS)
	// Set the user ID in the response header
	c.Header("X-User-ID", user.GetID())
	// Set the user role in the response header
	c.Header("X-User-Role", user.GetRoleID())
	// Set the user ID in the response body
	uc.APIAuthWrapper.JSONResponse(
		c,
		"success",
		"User authenticated successfully",
		"",
		gin.H{
			"user_id":            user.GetID(),
			"username":           user.GetUsername(),
			"email":              user.GetEmail(),
			"name":               user.GetName(),
			"role":               user.GetRoleID(),
			"expires_in":         idExpirationSecs,
			"refresh_expires_in": refreshExpirationSecs,
			"token_type":         "Bearer",
			"refresh_token":      token.RefreshToken.SS,
			"id_token":           token.IDToken.SS,
		},
		nil,
		http.StatusOK,
	)
	//c.JSON(http.StatusOK, gin.H{"token": token.IDToken, "refresh_token": token.RefreshToken})
}

// @Summary Logout
// @Description Logs out the user by invalidating the refresh token.
// @Accept json
// @Produce json
// @Success 204 {object} types.APIResponse[string]
// @Failure 400 {object} types.APIResponse[string]
// @Failure 401 {object} types.APIResponse[string]
// @Failure 500 {object} types.APIResponse[string]
// @Router /users/logout [post]
func (uc *UserController) Logout(c *gin.Context) {
	refreshTk := strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")
	if refreshTk == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing refresh token"})
		return
	}

	tkClient := sau.NewTokenClient(crt.NewCertService("", ""), uc.userService.GetContextDBService())
	tokenService, _, _, err := tkClient.LoadTokenCfg()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = tokenService.SignOut(c, refreshTk)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Get User By Email
// @Description Retrieves a user by their email address.
// @Accept json
// @Produce json
// @Success 200 {object} types.APIResponse[string]
// @Failure 400 {object} types.APIResponse[string]
// @Failure 404 {object} types.APIResponse[string]
// @Router /users/email/{email} [get]
func (uc *UserController) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	user, err := uc.userService.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// @Summary Get User By Username
// @Description Retrieves a user by their username.
// @Accept json
// @Produce json
// @Success 200 {object} types.APIResponse[string]
// @Failure 400 {object} types.APIResponse[string]
// @Failure 404 {object} types.APIResponse[string]
// @Router /users/username/{username} [get]
func (uc *UserController) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	user, err := uc.userService.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// @Summary Update User
// @Description Updates a user's information.
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body user.UserModel true "User information"
// @Success 200 {object} types.APIResponse[string]
// @Failure 400 {object} types.APIResponse[string]
// @Failure 404 {object} types.APIResponse[string]
// @Failure 500 {object} types.APIResponse[string]
// @Router /users/{id} [put]
func (uc *UserController) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var userRequest user.UserModel
	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userRequest.SetID(id)
	updatedUser, err := uc.userService.UpdateUser(userRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedUser)
}

// @Summary Delete User
// @Description Deletes a user by their ID.
// @Accept json
// @Produce json
// @Success 204 {object} types.APIResponse[string]
// @Failure 400 {object} types.APIResponse[string]
// @Failure 404 {object} types.APIResponse[string]
// @Failure 500 {object} types.APIResponse[string]
// @Router /users/{id} [delete]
func (uc *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	err := uc.userService.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
