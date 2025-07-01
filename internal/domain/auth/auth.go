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

type Tokenizer interface {
	CreateAccessToken(user *user.User) (string, error)
	CreateRefreshToken(user *user.User) (string, error)
}

type PasswordGenerator interface {
	Generate(string) (string, error)
	Compare(hash, password string) bool
}

type Login struct {
	Email    string
	Password string
}

type SignUp struct {
	Name     string
	Email    string
	Password string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	User         *user.User
}
