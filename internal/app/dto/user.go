package dto

// RegisterRequest represents user registration request
// @Description User registration request body
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email" example:"user@example.com"`
	Password  string `json:"password" validate:"required,min=6" example:"password123"`
	FirstName string `json:"first_name" validate:"required" example:"Ivan"`
	LastName  string `json:"last_name" validate:"required" example:"Ivanov"`
}

// LoginRequest represents user login request
// @Description User login request body
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}

// LoginResponse represents successful login response
// @Description Login response with JWT token and user info
type LoginResponse struct {
	Token string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  UserResponse `json:"user"`
}

// UserResponse represents user data in responses
// @Description User information response
type UserResponse struct {
	ID        int64  `json:"id" example:"1"`
	Email     string `json:"email" example:"user@example.com"`
	FirstName string `json:"first_name" example:"Ivan"`
	LastName  string `json:"last_name" example:"Ivanov"`
	Role      string `json:"role" example:"user" enums:"user,moderator,admin"`
}

// UpdateUserRequest represents user profile update request
// @Description User profile update request body
type UpdateUserRequest struct {
	FirstName string `json:"first_name" example:"Ivan"`
	LastName  string `json:"last_name" example:"Ivanov"`
}

// UpdateUserRoleRequest represents user role update request
// @Description Admin request to update user role
type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=user moderator admin" example:"moderator" enums:"user,moderator,admin"`
}

// ChangePasswordRequest represents password change request
// @Description Change password request body
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required" example:"oldpassword123"`
	NewPassword string `json:"new_password" validate:"required,min=6" example:"newpassword123"`
}
