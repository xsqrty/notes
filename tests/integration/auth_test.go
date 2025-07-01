package integration

import (
	"bytes"
	"encoding/json"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"net/http"
	"testing"
)

func TestIntegrationAuth_Login(t *testing.T) {
	t.Parallel()
	cases := []integrationCase[dto.LoginRequest, dto.TokenResponse]{
		{
			name: "successful_login",
			req: &dto.LoginRequest{
				Email:    rootEmail,
				Password: rootPassword,
			},
			statusCode: http.StatusCreated,
			expected: &dto.TokenResponse{
				User: &dto.UserResponse{
					Name:  rootName,
					Email: rootEmail,
				},
			},
		},
		{
			name: "user_not_found",
			req: &dto.LoginRequest{
				Email:    gofakeit.Email(),
				Password: rootPassword,
			},
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
		{
			name: "validation_error",
			req: &dto.LoginRequest{
				Email:    gofakeit.Name(),
				Password: rootPassword,
			},
			statusCode: http.StatusBadRequest,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t, http.MethodPost, "/api/v1/auth/login", func(expected, actual *dto.TokenResponse) {
				require.NotEmpty(t, actual.AccessToken)
				require.NotEmpty(t, actual.RefreshToken)
				require.Equal(t, expected.User.Name, actual.User.Name)
				require.Equal(t, expected.User.Email, actual.User.Email)
			})
		})
	}
}

func TestIntegrationAuth_SignUp(t *testing.T) {
	t.Parallel()

	email := gofakeit.Email()
	name := gofakeit.Name()
	password := gofakeit.Password(true, true, true, true, true, 10)

	cases := []integrationCase[dto.SignUpRequest, dto.TokenResponse]{
		{
			name: "successful_signup",
			req: &dto.SignUpRequest{
				Name:     name,
				Email:    email,
				Password: password,
			},
			statusCode: http.StatusCreated,
			expected: &dto.TokenResponse{
				User: &dto.UserResponse{
					Name:  name,
					Email: email,
				},
			},
			onSuccess: func() {
				require.NotEmpty(t, login(t, &dto.LoginRequest{
					Email:    email,
					Password: password,
				}).AccessToken)
			},
		},
		{
			name: "email_already_exists",
			req: &dto.SignUpRequest{
				Name:     gofakeit.Name(),
				Email:    rootEmail,
				Password: rootPassword,
			},
			statusCode: http.StatusBadRequest,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeEmailExists,
				},
			},
		},
		{
			name: "validation_error",
			req: &dto.SignUpRequest{
				Email:    gofakeit.Name(),
				Password: rootPassword,
			},
			statusCode: http.StatusBadRequest,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t, http.MethodPost, "/api/v1/auth/signup", func(expected, actual *dto.TokenResponse) {
				require.NotEmpty(t, actual.AccessToken)
				require.NotEmpty(t, actual.RefreshToken)
				require.Equal(t, expected.User.Name, actual.User.Name)
				require.Equal(t, expected.User.Email, actual.User.Email)
			})
		})
	}
}

func TestIntegrationAuth_RefreshToken(t *testing.T) {
	t.Parallel()

	cases := []integrationCase[any, dto.TokenResponse]{
		{
			name:       "successful_refresh",
			token:      rootTokens.RefreshToken,
			statusCode: http.StatusCreated,
			expected: &dto.TokenResponse{
				User: &dto.UserResponse{
					Name:  rootName,
					Email: rootEmail,
				},
			},
		},
		{
			name:       "incorrect_token",
			token:      rootTokens.AccessToken,
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
		{
			name:       "bad_request",
			token:      "",
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t, http.MethodPost, "/api/v1/auth/refresh", func(expected, actual *dto.TokenResponse) {
				require.NotEmpty(t, actual.AccessToken)
				require.NotEmpty(t, actual.RefreshToken)
				require.Equal(t, expected.User.Name, actual.User.Name)
				require.Equal(t, expected.User.Email, actual.User.Email)
			})
		})
	}
}

func login(t *testing.T, req *dto.LoginRequest) *dto.TokenResponse {
	t.Helper()
	jsonReq, err := json.Marshal(req)
	require.NoError(t, err)

	res, err := http.Post(withBaseUrl("/api/v1/auth/login"), "application/json", bytes.NewReader(jsonReq))
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusCreated, res.StatusCode)

	tokens := &dto.TokenResponse{}
	json.NewDecoder(res.Body).Decode(tokens)
	require.Equal(t, req.Email, tokens.User.Email)

	return tokens
}

func signUp(t *testing.T, req *dto.SignUpRequest) *dto.TokenResponse {
	t.Helper()
	jsonReq, err := json.Marshal(req)
	require.NoError(t, err)

	res, err := http.Post(withBaseUrl("/api/v1/auth/signup"), "application/json", bytes.NewReader(jsonReq))
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusCreated, res.StatusCode)

	tokens := &dto.TokenResponse{}
	json.NewDecoder(res.Body).Decode(tokens)
	require.Equal(t, req.Email, tokens.User.Email)

	return tokens
}
