package service

import (
	"context"
	"fmt"
	"github.com/xsqrty/notes/internal/domain/auth"
	"github.com/xsqrty/notes/internal/domain/role"
	"github.com/xsqrty/notes/internal/domain/tx"
	"github.com/xsqrty/notes/internal/domain/user"
	"time"
)

type AuthServiceDeps struct {
	UserRepo  user.Repository
	RoleRepo  role.Repository
	Tokenizer auth.Tokenizer
	PassGen   auth.PasswordGenerator
	TxManager tx.Manager
}

type authService struct {
	tokenizer auth.Tokenizer
	roleRepo  role.Repository
	userRepo  user.Repository
	passGen   auth.PasswordGenerator
	tx        tx.Manager
}

func NewAuthService(deps *AuthServiceDeps) auth.Service {
	return &authService{
		tokenizer: deps.Tokenizer,
		roleRepo:  deps.RoleRepo,
		userRepo:  deps.UserRepo,
		passGen:   deps.PassGen,
		tx:        deps.TxManager,
	}
}

func (s *authService) Login(ctx context.Context, login *auth.Login) (*auth.Tokens, error) {
	user, err := s.userRepo.GetByEmail(ctx, login.Email)
	if err != nil {
		return nil, fmt.Errorf("login: %w", auth.ErrUserNotFound)
	}

	if !s.passGen.Compare(user.HashedPassword, login.Password) {
		return nil, fmt.Errorf("login: %w", auth.ErrPasswordIncorrect)
	}

	return s.GenerateTokens(user)
}

func (s *authService) SignUp(ctx context.Context, data *auth.SignUp) (*auth.Tokens, error) {
	isExist, err := s.userRepo.EmailExists(ctx, data.Email)
	if err != nil {
		return nil, fmt.Errorf("signup check email: %w", err)
	}

	if isExist {
		return nil, fmt.Errorf("signup: %w (%s)", auth.ErrEmailAlreadyExists, data.Email)
	}

	pass, err := s.passGen.Generate(data.Password)
	if err != nil {
		return nil, fmt.Errorf("signup: %w", err)
	}

	user := &user.User{
		Name:           data.Name,
		Email:          data.Email,
		HashedPassword: pass,
		CreatedAt:      time.Now(),
	}

	err = s.tx.Transact(ctx, func(ctx context.Context) error {
		if err := s.userRepo.Save(ctx, user); err != nil {
			return fmt.Errorf("signup: %w", err)
		}

		err = s.roleRepo.AttachUserRolesByLabel(ctx, role.LabelOnCreated, user)
		if err != nil {
			return fmt.Errorf("signup: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GenerateTokens(user)
}

func (s *authService) GenerateTokens(user *user.User) (*auth.Tokens, error) {
	accessToken, err := s.tokenizer.CreateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("get access token by user: %w", err)
	}

	refreshToken, err := s.tokenizer.CreateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("get refresh token by user: %w", err)
	}

	return &auth.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}
