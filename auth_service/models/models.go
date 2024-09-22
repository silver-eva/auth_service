package models

import "time"

// RefreshRequest holds the refresh token for the request.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthRequest holds the roles and refresh request.
type AuthRequest struct {
	Roles []string `json:"roles" validate:"required"`
	RefreshRequest
}

// LoginRequest holds the login credentials.
type LoginRequest struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// SignUpRequest contains the signup information.
type SignUpRequest struct {
	Email string `json:"email" validate:"required,email"`
	LoginRequest
}

// User represents the user in the database.
type User struct {
	Id         string `db:"id"`
	Name       string `db:"name"`
	Password   string `db:"password"`
	Role       string `db:"role"`
	IsLoggedIn bool   `db:"is_logged_in"`
}

// Response struct for standard API responses.
type Response struct {
	Status       int    `json:"status,omitempty"`
	Message      string `json:"message,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// DecodedToken represents the decoded JWT token.
type DecodedToken struct {
	UserName string
	UserPass string
	UserRole string
	Expired  time.Time
	UserId   string
}
