package auth

import (
	"errors"

	"github.com/xsqrty/notes/internal/domain/user"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrPasswordIncorrect  = errors.New("password incorrect")
	ErrUserNotFound       = errors.New("user not found")
)

// Tokenizer defines methods for creating access and refresh tokens for user authentication.
type Tokenizer interface {
	CreateAccessToken(user *user.User) (string, error)
	CreateRefreshToken(user *user.User) (string, error)
}

// PasswordGenerator defines methods for generating and verifying hashed passwords.
type PasswordGenerator interface {
	Generate(string) (string, error)
	Compare(hash, password string) bool
}

// Login represents the user's credentials required for authentication.
type Login struct {
	Email    string
	Password string
}

// SignUp represents the structure used to hold user registration details.
type SignUp struct {
	Name     string
	Email    string
	Password string
}

// Tokens represent a pair of access and refresh tokens associated with a user.
type Tokens struct {
	AccessToken  string
	RefreshToken string
	User         *user.User
}
