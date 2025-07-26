package users

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	user "github.com/rafa-mori/gdbase/factory/models"
	cm "github.com/rafa-mori/gobe/internal/common"
	sau "github.com/rafa-mori/gobe/internal/security/authentication"
	crt "github.com/rafa-mori/gobe/internal/security/certificates"

	gl "github.com/rafa-mori/gobe/logger"

	"github.com/gin-gonic/gin"
	"github.com/rafa-mori/gobe/internal/types"
	"gorm.io/gorm"
)

type UserController struct {
	userService    user.UserService
	ApiWrapper     *types.ApiWrapper[user.UserModel]
	ApiAuthWrapper *types.ApiWrapper[user.AuthRequestDTO]
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{
		userService:    user.NewUserService(user.NewUserRepo(db)),
		ApiWrapper:     types.NewApiWrapper[user.UserModel](),
		ApiAuthWrapper: types.NewApiWrapper[user.AuthRequestDTO](),
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

func (uc *UserController) GetAllUsers(c *gin.Context) {
	users, err := uc.userService.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (uc *UserController) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, err := uc.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (uc *UserController) CreateUser(c *gin.Context) {
	var userRequest user.UserModel
	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createdUser, err := uc.userService.CreateUser(userRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, createdUser)
}

func (uc *UserController) AuthenticateUser(c *gin.Context) {
	// Define a DTO for authentication requests
	type UserRequestDTO struct {
		// If no username is provided, use email (will be required in this case)
		//Email string `json:"email" binding:"required,email"`
		// If no email is provided, use username (will be required in this case)
		Username string `json:"username" binding:"required,min=3,max=32"`
		Password string `json:"password" binding:"required,min=8,max=32"`
		Remember bool   `json:"remember,omitempty"`
	}
	type AuthRequestDTO struct {
		User UserRequestDTO `json:"user"`
	}
	var authReqT *AuthRequestDTO = &AuthRequestDTO{}
	if err := c.ShouldBindJSON(&authReqT); err != nil && authReqT == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Validate the request
	if authReqT == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	// Check if the request contains a valid username and password
	if authReqT.User.Username == "" && authReqT.User.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}
	// Check if the request contains a valid email and password
	authReq := authReqT.User
	if authReq.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}
	if authReq.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	userRequest := user.NewUserModel(authReq.Username, authReq.Username, "" /* authReq.Email */)
	userRequest.SetPassword(authReq.Password)
	user, err := uc.userService.GetUserByUsername(userRequest.GetUsername())
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	pwdValidation := !user.CheckPasswordHash(userRequest.GetPassword())
	if !pwdValidation {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	tokenClient := sau.NewTokenClient(
		crt.NewCertService(
			os.ExpandEnv(cm.DefaultGoBEKeyPath),
			os.ExpandEnv(cm.DefaultGoBECertPath),
		),
		uc.userService.GetContextDBService(),
	)

	tokenService, idExpirationSecs, refreshExpirationSecs, err := tokenClient.LoadTokenCfg() // Ta vindo zerado aqui os tempos de expiração
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the user is already logged in
	prevTokenID := c.GetHeader("Authorization")
	if prevTokenID != "" {
		prevTokenID = strings.ReplaceAll(prevTokenID, "Bearer ", "")
		userM, userMErr := tokenService.ValidateIDToken(prevTokenID)
		if userMErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if userM == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not found"})
			return
		}

		user = userM
	}

	// Generate a new token pair
	token, err := tokenService.NewPairFromUser(c, user, prevTokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if idExpirationSecs <= 0 || refreshExpirationSecs <= 0 {
		gl.Log("error", fmt.Sprintf("Invalid token expiration times: %d, %d", idExpirationSecs, refreshExpirationSecs))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token expiration time"})
		return
	}

	if token == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Insert the refresh token into user context
	c.Set("refresh_token", token.RefreshToken.ID)
	c.Set("user_id", user.GetID())

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

	uc.ApiAuthWrapper.JSONResponse(
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

	// Uncomment the following line if you want to return the token in the response body
	//c.JSON(http.StatusOK, gin.H{"token": token.IDToken, "refresh_token": token.RefreshToken})
}

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
	uc.ApiAuthWrapper.JSONResponse(
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

func (uc *UserController) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	user, err := uc.userService.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (uc *UserController) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	user, err := uc.userService.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

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

func (uc *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	err := uc.userService.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
