package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xsqrty/notes/internal/domain/auth"
	"github.com/xsqrty/notes/internal/domain/role"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/mocks/app/mock_tx"
	"github.com/xsqrty/notes/mocks/domain/mock_auth"
	"github.com/xsqrty/notes/mocks/domain/mock_role"
	"github.com/xsqrty/notes/mocks/domain/mock_user"
	"testing"
)

func TestAuthService_Login(t *testing.T) {
	t.Parallel()

	accessToken := gofakeit.LetterN(50)
	refreshToken := gofakeit.LetterN(50)
	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, 20)

	u := &user.User{
		ID:             uuid.Must(uuid.NewV7()),
		Email:          email,
		HashedPassword: gofakeit.LetterN(32),
	}

	cases := []struct {
		name        string
		expected    *auth.Tokens
		expectedErr string
		mocker      func(repo *mock_user.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator)
	}{
		{
			name: "successful_login",
			expected: &auth.Tokens{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				User:         u,
			},
			mocker: func(repo *mock_user.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator) {
				repo.EXPECT().GetByEmail(mock.Anything, email).Return(u, nil).Once()
				tokenizer.EXPECT().CreateAccessToken(u).Return(accessToken, nil).Once()
				tokenizer.EXPECT().CreateRefreshToken(u).Return(refreshToken, nil).Once()
				passgen.EXPECT().Compare(u.HashedPassword, password).Return(true).Once()
			},
		},
		{
			name:        "user_not_found",
			expected:    nil,
			expectedErr: "login: user not found",
			mocker: func(repo *mock_user.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator) {
				repo.EXPECT().GetByEmail(mock.Anything, email).Return(nil, errors.New("no rows")).Once()
			},
		},
		{
			name:        "incorrect_password",
			expected:    nil,
			expectedErr: "login: password incorrect",
			mocker: func(repo *mock_user.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator) {
				repo.EXPECT().GetByEmail(mock.Anything, email).Return(u, nil).Once()
				passgen.EXPECT().Compare(u.HashedPassword, password).Return(false).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := mock_user.NewRepository(t)
			tokenizer := mock_auth.NewTokenizer(t)
			passgen := mock_auth.NewPasswordGenerator(t)
			tc.mocker(repo, tokenizer, passgen)

			service := NewAuthService(&AuthServiceDeps{
				UserRepo:  repo,
				Tokenizer: tokenizer,
				PassGen:   passgen,
			})

			result, err := service.Login(context.Background(), &auth.Login{
				Email:    email,
				Password: password,
			})

			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			}

			require.Equal(t, tc.expected, result)
			mock.AssertExpectationsForObjects(t, repo, tokenizer, passgen)
		})
	}
}

func TestAuthService_SignUp(t *testing.T) {
	t.Parallel()

	accessToken := gofakeit.LetterN(50)
	refreshToken := gofakeit.LetterN(50)
	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, 20)

	u := &user.User{
		ID:             uuid.Must(uuid.NewV7()),
		Email:          email,
		HashedPassword: gofakeit.LetterN(32),
	}

	cases := []struct {
		name        string
		expected    *auth.Tokens
		expectedErr string
		mocker      func(repo *mock_user.Repository, roleRepo *mock_role.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator)
	}{
		{
			name: "successful_signup",
			expected: &auth.Tokens{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				User:         u,
			},
			mocker: func(repo *mock_user.Repository, roleRepo *mock_role.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator) {
				repo.EXPECT().EmailExists(mock.Anything, email).Return(false, nil).Once()
				passgen.EXPECT().Generate(password).Return(password, nil).Once()
				repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil).Once()
				roleRepo.EXPECT().AttachUserRolesByLabel(mock.Anything, role.LabelOnCreated, mock.Anything).Return(nil).Once()
				tokenizer.EXPECT().CreateRefreshToken(mock.Anything).Return(refreshToken, nil).Once()
				tokenizer.EXPECT().CreateAccessToken(mock.Anything).Return(accessToken, nil).Once()
			},
		},
		{
			name:        "email_exists_error",
			expected:    nil,
			expectedErr: "signup check email: email error",
			mocker: func(repo *mock_user.Repository, roleRepo *mock_role.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator) {
				repo.EXPECT().EmailExists(mock.Anything, email).Return(false, errors.New("email error")).Once()
			},
		},
		{
			name:        "email_exists",
			expected:    nil,
			expectedErr: fmt.Sprintf("signup: %s (%s)", auth.ErrEmailAlreadyExists.Error(), email),
			mocker: func(repo *mock_user.Repository, roleRepo *mock_role.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator) {
				repo.EXPECT().EmailExists(mock.Anything, email).Return(true, nil).Once()
			},
		},
		{
			name:        "password_gen_error",
			expected:    nil,
			expectedErr: "signup: gen error",
			mocker: func(repo *mock_user.Repository, roleRepo *mock_role.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator) {
				repo.EXPECT().EmailExists(mock.Anything, email).Return(false, nil).Once()
				passgen.EXPECT().Generate(password).Return("", errors.New("gen error")).Once()
			},
		},
		{
			name:        "save_user_err",
			expected:    nil,
			expectedErr: "signup: save user error",
			mocker: func(repo *mock_user.Repository, roleRepo *mock_role.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator) {
				repo.EXPECT().EmailExists(mock.Anything, email).Return(false, nil).Once()
				repo.EXPECT().Save(mock.Anything, mock.Anything).Return(errors.New("save user error")).Once()
				passgen.EXPECT().Generate(password).Return(password, nil).Once()
			},
		},
		{
			name:        "attach_roles_error",
			expected:    nil,
			expectedErr: "signup: attach error",
			mocker: func(repo *mock_user.Repository, roleRepo *mock_role.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator) {
				repo.EXPECT().EmailExists(mock.Anything, email).Return(false, nil).Once()
				passgen.EXPECT().Generate(password).Return(password, nil).Once()
				repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil).Once()
				roleRepo.EXPECT().AttachUserRolesByLabel(mock.Anything, role.LabelOnCreated, mock.Anything).Return(errors.New("attach error")).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := mock_user.NewRepository(t)
			roleRepo := mock_role.NewRepository(t)
			tokenizer := mock_auth.NewTokenizer(t)
			passgen := mock_auth.NewPasswordGenerator(t)
			tc.mocker(repo, roleRepo, tokenizer, passgen)

			service := NewAuthService(&AuthServiceDeps{
				TxManager: mock_tx.NewMockTxManager(),
				RoleRepo:  roleRepo,
				UserRepo:  repo,
				Tokenizer: tokenizer,
				PassGen:   passgen,
			})

			result, err := service.SignUp(context.Background(), &auth.SignUp{
				Email:    email,
				Password: password,
			})

			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			}

			if tc.expected != nil {
				require.NotNil(t, result)
				require.Equal(t, tc.expected.AccessToken, result.AccessToken)
				require.Equal(t, tc.expected.RefreshToken, result.RefreshToken)
				require.Equal(t, tc.expected.User.Name, result.User.Name)
				require.Equal(t, tc.expected.User.Email, result.User.Email)
				require.Equal(t, password, result.User.HashedPassword)
				require.NotZero(t, result.User.CreatedAt)
			}

			mock.AssertExpectationsForObjects(t, repo, roleRepo, tokenizer, passgen)
		})
	}
}

func TestAuthService_GenerateTokens(t *testing.T) {
	t.Parallel()

	accessToken := gofakeit.LetterN(50)
	refreshToken := gofakeit.LetterN(50)
	email := gofakeit.Email()

	u := &user.User{
		ID:             uuid.Must(uuid.NewV7()),
		Email:          email,
		HashedPassword: gofakeit.LetterN(32),
	}

	cases := []struct {
		name        string
		expected    *auth.Tokens
		expectedErr string
		mocker      func(repo *mock_user.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator)
	}{
		{
			name: "successful_tokens_generated",
			expected: &auth.Tokens{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				User:         u,
			},
			mocker: func(repo *mock_user.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator) {
				tokenizer.EXPECT().CreateAccessToken(u).Return(accessToken, nil).Once()
				tokenizer.EXPECT().CreateRefreshToken(u).Return(refreshToken, nil).Once()
			},
		},
		{
			name:        "access_token_error",
			expected:    nil,
			expectedErr: "get access token by user: access token error",
			mocker: func(repo *mock_user.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator) {
				tokenizer.EXPECT().CreateAccessToken(u).Return("", errors.New("access token error")).Once()
			},
		},
		{
			name:        "refresh_token_error",
			expected:    nil,
			expectedErr: "get refresh token by user: refresh token error",
			mocker: func(repo *mock_user.Repository, tokenizer *mock_auth.Tokenizer, passgen *mock_auth.PasswordGenerator) {
				tokenizer.EXPECT().CreateAccessToken(u).Return(accessToken, nil).Once()
				tokenizer.EXPECT().CreateRefreshToken(u).Return("", errors.New("refresh token error")).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := mock_user.NewRepository(t)
			tokenizer := mock_auth.NewTokenizer(t)
			passgen := mock_auth.NewPasswordGenerator(t)
			tc.mocker(repo, tokenizer, passgen)

			service := NewAuthService(&AuthServiceDeps{
				UserRepo:  repo,
				Tokenizer: tokenizer,
				PassGen:   passgen,
			})

			result, err := service.GenerateTokens(u)

			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			}

			require.Equal(t, tc.expected, result)
			mock.AssertExpectationsForObjects(t, tokenizer)
		})
	}
}
