package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xsqrty/notes/internal/adapter/dtoadapter"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/domain/auth"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/internal/middleware"
	"github.com/xsqrty/notes/mocks/app/mock_app"
	"github.com/xsqrty/notes/mocks/domain/mock_auth"
	"github.com/xsqrty/notes/mocks/middleware/mock_middleware"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
)

func TestAuthHandler_Login(t *testing.T) {
	t.Parallel()

	tokens := &auth.Tokens{
		AccessToken:  gofakeit.LetterN(50),
		RefreshToken: gofakeit.LetterN(50),
		User: &user.User{
			ID: uuid.Must(uuid.NewV7()),
		},
	}

	cases := []struct {
		name        string
		req         *dto.LoginRequest
		statusCode  int
		expected    *dto.TokenResponse
		expectedErr *httpio.ErrorResponse
		mocker      func(req *dto.LoginRequest, service *mock_auth.Service)
	}{
		{
			name:       "successful_login",
			statusCode: http.StatusCreated,
			req: &dto.LoginRequest{
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 10),
			},
			expected: dtoadapter.TokensToResponseDto(tokens),
			mocker: func(req *dto.LoginRequest, service *mock_auth.Service) {
				service.EXPECT().
					Login(mock.Anything, dtoadapter.LoginRequestDtoToEntity(req)).
					Return(tokens, nil).
					Once()
			},
		},
		{
			name:       "request_error",
			statusCode: http.StatusBadRequest,
			req:        &dto.LoginRequest{},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
		},
		{
			name:       "login_error",
			statusCode: http.StatusUnauthorized,
			req: &dto.LoginRequest{
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 10),
			},
			expected: nil,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			mocker: func(req *dto.LoginRequest, service *mock_auth.Service) {
				service.EXPECT().
					Login(mock.Anything, dtoadapter.LoginRequestDtoToEntity(req)).
					Return(nil, errors.New("login error")).
					Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := mock_auth.NewService(t)
			deps := mock_app.NewDeps(t, func(deps *app.Deps) {
				deps.Service.AuthService = service
			})

			handler := NewAuthHandler(deps)
			if tc.mocker != nil {
				tc.mocker(tc.req, service)
			}

			body, err := json.Marshal(tc.req)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
			w := httptest.NewRecorder()

			middleware.Logger(deps.Logger)(http.HandlerFunc(handler.Login)).ServeHTTP(w, r)
			res := w.Result()

			require.Equal(t, tc.statusCode, res.StatusCode)
			if tc.expected != nil {
				result := dto.TokenResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
				require.Equal(t, tc.expected, &result)
				require.NotEmpty(t, result.AccessToken)
				require.NotEmpty(t, result.RefreshToken)
			} else {
				err := httpio.ErrorResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&err))
				require.Equal(t, tc.expectedErr.Error.Code, err.Error.Code)
				require.NotEmpty(t, err.Error.Message)
			}

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

	cases := []struct {
		name        string
		req         *dto.SignUpRequest
		statusCode  int
		expected    *dto.TokenResponse
		expectedErr *httpio.ErrorResponse
		mocker      func(req *dto.SignUpRequest, service *mock_auth.Service)
	}{
		{
			name:       "successful_signup",
			statusCode: http.StatusCreated,
			req: &dto.SignUpRequest{
				Name:     gofakeit.Name(),
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 10),
			},
			expected: dtoadapter.TokensToResponseDto(tokens),
			mocker: func(req *dto.SignUpRequest, service *mock_auth.Service) {
				service.EXPECT().
					SignUp(mock.Anything, dtoadapter.SignUpRequestDtoToEntity(req)).
					Return(tokens, nil).
					Once()
			},
		},
		{
			name:       "request_error",
			statusCode: http.StatusBadRequest,
			req:        &dto.SignUpRequest{},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
		},
		{
			name:       "email_already_exists",
			statusCode: http.StatusBadRequest,
			req: &dto.SignUpRequest{
				Name:     gofakeit.Name(),
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 10),
			},
			expected: nil,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeEmailExists,
				},
			},
			mocker: func(req *dto.SignUpRequest, service *mock_auth.Service) {
				service.EXPECT().
					SignUp(mock.Anything, dtoadapter.SignUpRequestDtoToEntity(req)).
					Return(nil, auth.ErrEmailAlreadyExists).
					Once()
			},
		},
		{
			name:       "unknown_error",
			statusCode: http.StatusInternalServerError,
			req: &dto.SignUpRequest{
				Name:     gofakeit.Name(),
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, 10),
			},
			expected: nil,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnknown,
				},
			},
			mocker: func(req *dto.SignUpRequest, service *mock_auth.Service) {
				service.EXPECT().
					SignUp(mock.Anything, dtoadapter.SignUpRequestDtoToEntity(req)).
					Return(nil, errors.New("some error")).
					Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := mock_auth.NewService(t)
			deps := mock_app.NewDeps(t, func(deps *app.Deps) {
				deps.Service.AuthService = service
			})

			handler := NewAuthHandler(deps)
			if tc.mocker != nil {
				tc.mocker(tc.req, service)
			}

			body, err := json.Marshal(tc.req)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(body))
			w := httptest.NewRecorder()

			middleware.Logger(deps.Logger)(http.HandlerFunc(handler.SignUp)).ServeHTTP(w, r)
			res := w.Result()

			require.Equal(t, tc.statusCode, res.StatusCode)
			if tc.expected != nil {
				result := dto.TokenResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
				require.Equal(t, tc.expected, &result)
				require.NotEmpty(t, result.AccessToken)
				require.NotEmpty(t, result.RefreshToken)
			} else {
				err := httpio.ErrorResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&err))
				require.Equal(t, tc.expectedErr.Error.Code, err.Error.Code)
				require.NotEmpty(t, err.Error.Message)
			}

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

	cases := []struct {
		name        string
		statusCode  int
		expected    *dto.TokenResponse
		expectedErr *httpio.ErrorResponse
		mocker      func(mw *mock_middleware.JWTAuthentication, service *mock_auth.Service)
	}{
		{
			name:       "successful_refresh_token",
			statusCode: http.StatusCreated,
			expected:   dtoadapter.TokensToResponseDto(tokens),
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_auth.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().GenerateTokens(u).Return(tokens, nil).Once()
			},
		},
		{
			name:       "user_unauthorized",
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_auth.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(nil, errors.New("no user")).Once()
			},
		},
		{
			name:       "generate_token_error",
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_auth.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().GenerateTokens(u).Return(nil, errors.New("generate token error")).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := mock_auth.NewService(t)
			mw := mock_middleware.NewJWTAuthentication(t)

			deps := mock_app.NewDeps(t, func(deps *app.Deps) {
				deps.Service.AuthService = service
				deps.JWTAuthentication = mw
			})

			handler := NewAuthHandler(deps)
			if tc.mocker != nil {
				tc.mocker(mw, service)
			}

			r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
			w := httptest.NewRecorder()

			middleware.Logger(deps.Logger)(http.HandlerFunc(handler.RefreshToken)).ServeHTTP(w, r)
			res := w.Result()

			require.Equal(t, tc.statusCode, res.StatusCode)
			if tc.expected != nil {
				result := dto.TokenResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
				require.Equal(t, tc.expected, &result)
				require.NotEmpty(t, result.AccessToken)
				require.NotEmpty(t, result.RefreshToken)
			} else {
				err := httpio.ErrorResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&err))
				require.Equal(t, tc.expectedErr.Error.Code, err.Error.Code)
				require.NotEmpty(t, err.Error.Message)
			}

			mock.AssertExpectationsForObjects(t, service, mw)
		})
	}
}
