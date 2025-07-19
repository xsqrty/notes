package dto

import (
	"github.com/google/uuid"
)

// LoginRequest represents the payload for user login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

// UserResponse represents the response containing basic user information.
type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
}

// TokenResponse represents the structure of the response containing authentication tokens and associated user details.
type TokenResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	User         *UserResponse `json:"user"`
}

// SignUpRequest represents the data structure for user sign-up requests.
type SignUpRequest struct {
	Name     string `json:"name"     validate:"required,min=2"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}
