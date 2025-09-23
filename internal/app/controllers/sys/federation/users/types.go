package users

import t "github.com/kubex-ecosystem/gobe/internal/contracts/types"

type (
	// ErrorResponse padroniza respostas de erro no módulo de usuários.
	ErrorResponse = t.ErrorResponse
)

// UserSummary destaca campos principais do usuário.
type UserSummary struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
}

// UserResponse encapsula o retorno de operações unitárias.
type UserResponse struct {
	User UserSummary `json:"user"`
}

// UserListResponse agrega múltiplos usuários.
type UserListResponse struct {
	Users []UserSummary `json:"users"`
}

// CreateUserRequest descreve o payload de criação/registro.
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	RoleID   string `json:"role_id"`
}

// UpdateUserRequest descreve alterações permitidas em usuários.
type UpdateUserRequest struct {
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	RoleID   string `json:"role_id,omitempty"`
	Password string `json:"password,omitempty"`
	Active   *bool  `json:"active,omitempty"`
}

// AuthRequest representa os dados básicos para autenticação.
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember,omitempty"`
}

// AuthResponse retorna tokens e metadados após autenticação.
type AuthResponse struct {
	TokenType        string      `json:"token_type"`
	AccessToken      string      `json:"access_token"`
	RefreshToken     string      `json:"refresh_token"`
	ExpiresIn        int64       `json:"expires_in"`
	RefreshExpiresIn int64       `json:"refresh_expires_in"`
	User             UserSummary `json:"user"`
}

// RefreshRequest aceita dados para renovação de tokens.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}

// LogoutRequest aceita dados para realizar logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// DeleteResponse informa o resultado de exclusão.
type DeleteResponse struct {
	Message string `json:"message"`
}
