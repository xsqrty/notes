package handler

import (
	"errors"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/xsqrty/notes/internal/adapter/dtoadapter"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/domain/auth"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/mocks/app/mock_app"
	"github.com/xsqrty/notes/mocks/domain/mock_auth"
	"github.com/xsqrty/notes/mocks/middleware/mock_middleware"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"github.com/xsqrty/notes/tests/testutil"
)

type authDeps struct {
	mw      *mock_middleware.JWTAuthentication
	service *mock_auth.Service
}

func TestAuthHandler_Login(t *testing.T) {
	t.Parallel()

	tokens := &auth.Tokens{
		AccessToken:  gofakeit.LetterN(50),
		RefreshToken: gofakeit.LetterN(50),
		User: &user.User{
			ID: uuid.Must(uuid.NewV7()),
		},
	}

	cases := []testutil.HandlerCase[*dto.LoginRequest, *dto.TokenResponse, *authDeps]{
		{
			Name:       "successful_login",
			StatusCode: http.StatusCreated,
			Req: &dto.LoginRequest{
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 10),
			},
			Expected: dtoadapter.TokensToResponseDto(tokens),
			Mocker: func(req *dto.LoginRequest, d *authDeps) {
				d.service.EXPECT().
					Login(mock.Anything, dtoadapter.LoginRequestDtoToEntity(req)).
					Return(tokens, nil).
					Once()
			},
		},
		{
			Name:       "request_error",
			StatusCode: http.StatusBadRequest,
			Req:        &dto.LoginRequest{},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
		},
		{
			Name:       "login_error",
			StatusCode: http.StatusUnauthorized,
			Req: &dto.LoginRequest{
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 10),
			},
			Expected: nil,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			Mocker: func(req *dto.LoginRequest, d *authDeps) {
				d.service.EXPECT().
					Login(mock.Anything, dtoadapter.LoginRequestDtoToEntity(req)).
					Return(nil, errors.New("login error")).
					Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			service := mock_auth.NewService(t)
			tc.Run(t, http.MethodPost, "/api/v1/auth/login", func() *authDeps {
				return &authDeps{
					service: service,
				}
			}, func(d *authDeps) http.HandlerFunc {
				return NewAuthHandler(mock_app.NewDeps(t, func(deps *app.Deps) {
					deps.Service.AuthService = service
				})).Login
			})

			mock.AssertExpectationsForObjects(t, service)
		})
	}
}

func TestAuthHandler_SignUp(t *testing.T) {
	t.Parallel()

	tokens := &auth.Tokens{
		AccessToken:  gofakeit.LetterN(50),
		RefreshToken: gofakeit.LetterN(50),
		User: &user.User{
			ID:    uuid.Must(uuid.NewV7()),
			Name:  gofakeit.Name(),
			Email: gofakeit.Email(),
		},
	}

	cases := []testutil.HandlerCase[*dto.SignUpRequest, *dto.TokenResponse, *authDeps]{
		{
			Name:       "successful_signup",
			StatusCode: http.StatusCreated,
			Req: &dto.SignUpRequest{
				Name:     gofakeit.Name(),
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 10),
			},
			Expected: dtoadapter.TokensToResponseDto(tokens),
			Mocker: func(req *dto.SignUpRequest, d *authDeps) {
				d.service.EXPECT().
					SignUp(mock.Anything, dtoadapter.SignUpRequestDtoToEntity(req)).
					Return(tokens, nil).
					Once()
			},
		},
		{
			Name:       "request_error",
			StatusCode: http.StatusBadRequest,
			Req:        &dto.SignUpRequest{},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
		},
		{
			Name:       "email_already_exists",
			StatusCode: http.StatusBadRequest,
			Req: &dto.SignUpRequest{
				Name:     gofakeit.Name(),
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 10),
			},
			Expected: nil,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeEmailExists,
				},
			},
			Mocker: func(req *dto.SignUpRequest, d *authDeps) {
				d.service.EXPECT().
					SignUp(mock.Anything, dtoadapter.SignUpRequestDtoToEntity(req)).
					Return(nil, auth.ErrEmailAlreadyExists).
					Once()
			},
		},
		{
			Name:       "unknown_error",
			StatusCode: http.StatusInternalServerError,
			Req: &dto.SignUpRequest{
				Name:     gofakeit.Name(),
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 10),
			},
			Expected: nil,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnknown,
				},
			},
			Mocker: func(req *dto.SignUpRequest, d *authDeps) {
				d.service.EXPECT().
					SignUp(mock.Anything, dtoadapter.SignUpRequestDtoToEntity(req)).
					Return(nil, errors.New("some error")).
					Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			service := mock_auth.NewService(t)
			tc.Run(t, http.MethodPost, "/api/v1/auth/signup", func() *authDeps {
				return &authDeps{
					service: service,
				}
			}, func(d *authDeps) http.HandlerFunc {
				return NewAuthHandler(mock_app.NewDeps(t, func(deps *app.Deps) {
					deps.Service.AuthService = service
				})).SignUp
			})

			mock.AssertExpectationsForObjects(t, service)
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	t.Parallel()

	u := &user.User{
		ID: uuid.Must(uuid.NewV7()),
	}

	tokens := &auth.Tokens{
		AccessToken:  gofakeit.LetterN(50),
		RefreshToken: gofakeit.LetterN(50),
		User:         u,
	}

	cases := []testutil.HandlerCase[struct{}, *dto.TokenResponse, *authDeps]{
		{
			Name:       "successful_refresh_token",
			StatusCode: http.StatusCreated,
			Expected:   dtoadapter.TokensToResponseDto(tokens),
			Mocker: func(_ struct{}, d *authDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().GenerateTokens(u).Return(tokens, nil).Once()
			},
		},
		{
			Name:       "user_unauthorized",
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			Mocker: func(_ struct{}, d *authDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(nil, errors.New("no user")).Once()
			},
		},
		{
			Name:       "generate_token_error",
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			Mocker: func(_ struct{}, d *authDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().GenerateTokens(u).Return(nil, errors.New("generate token error")).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			service := mock_auth.NewService(t)
			mw := mock_middleware.NewJWTAuthentication(t)

			tc.Run(t, http.MethodPost, "/api/v1/auth/refresh", func() *authDeps {
				return &authDeps{
					mw:      mw,
					service: service,
				}
			}, func(d *authDeps) http.HandlerFunc {
				return NewAuthHandler(mock_app.NewDeps(t, func(deps *app.Deps) {
					deps.Service.AuthService = service
					deps.JWTAuthentication = mw
				})).RefreshToken
			})

			mock.AssertExpectationsForObjects(t, service)
		})
	}
}
