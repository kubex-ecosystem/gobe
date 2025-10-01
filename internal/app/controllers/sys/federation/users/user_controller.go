// Package users provides the UserController for managing user-related operations.
package users

import (
	"net/http"
	"os"
	"strings"

	user "github.com/kubex-ecosystem/gdbase/factory/models"
	sau "github.com/kubex-ecosystem/gobe/internal/app/security/authentication"
	crt "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	cm "github.com/kubex-ecosystem/gobe/internal/commons"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/contracts/types"
)

type UserController struct {
	userService user.UserService
	APIWrapper  *types.APIWrapper[user.UserModel]
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

func NewUserController(bridge *svc.Bridge) *UserController {
	return &UserController{
		userService: bridge.UserService(),
		APIWrapper:  types.NewAPIWrapper[user.UserModel](),
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
	summaries := make([]UserSummary, 0, len(users))
	for _, u := range users {
		if summary, ok := summaryFromUser(u); ok {
			summaries = append(summaries, summary)
		}
	}
	c.JSON(http.StatusOK, UserListResponse{Users: summaries})
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

// RefreshToken emite um novo par de tokens.
//
// @Summary     Renovar tokens
// @Description Gera um novo par de tokens a partir do refresh token válido. [Em desenvolvimento]
// @Tags        users beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body RefreshRequest false "Tokens atuais"
// @Success     200 {object} AuthResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/check [post]
func (uc *UserController) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	_ = c.ShouldBindJSON(&req)
	if req.RefreshToken == "" {
		req.RefreshToken = strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")
	}
	if req.RefreshToken == "" {
		req.RefreshToken = c.GetHeader("X-Refresh-Token")
	}
	if req.IDToken == "" {
		req.IDToken = c.GetHeader("X-ID-Token")
	}
	if req.RefreshToken == "" || req.IDToken == "" {
		respondUserError(c, http.StatusBadRequest, "missing tokens for refresh")
		return
	}
	tokenClient := sau.NewTokenClient(crt.NewCertService("", ""), uc.userService.GetContextDBService())
	tokenService, idExpirationSecs, refreshExpirationSecs, err := tokenClient.LoadTokenCfg()
	if err != nil {
		respondUserError(c, http.StatusInternalServerError, err.Error())
		return
	}
	usr, err := tokenService.ValidateIDToken(req.IDToken)
	if err != nil || usr == nil {
		respondUserError(c, http.StatusUnauthorized, "invalid id token")
		return
	}
	tokenPair, err := tokenService.NewPairFromUser(c, usr, req.RefreshToken)
	if err != nil || tokenPair == nil {
		respondUserError(c, http.StatusInternalServerError, "failed to refresh tokens")
		return
	}
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

// Logout invalida o refresh token atual.
//
// @Summary     Encerrar sessão
// @Description Invalida o refresh token ativo do usuário. [Em desenvolvimento]
// @Tags        users beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body LogoutRequest false "Token de refresh"
// @Success     200 {object} DeleteResponse
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/sign-out [post]
func (uc *UserController) Logout(c *gin.Context) {
	var req LogoutRequest
	_ = c.ShouldBindJSON(&req)
	if req.RefreshToken == "" {
		req.RefreshToken = strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")
	}
	if req.RefreshToken == "" {
		req.RefreshToken = c.GetHeader("X-Refresh-Token")
	}
	if req.RefreshToken == "" {
		respondUserError(c, http.StatusBadRequest, "missing refresh token")
		return
	}
	tkClient := sau.NewTokenClient(crt.NewCertService("", ""), uc.userService.GetContextDBService())
	tokenService, _, _, err := tkClient.LoadTokenCfg()
	if err != nil {
		respondUserError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if err := tokenService.SignOut(c, req.RefreshToken); err != nil {
		respondUserError(c, http.StatusInternalServerError, "failed to revoke token")
		return
	}
	c.JSON(http.StatusOK, DeleteResponse{Message: "signed out successfully"})
}

// GetUserByEmail recupera usuário pelo email.
//
// @Summary     Buscar usuário por email
// @Description Retorna o usuário associado ao email informado. [Em desenvolvimento]
// @Tags        users beta
// @Security    BearerAuth
// @Produce     json
// @Param       email path string true "E-mail"
// @Success     200 {object} UserResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users/email/{email} [get]
func (uc *UserController) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	if strings.TrimSpace(email) == "" {
		respondUserError(c, http.StatusBadRequest, "email is required")
		return
	}
	if _, err := uc.APIWrapper.GetContext(c); err != nil {
		respondUserError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	usr, err := uc.userService.GetUserByEmail(email)
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

// GetUserByUsername recupera usuário pelo username.
//
// @Summary     Buscar usuário por username
// @Description Retorna usuário associado ao username informado. [Em desenvolvimento]
// @Tags        users beta
// @Security    BearerAuth
// @Produce     json
// @Param       username path string true "Username"
// @Success     200 {object} UserResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users/username/{username} [get]
func (uc *UserController) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if strings.TrimSpace(username) == "" {
		respondUserError(c, http.StatusBadRequest, "username is required")
		return
	}
	if _, err := uc.APIWrapper.GetContext(c); err != nil {
		respondUserError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	usr, err := uc.userService.GetUserByUsername(username)
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

// UpdateUser atualiza dados de um usuário.
//
// @Summary     Atualizar usuário
// @Description Atualiza os campos permitidos para o usuário informado. [Em desenvolvimento]
// @Tags        users beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string             true "ID do usuário"
// @Param       payload body UpdateUserRequest true "Dados a atualizar"
// @Success     200 {object} UserResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users/{id} [put]
func (uc *UserController) UpdateUser(c *gin.Context) {
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
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondUserError(c, http.StatusBadRequest, "invalid request payload")
		return
	}
	if strings.TrimSpace(req.Email) != "" {
		usr.SetEmail(req.Email)
	}
	if strings.TrimSpace(req.Name) != "" {
		usr.SetName(req.Name)
	}
	if strings.TrimSpace(req.RoleID) != "" {
		usr.SetRoleID(req.RoleID)
	}
	if req.Active != nil {
		usr.SetActive(*req.Active)
	}
	if strings.TrimSpace(req.Password) != "" {
		if err := usr.SetPassword(req.Password); err != nil {
			respondUserError(c, http.StatusBadRequest, "failed to hash password")
			return
		}
	}
	updated, err := uc.userService.UpdateUser(usr)
	if err != nil {
		respondUserError(c, http.StatusInternalServerError, "failed to update user")
		return
	}
	if summary, ok := summaryFromUser(updated); ok {
		c.JSON(http.StatusOK, UserResponse{User: summary})
		return
	}
	respondUserError(c, http.StatusInternalServerError, "failed to serialize user")
}

// DeleteUser remove usuário existente.
//
// @Summary     Remover usuário
// @Description Remove o usuário identificado pelo ID. [Em desenvolvimento]
// @Tags        users beta
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do usuário"
// @Success     200 {object} DeleteResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users/{id} [delete]
func (uc *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
		respondUserError(c, http.StatusBadRequest, "id is required")
		return
	}
	if _, err := uc.APIWrapper.GetContext(c); err != nil {
		respondUserError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	if err := uc.userService.DeleteUser(id); err != nil {
		respondUserError(c, http.StatusInternalServerError, "failed to delete user")
		return
	}
	c.JSON(http.StatusOK, DeleteResponse{Message: "User deleted successfully"})
}
