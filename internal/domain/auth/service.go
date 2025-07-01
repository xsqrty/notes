package auth

import (
	"context"
	"github.com/xsqrty/notes/internal/domain/user"
)

type Service interface {
	Login(ctx context.Context, login *Login) (*Tokens, error)
	SignUp(ctx context.Context, user *SignUp) (*Tokens, error)
	GenerateTokens(user *user.User) (*Tokens, error)
}
